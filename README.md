# refs

A per-project key-value store for the pillars of an agent's session (requirements, PRs, decisions), read on demand, shipped as a Go CLI and a Claude Code plugin.

```
$ refs
requirement  R1     Users can export a report as CSV  @ docs/spec.md:1
pr           pr#12  Add CSV export — https://github.com/acme/app/pull/12
unknown      D1     Store dates as UTC
```

Each ref is a key, a value and, optionally, a home: the file and line that is its source of truth. Nothing is injected into the context. A SessionStart hook tells the agent the refs exist and how many there are, and the agent runs `refs` when it needs them. On top of the store, two skills build and maintain a project's canon (requirements, truths, nevers, decisions, inquiry, open questions) and mirror every entry into refs.

## Install

As a Claude Code plugin, for every project:

```
claude plugin marketplace add eighteyes/refs
claude plugin install refs@refs --scope user
```

On your PATH, so agents can call it from Bash:

```
go install github.com/eighteyes/refs@latest
refs --help
```

- **Go is required.** The plugin's hook runs `bin/refs`, a sh launcher that builds the binary into `bin/.build/` on first run and rebuilds when a `.go` file changes. Without Go, the hook tells the agent that refs needs Go.
- **Both, usually.** The plugin gives you the hook and the skills; `go install` puts `refs` on PATH for the agent's Bash calls.
- **The pane** is a second plugin: `claude plugin install refs-pane@refs --scope user`, then `/refs-pane`.

## Quick start

Name a key prefix, set refs, read them back.

```
refs kind R requirement                                   # R1, R2... are requirements
refs R1 "Users can export a report as CSV" --home docs/spec.md:1
refs 'pr#12' '[Add CSV export](https://github.com/acme/app/pull/12)' --kind pr
refs                                                      # list
refs R1                                                   # show one
refs --kind pr                                            # list one kind
```

```
R1  [requirement]
  value   Users can export a report as CSV
  home    docs/spec.md:1
  open    file:///…/docs/spec.md
  set_at  2026-11-10T08:56:23-0800
```

- **`prefix 'D' has no kind`**: the key's prefix (`D` in `D1`) has no name yet. The ref saves as `unknown`. Run `refs kind D decision`; it relabels the existing `unknown` refs too.
- **Keys can't start with a digit or `-`**, and `rm kind kinds home seed hook ids sync` are reserved.
- **Read the home before trusting the value.** The value is a note; the home is the source.

## For AI agents

`refs --help` is written for a model: when to run it, every command, and the rules for keys and homes. The SessionStart hook runs on `startup|resume|clear|compact`:

- **No `.ai/refs.json` yet**: it tells the agent to populate refs first, with `/refs:harvest` or by hand. `refs seed` prints the same text.
- **Otherwise**: one line on how to use refs, the ref count, the canon files present with a count per kind, how to look one up, and whether canon IDs are fixed or fluid. It also tells the agent to cite refs on cactus questions with `--refs KEY` instead of restating them.

## Canon

Six files in `docs/canon/`. Each entry is a `##` heading with an ID; refs points at the heading line.

| File | ID | Holds |
|---|---|---|
| `REQUIREMENTS.md` | `REQ-NN` | what a person can do with the product, and what they see |
| `TRUTHS.md` | `TRU-NN` | predicates that hold before and after every change, each with a check |
| `NEVERS.md` | `NEV-NN` | shapes the code must not take, with the pull toward each |
| `DECISIONS.md` | `DEC-NN` | choices with a rejected alternative and a costly reversal |
| `INQUIRY.md` | `INQ-NN` | what building this is meant to answer |
| `OPEN_QUESTIONS.md` | `OQ-NN` | unresolved frontier questions; each closes with `Resolved: date → ID` |

