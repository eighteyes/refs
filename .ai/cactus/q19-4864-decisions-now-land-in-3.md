# q19 — Decisions now land in 3 places: canon DECISIONS.md, know graph, cactus records. How should refs tie them?

status: answered
act: ask
kind: multi
thread: integrate
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T06:37:36.637396+00:00

## Context

**Default if unanswered: cactus-promote. Nothing is built until you pick.**

**Today**
- refs Go code knows nothing of cactus or know
- harvest (feat/harvest, other session) uses cactus only to promote Proposed items
- the cactus hook tells agents: know add decision qN after an answer
- this repo has no .ai/know graph

**Note**: harvest and sync.sh belong to the other session; changes there go through it

— slugs —
touched: .claude-plugin/marketplace.json, CLAUDE.md, README.md, bin/refs

## Files

- /Users/god/projects/ai-tooling/utils/refs/skills/harvest/SKILL.md

## Options

- cactus-records — sync.sh mirrors answered .ai/cactus/qN-*.md as refs qN, home at the record
+ every human answer becomes recallable by key
- many rows are process noise (merge yes/no)
- cactus-promote — harvest reads cactus records as a 5th source of DEC candidates
+ answers flow into canon; refs follows from the canon
- relies on the harvest judge to filter noise  (★○)
- know-mirror — sync.sh mirrors know decision/requirement nodes into refs
+ refs becomes one index over both stores
- know nodes have no file:line home; home is the graph file
- canon-only — canon files are the one source; know and cactus stay separate tools
+ no coupling
- the cactus hook keeps writing decisions into know

## Recommendation

cactus-promote — low
keeps canon as the one source; refs already mirrors canon

## Answer

cactus-promote, cactus-records — there should be a --pr flag to skep on the process noise
answered at: 2026-11-08T06:39:01.783412+00:00
