# q44 — How should /refs-pane reach every session?

status: answered
act: ask
kind: choice
thread: canon6
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-09T19:15:10.356851+00:00

## Context

**Default: entry.**

- today the pane loads only in this session, from its dev-mods folder
- other sessions never saw it; the repo copy is not installed anywhere

— slugs —
touched: main.go, hook.go, cmd.go, skills/tighten/SKILL.md

## Files

- /Users/god/projects/ai-tooling/utils/refs/.claude-plugin/marketplace.json
- /Users/god/projects/ai-tooling/utils/refs/mods/refs-pane/.claude-plugin/plugin.json

## Options

- entry — add refs-pane as a 2nd plugin in the refs marketplace, install refs-pane@refs at user scope
+ same pattern as refs itself
- a second plugin to bump on change  (★●)
- fold — merge the mod into the refs plugin (one install)
- unverified that a plugin can mix command hooks and mod modules
- plugin-dir — point CLAUDE_CODE_PLUGIN_DIRS at mods/refs-pane in user settings
+ edits go live without a bump
- dev-style; loads straight from the working tree

## Recommendation

entry — high

## Answer

entry
answered at: 2026-11-09T21:41:47.400131+00:00