| Skill | Does |
|---|---|
| `/refs:harvest` | reads docs, git history, GitHub issues and PRs, code and answered cactus rows; writes stated entries with IDs and unstated ones under `## Proposed` for you to promote; mirrors everything into refs |
| `/refs:tighten` | makes each truth one checkable predicate and each never a real pull with a real alternative; merges duplicates, moves misfiled entries, retires dead ones, resolves answered open questions; you approve each batch before it writes |

`refs sync --github --cactus` re-mirrors the canon, open issues and PRs (`issue#N`, `pr#N`) and answered cactus rows (`qN`) into refs. It is idempotent and leaves `set_at` alone on unchanged refs; `--dry-run` prints the changes and writes nothing. The canon part runs by itself: any refs command and the session hook re-read `docs/canon/*.md` first when a file is newer than `.ai/refs.json`. GitHub and cactus sync only when asked.

**Fixed or fluid IDs.** `refs ids` prints the project's mode. Fluid, the default: tighten may renumber, delete, reuse and reorder IDs while the canon is still taking shape. Fixed: IDs never change and are never reused. Run `refs ids fixed` once the canon settles.

## Reference

| Command | Does |
|---|---|
| `refs` | list refs |
| `refs --kind NAME\|PREFIX` | list one kind, by name or key prefix |
| `refs KEY` | show one |
| `refs KEY VALUE... [--kind NAME] [--home PATH[:LINE]]` | set or overwrite |
| `refs rm KEY` | remove |
| `refs kind PREFIX NAME` / `refs kinds` | name a key prefix / list prefixes |
| `refs home KEY PATH[:LINE]` / `refs home KEY --clear` | set or clear the home |
| `refs ids [fixed\|fluid]` | show or set whether canon IDs may change |
| `refs sync [--github] [--cactus] [--dry-run]` | mirror `docs/canon/*.md` (and open issues/PRs, answered cactus rows) into refs |
| `refs seed` | print how to populate an empty project |
| `refs hook` | the SessionStart hook: JSON payload on stdin, hook JSON out |

**Store.** `<root>/.ai/refs.json`, shape `{"kinds": {}, "refs": {}}`, plus `"ids"` once set. Root is the git toplevel of the current directory, else the directory. Writes are atomic. Key order is insertion order, and unknown fields survive.

**Values.** Kept verbatim. A value written as a markdown link (`[title]` then `(url)`) also stores `title` and `url`.

**Homes.** Stored relative to the root, `:LINE` kept. A path outside the project is refused; a missing file only warns. Setting a value without `--home` keeps the old home.

## refs-pane

A live pane beside the chat, opened with `/refs-pane`. Tabs: `all`, then one per kind, each with its count. A detail card shows the selected ref; `o` opens its home. It polls `refs.json` every two seconds and never writes it.

| Key | Does |
|---|---|
| `h` / `l` | previous / next tab |
| `j` / `k` | move within the tab |
| `o`, Enter | open the selected ref's home |
| ←/→, ↑/↓ | tabs and rows, after you click the hint line; Esc releases |

## Development

```
go build ./...
go vet ./...
gofmt -l .                       # prints nothing
go test ./...
claude plugin test mods/refs-pane
```

Tests run inside a temp git repo, so this repo's own `.ai/refs.json` is untouched. Usage is written twice, in `help` in `main.go` and in this README; change both.

## Documentation

| I want to... | Read |
|---|---|
| see every command and rule, as an agent does | `refs --help` |
| build canon for an existing repo | [skills/harvest/SKILL.md](skills/harvest/SKILL.md) |
| tighten canon that already exists | [skills/tighten/SKILL.md](skills/tighten/SKILL.md) |
| write a canon entry by hand | [skills/harvest/formats.md](skills/harvest/formats.md) |
| know why refs is built the way it is | [docs/canon/DECISIONS.md](docs/canon/DECISIONS.md) |
| work on the code | [CLAUDE.md](CLAUDE.md) |
| install or change the pane | [mods/refs-pane/README.md](mods/refs-pane/README.md) |

## License

MIT. See [LICENSE](LICENSE).
