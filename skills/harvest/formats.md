# Fact formats

Three files in the repo's `docs/`. Each entry is a `##` heading with a stable ID; `bin/sync.sh` reads the headings.

	file	ID	holds
	docs/INVARIANTS.md	INV-NN	predicates that hold before and after every change
	docs/NEVERS.md	NEV-NN	shapes the code must not take, with the pull toward each
	docs/DECISIONS.md	DEC-NN	choices with a rejected alternative and a costly reversal

## Invariant

    ## INV-03 <predicate, true or false, one line>
    Why: <what breaks when it does not hold>
    Applies to: <path globs, or all>
    Check: <test name | command and expected output | review>
    Source: <path:line, commit or PR that states it>

## Never

    ## NEV-02 <the failure mode, one line>
    Temptation: <the real pull toward it>
    Instead: <the shape to take>
    Applies to: <path globs, or all>
    Seen: <path:line, commit or PR where it showed up>

## Decision

    ## DEC-07 · 2026-11-07 — <decision as one imperative clause>
    - **Decision** — what is now true, present tense
    - **Forces** — constraints that made this the answer, with numbers where they exist
    - **Rejected** — each alternative, one line, with why it lost
    - **Cost** — what this makes harder
    - **Reversal** — the signal to revisit and what undoing takes; "one-way" if it cannot be undone
    - **Anchors** — path:line, commit or PR

## Rules

- IDs never change and are never reused.
- Retire an entry with a `Retired: <date> <why>` line in its block. Decisions are append-only: a reversal is a new DEC that says `Supersedes DEC-NN`, and the old one gets `Retired: <date> superseded by DEC-MM`.
- Unconfirmed candidates live under `## Proposed` at the end of the file, without IDs.

## New file header

    # Invariants
    What must stay true in this codebase, before and after every change. Cite by ID.

(`# Nevers` and `# Decisions` likewise.)
