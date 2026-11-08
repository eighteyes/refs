# q21 — What does the --pr flag mean for skipping process noise in cactus records?

status: answered
act: ask
kind: choice
thread: integrate
parent: q19
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T06:39:28.628151+00:00

## Context

**Default if unanswered: wait; nothing is built.**

- records carry: status, act (ask/steer/run/review), kind (confirm/choice/multi), thread
- cactus already has --pr STATE (draft|review|changes|ci-fail|ready|merged)
- here, the noise rows are the merge confirms and the denied-command run row

— slugs —
touched: .claude-plugin/marketplace.json, CLAUDE.md, README.md, bin/refs

## Options

- cactus-tag — agents tag process rows with cactus's existing --pr STATE; the mirror skips any record with a pr tag
+ flag exists today
- --pr means "work sits in a PR"; rows not about a PR stay in
- new-flag — a new mirror flag --pr that keeps only records that settle intent
- you name what "settles intent" means
- by-shape — no flag; skip act=run, thread=denied, and confirm rows by default; --all keeps them
+ works on the 10 records here: drops q11,q12,q14,q15,q16
- a real yes/no decision gets skipped
- mark-at-ask — add a cactus flag (e.g. --process) agents set when posting; mirror skips those
+ exact
- a cactus change, and agents must remember it

## Answer

cactus-tag
answered at: 2026-11-08T06:43:36.400967+00:00
