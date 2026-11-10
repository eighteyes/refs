# q22 — Cactus records don't write the --pr tag. Add a pr: line to the record, or have sync ask the cactus DB?

status: answered
act: ask
kind: choice
thread: integrate
parent: q21
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T06:44:16.806362+00:00

## Context

**Default if unanswered: record-field.**

- checked: src/cactus/record.py _field_block writes status, act, kind, thread, parent, agent, cwd, asked at; no pr
- cactus --pr STATE exists on ask/run/edit

— slugs —
touched: .claude-plugin/marketplace.json, CLAUDE.md, README.md, bin/refs

## Files

- ~/projects/cactus/src/cactus/record.py

## Options

- record-field — cactus record.py writes pr: STATE in the field block (1 line + test)
+ records alone carry the tag; sync reads only files
- a cactus change; I send it to session cactus-b2  (★●)
- db-lookup — sync.sh runs cactus get qN --json per record for pr
+ no cactus change
- sync needs cactus installed and the DB row alive
- both — write the field, fall back to the DB for old records

## Recommendation

record-field — high
the record is the home; it should hold the filter

## Answer

db-lookup
answered at: 2026-11-08T06:46:52.672382+00:00
