// refs_test.go — tests for the refs CLI and hook.
//
// Responsibilities:
//   - run the CLI in-process inside a temp dir with its own `git init`
//   - cover list/show/set/rm/kind/kinds/home/--clear, key validation, link parsing
//   - check insertion order, unknown-field preservation, home rules
//   - check hook output for empty, populated and garbage input
//   - cover `refs sync` (canon, dry-run, set_at stability, fake cactus) and the canon auto-sync
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
	"time"
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
	guards := "\nGuards (read the home before changing what they cover):\nTRU-01 a\nTRU-02 b\nNEV-01 c"
	if got := hookCtx(t, payload); got != base+guards {
		t.Errorf("no canon: %q", got)
	}
	canon := filepath.Join(root, "docs", "canon")
	os.MkdirAll(canon, 0o755)
	// auto-sync mirrors the md into refs, so the files must agree with the refs set above
	for f, body := range map[string]string{
		"NEVERS.md": "## NEV-01 c\n", "TRUTHS.md": "## TRU-01 a\n## TRU-02 b\n", "OPEN_QUESTIONS.md": "x",
	} {
		os.WriteFile(filepath.Join(canon, f), []byte(body), 0o644)
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

// ---- sync ----

func writeCanon(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSyncCanon(t *testing.T) {
	root := sandbox(t)
	mustOK(t, refs("REQ-02", "stale", "--kind", "requirement")) // set before any canon file exists
	mustOK(t, refs("REQ-77", "gone from md", "--kind", "requirement"))
	writeCanon(t, root, "docs/canon/REQUIREMENTS.md", "# Reqs\n\n## REQ-01 Do the thing\nbody\n\n## REQ-02 Old one\nRetired: 2026-01-01 gone\n")
	writeCanon(t, root, "docs/canon/DECISIONS.md", "## DEC-01 · 2026-05-01 — Use Go\nbody\n")
	writeCanon(t, root, "docs/canon/OPEN_QUESTIONS.md", "## OQ-01 Is it?\nResolved: 2026-02-02 → DEC-01\n\n## OQ-02 Still open\n")
	writeCanon(t, root, "docs/canon/TRUTHS.md", "## TRU-01 A truth\n\n## Proposed\n\n### TRU-09 not an entry\n")

	r := mustOK(t, refs("sync"))
	if !strings.HasPrefix(r.out, "sync: 4 set, 2 removed") {
		t.Errorf("summary: %q", r.out)
	}
	list := mustOK(t, refs()).out
	for _, want := range []string{"Do the thing  @ docs/canon/REQUIREMENTS.md:3", "Use Go  @ docs/canon/DECISIONS.md:1",
		"Still open  @ docs/canon/OPEN_QUESTIONS.md:4", "A truth  @ docs/canon/TRUTHS.md:1"} {
		if !strings.Contains(list, want) {
			t.Errorf("missing %q in\n%s", want, list)
		}
	}
	for _, bad := range []string{"REQ-02", "REQ-77", "OQ-01", "TRU-09"} {
		if strings.Contains(list, bad) {
			t.Errorf("%s should not be listed:\n%s", bad, list)
		}
	}
	if r := mustOK(t, refs("TRU-01")); !strings.Contains(r.out, "[truth]") {
		t.Errorf("kind: %q", r.out)
	}
}

func TestSyncLegacyInvariants(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/INVARIANTS.md", "## INV-01 Always valid\n")
	mustOK(t, refs("sync"))
	r := mustOK(t, refs("INV-01"))
	if !strings.Contains(r.out, "[invariant]") || !strings.Contains(r.out, "docs/INVARIANTS.md:1") {
		t.Errorf("legacy: %q", r.out)
	}
	// TRUTHS.md wins over the legacy name once it exists
	writeCanon(t, root, "docs/canon/TRUTHS.md", "## TRU-01 New\n")
	mustOK(t, refs("sync"))
	if r := mustOK(t, refs("TRU-01")); !strings.Contains(r.out, "[truth]") {
		t.Errorf("TRUTHS.md not preferred: %q", r.out)
	}
}

func TestSyncDryRunWritesNothing(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 Never X\n")
	r := mustOK(t, refs("sync", "--dry-run"))
	if !strings.Contains(r.out, "would set NEV-01") || !strings.Contains(r.out, "1 set") {
		t.Errorf("dry output: %q", r.out)
	}
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json")); err == nil {
		t.Error("dry-run wrote refs.json")
	}
}

func TestSyncSetAtStable(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 Never X\n## NEV-02 Never Y\n")
	mustOK(t, refs("sync"))
	first := readStore(t, root)
	// age set_at so a rewrite would show
	aged := strings.ReplaceAll(first, `"set_at": "`, `"set_at": "1999-`)
	os.WriteFile(filepath.Join(root, ".ai", "refs.json"), []byte(aged), 0o644)
	r := mustOK(t, refs("sync"))
	if !strings.HasPrefix(r.out, "sync: 0 set, 0 removed, 2 unchanged") {
		t.Errorf("resync: %q", r.out)
	}
	if got := readStore(t, root); got != aged {
		t.Errorf("refs.json churned:\n%s", got)
	}
	// change one: only it gets a fresh set_at
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 Never X\n## NEV-02 Never Z\n")
	mustOK(t, refs("sync"))
	got := readStore(t, root)
	if strings.Count(got, `"set_at": "1999-`) != 1 {
		t.Errorf("expected one aged set_at left:\n%s", got)
	}
}

func TestSyncUsage(t *testing.T) {
	sandbox(t)
	if r := refs("sync", "--bogus"); r.code == 0 || !strings.Contains(r.err, "usage: refs sync") {
		t.Errorf("bogus flag: %+v", r)
	}
	if !reserved["sync"] {
		t.Error("sync is not reserved")
	}
}

func TestAutoSync(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 Never X\n")
	// refs.json missing, canon present: any read command syncs first
	if r := mustOK(t, refs()); !strings.Contains(r.out, "NEV-01") || r.err != "" {
		t.Fatalf("missing store: %+v", r)
	}
	// md edited but older than refs.json: not re-read
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 Never X\n## NEV-02 Never Y\n")
	old := time.Now().Add(-time.Hour)
	os.Chtimes(filepath.Join(root, "docs/canon/NEVERS.md"), old, old)
	if r := mustOK(t, refs()); strings.Contains(r.out, "NEV-02") {
		t.Errorf("synced although md is older: %q", r.out)
	}
	// md newer than refs.json: re-read, for show too
	past := time.Now().Add(-2 * time.Hour)
	store := filepath.Join(root, ".ai", "refs.json")
	os.Chtimes(store, past, past)
	os.Chtimes(filepath.Join(root, "docs/canon/NEVERS.md"), past.Add(time.Hour), past.Add(time.Hour))
	if r := mustOK(t, refs("NEV-02")); !strings.Contains(r.out, "Never Y") {
		t.Errorf("not synced although md is newer: %q", r.out)
	}
	// a sync with no changes still bumps refs.json past the md, so the next call is stat-only
	os.Chtimes(store, past, past)
	mustOK(t, refs("NEV-01"))
	if st, _ := os.Stat(store); !st.ModTime().After(past.Add(time.Hour)) {
		t.Errorf("refs.json mtime not advanced: %v", st.ModTime())
	}
}

func TestAutoSyncNotForSeedOrHelp(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 Never X\n")
	mustOK(t, refs("seed"))
	mustOK(t, refs("--help"))
	if _, err := os.Stat(filepath.Join(root, ".ai", "refs.json")); err == nil {
		t.Error("seed/help synced")
	}
}

func TestHookAutoSyncAndMalformedCanon(t *testing.T) {
	root := sandbox(t)
	payload := `{"cwd":"` + root + `"}`
	writeCanon(t, root, "docs/canon/TRUTHS.md", "## TRU-01 One\n")
	if got := hookCtx(t, payload); !strings.Contains(got, "This project has 1 ref.") {
		t.Errorf("hook did not sync: %q", got)
	}
	// binary junk, a directory where a file belongs, huge line: still valid JSON, exit 0
	writeCanon(t, root, "docs/canon/TRUTHS.md", "\xff\xfe## TRU-01\n## TRU-02\x00 \xc3\n"+strings.Repeat("z", 1<<20))
	future := time.Now().Add(time.Hour)
	os.Chtimes(filepath.Join(root, "docs/canon/TRUTHS.md"), future, future)
	os.MkdirAll(filepath.Join(root, "docs/canon/NEVERS.md"), 0o755)
	hookCtx(t, payload)
	hookCtx(t, `{"cwd":"`+root+`"}`)
}

func TestSyncCactusFake(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, ".ai/cactus/q1-aaa-first.md", "x")
	writeCanon(t, root, ".ai/cactus/q2-bbb-second.md", "x")
	writeCanon(t, root, ".ai/cactus/q3-ccc-third.md", "x")
	bin := t.TempDir()
	rows := `[{"key":"q1","text":"Pick\none","answer":{"selected":["a","b"],"text":"why"},"pr":null,"thread":"x"},` +
		`{"key":"q2","text":"Skip","answer":{"skipped":true},"pr":null},` +
		`{"key":"q3","text":"Denied","answer":{"selected":["a"]},"pr":null,"thread":"denied"}]`
	os.WriteFile(filepath.Join(bin, "cactus"), []byte("#!/bin/sh\ncat <<'EOF'\n"+rows+"\nEOF\n"), 0o755)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	mustOK(t, refs("q9", "stale", "--kind", "answer"))
	r := mustOK(t, refs("sync", "--cactus"))
	if !strings.HasPrefix(r.out, "sync: 1 set, 1 removed") {
		t.Errorf("summary: %q", r.out)
	}
	got := mustOK(t, refs("q1")).out
	if !strings.Contains(got, "Pick one → a, b — why") || !strings.Contains(got, ".ai/cactus/q1-aaa-first.md:1") {
		t.Errorf("q1: %q", got)
	}
	if strings.Contains(mustOK(t, refs()).out, "q2") {
		t.Error("skipped answer was mirrored")
	}
}

