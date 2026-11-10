# Truths
What must stay true in this codebase, before and after every change. Cite by ID.

## TRU-01 `refs hook` always exits 0 and prints valid hook JSON
Why: a failing SessionStart hook breaks session start, and its payload can be garbage.
Applies to: hook.go, bin/refs
Check: TestHookGarbageStdin, TestHookIgnoresMalformedStore
Source: hook.go:6, CLAUDE.md:46

## TRU-02 Writes to refs.json are atomic
Why: a crash or a reader mid-write would see a truncated store, which loads as empty.
Applies to: store.go
Check: review (saveRefs writes a tmp file, then renames it)
Source: store.go:265, README.md:22

## TRU-03 Key order in refs.json is insertion order; a set replaces in place
Why: list order and refs.json diffs depend on it.
Applies to: store.go, cmd.go
Check: TestInsertionOrderAndUnknownFields
Source: store.go:43

## TRU-04 A stored home is root-relative and inside the project
Why: a home outside the project points at a file the repo does not version.
Applies to: store.go
Check: TestHomeOutsideRoot
Source: main.go:44, store.go:354

## TRU-05 Reserved words (rm kind kinds home seed hook) are never keys
Why: a key with a subcommand's name could never be shown or set.
Applies to: cmd.go, main.go
Check: TestReservedAndInvalidKeys
Source: main.go:37

## TRU-06 A key never starts with a digit or `-`
Why: a leading `-` parses as a flag, and a leading digit leaves no prefix to be the kind.
Applies to: cmd.go
Check: TestReservedAndInvalidKeys
Source: cmd.go:223, README.md:26

## TRU-07 A rejected set never writes refs.json
Why: a refused key or home must leave the store exactly as it was.
Applies to: cmd.go, store.go
Check: TestReservedAndInvalidKeys
Source: refs_test.go:156

## TRU-08 A missing or malformed refs.json loads as an empty store
Why: a bad hand edit must not crash list, set or the hook.
Applies to: store.go
Check: TestMalformedStoreLoadsEmpty
Source: store.go:244

## TRU-09 refs.json output is ASCII-escaped
Why: files stay byte-compatible with the 0.1.x Python store.
Applies to: store.go
Check: TestNoHTMLEscapeAndASCII
Source: store.go:137, CLAUDE.md:41

## TRU-10 The project root is the git toplevel of cwd, else cwd
Why: running from a subdirectory would otherwise write a second `.ai/refs.json`.
Applies to: store.go
Check: TestNonGitDirUsesCwd
Source: README.md:21

## TRU-11 The bin/refs launcher writes build output to stderr only
Why: stdout carries the hook JSON; build chatter there corrupts it.
Applies to: bin/refs
Check: review (with no binary, `echo '{}' | bin/refs hook` prints one JSON line)
Source: bin/refs:7

## TRU-12 Canon IDs never change and are never reused
Why: refs and commit messages that cite an ID would point at a different rule.
Applies to: docs/canon/*.md
Check: review
Source: skills/harvest/formats.md:47

## TRU-13 `refs sync` is idempotent and drops refs whose source is gone or retired
Why: harvest is re-run often; drift or stale refs mislead every session.
Applies to: sync.go
Check: go test -run Sync ./...
Source: skills/harvest/SKILL.md:54, sync.go:1

## TRU-14 `refs sync --cactus` never mirrors a skipped answer, a pr-tagged row or a denied-thread row
Why: those rows are process or blocked commands, not decisions.
Applies to: sync.go
Check: go test -run Cactus ./...
Source: sync.go:1
