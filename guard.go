// guard.go — `refs guard`, a PreToolUse hook that surfaces the truths and nevers covering a file about to be edited.
//
// Responsibilities:
//   - read the PreToolUse payload (cwd, tool_input.file_path or notebook_path) from stdin
//   - resolve the file against the project root; skip paths outside it and under .ai/
//   - match each truth/never's `applies` globs against the root-relative path
//   - print hookSpecificOutput.additionalContext listing the matches; never a permissionDecision
//   - print nothing and exit 0 on no match or any error; never panic
package main

import (
	"encoding/json"
	"io"
	"path"
	"path/filepath"
	"strings"
)

// globMatch reports whether rel matches pattern. `all` matches everything; a pattern
// without "/" also matches the basename; `**` spans directories.
func globMatch(pattern, rel string) bool {
	pattern = strings.TrimPrefix(pattern, "./")
	if pattern == "all" {
		return true
	}
	if i := strings.Index(pattern, "**"); i >= 0 {
		head, tail := pattern[:i], strings.TrimPrefix(pattern[i+2:], "/")
		rest := rel
		if head != "" {
			if !strings.HasPrefix(rel, head) {
				return false
			}
			rest = rel[len(head):]
		}
		if tail == "" {
			return true
		}
		for start := 0; start <= len(rest); {
			if globMatch(tail, rest[start:]) {
				return true
			}
			n := strings.Index(rest[start:], "/")
			if n < 0 {
				break
			}
			start += n + 1
		}
		return false
	}
	if ok, _ := path.Match(pattern, rel); ok {
		return true
	}
	if !strings.Contains(pattern, "/") {
		if ok, _ := path.Match(pattern, path.Base(rel)); ok {
			return true
		}
	}
	return false
}

// guardsFor lists "KEY title @ home" for each truth/never whose applies match rel.
func guardsFor(refs *omap, rel string) []string {
	var out []string
	for _, k := range refs.keys {
		r := asObj(refs.Get(k))
		if kind := pystr(r.Get("kind")); kind != "truth" && kind != "never" {
			continue
		}
		for _, g := range appliesOf(r) {
			if globMatch(g, rel) {
				s := k + " " + pystr(r.Get("value"))
				if h := r.Get("home"); truthy(h) {
					s += " @ " + pystr(h)
				}
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func (a *app) cmdGuard() (code int) {
	defer func() {
		if recover() != nil {
			code = 0
		}
	}()
	data, err := io.ReadAll(a.stdin)
	if err != nil {
		return 0
	}
	var payload struct {
		Cwd       string `json:"cwd"`
		ToolInput struct {
			FilePath     string `json:"file_path"`
			NotebookPath string `json:"notebook_path"`
		} `json:"tool_input"`
	}
	if json.Unmarshal(data, &payload) != nil {
		return 0
	}
	file := payload.ToolInput.FilePath
	if file == "" {
		file = payload.ToolInput.NotebookPath
	}
	if file == "" {
		return 0
	}
	root := projectRoot(payload.Cwd)
	if !filepath.IsAbs(file) {
		base := payload.Cwd
		if base == "" {
			base = getcwd()
		}
		file = filepath.Join(base, file)
	}
	rel, err := filepath.Rel(realpath(root), realpath(file))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return 0
	}
	rel = filepath.ToSlash(rel)
	if rel == ".ai" || strings.HasPrefix(rel, ".ai/") {
		return 0
	}
	autoSync(root, io.Discard)
	hits := guardsFor(asObj(loadRefs(root).Get("refs")), rel)
	if len(hits) == 0 {
		return 0
	}
	text := "refs guards for " + rel + ": " + strings.Join(hits, "; ")
	io.WriteString(a.stdout, `{"hookSpecificOutput": {"hookEventName": "PreToolUse", "additionalContext": `+quote(text, false)+"}}\n")
	return 0
}
