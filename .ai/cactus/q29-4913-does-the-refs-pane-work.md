# q29 — Does the refs pane work live?

status: passed
act: review
kind: confirm
thread: refs-pane
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:32:51.477826+00:00

## Context

— slugs —
touched: skills/harvest/SKILL.md, skills/harvest/bin/sync.sh, .claude-plugin/marketplace.json, CLAUDE.md

## Files

- /Users/god/.claude/dev-mods/72f7858d-f45b-4670-a659-3b5441573403/refs-pane/hooks/register.tsx

## Options

- pass
- fail

## Verdicts

look at: /refs-pane in this session (mod must be turned on), in a project that has .ai/refs.json
run: skills/harvest/bin/sync.sh --cactus
pass: pane shows counts header, an answer group with q8..q25, j/k moves, o opens the record file
fail: pane missing, blank, or o does nothing
then: copy into mods/refs-pane and commit

- 2026-11-08T07:36:10.411841+00:00: pass
