# q60 — How should Claude learn ref content, not just that refs exist?

status: answered
act: ask
kind: multi
thread: context
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-10T17:03:07.767019+00:00

## Context

**Today:** the hook says refs exist, the count, the canon files and how to look one up. No content is injected (INQ-01 bets on read-on-demand). Claude sees content only when it runs `refs` / `refs KEY`, or when a cactus row cites --refs.

- titles are short; full entries stay in the .md (the home)

— slugs —
touched: .ai/readme/README.md, main.go, hook.go, store.go

## Files

- hook.go

## Options

- titles — hook lists every ref key + title, one line each
+ always in context
- 68 refs ≈ 1.5k tokens every session, grows
- guards — hook injects only truths + nevers (titles)
+ the rules agents break by not knowing
- ~20 lines here  (★◐)
- on-edit — before Edit/Write, inject the truths and nevers whose Applies to matches the file
+ right rule at the right moment, tiny
- needs Applies-to globs filled in; new PreToolUse hook  (★◐)
- pull-only — keep today: count + how to look up; the agent runs `refs`
- agents often skip it

## Recommendation

guards, on-edit — med

## Answer

guards, on-edit
answered at: 2026-11-10T17:05:31.737639+00:00
