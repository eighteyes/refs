# q28 — Where does the refs-pane mod live?

status: answered
act: ask
kind: choice
thread: refs-pane
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:20:37.626844+00:00

## Context

**Default if unanswered: both.**

- cactus-pane lives at ~/projects/cactus/mods/cactus-pane (README, register.tsx, tests, types)
- /refs-pane is the slash command that opens it

## Options

- repo-mods — refs repo mods/refs-pane/, like cactus/mods/cactus-pane
+ versioned with refs, shareable
- the repo is on feat/harvest; the commit lands there
- dev-mods — this session's ~/.claude/dev-mods folder only
+ hot-reloads now, nothing committed
- gone with the session unless copied
- both — build in dev-mods to hot-reload, copy into mods/refs-pane when done  (★●)

## Recommendation

both — high
fast iteration, ends versioned with refs

## Answer

both
answered at: 2026-11-08T07:26:07.043393+00:00
