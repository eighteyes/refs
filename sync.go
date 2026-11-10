// sync.go — `refs sync` and the canon auto-sync: mirror docs/canon/*.md (and optionally GitHub and cactus) into refs.json.
//
// Responsibilities:
//   - parse `## <PREFIX>NN text` headings of the six canon files into refs (home = file:line)
//   - read each block's `Applies to:` globs into an `applies` list
//   - drop refs for entries marked Retired/Superseded/Resolved and refs whose entry is gone
//   - --github: open issues and PRs via `gh`; --cactus: answered cactus rows via `cactus get --json`
//   - apply all changes in-process against one loaded store and save once, atomically
//   - keep set_at stable for refs whose kind/value/home did not change
//   - autoSync: run the canon part silently when a canon file is newer than refs.json
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// canonSpec names one canon file; legacy* apply when neither the file nor its docs/ twin exists.
type canonSpec struct {
	prefix, kind, name                   string
	legacyPrefix, legacyKind, legacyName string
}

var canonSpecs = []canonSpec{
	{"REQ-", "requirement", "REQUIREMENTS", "", "", ""},
	{"TRU-", "truth", "TRUTHS", "INV-", "invariant", "INVARIANTS"},
	{"NEV-", "never", "NEVERS", "", "", ""},
	{"DEC-", "decision", "DECISIONS", "", "", ""},
	{"INQ-", "inquiry", "INQUIRY", "", "", ""},
	{"OQ-", "open", "OPEN_QUESTIONS", "", "", ""},
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Mode().IsRegular()
}

// resolve picks the file, prefix and kind for a spec, or ok=false when no file exists.
func (c canonSpec) resolve(root string) (rel, prefix, kind string, ok bool) {
	pick := func(name string) (string, bool) {
		for _, r := range []string{"docs/canon/" + name + ".md", "docs/" + name + ".md"} {
			if isFile(filepath.Join(root, r)) {
				return r, true
			}
		}
		return "", false
	}
	if r, found := pick(c.name); found {
		return r, c.prefix, c.kind, true
	}
	if c.legacyName != "" {
		if r, found := pick(c.legacyName); found {
			return r, c.legacyPrefix, c.legacyKind, true
		}
	}
	return "", "", "", false
}

var (
	appliesRE  = regexp.MustCompile(`^(?:- )?\**Applies to\**:\**\s*(.*)$`)
	retiredRE  = regexp.MustCompile(`^(Retired:|- \*\*Retired|Superseded by|Resolved:)`)
	decTitleRE = regexp.MustCompile(`^[^—]*— `)
)

// canonEntry is one `## ` heading with its block's verdict.
type canonEntry struct {
	line    int
	id      string
	text    string
	retired bool
	applies []string
}

// parseApplies splits an `Applies to:` value into globs; backticks and blanks drop out.
func parseApplies(v string) []string {
	var out []string
	for _, g := range strings.Split(v, ",") {
		if g = strings.TrimSpace(strings.Trim(strings.TrimSpace(g), "`")); g != "" {
			out = append(out, g)
		}
	}
	return out
}

// parseCanon reads headings that start with prefix+digits; a block runs to the next `## ` heading.
func parseCanon(content, prefix string) []canonEntry {
	headRE := regexp.MustCompile(`^## ` + regexp.QuoteMeta(prefix) + `[0-9]+`)
	var out []canonEntry
	var cur *canonEntry
	flush := func() {
		if cur != nil {
			out = append(out, *cur)
			cur = nil
		}
	}
	for i, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(ln, "## ") {
			flush()
			if loc := headRE.FindStringIndex(ln); loc != nil {
				rest := ln[loc[1]:]
				if rest != "" {
					rest = rest[1:] // the separator after the ID, whatever it is
				}
				if prefix == "DEC-" {
					rest = decTitleRE.ReplaceAllString(rest, "")
				}
				cur = &canonEntry{line: i + 1, id: ln[3:loc[1]], text: rest}
			}
			continue
		}
		if cur == nil {
			continue
		}
		if retiredRE.MatchString(ln) {
			cur.retired = true
		}
		if m := appliesRE.FindStringSubmatch(ln); m != nil && cur.applies == nil {
			cur.applies = parseApplies(m[1])
		}
	}
	flush()
	return out
}

