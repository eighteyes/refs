// refs_test.go — tests for the refs CLI and hook.
//
// Responsibilities:
//   - run the CLI in-process inside a temp dir with its own `git init`
//   - cover list/show/set/rm/kind/kinds/home/--clear, key validation, link parsing
//   - check insertion order, unknown-field preservation, home rules
//   - check hook output for empty, populated and garbage input
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type result struct {
	out, err string
	code     int
}

// sandbox makes a temp git repo, chdirs into it and returns its resolved path.
func sandbox(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	t.Chdir(dir)
	return dir
}

func refs(args ...string) result {
	return refsIn("", args...)
}

func refsIn(stdin string, args ...string) result {
	var o, e bytes.Buffer
	code := run(args, strings.NewReader(stdin), &o, &e)
	return result{o.String(), e.String(), code}
}

func mustOK(t *testing.T, r result) result {
	t.Helper()
	if r.code != 0 {
		t.Fatalf("exit %d, stderr %q", r.code, r.err)
	}
	return r
}

func readStore(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".ai", "refs.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestEmptyList(t *testing.T) {
	root := sandbox(t)
	r := mustOK(t, refs())
	want := "no refs in " + filepath.Join(root, ".ai", "refs.json") + "\n"
	if r.out != want {
		t.Errorf("got %q want %q", r.out, want)
	}
	if r := mustOK(t, refs("kinds")); r.out != "no kinds defined\n" {
		t.Errorf("kinds: %q", r.out)
	}
}

func TestSetShowListRm(t *testing.T) {
	sandbox(t)
	r := mustOK(t, refs("R1", "first", "req", "--kind", "req"))
	if r.out != "set R1 [req]\n" || r.err != "" {
		t.Errorf("set: %+v", r)
	}
	r = mustOK(t, refs("R1"))
	if !strings.HasPrefix(r.out, "R1  [req]\n  value   first req\n  set_at  ") {
		t.Errorf("show: %q", r.out)
	}
	r = mustOK(t, refs())
	if r.out != "req  R1  first req\n" {
		t.Errorf("list: %q", r.out)
	}
	if r := mustOK(t, refs("rm", "R1")); r.out != "rm R1\n" {
		t.Errorf("rm: %q", r.out)
	}
	if r := refs("R1"); r.code != 1 || !strings.HasPrefix(r.err, "refs: no ref 'R1' in ") {
		t.Errorf("show after rm: %+v", r)
	}
	if r := refs("rm", "R1"); r.code != 1 {
		t.Errorf("rm twice: %+v", r)
	}
}

func TestStoreFormat(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("R1", "v", "--kind=req"))
	s := readStore(t, root)
	if !strings.HasPrefix(s, "{\n  \"kinds\": {\n    \"R\": \"req\"\n  },\n  \"refs\": {\n    \"R1\": {\n") || !strings.HasSuffix(s, "}\n") {
		t.Errorf("format: %q", s)
	}
	if !regexp.MustCompile(`"set_at": "\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d[+-]\d{4}"`).MatchString(s) {
		t.Errorf("set_at format: %q", s)
	}
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json.tmp")); err == nil {
		t.Error("tmp file left behind")
	}
}

func TestNoHTMLEscapeAndASCII(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("R1", "a<b>&c é"))
	s := readStore(t, root)
	if !strings.Contains(s, `"a<b>&c \u00e9"`) {
		t.Errorf("escaping: %q", s)
	}
	if r := mustOK(t, refs("R1")); !strings.Contains(r.out, "a<b>&c é") {
		t.Errorf("roundtrip: %q", r.out)
	}
}

