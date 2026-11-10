# q9 — go install, build-on-first-run, or both: which one does the hook rely on?

status: answered
act: ask
kind: choice
thread: go-port
parent: q8
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-07T05:56:32.304948+00:00

## Context

**Default if unanswered: both.**

- agents run refs from Bash, so it must be on PATH either way
- checked: no plugin bin/ dir is on PATH in this session; cactus itself lives in ~/.local/bin
- so first-run alone covers the hook, not Bash calls

## Files

- hooks/hooks.json

## Options

- both — bin/refs shim builds into the plugin dir on first run; go install also works for PATH use
+ plugin works with no extra step, CLI on PATH for humans
- two ways in, one shim to keep  (★◐)
- first-run — hook calls ${CLAUDE_PLUGIN_ROOT}/bin/refs, shim builds once
+ plugin is self-contained
- agents calling refs in Bash still need it on PATH
- go-install — hook calls refs on PATH, install hint if missing
+ no shim
- plugin is dead until the user runs go install

## Recommendation

both — med
plugin works out of the box, PATH stays clean

## Answer

both
answered at: 2026-11-07T06:16:21.778084+00:00
