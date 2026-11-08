// hook.go — the `refs hook` SessionStart hook.
//
// Responsibilities:
//   - read the hook payload JSON from stdin and take `cwd`
//   - print hookSpecificOutput.additionalContext: seed text for an empty project, else a standing line + ref count
//   - never fail: garbage input or errors still yield valid JSON and exit 0
package main

import (
	"encoding/json"
	"io"
	"os"
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
	if _, err := os.Stat(refsPath(root)); err != nil {
		return hookEmpty
	}
	n := asObj(loadRefs(root).Get("refs")).Len()
	s := hookText + " This project has " + itoa(n) + " ref"
	if n != 1 {
		s += "s"
	}
	return s + "."
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
