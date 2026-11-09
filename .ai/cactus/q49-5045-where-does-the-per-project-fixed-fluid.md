# q49 — Where does the per-project fixed/fluid flag live?

status: answered
act: ask
kind: choice
thread: ids
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-09T21:40:35.943851+00:00

## Context

**Default: refs-json.**

- the hook could say "IDs are fluid; cite by title" so agents cite carefully

## Options

- refs-json — "ids": "fixed"|"fluid" in .ai/refs.json, set by `refs ids fixed`
+ the hook and pane can show it
- a new subcommand, help+README  (★◐)
- canon-readme — a line in docs/canon/README.md the skills read
+ no Go change
- per-file — a header line in each canon file
+ truths can be fixed while open questions stay fluid

## Recommendation

refs-json — med

## Answer

refs-json
answered at: 2026-11-09T21:51:41.131063+00:00
