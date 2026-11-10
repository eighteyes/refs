# Requirements
What the product must do, each with acceptance criteria. Cite by ID.

## REQ-01 Set or overwrite a ref with `refs KEY VALUE`
Why: agents record a requirement, PR or decision in one command.
Acceptance:
- `refs KEY VALUE...` stores the value verbatim, args joined by spaces; check with `refs KEY`.
- Setting the same KEY again overwrites the value; check with `refs KEY`.
Applies to: cmd.go
Source: README.md:11

## REQ-02 An unknown key prefix saves as kind `unknown` and warns
Why: a mistyped or new prefix must not pass silently.
Acceptance:
- The ref saves with kind `unknown`; check with `refs KEY`.
- stderr names the prefix and `refs kind PREFIX NAME`; check with TestUnknownKindNagAndRelabel.
Applies to: cmd.go
Source: README.md:29

## REQ-03 The SessionStart hook seeds an empty project, else reminds
Why: a new project gets its pillars written instead of running without them.
Acceptance:
- With no `.ai/refs.json`, `refs hook` says to populate the refs; check with TestHookEmpty.
- With refs present, it prints a one-line reminder and the ref count; check with TestHookCounts.
Applies to: hook.go
Source: README.md:60

## REQ-04 Setting a value without `--home` keeps the existing home
Why: a re-run of `refs sync` or a manual set must not drop the source-of-truth link.
Acceptance:
- Set a ref with `--home`, set it again without; `refs KEY` still shows the home. Check with TestHome.
Applies to: cmd.go
Source: main.go:45

## REQ-05 `refs seed` prints how to populate an empty project
Why: an agent can ask for the populate instructions at any time, not only at session start.
Acceptance:
- `refs seed` prints the empty-project hook text and exits 0, and never touches refs.json; check with TestSeed.
- `refs seed ARG` is a usage error, exit 1; check with TestSeedRejectsArgs.
Applies to: main.go
Source: main.go:32

## REQ-06 `refs sync` mirrors live canon, open GitHub work and qualifying cactus answers into refs
Why: sessions recall canon by ID and read its source line before trusting it.
Acceptance:
- `refs sync --github --cactus --dry-run` prints one change per live entry, open issue, open PR and qualifying cactus answer; check by running it.
- Each canon ref's home is its heading line; each answer's home is its record, line 1.
Applies to: skills/harvest/**
Source: skills/harvest/SKILL.md:54
