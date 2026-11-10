// hook.go — the `refs hook` SessionStart hook.
//
// Responsibilities:
//   - read the hook payload JSON from stdin and take `cwd`
//   - print hookSpecificOutput.additionalContext: seed text for an empty project, else a standing line + ref count + canon blurb (canon files present, counts by kind) + the ids mode sentence
//   - auto-sync the canon first when a canon file is newer than refs.json
//   - never fail: garbage input or errors still yield valid JSON and exit 0
package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const hookText = "refs: this project keeps session pillars (requirements, PRs, decisions) " +
	"— run `refs` to list them, `refs KEY VALUE` to add one. " +
	"When you post a cactus question, cite the refs it rests on with `--refs KEY[,KEY]` " +
	"instead of restating them in --context. See `refs --help`."

const hookEmpty = "refs: this project has no refs yet. Your first step is to populate them: " +
	"run the `harvest` skill (`/refs:harvest`) to turn the repo into canon " +
	"(requirements, truths, nevers, decisions) and refs, or add pillars by hand with " +
	"`refs KEY VALUE --kind NAME [--home PATH]`, defining key prefixes with " +
	"`refs kind PREFIX NAME`. See `refs --help`."

// hookContext picks the additionalContext text for the project at cwd.
func hookContext(cwd string) string {
	root := projectRoot(cwd)
	autoSync(root, io.Discard) // best effort; errors must not break the hook
	if _, err := os.Stat(refsPath(root)); err != nil {
		return hookEmpty
	}
	store := loadRefs(root)
	all := asObj(store.Get("refs"))
	n := all.Len()
	s := hookText + " This project has " + itoa(n) + " ref"
	if n != 1 {
		s += "s"
	}
	s += "." + canonBlurb(root, all)
	if strings.Contains(s, " Canon in docs/canon/") {
		if idsMode(store) == "fixed" {
			s += " Canon IDs are fixed."
		} else {
			s += " Canon IDs are fluid and may be renumbered; cite an ID with its title."
		}
	}
	return s
}

// canonFiles are the canon docs under docs/canon/, in blurb order.
var canonFiles = []struct{ file, kind string }{
	{"REQUIREMENTS.md", "requirement"},
	{"TRUTHS.md", "truth"},
	{"NEVERS.md", "never"},
	{"DECISIONS.md", "decision"},
	{"INQUIRY.md", "inquiry"},
	{"OPEN_QUESTIONS.md", "open"},
}

// canonBlurb names the canon files present (stat only) with their ref counts by kind.
func canonBlurb(root string, refs *omap) string {
	counts := map[string]int{}
	first := map[string]string{} // kind -> first key of that kind
	for _, k := range refs.keys {
		kind := pystr(asObj(refs.Get(k)).Get("kind"))
		counts[kind]++
		if _, ok := first[kind]; !ok {
			first[kind] = k
		}
	}
	var parts []string
	exKey, exKind := "", ""
	for _, c := range canonFiles {
		if _, err := os.Stat(filepath.Join(root, "docs", "canon", c.file)); err == nil {
			parts = append(parts, c.file+" ("+itoa(counts[c.kind])+" "+c.kind+")")
			if exKey == "" && first[c.kind] != "" {
				exKey, exKind = first[c.kind], c.kind
			}
		}
	}
	if exKey == "" {
		exKey, exKind = "KEY", "NAME"
	}
	if len(parts) == 0 {
		return ""
	}
	return " Canon in docs/canon/: " + strings.Join(parts, ", ") +
		". Read one with `refs " + exKey + "` (its home is the source line); list a kind with `refs --kind " + exKind + "`."
}

func (a *app) cmdHook() (code int) {
	text := hookText
	func() {
		defer func() { recover() }()
		cwd := ""
		if data, err := io.ReadAll(a.stdin); err == nil {
			var payload struct {
				Cwd any `json:"cwd"`
			}
			if json.Unmarshal(data, &payload) == nil {
				cwd, _ = payload.Cwd.(string)
			}
		}
		text = hookContext(cwd)
	}()
	io.WriteString(a.stdout, `{"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": `+
		quote(text, false)+"}}\n")
	return 0
}
