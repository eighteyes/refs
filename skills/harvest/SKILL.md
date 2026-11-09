---
name: harvest
description: Turn an existing repo into canon and refs — read its docs, git history, GitHub issues and PRs, and code; write what holds as requirements (docs/canon/REQUIREMENTS.md, REQ-NN), truths (docs/canon/TRUTHS.md, TRU-NN), nevers (docs/canon/NEVERS.md, NEV-NN) and decisions (docs/canon/DECISIONS.md, DEC-NN), inquiry (docs/canon/INQUIRY.md, INQ-NN) and open questions (docs/canon/OPEN_QUESTIONS.md, OQ-NN); then mirror every entry and the open issues and PRs into refs with a home at its line. Use when the refs hook says a project has no refs yet, or on "harvest", "/refs:harvest", "turn this repo into refs", "harvest the canon", "bootstrap the canon". Do NOT use to add one ref by hand (use `refs KEY VALUE`).
---

# harvest

The canon is the set of calls that outlive one issue: what the product must do (requirements), what always holds (truths), what the code must never become (nevers), what was chosen over what (decisions), what building this is meant to answer (inquiry), and what is still unresolved (open questions). They live in six files in the repo, versioned with the code. refs points at them, so every session can recall them by ID and read the source line before trusting the value.

Formats: `formats.md` beside this file.

## 1. Read what exists

- `refs` (current refs) and `docs/canon/REQUIREMENTS.md`, `docs/canon/TRUTHS.md`, `docs/canon/NEVERS.md`, `docs/canon/DECISIONS.md`, `docs/canon/INQUIRY.md`, `docs/canon/OPEN_QUESTIONS.md`. Missing files get created in step 4 from `formats.md`. Files at the legacy `docs/<NAME>.md`: `git mv` them into `docs/canon/` first (say so to the human), never keep two copies.
- Note the next free ID per file. IDs never change and are never reused.

## 2. Harvest, five sources in parallel

One subagent per source, each brief carrying `formats.md` and the existing entries so it skips duplicates. Each returns candidates, one per line:

    {"type":"REQ|TRU|NEV|DEC|INQ|OQ","title":"<one line>","fields":{...},"source":"<path:line | commit sha | PR url>","stated":true|false,"quote":"<the words that state it>"}

- **docs**: README, CLAUDE.md, AGENTS.md (a `## Shape` block especially), CONTRIBUTING, `docs/`, an intent-shape doc (its Truths and Nevers sections, or Inviolates in older docs, are stated entries; skip any line already tagged with a TRU- or NEV- ID). "Never", "always", "must", "do not", "we chose", "instead of". For requirements, return what a person does and sees; sort caps, gates and infrastructure by the step 3 rules. Goals, "why this exists", research questions → INQ; "TBD", "open question", "not sure yet", TODOs that name a choice → OQ.
- **history**: `git log` for reverts, large refactors, and messages that choose ("switch to", "instead of", "drop", "replace"). Each is a DEC candidate; the forces and the rejected option come from the message or the diff.
- **github**: closed PR and issue discussions with a settled choice (`gh pr list --state merged --limit 50`, then `gh pr view N --comments` on the promising ones). Review comments that block a pattern → NEV candidates.
- **code**: schemas, migrations, public API, config validation, tests named for a guarantee ("never", "always", "must", "rejects"). A test is a ready-made `Check:` line.
- **cactus**: the answered records in `.ai/cactus/` that qualify: the row has an answer, the answer is not skipped, the row has no `pr` tag, and it is not in thread `denied` (read rows with `cactus get qN ... --json`). Each is a DEC candidate; the forces and rejected options come from the record's choices and context.

`stated: true` only when the source says it in words, quoted. Anything read off the code's shape alone is `stated: false`.

## 3. Judge (the lead, not the subagents)

- Dedupe across sources; one entry cited by two sources keeps both.
- **Requirement**: product level. A person who uses the product does something and sees a result: "An operator prices a building from its measurements." The title names who and what they do; Acceptance names what they see. A feature list, spec or README "what it does" section is a source; an issue's acceptance criteria is one only when the work shipped. A row in a requirements doc is a candidate to sort, not a requirement by default. Sort out what is not product level:
  - a cap, limit, or "every / no path / at most N" rule → truth
  - a switch, gate or phase ("off until", "behind a flag", "in phase 3") → decision, or drop
  - hosting, database, deploy, security plumbing, test coverage → drop, or a truth when it states a guarantee
  - the source names no outcome the person sees → write one Acceptance line from the source's words and mark it `stated: false`
- **Truth**: a predicate, true or false, with a check. No check → find one or mark it `Check: review`.
- **Never**: names a pull someone actually felt (a revert, a blocked PR, a "don't" in the docs). A generic warning is not an entry.
- **Decision**: has a rejected alternative and a costly reversal. Otherwise drop it.
- **Inquiry**: a question building this answers, phrased as a question, with evidence that would settle it. A goal or feature list is a source, not an entry.
- **Open question**: an unresolved choice or unknown that blocks or bends work. One already answered (a DEC, a TRU, an answered cactus row) is not an entry.
- Conflicting sources → keep both quotes, mark it `stated: false`, and ask.

## 4. Write

- `stated: true` → append to its file with the next free ID. `Source:` (or `Anchors` for a DEC) is the cited line, commit or PR.
- `stated: false` → under `## Proposed` at the end of its file, no ID, with its evidence.
- Never rewrite an existing entry. An entry that changed meaning is a new entry plus `Retired: <date> <why>` on the old one.

## 5. The human promotes

Show the Proposed list (a cactus ask when cactus is installed, else in chat): each item, its evidence, promote or drop. Promoted items move up with the next free ID; dropped ones are deleted from Proposed. Never promote unasked.

## 6. Mirror into refs

    bin/sync.sh --github --cactus

One ref per live entry (`REQ-04`, `TRU-03`, `NEV-02`, `DEC-07`, `INQ-01`, `OQ-05`) with its home at the heading line, one per open issue (`issue#12`) and PR (`pr#40`), one per qualifying cactus answer (`q19`, home at the record). Retired entries, resolved open questions, closed issues and answers that stop qualifying lose their refs. Re-run it after any edit to the six files; it is idempotent. `--dry-run` prints the calls.

Leave committing the six files to the human or the session's commit flow.
