# q51 — eighteyes/refs on GitHub: public or private?

status: answered
act: ask
kind: choice
thread: gh
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-10T07:18:23.399104+00:00

## Context

**Nothing is created until you answer.**

- secret scan of the tracked files: no hits
- gh is logged in as lucidbots-sean (active) and eighteyes; I switch to eighteyes to create the repo, then switch back
- pushes main only; feat/harvest stays local

— slugs —
touched: main.go, hook.go, store.go, cmd.go

## Options

- public — anyone can read the code and install the plugin
+ `go install github.com/eighteyes/refs@latest` works for anyone
+ the marketplace installs with one line  (★◐)
- private — only people you invite; they need GitHub set up to install
- go install needs GOPRIVATE and git auth

## Recommendation

public — med

## Answer

public
answered at: 2026-11-10T07:28:56.935320+00:00