func TestReservedAndInvalidKeys(t *testing.T) {
	root := sandbox(t)
	cases := map[string]string{
		"-x":   "looks like a flag",
		"9abc": "must start with a non-digit",
	}
	for key, msg := range cases {
		r := refs(key, "v")
		if r.code != 1 || !strings.Contains(r.err, msg) || !strings.HasPrefix(r.err, "refs: ") {
			t.Errorf("%s: %+v", key, r)
		}
	}
	// "seed" with a value is a usage error, not a key
	if r := refs("seed", "v"); r.code != 1 || !strings.Contains(r.err, "usage: refs seed") {
		t.Errorf("seed as key: %+v", r)
	}
	// "hook" with a value is the hook subcommand; check the others reserved by name
	for _, key := range []string{"rm", "kind", "kinds", "home"} {
		if r := refs(key); r.code != 1 && key != "kinds" {
			t.Errorf("%s alone should fail usage: %+v", key, r)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json")); err == nil {
		t.Error("rejected keys wrote a store")
	}
}

func TestUnknownKindNagAndRelabel(t *testing.T) {
	sandbox(t)
	r := mustOK(t, refs("zz9", "a"))
	if r.out != "set zz9 [unknown]\n" || r.err != "refs: prefix 'zz' has no kind. Define it: refs kind zz NAME\n" {
		t.Errorf("nag: %+v", r)
	}
	mustOK(t, refs("zz10", "b"))
	mustOK(t, refs("other1", "c", "--kind", "o"))
	if r := mustOK(t, refs("kind", "zz", "thing")); r.out != "kind zz = thing\n" {
		t.Errorf("kind: %q", r.out)
	}
	want := "thing  zz9     a\nthing  zz10    b\no      other1  c\n"
	if r := mustOK(t, refs()); r.out != want {
		t.Errorf("relabel list:\n%q\nwant\n%q", r.out, want)
	}
	if r := mustOK(t, refs("kinds")); r.out != "other  o\nzz  thing\n" {
		t.Errorf("kinds: %q", r.out)
	}
	if r := mustOK(t, refs("zz11", "d")); r.err != "" || r.out != "set zz11 [thing]\n" {
		t.Errorf("known prefix: %+v", r)
	}
	if r := refs("kind", "a"); r.code != 1 || r.err != "refs: usage: refs kind PREFIX NAME\n" {
		t.Errorf("kind usage: %+v", r)
	}
}

func TestPrefixRule(t *testing.T) {
	for key, want := range map[string]string{"R20": "R", "pr#131": "pr#", "adr-3": "adr-", "abc": "abc", "a1b2": "a"} {
		if got := prefixOf(key); got != want {
			t.Errorf("prefixOf(%q) = %q want %q", key, got, want)
		}
	}
}

func TestLinkParsing(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("pr#1", "[Fix it](https://x.io/1)", "--kind", "pr#"))
	r := mustOK(t, refs("pr#1"))
	if !strings.Contains(r.out, "  value   [Fix it](https://x.io/1)\n  title   Fix it\n  url     https://x.io/1\n") {
		t.Errorf("show: %q", r.out)
	}
	if r := mustOK(t, refs()); r.out != "pr#  pr#1  Fix it — https://x.io/1\n" {
		t.Errorf("list: %q", r.out)
	}
	for i, v := range []string{"[t](u v)", "[](u)", "see [t](u)", "[t](u) tail"} {
		key := "n" + strings.Repeat("x", i)
		mustOK(t, refs(key, v))
		if strings.Contains(readStore(t, root), `"`+key+`": {`+"\n      \"kind\": \"unknown\",\n      \"value\": \""+v+`",`+"\n      \"set_at\": \"") == false {
			t.Errorf("%q should not parse as link", v)
		}
	}
	if s := readStore(t, root); strings.Count(s, `"url"`) != 1 {
		t.Errorf("url count: %s", s)
	}
}

