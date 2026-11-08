# q18 — Where does the new .claude-plugin/marketplace.json get committed?

status: answered
act: ask
kind: choice
thread: install
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T04:47:03.956774+00:00

## Context

**Default if unanswered: feat-harvest.**

- plugin is installed (user scope) and the hook works from the cache
- the cache is a copy at version 0.2.0, taken from the feat/harvest checkout, harvest skill included
- after code changes: bump version and run claude plugin update refs@refs

— slugs —
touched: .claude-plugin/marketplace.json, CLAUDE.md, README.md, bin/refs

## Files

- /Users/god/projects/ai-tooling/utils/refs/.claude-plugin/marketplace.json

## Options

- feat-harvest — commit on the current branch, next to the harvest work
+ no branch switch; lands with the skill it installs  (★◐)
- main — tell the harvest session, it commits on main when it merges
- needs the other session to carry it
- leave — keep it uncommitted for now

## Recommendation

feat-harvest — med
the install already includes the harvest work

## Answer

feat-harvest
answered at: 2026-11-08T06:34:26.089922+00:00
