# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`refs`: a per-project key-value store for an agent session's pillars (requirements, PRs, decisions). It ships as a Claude Code plugin and as a Go program: stdlib only, no dependencies. Keep it that way.

## Layout

- `main.go`: entry point, `run`, `die`, the `help` text.
- `cmd.go`: the refs command (list, show, set, rm, kind, kinds, home).
- `store.go`: ordered JSON (`omap`), load/save, project root, home normalization.
- `hook.go`: the SessionStart hook (`refs hook`).
- `refs_test.go`: tests; each runs in a temp dir with its own `git init`.
- `bin/refs`: POSIX sh launcher. Builds `bin/.build/refs` on first run and when a `.go` file or `go.mod` is newer, then execs it. Build output goes to stderr only.
- `hooks/hooks.json`: registers `${CLAUDE_PLUGIN_ROOT}/bin/refs hook` on `startup|resume|clear|compact`.
- `.claude-plugin/plugin.json`: plugin manifest. Bump `version` here on release.

## Commands

```
go build ./...                  # compile
go vet ./...
gofmt -l .                      # must print nothing
go test ./...                   # all tests
go test -run TestHome ./...     # one test
```

Tests call `run` in-process inside a temp git repo, so the repo's own `.ai/refs.json` stays untouched. To try the launcher by hand, use a scratch project:

```
d=$(mktemp -d) && cd "$d" && git init -q
/path/to/refs/bin/refs R1 "some requirement" --kind req
/path/to/refs/bin/refs
echo '{"cwd":"'"$d"'"}' | /path/to/refs/bin/refs hook
```

## How it works

- **Storage**: `<root>/.ai/refs.json`, shape `{"kinds": {prefix: name}, "refs": {key: entry}}`. Root is the git toplevel of cwd, else cwd. `loadRefs` repairs a missing or malformed file to empty instead of failing. `saveRefs` writes a tmp file and renames it. Key order of `kinds` and `refs` is insertion order, so `omap` decodes and encodes in order; unknown fields on entries survive. Output is ASCII-escaped, so files stay byte-compatible with 0.1.x.
- **Kinds come from key prefixes**: `prefixOf` takes the leading non-digit characters (`R20` → `R`, `pr#131` → `pr#`). An unknown prefix saves as `unknown` and warns on stderr. `refs kind PREFIX NAME` relabels existing `unknown` refs that share the prefix.
- **Dispatch**: `run` handles `--help`, `hook` and `seed` (`seed` prints `hookEmpty` as plain text; extra args fail with `usage: refs seed`). Everything else goes to `cmdRefs`, which parses flags (`--kind`, `--home`, `--clear`) by hand and then matches the first positional against subcommands. If none match, it is a KEY: show it with no value args, set it with value args. Any new subcommand name must go into `reserved` so it can't be used as a key.
- **Entry fields**: `kind`, `value` (verbatim, args joined by spaces), `set_at`. Optional `title`/`url` when the value matches `[title](url)`. Optional `home`. Overwriting a ref without `--home` keeps its old home.
- **Home**: `normalizeHome` resolves the path relative to cwd, stores it relative to root, keeps a `:LINE` suffix, refuses paths outside the root, and only warns when the file is missing.
- **Hook contract**: `cmdHook` reads the SessionStart JSON from stdin (it uses `cwd`), prints `hookSpecificOutput.additionalContext`, and must always exit 0 and never raise. It prints `hookEmpty` when no `refs.json` exists, otherwise `hookText` plus the ref count, plus a canon blurb listing which of the six `docs/canon/*.md` files exist (stat only) with ref counts by kind (none present, no blurb).

## Keep in sync

Usage is written down twice: the `help` string in `main.go` and the usage block in `README.md`. When you change behaviour, update both. `help` is written for a model reader.