func TestHome(t *testing.T) {
	root := sandbox(t)
	os.MkdirAll(filepath.Join(root, "docs", "sub"), 0o755)
	os.WriteFile(filepath.Join(root, "docs", "a.md"), []byte("x"), 0o644)
	mustOK(t, refs("R1", "v", "--kind", "req"))

	r := mustOK(t, refs("home", "R1", "docs/a.md:12"))
	if r.out != "home R1 = docs/a.md:12\n" || r.err != "" {
		t.Errorf("home: %+v", r)
	}
	if r := mustOK(t, refs()); r.out != "req  R1  v  @ docs/a.md:12\n" {
		t.Errorf("list: %q", r.out)
	}
	show := mustOK(t, refs("R1")).out
	if !strings.Contains(show, "  home    docs/a.md:12\n  open    file://"+filepath.Join(root, "docs/a.md")+"\n") {
		t.Errorf("show: %q", show)
	}

	// cwd-relative input from a subdirectory, stored root-relative
	t.Chdir(filepath.Join(root, "docs", "sub"))
	if r := mustOK(t, refs("home", "R1", "../a.md")); r.out != "home R1 = docs/a.md\n" {
		t.Errorf("cwd-relative: %+v", r)
	}
	t.Chdir(root)

	// setting a value without --home keeps the home; --home overrides
	mustOK(t, refs("R1", "v2"))
	if !strings.Contains(mustOK(t, refs("R1")).out, "  home    docs/a.md\n") {
		t.Error("home dropped on overwrite")
	}
	mustOK(t, refs("R1", "v3", "--home", "docs/a.md:3"))
	if !strings.Contains(mustOK(t, refs("R1")).out, "  home    docs/a.md:3\n") {
		t.Error("--home on set ignored")
	}

	// missing file warns but stores
	r = mustOK(t, refs("home", "R1", "nope.txt"))
	if r.out != "home R1 = nope.txt\n" || r.err != "refs: home 'nope.txt' does not exist\n" {
		t.Errorf("missing: %+v", r)
	}

	// --clear
	if r := mustOK(t, refs("home", "R1", "--clear")); r.out != "home R1 cleared\n" {
		t.Errorf("clear: %q", r.out)
	}
	if strings.Contains(mustOK(t, refs("R1")).out, "home") {
		t.Error("home not cleared")
	}
	if r := refs("home", "R1", "--clear", "x"); r.code != 1 || r.err != "refs: usage: refs home KEY --clear\n" {
		t.Errorf("clear usage: %+v", r)
	}
	if r := refs("R1", "--clear"); r.code != 1 || r.err != "refs: --clear applies only to: refs home KEY --clear\n" {
		t.Errorf("stray clear: %+v", r)
	}
	if r := refs("home", "R9", "--clear"); r.code != 1 || !strings.HasPrefix(r.err, "refs: no ref 'R9'") {
		t.Errorf("clear unknown: %+v", r)
	}
	if r := refs("home", "R1"); r.code != 1 || !strings.Contains(r.err, "usage: refs home KEY PATH") {
		t.Errorf("home usage: %+v", r)
	}
}

func TestHomeOutsideRoot(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("R1", "v", "--kind", "req"))
	outside := t.TempDir()
	for _, p := range []string{"../elsewhere.txt", outside, "/etc/hosts"} {
		r := refs("home", "R1", p)
		if r.code != 1 || !strings.Contains(r.err, "is outside the project root "+root+"; homes must live inside the project") {
			t.Errorf("%s: %+v", p, r)
		}
		r = refs("R2", "v", "--kind", "req", "--home", p)
		if r.code != 1 {
			t.Errorf("set with outside home %s: %+v", p, r)
		}
	}
	if strings.Contains(readStore(t, root), "R2") {
		t.Error("failed set was saved")
	}
	if r := refs("home", "R1", ":12"); r.code != 1 || r.err != "refs: home needs a PATH\n" {
		t.Errorf("empty path: %+v", r)
	}
}

func TestFlagErrors(t *testing.T) {
	sandbox(t)
	if r := refs("R1", "v", "--kind"); r.code != 1 || r.err != "refs: --kind needs a NAME\n" {
		t.Errorf("%+v", r)
	}
	if r := refs("R1", "v", "--home"); r.code != 1 || r.err != "refs: --home needs a PATH\n" {
		t.Errorf("%+v", r)
	}
	if r := refs("R1", "v", "--home="); r.code != 1 || r.err != "refs: home needs a PATH\n" {
		t.Errorf("%+v", r)
	}
	if r := refs("--bogus", "v"); r.code != 1 || !strings.Contains(r.err, "looks like a flag") {
		t.Errorf("%+v", r)
	}
}

