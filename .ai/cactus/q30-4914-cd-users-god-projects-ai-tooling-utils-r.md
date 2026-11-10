# q30 — cd .ai/cactus && for f in q22-*.md q23-*.md q25-*.md q26-*.md q27-*.md q28-*.md; do echo "=== $f"; rtk read "$f"; echo; done

status: answered
act: run
kind: confirm
thread: denied
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:35:30.442357+00:00

## Context

Read second batch of qualifying cactus records
(denied by permission mode)

## Options

- approve
- deny

## Answer

approve
answered at: 2026-11-08T07:36:12.075276+00:00

## Result

exit: 0

```

## Recommendation

open-home — med
the home is the source of truth; reading it is the main move

## Answer

open-home
answered at: 2026-11-08T07:26:01.371103+00:00

=== q28-4904-where-does-the-refs-pane-mod.md
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

— exit 0 —
```

log: /var/folders/vy/jl7086td5nv4_6ms16c6g1rr0000gn/T/cactus-q30-2824484252.log
