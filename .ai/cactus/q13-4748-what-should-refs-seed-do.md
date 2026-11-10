# q13 — What should 'refs seed' do (it replaces the reserved 'install')?

status: answered
act: ask
kind: choice
thread: seed
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-07T06:47:53.625088+00:00

## Context

**Default if unanswered: prompt.**

- install is reserved today but does nothing
- the empty-project hook already tells the agent to populate refs
- either way: rename in reserved list, help, README, CLAUDE.md, tests

— slugs —
touched: CLAUDE.md, README.md, bin/refs

## Files

- cmd.go
- hook.go
- main.go

## Options

- prompt — print the populate-your-refs instructions (the empty-project hook text)
+ agent can call it any time, not only at session start
+ tiny change  (★◐)
- scaffold — create an empty .ai/refs.json plus starter kinds (R requirement, pr# pr, adr- decision)
+ new project is ready to use
- picks kinds for you
- both — create the file with starter kinds, then print the instructions
- reserve — just rename the reserved word to seed, build nothing yet

## Recommendation

prompt — med
reuses existing text, gives agents a way to seed on demand

## Answer

prompt
answered at: 2026-11-07T06:48:08.786428+00:00