func TestInsertionOrderAndUnknownFields(t *testing.T) {
	root := sandbox(t)
	os.MkdirAll(filepath.Join(root, ".ai"), 0o755)
	seed := `{
  "kinds": {
    "Z": "zed",
    "A": "aye"
  },
  "refs": {
    "Z2": {
      "kind": "zed",
      "value": "v",
      "custom": 1.5,
      "x": [
        true,
        null
      ]
    },
    "A1": {
      "kind": "unknown",
      "value": "w",
      "extra": {
        "k": "v"
      }
    },
    "B1": {
      "kind": "aye",
      "value": "u"
    }
  },
  "top": 1
}
`
	os.WriteFile(filepath.Join(root, ".ai", "refs.json"), []byte(seed), 0o644)

	want := "zed      Z2  v\nunknown  A1  w\naye      B1  u\n"
	if r := mustOK(t, refs()); r.out != want {
		t.Errorf("order: %q", r.out)
	}
	if r := mustOK(t, refs("kinds")); r.out != "Z  zed\nA  aye\n" {
		t.Errorf("kinds order: %q", r.out)
	}

	// relabel rewrites kind in place and keeps unknown fields and position
	mustOK(t, refs("kind", "A", "ay"))
	got := readStore(t, root)
	wantStore := strings.Replace(seed, `"A": "aye"`, `"A": "ay"`, 1)
	wantStore = strings.Replace(wantStore, `"kind": "unknown"`, `"kind": "ay"`, 1)
	if got != wantStore {
		t.Errorf("relabel changed more than it should:\n%s", got)
	}

	// new keys append; removal keeps the rest in order
	mustOK(t, refs("Z3", "new"))
	mustOK(t, refs("rm", "A1"))
	if r := mustOK(t, refs()); r.out != "zed  Z2  v\naye  B1  u\nzed  Z3  new\n" {
		t.Errorf("after edits: %q", r.out)
	}
	// overwrite keeps position
	mustOK(t, refs("Z2", "v2"))
	if r := mustOK(t, refs()); !strings.HasPrefix(r.out, "zed  Z2  v2\n") {
		t.Errorf("overwrite moved: %q", r.out)
	}
	if s := readStore(t, root); !strings.Contains(s, `"top": 1`) {
		t.Errorf("top-level field lost: %s", s)
	}
}

func TestMalformedStoreLoadsEmpty(t *testing.T) {
	root := sandbox(t)
	os.MkdirAll(filepath.Join(root, ".ai"), 0o755)
	for _, body := range []string{"not json", "[1,2]", "", `{"kinds":[],"refs":5}`, `{"a":1} trailing`} {
		os.WriteFile(filepath.Join(root, ".ai", "refs.json"), []byte(body), 0o644)
		if r := mustOK(t, refs()); !strings.HasPrefix(r.out, "no refs in ") {
			t.Errorf("%q: %q", body, r.out)
		}
		mustOK(t, refs("Q1", "v", "--kind", "q"))
		if r := mustOK(t, refs()); r.out != "q  Q1  v\n" {
			t.Errorf("%q after set: %q", body, r.out)
		}
	}
}

func TestNonGitDirUsesCwd(t *testing.T) {
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	t.Chdir(dir)
	mustOK(t, refs("N1", "v", "--kind", "n"))
	if _, err := os.Stat(filepath.Join(dir, ".ai", "refs.json")); err != nil {
		t.Error(err)
	}
}

func TestHelp(t *testing.T) {
	for _, a := range []string{"--help", "-h", "help"} {
		r := mustOK(t, refs(a))
		if !strings.HasPrefix(r.out, "refs — per-project key-value store") || strings.Contains(strings.ToLower(r.out), "python") {
			t.Errorf("%s: %q", a, r.out)
		}
	}
}

// ---- hook ----

type hookOut struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

func hookCtx(t *testing.T, stdin string) string {
	t.Helper()
	r := mustOK(t, refsIn(stdin, "hook"))
	var h hookOut
	if err := json.Unmarshal([]byte(r.out), &h); err != nil {
		t.Fatalf("hook output is not JSON: %q", r.out)
	}
	if h.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Errorf("event name: %q", r.out)
	}
	return h.HookSpecificOutput.AdditionalContext
}

func TestHookEmpty(t *testing.T) {
	root := sandbox(t)
	if got := hookCtx(t, `{"cwd":"`+root+`"}`); got != hookEmpty {
		t.Errorf("got %q", got)
	}
}

