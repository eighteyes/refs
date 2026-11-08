# Decisions
Choices with a rejected alternative and a costly reversal. Cite by ID.

## DEC-01 · 2026-11-06 — Write refs in Go, stdlib only
- **Decision** — refs is a Go program with the same subcommands, flags, output, exit codes and refs.json format as the 0.1.x Python script.
- **Forces** — Sean did not want to maintain Python. A parity run of ~130 commands matched the Python output byte for byte.
- **Rejected** — stay on python3: the reason for the port. TypeScript on bun: picked first, then dropped for Go.
- **Cost** — a Go toolchain on the machine, or a prebuilt binary; the hook prints an install hint without one.
- **Reversal** — restore bin/refs from effbd15. The data survives because refs.json stays byte-compatible.
- **Anchors** — c5d6d26

## DEC-02 · 2026-11-06 — Keep the store an insertion-ordered JSON codec, not Go maps
- **Decision** — store.go decodes and encodes refs.json through `omap`, which keeps key order and unknown fields.
- **Forces** — Go maps lose order. The Python store listed refs in insertion order and kept unknown fields.
- **Rejected** — `map[string]any` with encoding/json: keys come back sorted and unknown fields are dropped.
- **Cost** — about 400 lines of hand-written JSON in store.go.
- **Reversal** — accept sorted keys and switch to encoding/json; every refs.json diff churns once.
- **Anchors** — store.go:4, c5d6d26

## DEC-03 · 2026-11-06 — Package main at the repo root, module github.com/eighteyes/refs
- **Decision** — the Go sources sit at the repo root beside the plugin files; install is `go install github.com/eighteyes/refs@latest`.
- **Forces** — a ~400-line tool; the shortest install line.
- **Rejected** — cmd/refs/main.go with logic in internal/: more directories than the tool needs.
- **Cost** — plugin files and Go sources share the root.
- **Reversal** — move to cmd/refs; the install path changes for every user.
- **Anchors** — .ai/cactus/q8-4718-go-layout-module-github-com-eighteyes-re.md

## DEC-04 · 2026-11-06 — Ship both a first-run launcher and `go install`
- **Decision** — hooks.json runs bin/refs, a sh launcher that builds bin/.build/refs on first run and rebuilds on change; `go install` puts refs on PATH for Bash calls.
- **Forces** — agents call refs from Bash, so it must be on PATH; no plugin bin/ directory was on PATH in the session.
- **Rejected** — launcher only: Bash calls fail. go install only: the plugin hook is dead until the user installs. Committed per-platform binaries: binaries in git.
- **Cost** — two install paths and a launcher to keep working.
- **Reversal** — point hooks.json at refs on PATH and delete the launcher.
- **Anchors** — .ai/cactus/q9-4719-go-install-build-on-first-run-or-both.md, c5d6d26

## DEC-05 · 2026-11-06 — `refs seed` prints the populate instructions and replaces the reserved `install`
- **Decision** — `seed` prints the empty-project hook text; `install` is an ordinary key again.
- **Forces** — `install` was reserved and did nothing; the hook already tells the agent how to populate refs.
- **Rejected** — scaffold a refs.json with starter kinds: it picks kinds for the user. Scaffold and print: same objection. Reserve the word and build nothing.
- **Cost** — new projects get instructions, not a starter file.
- **Reversal** — any key named `seed` stops working if the subcommand changes meaning.
- **Anchors** — .ai/cactus/q13-4748-what-should-refs-seed-do.md, f5ec2ca

## DEC-06 · 2026-11-07 — Canon lives in docs/canon/, with a legacy docs/ fallback
- **Decision** — the four canon files live in docs/canon/; sync.sh and harvest read docs/<NAME>.md when docs/canon/ has none; long forms go in docs/canon/adrs/.
- **Forces** — canon is versioned with the code and read by sync.sh.
- **Rejected** — flat docs/<NAME>.md, the earlier layout.
- **Cost** — a legacy read path until old repos migrate.
- **Reversal** — `git mv` the files back and drop the fallback; every home in refs changes.
- **Anchors** — 3d22498

## DEC-07 · 2026-11-07 — Call invariants truths (TRU-NN) and keep legacy INV-NN IDs
- **Decision** — predicates that always hold are TRU-NN in TRUTHS.md; INVARIANTS.md with INV-NN is read when TRUTHS.md is missing.
- **Forces** — IDs never change, so old INV IDs must stay valid.
- **Rejected** — keep the invariant name and INV prefix.
- **Cost** — two prefixes live during migration.
- **Reversal** — rename back; refs keyed TRU- would all change.
- **Anchors** — b81fdc0

