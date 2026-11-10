# q52 — Publish .ai/ (cactus records + refs.json) with the repo?

status: answered
act: ask
kind: choice
thread: gh
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-10T07:18:23.437719+00:00

## Context

**Default: keep.**

- no secrets found, just home-dir paths and session ids
- history already holds them; scrub only cleans the tip

## Options

- keep — push them as they are
+ DEC anchors and q refs point at records
- 36 records hold local paths (/Users/god/...) and agent ids  (★○)
- scrub — rewrite local paths to repo-relative in a new commit, then push
- untrack — git rm --cached .ai/cactus, gitignore it
- canon Anchors pointing at records dangle

## Recommendation

keep — low

## Answer

scrub
answered at: 2026-11-10T07:29:05.527438+00:00
