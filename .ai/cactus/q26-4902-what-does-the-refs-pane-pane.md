# q26 — What does the /refs-pane pane show?

status: answered
act: ask
kind: multi
thread: refs-pane
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:20:37.561752+00:00

## Context

**Default if unanswered: grouped + detail.**

- source: <git root of the session cwd>/.ai/refs.json
- this repo has no refs.json yet; other projects may

— slugs —
touched: skills/harvest/SKILL.md, skills/harvest/bin/sync.sh, .claude-plugin/marketplace.json, CLAUDE.md

## Options

- list — every ref, one line each: kind, key, value, home
+ same as `refs`, live beside the chat
- grouped — refs grouped under kind headings (requirement, decision, answer, pr...)
+ scans faster with canon + cactus answers mixed in  (★◐)
- detail — select a ref to see its full value, title/url, home, set_at
- counts — a header line with counts per kind and the refs.json path

## Recommendation

grouped — med
kinds are the natural sections now that answers and canon share the store

## Answer

counts, detail, grouped
answered at: 2026-11-08T07:25:50.011143+00:00
