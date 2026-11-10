# q41 — What does the SessionStart blurb do?

status: answered
act: ask
kind: choice
thread: canon6
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-09T17:15:07.960615+00:00

## Context

**Default: dynamic+filter.**

- today `refs` lists all 55; no kind filter
- hook must stay exit 0, cheap

## Files

- hook.go

## Options

- static — fixed text naming the 6 canon files + `refs KEY` to read one
- dynamic — names only canon files present, with per-kind counts
- dynamic+filter — dynamic, plus new `refs --kind NAME` list filter
+ agent can list just truths
- new CLI flag, help+README sync  (★◐)

## Recommendation

dynamic+filter — med

## Answer

dynamic+filter
answered at: 2026-11-09T18:48:57.385475+00:00
