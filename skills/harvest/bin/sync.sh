#!/usr/bin/env bash
# sync.sh — mirror a repo's canon (and optionally open GitHub work) into refs
# Responsibilities:
#   - read docs/canon/{REQUIREMENTS,INVARIANTS,NEVERS,DECISIONS}.md (legacy fallback: docs/<NAME>.md) headings: "## INV-03 <text>", "## DEC-07 · <date> — <clause>"
#   - set one ref per live entry, home = file:line of its heading; kinds REQ- requirement, INV- invariant, NEV- never, DEC- decision
#   - remove the ref for an entry marked "Retired:" or "Superseded by"; remove refs whose entry is gone
#   - --github: one ref per open issue (issue#N) and PR (pr#N) as [title](url); drop refs for closed ones
#   - --dry-run prints the refs calls only
# Usage: sync.sh [--github] [--dry-run]   (run from anywhere inside the repo)
set -euo pipefail
[[ ${1:-} == -h ]] && { sed -n '2,9p' "$0"; exit 0; }

gh_too=0; dry=0
for a in "$@"; do case $a in --github) gh_too=1 ;; --dry-run) dry=1 ;; *) echo "unknown arg: $a" >&2; exit 2 ;; esac; done
here=$(cd "$(dirname "$0")" && pwd)
refs_bin=$(command -v refs || echo "$here/../../../bin/refs")
root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$root"

run() { if ((dry)); then printf 'DRY refs'; printf ' %q' "$@"; echo; else "$refs_bin" "$@" >/dev/null; fi; }
existing() { "$refs_bin" 2>/dev/null | awk '{print $2}' | grep -E "^$1" || true; }

for spec in REQ-:requirement:REQUIREMENTS INV-:invariant:INVARIANTS NEV-:never:NEVERS DEC-:decision:DECISIONS; do
  IFS=: read -r p kind name <<<"$spec"
  f=docs/canon/$name.md
  [[ -f $f ]] || f=docs/$name.md      # legacy layout, before docs/canon/
  [[ -f $f ]] || continue
  run kind "$p" "$kind"
  live=()
  # entries: heading line, id, text; retired when its block says so
  while IFS=$'\t' read -r line id text retired; do
    if [[ $retired == 1 ]]; then run rm "$id" 2>/dev/null || true; continue; fi
    run "$id" "$text" --kind "$kind" --home "$f:$line"; live+=("$id")
  done < <(awk -v p="$p" '
    function flush() { if (id != "") print ln "\t" id "\t" txt "\t" ret }
    /^## / { flush(); id=""; ret=0
             if (match($0, "^## " p "[0-9]+")) {
               id=substr($0, 4, RLENGTH-3); txt=substr($0, RLENGTH+2)
               if (p == "DEC-") sub(/^[^—]*— /, "", txt)   # drop "· date — "
               ln=NR } next }
    /^(Retired:|- \*\*Retired|Superseded by)/ { ret=1 }
    END { flush() }' "$f")
  for k in $(existing "$p"); do
    [[ " ${live[*]-} " == *" $k "* ]] || run rm "$k"
  done
done

if ((gh_too)); then
  run kind "issue#" issue; run kind "pr#" pr
  open_i=$(gh issue list --state open --limit 200 --json number,title,url -q '.[] | "\(.number)\t\(.title)\t\(.url)"')
  open_p=$(gh pr list --state open --limit 200 --json number,title,url -q '.[] | "\(.number)\t\(.title)\t\(.url)"')
  while IFS=$'\t' read -r n t u; do [[ -n $n ]] && run "issue#$n" "[$t]($u)" --kind issue; done <<<"$open_i"
  while IFS=$'\t' read -r n t u; do [[ -n $n ]] && run "pr#$n" "[$t]($u)" --kind pr; done <<<"$open_p"
  for k in $(existing 'issue#'); do grep -q "^${k#issue#}"$'\t' <<<"$open_i" || run rm "$k"; done
  for k in $(existing 'pr#'); do grep -q "^${k#pr#}"$'\t' <<<"$open_p" || run rm "$k"; done
fi
echo "refs: $("$refs_bin" 2>/dev/null | wc -l | tr -d " ") total" >&2
