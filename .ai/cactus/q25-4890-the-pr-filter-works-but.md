# q25 — The --pr filter works, but 5 old merge rows (q11 q12 q14 q15 q16) are untagged and cactus won't tag answered rows. Handle them how?

status: answered
act: ask
kind: choice
thread: integrate
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T06:49:39.031471+00:00

## Context

**Default if unanswered: skip-denied.**

- tested: a row posted with --pr ready, then answered, is excluded; untagged q1 mirrored
- --pr must be set at ask time; cactus edit refuses answered rows
- the rule for agents: put --pr STATE on merge/ship rows

— slugs —
touched: skills/harvest/SKILL.md, skills/harvest/bin/sync.sh, .claude-plugin/marketplace.json, CLAUDE.md

## Files

- skills/harvest/bin/sync.sh

## Options

- skip-denied — also skip thread=denied rows (q14, the raw command); keep the rest
+ one rule, kills the ugliest row
- the 4 merge confirms stay as refs
- accept — mirror them; tag future process rows with --pr at ask time
+ no new rule
- 5 noise refs here today
- cactus-allow — ask cactus-b2 to let --pr edit answered rows, then tag the 5
+ fixes the data, keeps one rule
- a cactus change
- skip-denied+allow — both of the above  (★◐)

## Recommendation

skip-denied+allow — med
denied rows are never decisions; old rows get tagged properly

## Answer

skip-denied+allow
answered at: 2026-11-08T06:50:45.153867+00:00
