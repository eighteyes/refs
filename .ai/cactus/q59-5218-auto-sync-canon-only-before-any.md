# q59 — auto-sync: canon only, before any refs command and in the hook, when a canon .md is newer than refs.json

status: answered
act: steer
kind: choice
thread: readme
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-10T17:02:59.189212+00:00

## Context

- sync.sh stays as a one-line wrapper around `refs sync` so old docs keep working
- skills switch to `refs sync`

— slugs —
touched: .ai/readme/README.md, main.go, hook.go, store.go

## Options

- canon-on-stale — stat the 6 files; newer than refs.json → re-mirror canon first
+ local, a few stat calls
+ --github and --cactus stay explicit (`refs sync --github --cactus`)  (doing)
- hook-only — auto-sync only at SessionStart
- everything — auto-sync also pulls GitHub and cactus
- network on every call

## Answer

canon-on-stale
answered at: 2026-11-10T17:05:25.104724+00:00
