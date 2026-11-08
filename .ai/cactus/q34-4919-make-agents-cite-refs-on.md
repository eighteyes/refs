# q34 — Make agents cite refs on cactus asks: where does the rule live?

status: answered
act: ask
kind: multi
thread: cactus-refs
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:37:24.028599+00:00

## Context

**Default if unanswered: refs-hook + cactus-skill + memory.**

- I start citing --refs on my own asks now either way
- cactus --agent-help already says "Cite refs instead of restating facts in --context"

— slugs —
touched: refs_test.go, skills/harvest/bin/sync.sh, bin/refs, .ai/cactus/q28-4904-where-does-the-refs-pane-mod.md

## Options

- refs-hook — refs SessionStart text adds "cite refs in cactus ask --no-wait --refs KEY instead of restating them"
+ every session in a project with refs sees it; I own this code  (★◐)
- cactus-skill — cactus skill + hook say "pass --refs for any ref the question rests on" (brief cactus-b2)
+ reaches agents even where refs hook is off
- memory — save it as my feedback memory for this project only

## Recommendation

refs-hook — med
refs owns the nudge; it only fires where refs exist

## Answer

cactus-skill, refs-hook
answered at: 2026-11-08T07:39:07.577163+00:00
