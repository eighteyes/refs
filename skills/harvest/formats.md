# Canon formats

Four files in the repo's `docs/canon/`. Each entry is a `##` heading with a stable ID; `bin/sync.sh` reads the headings.

	file	ID	holds
	docs/canon/TRUTHS.md	TRU-NN	predicates that hold before and after every change (broader)
	docs/canon/REQUIREMENTS.md	REQ-NN	what the product must do, each with acceptance criteria (granular)
	docs/canon/NEVERS.md	NEV-NN	shapes the code must not take, with the pull toward each
	docs/canon/DECISIONS.md	DEC-NN	choices with a rejected alternative and a costly reversal

## Requirement

    ## REQ-04 <what the product must do, one line>
    Why: <who needs it and what it serves>
    Acceptance: <observable criteria, one per line, each with how to check it>
    Applies to: <path globs, or all>
    Source: <path:line, issue, PR or spec that states it>

## Truth

    ## TRU-03 <predicate, true or false, one line>
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
- A decision too big for its entry also gets `docs/canon/adrs/DEC-07-<slug>.md` (full context, options weighed, consequences). The entry stays the summary and names it under Anchors: `Long form: adrs/DEC-07-<slug>.md`. refs points at the entry, not the long form.
- Legacy: `INVARIANTS.md` with `INV-NN` IDs is the old name for truths; read it when `TRUTHS.md` is missing and keep its IDs.
- Legacy layout: files at `docs/<NAME>.md` (before `docs/canon/`). Read them when `docs/canon/` has none; `sync.sh` does the same.

## New file header

    # Truths
    What must stay true in this codebase, before and after every change. Cite by ID.

(`# Requirements`, `# Nevers` and `# Decisions` likewise.)