func TestHookCounts(t *testing.T) {
	root := sandbox(t)
	other, _ := filepath.EvalSymlinks(t.TempDir())
	t.Chdir(other) // prove the hook follows payload cwd, not process cwd
	mustOK(t, refsIn("", "hook"))

	t.Chdir(root)
	mustOK(t, refs("R1", "a", "--kind", "r"))
	t.Chdir(other)
	payload := `{"cwd":"` + root + `"}`
	if got := hookCtx(t, payload); got != hookText+" This project has 1 ref." {
		t.Errorf("one: %q", got)
	}
	t.Chdir(root)
	mustOK(t, refs("R2", "b"))
	t.Chdir(other)
	if got := hookCtx(t, payload); got != hookText+" This project has 2 refs." {
		t.Errorf("two: %q", got)
	}
	// a store with zero refs still counts as populated: plural
	os.WriteFile(filepath.Join(root, ".ai", "refs.json"), []byte(`{"kinds":{},"refs":{}}`), 0o644)
	if got := hookCtx(t, payload); got != hookText+" This project has 0 refs." {
		t.Errorf("zero: %q", got)
	}
}

func TestHookGarbageStdin(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("R1", "a", "--kind", "r"))
	for _, in := range []string{"", "garbage", "[1]", "null", `{"cwd":5}`, `{"cwd":""}`, `{"cwd":`} {
		// falls back to process cwd, which is the populated sandbox
		if got := hookCtx(t, in); got != hookText+" This project has 1 ref." {
			t.Errorf("%q: %q", in, got)
		}
	}
	// payload cwd that does not exist: empty-project text, still valid JSON, exit 0
	if got := hookCtx(t, `{"cwd":"`+filepath.Join(root, "missing")+`"}`); got != hookEmpty {
		t.Errorf("missing cwd: %q", got)
	}
}

func TestHookIgnoresMalformedStore(t *testing.T) {
	root := sandbox(t)
	os.MkdirAll(filepath.Join(root, ".ai"), 0o755)
	os.WriteFile(filepath.Join(root, ".ai", "refs.json"), []byte("junk"), 0o644)
	if got := hookCtx(t, `{"cwd":"`+root+`"}`); got != hookText+" This project has 0 refs." {
		t.Errorf("got %q", got)
	}
}

func TestSeed(t *testing.T) {
	root := sandbox(t)
	r := mustOK(t, refs("seed"))
	if r.out != hookEmpty+"\n" || r.err != "" {
		t.Errorf("seed without store: %+v", r)
	}
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json")); err == nil {
		t.Error("seed wrote a store")
	}
	mustOK(t, refs("R1", "x", "--kind", "req"))
	before := readStore(t, root)
	if r := mustOK(t, refs("seed")); r.out != hookEmpty+"\n" || r.err != "" {
		t.Errorf("seed with store: %+v", r)
	}
	if readStore(t, root) != before {
		t.Error("seed changed the store")
	}
}

func TestSeedRejectsArgs(t *testing.T) {
	sandbox(t)
	r := refs("seed", "extra")
	if r.code != 1 || r.out != "" || r.err != "refs: usage: refs seed\n" {
		t.Errorf("seed args: %+v", r)
	}
}

func TestInstallIsAKey(t *testing.T) {
	sandbox(t)
	if r := mustOK(t, refs("install", "v", "--kind", "k")); r.out != "set install [k]\n" {
		t.Errorf("install key: %+v", r)
	}
}

// ---- kind filter ----

func TestKindFilter(t *testing.T) {
	sandbox(t)
	mustOK(t, refs("TRU-01", "t one", "--kind", "truth"))
	mustOK(t, refs("TRU-02", "t two"))
	mustOK(t, refs("NEV-01", "n one", "--kind", "never"))
	for _, f := range []string{"truth", "TRU-"} {
		r := mustOK(t, refs("--kind", f))
		if !strings.Contains(r.out, "TRU-01") || !strings.Contains(r.out, "TRU-02") || strings.Contains(r.out, "NEV-01") || r.err != "" {
			t.Errorf("--kind %s: %+v", f, r)
		}
	}
	r := mustOK(t, refs("--kind=never"))
	if strings.Count(r.out, "\n") != 1 || !strings.Contains(r.out, "NEV-01") {
		t.Errorf("--kind=never: %+v", r)
	}
	r = mustOK(t, refs("--kind", "nope"))
	if r.out != "" || strings.Count(r.err, "\n") != 1 {
		t.Errorf("no match: %+v", r)
	}
}

