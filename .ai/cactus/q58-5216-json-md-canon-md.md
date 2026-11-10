# q58 — JSON ↔ md: canon .md is the source and sync.sh copies it into refs.json. Change what?

status: answered
act: ask
kind: multi
thread: readme
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-10T16:57:40.477023+00:00

## Context

**Today:** you or a skill edit docs/canon/*.md → someone runs skills/harvest/bin/sync.sh → refs.json mirrors each ## ID heading (value = title, home = file:line). cactus answers and GitHub issues/PRs come in the same way.

- nothing runs sync for you, so refs.json goes stale after a hand edit
- the script hides under skills/harvest/bin
- `refs KEY VALUE` writes refs.json only; it never touches the .md

— slugs —
touched: .ai/readme/README.md, main.go, hook.go, store.go

## Files

- skills/harvest/bin/sync.sh
- skills/harvest/SKILL.md

## Options

- explain — keep the flow; README gets a short "how data moves" section + diagram
- refs-sync — move sync.sh into Go as `refs sync`
+ one tool, one name; no hidden script path
- Go grows a canon parser  (★◐)
- auto-sync — refs sync runs itself (SessionStart hook, or on any refs read when an .md is newer)
+ refs.json never stale
- hook gets slower; --github needs network  (★◐)
- invert — refs.json is the source; the .md files are generated from it
- canon edits go through the CLI, not the editor

## Recommendation

refs-sync, auto-sync — med

## Answer

refs-sync, auto-sync
answered at: 2026-11-10T17:02:31.012653+00:00
