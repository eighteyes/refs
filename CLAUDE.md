# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`refs`: a per-project key-value store for an agent session's pillars (requirements, PRs, decisions). It ships as a Claude Code plugin and as a Go program: stdlib only, no dependencies. Keep it that way.

## Layout

- `main.go`: entry point, `run`, `die`, the `help` text.
- `cmd.go`: the refs command (list, show, set, rm, kind, kinds, home, ids).
- `store.go`: ordered JSON (`omap`), load/save, project root, home normalization.
- `sync.go`: `refs sync` and the canon auto-sync: heading parser, in-process syncer, `gh` and `cactus` shell-outs.
- `guard.go`: `refs guard`, the PreToolUse hook; glob matching of `applies`.
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

- **Storage**: `<root>/.ai/refs.json`, shape `{"kinds": {prefix: name}, "refs": {key: entry}}`, plus an optional top-level `"ids": "fixed"|"fluid"` written only by `refs ids VALUE` (absent reads as fluid; unknown top-level fields survive). Root is the git toplevel of cwd, else cwd. `loadRefs` repairs a missing or malformed file to empty instead of failing. `saveRefs` writes a tmp file and renames it. Key order of `kinds` and `refs` is insertion order, so `omap` decodes and encodes in order; unknown fields on entries survive. Output is ASCII-escaped, so files stay byte-compatible with 0.1.x.
- **Kinds come from key prefixes**: `prefixOf` takes the leading non-digit characters (`R20` → `R`, `pr#131` → `pr#`). An unknown prefix saves as `unknown` and warns on stderr. `refs kind PREFIX NAME` relabels existing `unknown` refs that share the prefix.
- **Dispatch**: `run` handles `--help`, `hook`, `sync` and `seed` (`seed` prints `hookEmpty` as plain text; extra args fail with `usage: refs seed`). Everything else goes to `cmdRefs`, which parses flags (`--kind`, `--home`, `--clear`) by hand and then matches the first positional against subcommands. If none match, it is a KEY: show it with no value args, set it with value args. Any new subcommand name must go into `reserved` so it can't be used as a key.
- **Sync**: `run` hands `sync` to `cmdSync` (no auto-sync). It loads the store once, applies canon (and `--github`, `--cactus`) changes through a `syncer`, and saves once; `--dry-run` saves nothing. A ref whose kind, value and home already match keeps its `set_at`. `cmdRefs` and `hookContext` call `autoSync` first: if a canon file (canon path or legacy `docs/` path) is newer than `.ai/refs.json`, or `refs.json` is missing, it runs the canon part and saves, or bumps the mtime when nothing changed. Failures warn on stderr in commands and are swallowed in the hook. `seed` and `--help` never sync.
- **Guards**: sync stores a block's `Applies to:` line as an `applies` list (an optional entry field; a change bumps `set_at`; unknown entry fields survive a re-sync). `hookContext` ends with a Guards list (truths then nevers, 40 max). `refs guard` is the PreToolUse hook: reads `cwd` and `tool_input.file_path`/`notebook_path`, skips paths outside the root and under `.ai/`, auto-syncs, and prints `additionalContext` for matching truths/nevers. It never emits `permissionDecision`; on no match or any error it prints nothing and exits 0.
- **Entry fields**: `kind`, `value` (verbatim, args joined by spaces), `set_at`. Optional `title`/`url` when the value matches `[title](url)`. Optional `home`. Overwriting a ref without `--home` keeps its old home.
- **Home**: `normalizeHome` resolves the path relative to cwd, stores it relative to root, keeps a `:LINE` suffix, refuses paths outside the root, and only warns when the file is missing.
- **Hook contract**: `cmdHook` reads the SessionStart JSON from stdin (it uses `cwd`), prints `hookSpecificOutput.additionalContext`, and must always exit 0 and never raise. It prints `hookEmpty` when no `refs.json` exists, otherwise `hookText` plus the ref count, plus a canon blurb listing which of the six `docs/canon/*.md` files exist (stat only) with ref counts by kind (none present, no blurb). When the blurb prints it is followed by one sentence on the `ids` mode (fluid: IDs may be renumbered, cite with title; fixed: IDs are fixed). Then comes the Guards list. `refs guard` follows the same rule: always exit 0, never raise.

## Keep in sync

Usage is written down twice: the `help` string in `main.go` and the usage block in `README.md`. When you change behaviour, update both. `help` is written for a model reader.
