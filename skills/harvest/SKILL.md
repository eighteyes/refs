---
name: harvest
description: Turn an existing repo into canon and refs — read its docs, git history, GitHub issues and PRs, and code; write what holds as invariants (docs/canon/INVARIANTS.md, INV-NN), nevers (docs/canon/NEVERS.md, NEV-NN) and decisions (docs/canon/DECISIONS.md, DEC-NN); then mirror every entry and the open issues and PRs into refs with a home at its line. Use when the refs hook says a project has no refs yet, or on "harvest", "/refs:harvest", "turn this repo into refs", "harvest the canon", "bootstrap the canon". Do NOT use to add one ref by hand (use `refs KEY VALUE`).
---

# harvest

The canon is the set of calls that outlive one issue: what always holds (invariants), what the code must never become (nevers), and what was chosen over what (decisions). They live in three files in the repo, versioned with the code. refs points at them, so every session can recall them by ID and read the source line before trusting the value.

Formats: `formats.md` beside this file.

## 1. Read what exists

- `refs` (current refs) and `docs/canon/INVARIANTS.md`, `docs/canon/NEVERS.md`, `docs/canon/DECISIONS.md`. Missing files get created in step 4 from `formats.md`. Files at the legacy `docs/<NAME>.md`: `git mv` them into `docs/canon/` first (say so to the human), never keep two copies.
- Note the next free ID per file. IDs never change and are never reused.

## 2. Harvest, four sources in parallel

One subagent per source, each brief carrying `formats.md` and the existing entries so it skips duplicates. Each returns candidates, one per line:

    {"type":"INV|NEV|DEC","title":"<one line>","fields":{...},"source":"<path:line | commit sha | PR url>","stated":true|false,"quote":"<the words that state it>"}

- **docs**: README, CLAUDE.md, AGENTS.md (a `## Shape` block especially), CONTRIBUTING, `docs/`, an intent-shape doc. "Never", "always", "must", "do not", "we chose", "instead of".
- **history**: `git log` for reverts, large refactors, and messages that choose ("switch to", "instead of", "drop", "replace"). Each is a DEC candidate; the forces and the rejected option come from the message or the diff.
- **github**: closed PR and issue discussions with a settled choice (`gh pr list --state merged --limit 50`, then `gh pr view N --comments` on the promising ones). Review comments that block a pattern → NEV candidates.
- **code**: schemas, migrations, public API, config validation, tests named for a guarantee ("never", "always", "must", "rejects"). A test is a ready-made `Check:` line.

`stated: true` only when the source says it in words, quoted. Anything read off the code's shape alone is `stated: false`.

## 3. Judge (the lead, not the subagents)

- Dedupe across sources; one entry cited by two sources keeps both.
- **Invariant**: a predicate, true or false, with a check. No check → find one or mark it `Check: review`.
- **Never**: names a pull someone actually felt (a revert, a blocked PR, a "don't" in the docs). A generic warning is not an entry.
- **Decision**: has a rejected alternative and a costly reversal. Otherwise drop it.
- Conflicting sources → keep both quotes, mark it `stated: false`, and ask.

## 4. Write

- `stated: true` → append to its file with the next free ID. `Source:` (or `Anchors` for a DEC) is the cited line, commit or PR.
- `stated: false` → under `## Proposed` at the end of its file, no ID, with its evidence.
- Never rewrite an existing entry. An entry that changed meaning is a new entry plus `Retired: <date> <why>` on the old one.

## 5. The human promotes

Show the Proposed list (a cactus ask when cactus is installed, else in chat): each item, its evidence, promote or drop. Promoted items move up with the next free ID; dropped ones are deleted from Proposed. Never promote unasked.

## 6. Mirror into refs

    bin/sync.sh --github

One ref per live entry (`INV-03`, `NEV-02`, `DEC-07`) with its home at the heading line, one per open issue (`issue#12`) and PR (`pr#40`). Retired entries and closed issues lose their refs. Re-run it after any edit to the three files; it is idempotent. `--dry-run` prints the calls.

Leave committing the three files to the human or the session's commit flow.