func TestHookCanonBlurb(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("TRU-01", "a", "--kind", "truth"))
	mustOK(t, refs("TRU-02", "b"))
	mustOK(t, refs("NEV-01", "c", "--kind", "never"))
	payload := `{"cwd":"` + root + `"}`
	base := hookText + " This project has 3 refs."
	if got := hookCtx(t, payload); got != base {
		t.Errorf("no canon: %q", got)
	}
	canon := filepath.Join(root, "docs", "canon")
	os.MkdirAll(canon, 0o755)
	for _, f := range []string{"NEVERS.md", "TRUTHS.md", "OPEN_QUESTIONS.md"} {
		os.WriteFile(filepath.Join(canon, f), []byte("x"), 0o644)
	}
	got := hookCtx(t, payload)
	want := base + " Canon in docs/canon/: TRUTHS.md (2 truth), NEVERS.md (1 never), OPEN_QUESTIONS.md (0 open)."
	if !strings.HasPrefix(got, want) || !strings.Contains(got, "`refs --kind truth`") || strings.Contains(got, "REQUIREMENTS.md") {
		t.Errorf("some canon: %q", got)
	}
}

// ---- ids ----

func TestIdsDefaultAndSet(t *testing.T) {
	root := sandbox(t)
	if r := mustOK(t, refs("ids")); r.out != "fluid\n" {
		t.Errorf("default: %q", r.out)
	}
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json")); err == nil {
		t.Error("reading ids wrote a store")
	}
	if r := mustOK(t, refs("ids", "fixed")); r.out != "ids = fixed\n" {
		t.Errorf("set fixed: %q", r.out)
	}
	if r := mustOK(t, refs("ids")); r.out != "fixed\n" {
		t.Errorf("read fixed: %q", r.out)
	}
	mustOK(t, refs("ids", "fluid"))
	if r := mustOK(t, refs("ids")); r.out != "fluid\n" {
		t.Errorf("read fluid: %q", r.out)
	}
}

func TestIdsBadValueAndReserved(t *testing.T) {
	root := sandbox(t)
	if !reserved["ids"] {
		t.Error("ids is not reserved")
	}
	for _, args := range [][]string{{"ids", "frozen"}, {"ids", "fixed", "x"}} {
		r := refs(args...)
		if r.code != 1 || !strings.Contains(r.err, "usage: refs ids [fixed|fluid]") {
			t.Errorf("%v: %+v", args, r)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json")); err == nil {
		t.Error("bad ids value wrote a store")
	}
}

func TestIdsPersistsAcrossSets(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("R1", "a", "--kind", "r"))
	if s := readStore(t, root); strings.Contains(s, `"ids"`) {
		t.Errorf("ids written before set: %s", s)
	}
	mustOK(t, refs("ids", "fixed"))
	mustOK(t, refs("R2", "b"))
	mustOK(t, refs("rm", "R1"))
	mustOK(t, refs("kind", "X", "x"))
	if r := mustOK(t, refs("ids")); r.out != "fixed\n" {
		t.Errorf("lost: %q", r.out)
	}
	if s := readStore(t, root); !strings.Contains(s, `"ids": "fixed"`) {
		t.Errorf("store: %s", s)
	}
}

func TestHookIdsSentence(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("TRU-01", "a", "--kind", "truth"))
	payload := `{"cwd":"` + root + `"}`
	if got := hookCtx(t, payload); strings.Contains(got, "Canon IDs") {
		t.Errorf("no canon files, sentence present: %q", got)
	}
	canon := filepath.Join(root, "docs", "canon")
	os.MkdirAll(canon, 0o755)
	os.WriteFile(filepath.Join(canon, "TRUTHS.md"), []byte("x"), 0o644)
	if got := hookCtx(t, payload); !strings.HasSuffix(got, " Canon IDs are fluid and may be renumbered; cite an ID with its title.") {
		t.Errorf("fluid: %q", got)
	}
	mustOK(t, refs("ids", "fixed"))
	if got := hookCtx(t, payload); !strings.HasSuffix(got, " Canon IDs are fixed.") {
		t.Errorf("fixed: %q", got)
	}
}