// syncer applies changes to a loaded store and counts them.
type syncer struct {
	store, kinds, refs *omap
	dry                bool
	out, warn          io.Writer
	set, rm, same      int
	dirty              bool
}

func newSyncer(store *omap, dry bool, out, warn io.Writer) *syncer {
	return &syncer{store: store, kinds: asObj(store.Get("kinds")), refs: asObj(store.Get("refs")), dry: dry, out: out, warn: warn}
}

func (s *syncer) say(format string, args ...any) {
	if s.dry {
		fmt.Fprintf(s.out, format, args...)
	}
}

// setKind names a prefix and relabels `unknown` refs under it, like `refs kind`.
func (s *syncer) setKind(prefix, name string) {
	if s.kinds.Get(prefix) != name {
		s.say("would name kind %s = %s\n", prefix, name)
		s.kinds.Set(prefix, name)
		s.dirty = true
	}
	for _, key := range s.refs.keys {
		if r, ok := s.refs.Get(key).(*omap); ok && r.Get("kind") == "unknown" && prefixOf(key) == prefix {
			r.Set("kind", name)
			s.dirty = true
		}
	}
}

// appliesOf reads a ref's stored applies list.
func appliesOf(r *omap) []string {
	var out []string
	if arr, ok := r.Get("applies").([]any); ok {
		for _, v := range arr {
			out = append(out, pystr(v))
		}
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// setRef writes a ref unless kind, value, home and applies already match. A changed ref keeps its unknown fields.
func (s *syncer) setRef(key, value, kind, home string, applies []string) {
	var resolved any = "unknown"
	if kv, ok := s.kinds.Lookup(prefixOf(key)); ok {
		resolved = kv
	}
	old, existed := s.refs.Get(key).(*omap)
	if existed && old.Get("kind") == resolved && old.Get("value") == value &&
		pystr(old.Get("home")) == home && sameStrings(appliesOf(old), applies) {
		s.same++
		return
	}
	entry := newOmap()
	if existed {
		entry = old
	}
	entry.Set("kind", resolved)
	entry.Set("value", value)
	entry.Set("set_at", now())
	entry.Delete("title")
	entry.Delete("url")
	if m := linkRE.FindStringSubmatch(value); m != nil {
		entry.Set("title", m[1])
		entry.Set("url", m[2])
	}
	entry.Delete("home")
	if home != "" {
		entry.Set("home", home)
	}
	entry.Delete("applies")
	if len(applies) > 0 {
		list := make([]any, len(applies))
		for i, g := range applies {
			list[i] = g
		}
		entry.Set("applies", list)
	}
	s.say("would set %s [%s] %s @ %s\n", key, pystr(resolved), value, home)
	s.refs.Set(key, entry)
	s.set++
	s.dirty = true
}

func (s *syncer) rmRef(key string) {
	if _, ok := s.refs.Lookup(key); !ok {
		return
	}
	s.say("would rm %s\n", key)
	s.refs.Delete(key)
	s.rm++
	s.dirty = true
}

// prune removes refs whose key matches keep-nothing: has prefix and (if re != nil) the regexp, and is not live.
func (s *syncer) prune(prefix string, re *regexp.Regexp, live map[string]bool) {
	for _, k := range append([]string(nil), s.refs.keys...) {
		if strings.HasPrefix(k, prefix) && (re == nil || re.MatchString(k)) && !live[k] {
			s.rmRef(k)
		}
	}
}

// canon mirrors the six canon files. The first read error is returned after the rest are done.
func (s *syncer) canon(root string) error {
	var first error
	for _, spec := range canonSpecs {
		rel, prefix, kind, ok := spec.resolve(root)
		if !ok {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		s.setKind(prefix, kind)
		live := map[string]bool{}
		for _, e := range parseCanon(string(data), prefix) {
			if e.retired {
				s.rmRef(e.id)
				continue
			}
			s.setRef(e.id, e.text, kind, fmt.Sprintf("%s:%d", rel, e.line), e.applies)
			live[e.id] = true
		}
		s.prune(prefix, nil, live)
	}
	return first
}

// ghRows lists open issues or PRs as number/title/url.
type ghRow struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

func ghRows(root, what string) ([]ghRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", what, "list", "--state", "open", "--limit", "200", "--json", "number,title,url")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	var rows []ghRow
	if err != nil {
		return nil, fmt.Errorf("gh %s list: %v %s", what, err, strings.TrimSpace(stderr.String()))
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, fmt.Errorf("gh %s list: %v", what, err)
	}
	return rows, nil
}

// github mirrors open issues and PRs. Any gh failure skips the source, removing nothing.
func (s *syncer) github(root string) {
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Fprintln(s.warn, "refs: gh not on PATH; skipping --github")
		return
	}
	issues, err := ghRows(root, "issue")
	if err == nil {
		var prs []ghRow
		prs, err = ghRows(root, "pr")
		if err == nil {
			s.setKind("issue#", "issue")
			s.setKind("pr#", "pr")
			liveI, liveP := map[string]bool{}, map[string]bool{}
			for _, r := range issues {
				k := fmt.Sprintf("issue#%d", r.Number)
				s.setRef(k, "["+r.Title+"]("+r.URL+")", "issue", "", nil)
				liveI[k] = true
			}
			for _, r := range prs {
				k := fmt.Sprintf("pr#%d", r.Number)
				s.setRef(k, "["+r.Title+"]("+r.URL+")", "pr", "", nil)
				liveP[k] = true
			}
			s.prune("issue#", nil, liveI)
			s.prune("pr#", nil, liveP)
			return
		}
	}
	fmt.Fprintln(s.warn, "refs: "+err.Error()+"; skipping --github")
}

var (
	cactusKeyRE = regexp.MustCompile(`^(q[0-9]+)-`)
	qKeyRE      = regexp.MustCompile(`^q[0-9]+$`)
	wsRunRE     = regexp.MustCompile(`[\r\n\t]+`)
)

func cactusGet(root string, keys []string) ([]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cactus", append(append([]string{"get"}, keys...), "--json")...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var rows []any
	if err := json.Unmarshal(out, &rows); err != nil {
		var one any
		if json.Unmarshal(out, &one) != nil {
			return nil, err
		}
		rows = []any{one}
	}
	return rows, nil
}

// cactusAnswer renders a qualifying row as "question → answer", or "" when it does not qualify.
func cactusAnswer(row map[string]any) string {
	ans, ok := row["answer"].(map[string]any)
	if !ok || truthy(ans["skipped"]) || row["pr"] != nil || row["thread"] == "denied" {
		return ""
	}
	var sel []string
	if arr, ok := ans["selected"].([]any); ok {
		for _, v := range arr {
			sel = append(sel, pystr(v))
		}
	}
	sels := strings.Join(sel, ", ")
	text := ""
	if t, ok := ans["text"].(string); ok {
		text = t
	}
	a := sels
	switch {
	case sels == "":
		a = text
	case text != "":
		a = sels + " — " + text
	}
	if a == "" {
		return ""
	}
	q := ""
	if t, ok := row["text"].(string); ok {
		q = t
	}
	return wsRunRE.ReplaceAllString(q+" → "+a, " ")
}

// cactus mirrors answered records under .ai/cactus. A failed lookup skips the source, removing nothing.
func (s *syncer) cactus(root string) {
	if _, err := exec.LookPath("cactus"); err != nil {
		fmt.Fprintln(s.warn, "refs: cactus not on PATH; skipping --cactus")
		return
	}
	recs, _ := filepath.Glob(filepath.Join(root, ".ai", "cactus", "q[0-9]*-*.md"))
	sort.Strings(recs)
	var keys, rels []string
	for _, f := range recs {
		if !isFile(f) {
			continue
		}
		m := cactusKeyRE.FindStringSubmatch(filepath.Base(f))
		if m == nil {
			continue
		}
		rel, _ := filepath.Rel(root, f)
		keys, rels = append(keys, m[1]), append(rels, filepath.ToSlash(rel))
	}
	var rows []any
	if len(keys) > 0 {
		var err error
		if rows, err = cactusGet(root, keys); err != nil { // one unknown key fails the batch
			ok := false
			for _, k := range keys {
				if r, err := cactusGet(root, []string{k}); err == nil {
					rows, ok = append(rows, r...), true
				}
			}
			if !ok {
				fmt.Fprintln(s.warn, "refs: cactus get failed; skipping --cactus")
				return
			}
		}
	}
	qual := map[string]string{}
	for _, r := range rows {
		if m, ok := r.(map[string]any); ok {
			if k, ok := m["key"].(string); ok {
				if v := cactusAnswer(m); v != "" {
					qual[k] = v
				}
			}
		}
	}
	s.setKind("q", "answer")
	live := map[string]bool{}
	for i, k := range keys {
		if v, ok := qual[k]; ok {
			s.setRef(k, v, "answer", rels[i]+":1", nil)
			live[k] = true
		}
	}
	s.prune("q", qKeyRE, live)
}

func (s *syncer) summary(label string) string {
	return fmt.Sprintf("%s: %d set, %d removed, %d unchanged", label, s.set, s.rm, s.same)
}

// finish saves once when anything changed; otherwise it bumps refs.json's mtime so auto-sync stops re-running.
func (s *syncer) finish(root string) error {
	if s.dry {
		return nil
	}
	if s.dirty {
		return saveRefs(root, s.store)
	}
	if _, err := os.Stat(refsPath(root)); err == nil {
		t := time.Now()
		return os.Chtimes(refsPath(root), t, t)
	}
	return nil
}

// cmdSync is `refs sync [--github] [--cactus] [--dry-run]`.
func (a *app) cmdSync(args []string) int {
	var gh, cactus, dry bool
	for _, arg := range args {
		switch arg {
		case "--github":
			gh = true
		case "--cactus":
			cactus = true
		case "--dry-run":
			dry = true
		default:
			a.die("usage: refs sync [--github] [--cactus] [--dry-run]")
		}
	}
	root := projectRoot("")
	s := newSyncer(loadRefs(root), dry, a.stdout, a.stderr)
	if err := s.canon(root); err != nil {
		fmt.Fprintln(a.stderr, "refs: sync: "+err.Error())
	}
	if gh {
		s.github(root)
	}
	if cactus {
		s.cactus(root)
	}
	if err := s.finish(root); err != nil {
		a.die("cannot write " + refsPath(root) + ": " + err.Error())
	}
	label := "sync"
	if dry {
		label = "sync (dry-run)"
	}
	a.out("%s\n", s.summary(label))
	return 0
}

// canonStale reports whether any existing canon file is newer than refs.json (or refs.json is missing).
func canonStale(root string) bool {
	var base time.Time
	haveStore := false
	if st, err := os.Stat(refsPath(root)); err == nil {
		base, haveStore = st.ModTime(), true
	}
	for _, c := range canonSpecs {
		for _, name := range []string{c.name, c.legacyName} {
			if name == "" {
				continue
			}
			for _, r := range []string{"docs/canon/" + name + ".md", "docs/" + name + ".md"} {
				if st, err := os.Stat(filepath.Join(root, r)); err == nil && st.Mode().IsRegular() {
					if !haveStore || st.ModTime().After(base) {
						return true
					}
				}
			}
		}
	}
	return false
}

// autoSync runs the canon part of sync when a canon file is newer than refs.json. Never panics.
func autoSync(root string, warn io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	if !canonStale(root) {
		return nil
	}
	s := newSyncer(loadRefs(root), false, io.Discard, warn)
	if err := s.canon(root); err != nil {
		return err
	}
	return s.finish(root)
}
