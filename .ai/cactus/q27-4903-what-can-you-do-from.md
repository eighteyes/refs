# q27 — What can you do from inside the pane?

status: answered
act: ask
kind: multi
thread: refs-pane
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:20:37.593190+00:00

## Context

**Default if unanswered: open-home + insert.**

- the pane re-reads refs.json when it changes (a refs call or sync.sh run)

## Options

- read-only — look, nothing else
+ smallest, no writes from the UI
- open-home — a key opens the ref's home file (file:// or editor)  (★◐)
- insert — a key puts the ref key (e.g. DEC-07) into the prompt
- edit — rm a ref, or set a value, from the pane
- writes refs.json from a UI; needs confirm

## Recommendation

open-home — med
the home is the source of truth; reading it is the main move

## Answer

open-home
answered at: 2026-11-08T07:26:01.371103+00:00