// ---- applies, guards ----

func TestSyncAppliesStoredAndChanged(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 No X\nApplies to: go.mod, `*.go`\n\n## NEV-02 No Y\nApplies to: all\n\n## NEV-03 No Z\n")
	mustOK(t, refs("sync"))
	got := readStore(t, root)
	if !strings.Contains(got, `"applies": [
        "go.mod",
        "*.go"
      ]`) || !strings.Contains(got, `"all"`) {
		t.Errorf("applies not stored:\n%s", got)
	}
	if strings.Count(got, `"applies"`) != 2 {
		t.Errorf("NEV-03 should have no applies:\n%s", got)
	}
	// unknown entry field survives; applies change bumps set_at
	aged := strings.ReplaceAll(got, `"set_at": "`, `"set_at": "1999-`)
	aged = strings.Replace(aged, `"kind": "never",`, `"kind": "never", "extra": 1,`, 1)
	os.WriteFile(filepath.Join(root, ".ai", "refs.json"), []byte(aged), 0o644)
	r := mustOK(t, refs("sync"))
	if !strings.HasPrefix(r.out, "sync: 0 set") {
		t.Errorf("resync: %q", r.out)
	}
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 No X\nApplies to: *.go\n\n## NEV-02 No Y\nApplies to: all\n\n## NEV-03 No Z\n")
	r = mustOK(t, refs("sync"))
	if !strings.HasPrefix(r.out, "sync: 1 set") {
		t.Errorf("applies change: %q", r.out)
	}
	got = readStore(t, root)
	if strings.Count(got, `"set_at": "1999-`) != 2 || !strings.Contains(got, `"extra": 1`) {
		t.Errorf("set_at/unknown field:\n%s", got)
	}
}

