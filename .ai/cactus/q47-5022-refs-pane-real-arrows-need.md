# q47 — refs-pane: real ←/→ arrows need a rewrite; worth it?

status: answered
act: ask
kind: choice
thread: canon6
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-09T19:18:43.907705+00:00

## Context

**Default: keep-hl.**

- the mod API takes only one letter or digit as a Button hotkey
- in a focused pane the engine owns the arrows (scrolling)
- only a Client element gets raw ←/→ keys, and only after a click

— slugs —
touched: main.go, hook.go, cmd.go, skills/tighten/SKILL.md

## Files

- mods/refs-pane/hooks/register.tsx

## Options

- keep-hl — h/l switch tabs (shipped in 3a4bab3)
+ works the moment the pane has focus
- not the arrow keys  (★◐)
- client — redraw the pane as a Client element that hears raw keys
+ ←/→ tabs, ↑/↓ rows
- you click the pane first; Esc gives focus back
- the drawing moves into a client module
- both — client for arrows, h/l kept as a fallback

## Recommendation

keep-hl — med

## Answer

both
answered at: 2026-11-09T21:42:10.689263+00:00
