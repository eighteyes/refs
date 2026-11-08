# q36 — Commit the new canon (docs/canon/*.md) and refs.json on feat/harvest?

status: answered
act: ask
kind: choice
thread: harvest
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:39:54.538071+00:00

## Context

**Default if unanswered: leave.**

- 6 REQ, 14 TRU, 5 NEV, 12 DEC written; 51 refs after sync (14 are cactus answers)
- harvest says committing the four files is yours

## Files

- /Users/god/projects/ai-tooling/utils/refs/docs/canon/REQUIREMENTS.md
- /Users/god/projects/ai-tooling/utils/refs/docs/canon/TRUTHS.md
- /Users/god/projects/ai-tooling/utils/refs/docs/canon/NEVERS.md
- /Users/god/projects/ai-tooling/utils/refs/docs/canon/DECISIONS.md

## Options

- canon+refs — commit docs/canon/ and .ai/refs.json with the new cactus records
+ refs.json travels with the code, like the records  (★◐)
- canon-only — commit docs/canon/ and records; gitignore .ai/refs.json
+ refs.json is rebuilt by sync.sh anyway
- a fresh clone has no refs until someone syncs
- leave — commit nothing yet

## Recommendation

canon+refs — med
sessions get refs without a sync step

## Answer

canon+refs
answered at: 2026-11-08T07:41:22.430714+00:00
