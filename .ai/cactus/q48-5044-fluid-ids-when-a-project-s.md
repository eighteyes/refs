# q48 — Fluid IDs: when a project's IDs are not fixed, what may change?

status: answered
act: ask
kind: multi
thread: ids
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-09T21:40:35.893235+00:00

## Context

**Reading "numbers" as the canon IDs (TRU-03).** Tap e if you meant something else.

- today: IDs never change, never reused, retired blocks stay
- fluid cost: refs, cactus --refs and commit messages citing an old ID go stale
- sync.sh already rewrites refs from the files

— slugs —
touched: main.go, hook.go, cmd.go, skills/tighten/SKILL.md

## Files

- /Users/god/projects/ai-tooling/utils/refs/skills/harvest/formats.md

## Options

- renumber — tighten renumbers each file 01..N, closing gaps  (★◐)
- reuse — a retired ID can be given to a new entry
- delete — retired entries are deleted, not kept with Retired:  (★◐)
- reorder — entries move in the file by importance

## Recommendation

renumber, delete — med

## Answer

renumber, delete, reuse, reorder
answered at: 2026-11-09T21:51:12.334291+00:00
