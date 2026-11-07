# refs

A per-project key-value store for the pillars of an agent session:
requirements, PRs, decisions. Agents read it on demand.

One file, [`bin/refs`](bin/refs). python3, stdlib only.

	refs                              list
	refs KEY                          show one
	refs KEY VALUE... [--kind NAME] [--home PATH[:LINE]]   set / overwrite
	refs rm KEY
	refs kind PREFIX NAME / refs kinds
	refs home KEY PATH[:LINE] / refs home KEY --clear
	refs hook                         the hook itself
	refs --help                       usage, written for a model

## Where it lives

`<root>/.ai/refs.json`, `{"kinds": {}, "refs": {}}`. Root is the git toplevel
of the current directory, else the directory itself. Writes are atomic.

## Keys and kinds

The key prefix is the kind: `R20` is `R`, `pr#131` is `pr#`. Keys start with a
non-digit, never `-`. Reserved: `rm kind kinds home install hook`.

An unknown prefix saves as kind `unknown` and nags. Name it with `--kind`, or
`refs kind PREFIX NAME`, which also relabels existing unknown refs.

A value of the form `[title](url)` is parsed into title and url. The value is
kept verbatim.

## Home

A ref may point at its source-of-truth file. Stored root-relative, `:LINE`
kept. Outside the project is refused. A missing file warns. Setting a value
without `--home` keeps the old home.

## Install

Add it as a Claude Code plugin from this directory. The plugin registers a
SessionStart hook (`startup|resume|clear|compact`) that runs `refs hook`.

In a project with no `.ai/refs.json`, the hook tells the agent to populate
the refs first. Otherwise it prints a one-line reminder and the ref count.