## DEC-08 · 2026-11-07 — Cactus answers flow into canon and into refs as qN
- **Decision** — harvest reads answered cactus records as a fifth source of decision candidates; `sync.sh --cactus` mirrors each qualifying record as ref qN, home at the record.
- **Forces** — decisions landed in three places (canon, the know graph, cactus records); canon stays the one source.
- **Rejected** — mirror know graph nodes: no file:line home. Canon only: answers never reach refs.
- **Cost** — sync needs cactus on PATH, and process answers need filtering.
- **Reversal** — drop `--cactus` and the fifth source; q refs disappear.
- **Anchors** — .ai/cactus/q19-4864-decisions-now-land-in-3.md, e26fa94

## DEC-09 · 2026-11-07 — Filter process noise with cactus's `--pr` tag, read from the cactus DB
- **Decision** — agents tag merge and ship rows with `--pr STATE`; sync.sh reads each row's tag with `cactus get --json` and skips tagged rows.
- **Forces** — the flag already exists; records do not write the tag.
- **Rejected** — a new mirror flag: needs a definition of "settles intent". Skip by shape: drops real yes/no decisions. A new cactus `--process` flag: a cactus change agents must remember. Write `pr:` into records: a cactus change.
- **Cost** — sync depends on the cactus DB row being alive; untagged process rows leak through.
- **Reversal** — move the tag into the record file and read files only.
- **Anchors** — .ai/cactus/q21-4872-what-does-the-pr-flag.md, .ai/cactus/q22-4879-cactus-records-don-t-write-the.md

## DEC-10 · 2026-11-07 — Skip denied-thread rows, and let cactus tag answered rows with `--pr`
- **Decision** — sync.sh skips thread `denied`; cactus allows `edit --pr` on answered and cleared rows (cactus-go eb5f87c), so old merge rows were tagged.
- **Forces** — denied rows are auto-posted blocked commands, never decisions; five old merge rows were untagged.
- **Rejected** — skip denied only: four merge confirms stay as refs. Accept the noise and tag future rows only. Cactus change only: q14 stays mirrored.
- **Cost** — a real choice filed in thread denied is dropped.
- **Reversal** — remove the thread filter from sync.sh's jq select.
- **Anchors** — .ai/cactus/q25-4890-the-pr-filter-works-but.md, 673c9b5

## DEC-11 · 2026-11-07 — refs-pane is read-only: grouped by kind, counts, detail, open home
- **Decision** — `/refs-pane` shows a counts header, refs grouped under kind headings and a detail block for the selection; its one action opens the selected ref's home. It never writes refs.json.
- **Forces** — kinds are the natural sections now that canon and cactus answers share the store; the home is the source of truth.
- **Rejected** — a flat list. Insert the key into the prompt. Edit or remove refs from the pane: writes refs.json from a UI.
- **Cost** — edits still go through the CLI.
- **Reversal** — adding writes means confirm flows and a write path in the mod.
- **Anchors** — .ai/cactus/q26-4902-what-does-the-refs-pane-pane.md, .ai/cactus/q27-4903-what-can-you-do-from.md, f64e6db

## DEC-12 · 2026-11-07 — The repo is its own plugin marketplace, installed at user scope
- **Decision** — .claude-plugin/marketplace.json names refs at source `./`; the plugin installs as refs@refs at user scope.
- **Forces** — the same pattern as cactus; the hook should run in every project.
- **Rejected** — copy into the eighteyes-plugins marketplace: the copy drifts. A bare SessionStart hook in settings.json: loses the plugin's skills.
- **Cost** — the install is a cached copy; changes need a version bump and `claude plugin update refs@refs`.
- **Reversal** — uninstall, remove the marketplace, and move to another distribution.
- **Anchors** — a04754e

## DEC-13 · 2026-11-07 — Requirements are a fourth canon kind (REQ-NN)
- **Decision** — capabilities with acceptance criteria are REQ-NN entries in REQUIREMENTS.md, mirrored into refs like the other kinds.
- **Forces** — a capability someone can check is narrower than a predicate about every case; truths are broader, requirements granular.
- **Rejected** — fold capabilities into truths (implied by the diff; promoted by the human in cactus q35).
- **Cost** — one more file to keep in sync; sources must be sorted into the right kind.
- **Reversal** — fold REQ entries into TRUTHS.md and retire every REQ ref.
- **Anchors** — a3fdb6a, .ai/cactus/q35-4920-promote-the-2-proposed-decisions.md

## DEC-14 · 2026-11-07 — A key's kind is its leading non-digit prefix
- **Decision** — `R20` is kind `R`, `pr#131` is `pr#`; `refs kind PREFIX NAME` names a prefix, and an unnamed prefix saves as `unknown`.
- **Forces** — keys like R20 and pr#131 already carry their kind; one command sets a ref.
- **Rejected** — a kind field required on every set (implied; no alternative recorded; promoted by the human in cactus q35).
- **Cost** — keys cannot start with a digit; new prefixes need naming.
- **Reversal** — every existing key and kind mapping would need migrating.
- **Anchors** — CLAUDE.md:42, README.md:26, .ai/cactus/q35-4920-promote-the-2-proposed-decisions.md
