# Nevers
Shapes the code must not take, with the pull toward each. Cite by ID.

## NEV-01 A third-party module in the Go program
Temptation: a library for ordered JSON or flag parsing.
Instead: stdlib only; ordered JSON is `omap` in store.go, flags are parsed by hand in cmd.go.
Applies to: go.mod, *.go
Seen: CLAUDE.md:7

## NEV-02 A new subcommand left out of `reserved`
Temptation: add the subcommand and forget the reserved list, so a key with that name stops working.
Instead: add the name to `reserved` in cmd.go in the same change.
Applies to: cmd.go
Seen: CLAUDE.md:43

## NEV-03 Rewriting an existing canon entry in place
Temptation: fix an old entry's wording when its meaning drifts.
Instead: a new entry with the next free ID, and `Retired: <date> <why>` on the old one.
Applies to: docs/canon/*.md
Seen: skills/harvest/SKILL.md:44

## NEV-04 Promoting a Proposed candidate without the human's answer
Temptation: move Proposed items up to save the human a step.
Instead: show them as a cactus ask; promote or drop only what the human answers.
Applies to: docs/canon/*.md
Seen: skills/harvest/SKILL.md:48

## NEV-05 A generic best-practice warning recorded as a never
Temptation: it sounds like a rule.
Instead: record only a pull someone felt: a revert, a blocked PR, a "don't" in the docs.
Applies to: docs/canon/NEVERS.md
Seen: skills/harvest/SKILL.md:36