func TestHookGuardsLine(t *testing.T) {
	root := sandbox(t)
	payload := `{"cwd":"` + root + `"}`
	mustOK(t, refs("R1", "req", "--kind", "requirement"))
	if got := hookCtx(t, payload); strings.Contains(got, "Guards") {
		t.Errorf("guards without truths: %q", got)
	}
	mustOK(t, refs("TRU-01", "t one", "--kind", "truth"))
	mustOK(t, refs("NEV-01", "n one", "--kind", "never"))
	got := hookCtx(t, payload)
	if !strings.HasSuffix(got, "\nGuards (read the home before changing what they cover):\nTRU-01 t one\nNEV-01 n one") {
		t.Errorf("guards: %q", got)
	}
	for i := 2; i <= 45; i++ {
		mustOK(t, refs("TRU-"+itoa(100+i), "t", "--kind", "truth"))
	}
	got = hookCtx(t, payload)
	// 45 truths + 1 never = 46 guards, 40 listed
	if strings.Count(got, "\nTRU-") != 40 || strings.Contains(got, "NEV-01 n one") || !strings.HasSuffix(got, "\n+6 more: refs --kind truth / never") {
		t.Errorf("cap: %q", got)
	}
}

func guardOut(t *testing.T, stdin string) string {
	t.Helper()
	r := mustOK(t, refsIn(stdin, "guard"))
	if strings.Contains(r.out, "permissionDecision") {
		t.Errorf("permissionDecision in output: %q", r.out)
	}
	if r.out != "" {
		var h struct {
			H struct {
				Event string `json:"hookEventName"`
				Ctx   string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(r.out), &h); err != nil || h.H.Event != "PreToolUse" {
			t.Fatalf("bad guard JSON: %q", r.out)
		}
		return h.H.Ctx
	}
	return ""
}

func TestGuard(t *testing.T) {
	root := sandbox(t)
	writeCanon(t, root, "docs/canon/TRUTHS.md", "## TRU-01 Store is JSON\nApplies to: store.go, internal/**\n\n## TRU-02 Everywhere\nApplies to: all\n")
	writeCanon(t, root, "docs/canon/NEVERS.md", "## NEV-01 No deps\nApplies to: *.go\n\n## NEV-02 Docs\nApplies to: docs/canon/*.md\n")
	call := func(file string) string {
		return guardOut(t, `{"cwd":"`+root+`","tool_name":"Edit","tool_input":{"file_path":"`+file+`"}}`)
	}
	// match: exact path, basename glob (*.go on a nested file), all
	got := call(filepath.Join(root, "store.go"))
	for _, w := range []string{"refs guards for store.go: ", "TRU-01 Store is JSON @ docs/canon/TRUTHS.md:1", "TRU-02 Everywhere", "NEV-01 No deps @ docs/canon/NEVERS.md:1"} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in %q", w, got)
		}
	}
	if strings.Contains(got, "NEV-02") {
		t.Errorf("NEV-02 should not match store.go: %q", got)
	}
	got = call(filepath.Join(root, "pkg", "deep", "x.go"))
	if !strings.Contains(got, "NEV-01") || strings.Contains(got, "TRU-01") || !strings.Contains(got, "TRU-02") {
		t.Errorf("basename glob: %q", got)
	}
	// ** prefix and directory form
	if got = call(filepath.Join(root, "internal", "a", "b.txt")); !strings.Contains(got, "TRU-01") {
		t.Errorf("internal/**: %q", got)
	}
	if got = call(filepath.Join(root, "docs", "canon", "NEVERS.md")); !strings.Contains(got, "NEV-02") {
		t.Errorf("docs/canon/*.md: %q", got)
	}
	// no match: only `all` truth would match, so drop it
	writeCanon(t, root, "docs/canon/TRUTHS.md", "## TRU-01 Store is JSON\nApplies to: store.go\n")
	future := time.Now().Add(time.Minute)
	os.Chtimes(filepath.Join(root, "docs/canon/TRUTHS.md"), future, future)
	if got = call(filepath.Join(root, "README.md")); got != "" {
		t.Errorf("no match should be empty: %q", got)
	}
	// outside root, .ai skipped, relative path, notebook_path
	if got = call("/etc/hosts"); got != "" {
		t.Errorf("outside root: %q", got)
	}
	if got = call(filepath.Join(root, ".ai", "refs.json")); got != "" {
		t.Errorf(".ai: %q", got)
	}
	if got = call("store.go"); !strings.Contains(got, "TRU-01") {
		t.Errorf("relative path: %q", got)
	}
	nb := guardOut(t, `{"cwd":"`+root+`","tool_input":{"notebook_path":"`+filepath.Join(root, "store.go")+`"}}`)
	if !strings.Contains(nb, "TRU-01") {
		t.Errorf("notebook_path: %q", nb)
	}
	// malformed stdin: empty output, exit 0
	for _, in := range []string{"", "garbage", "null", "[1]", `{"cwd":5}`, `{"tool_input":5}`, `{"cwd":`} {
		if r := mustOK(t, refsIn(in, "guard")); r.out != "" {
			t.Errorf("%q: %q", in, r.out)
		}
	}
	if !reserved["guard"] {
		t.Error("guard is not reserved")
	}
}

func TestHooksJSONRegistersGuard(t *testing.T) {
	b, err := os.ReadFile("hooks/hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	var h struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &h); err != nil {
		t.Fatal(err)
	}
	g := h.Hooks["PreToolUse"]
	if len(g) != 1 || g[0].Matcher != "Edit|Write|MultiEdit|NotebookEdit" || g[0].Hooks[0].Command != "${CLAUDE_PLUGIN_ROOT}/bin/refs guard" {
		t.Errorf("PreToolUse: %+v", g)
	}
	if len(h.Hooks["SessionStart"]) != 1 {
		t.Error("SessionStart entry lost")
	}
}
