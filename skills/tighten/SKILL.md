---
name: tighten
description: Consolidate a project's canon and refs — make every truth one checkable predicate and every never a real pull with a real alternative; merge duplicates, move misfiled entries, retire dead ones, resolve answered open questions, then re-sync refs. Use on "tighten", "/refs:tighten", "consolidate the refs", "tighten the canon", "clean up truths and nevers", or after a harvest. Do NOT use to build canon from scratch (use harvest) or to add one ref (use `refs KEY VALUE`).
---

# tighten

Harvest gets the canon written. Tighten makes it worth reading: fewer entries, each one sharp enough that an agent can act on it and a check can catch it breaking. Truths and nevers get the hardest look, because they are the entries agents cite while changing code.

Formats: `../harvest/formats.md`.

## 1. Read

- `refs`, and the six files in `docs/canon/`: REQUIREMENTS, TRUTHS, NEVERS, DECISIONS, INQUIRY, OPEN_QUESTIONS. Skip a file that does not exist; tighten never creates one.
- Note the next free ID per file. IDs never change and are never reused.

## 2. Audit, one subagent per file

Each brief carries `formats.md`, the file, the rules below for its kind, and the IDs of the other five files so it can flag overlaps. Each returns proposals, one per line:

    {"op":"reword|merge|split|move|retire|resolve|fix-source","ids":["TRU-03"],"to":"TRU|NEV|REQ|DEC|INQ|OQ","new":"<full new block, or null>","why":"<one line>","evidence":"<path:line, test name, commit>"}

**Truths**: the tightest file.
- One predicate per entry, true or false, one line. "and" joining two claims → split.
- No "should", "try to", "ideally", "usually". A truth that does not always hold is not a truth: narrow `Applies to` until it does, or retire it.
- `Check:` names a test that exists (grep for it), or a command with its expected output. `Check: review` only when no check can exist; say why.
- Two truths where one implies the other → keep the stronger, retire the weaker.
- A capability someone uses → move to requirement. A ban → move to never. A choice between options → move to decision.

**Nevers**:
- Names a failure mode someone actually felt. `Seen:` points at a real commit, PR, revert or doc line. No evidence → retire it as a generic warning.
- `Temptation:` is the concrete pull ("a map is two lines shorter"), not "it is easier".
- `Instead:` is a shape to take, not "don't".
- A never that is only the negation of a truth → keep the one with the better check, retire the other.

**Requirements**: product level, per harvest step 3. Caps → truths, gates → decisions, plumbing → retire.

**Decisions**: append-only. Check only that `Anchors` resolve and that `Supersedes` / `Retired: superseded by` pairs match. Never reword a decision's substance.

**Inquiry**: each entry is a question building this answers, not a goal or a feature. `Answered when` names evidence someone could gather. An answered inquiry gets `Status: answered <date> by <IDs>`.

**Open questions**: an OQ whose answer now exists (a DEC, a TRU, an answered cactus row) → `resolve`. An OQ nobody has touched across several sessions → list it in step 4 as stale.

**Every file**: `Source:`, `Seen:`, `Check:` and `Anchors` paths exist and their line still says what the entry claims (`fix-source` when the line moved).

## 3. Judge (the lead)

- Cross-file duplicates: a truth and a never saying one thing, a requirement restating a truth. Keep one.
- Apply the edit rule:
  - **reword** in place when the meaning holds: tighter words, a narrower `Applies to`, a real `Check:`.
  - **merge, split, move, or a change of meaning** retires the old ID (`Retired: <date> <why>, see <new ID>`) and adds a new entry with the next free ID in its file.
  - Decisions: a reversal is a new DEC with `Supersedes DEC-NN`.
- Drop proposals without evidence.

## 4. The human approves

Post the batch to cactus before writing, one ask per file with changes, `--multi`, one `-c` per proposal (`op ID: one line`), `-f` the canon file, `--refs` the IDs it touches, `--recommend` the ones you would bet on. List stale open questions as a separate ask: keep, retire, or post to cactus as its own question. Without cactus, show the batch in chat.

Write only what was approved.

## 5. Re-sync

    ../harvest/bin/sync.sh --github --cactus

Retired and resolved entries lose their refs; new IDs get theirs. Report the counts per kind before and after.

Leave committing to the human or the session's commit flow.
