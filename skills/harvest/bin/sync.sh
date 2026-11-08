#!/usr/bin/env bash
# sync.sh — mirror a repo's canon (and optionally open GitHub work) into refs
# Responsibilities:
#   - read docs/canon/{REQUIREMENTS,TRUTHS,NEVERS,DECISIONS}.md (legacy fallbacks: docs/<NAME>.md; INVARIANTS.md with INV- ids, kind invariant) headings: "## TRU-03 <text>", "## DEC-07 · <date> — <clause>"
#   - set one ref per live entry, home = file:line of its heading; kinds REQ- requirement, TRU- truth, NEV- never, DEC- decision
#   - remove the ref for an entry marked "Retired:" or "Superseded by"; remove refs whose entry is gone
#   - --github: one ref per open issue (issue#N) and PR (pr#N) as [title](url); drop refs for closed ones
#   - --cactus: one ref per answered cactus record (.ai/cactus/qN-*.md) as "<question> → <answer>", kind answer, home = the record, line 1; skips skipped answers and rows tagged with a pr and rows in thread denied; drops refs that no longer qualify; warns and skips when cactus is not on PATH
#   - --dry-run prints the refs calls only
# Usage: sync.sh [--github] [--cactus] [--dry-run]   (run from anywhere inside the repo)
set -euo pipefail
[[ ${1:-} == -h ]] && { sed -n '2,11p' "$0"; exit 0; }

gh_too=0; cactus_too=0; dry=0
for a in "$@"; do case $a in --github) gh_too=1 ;; --cactus) cactus_too=1 ;; --dry-run) dry=1 ;; *) echo "unknown arg: $a" >&2; exit 2 ;; esac; done
here=$(cd "$(dirname "$0")" && pwd)
refs_bin=$(command -v refs || echo "$here/../../../bin/refs")
root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$root"

run() { if ((dry)); then printf 'DRY refs'; printf ' %q' "$@"; echo; else "$refs_bin" "$@" >/dev/null; fi; }
existing() { "$refs_bin" 2>/dev/null | awk '{print $2}' | grep -E "^$1" || true; }

# prefix:kind:NAME[:legacy prefix:legacy kind:legacy NAME]
for spec in REQ-:requirement:REQUIREMENTS TRU-:truth:TRUTHS:INV-:invariant:INVARIANTS NEV-:never:NEVERS DEC-:decision:DECISIONS; do
  IFS=: read -r p kind name lp lkind lname <<<"$spec"
  f=docs/canon/$name.md
  [[ -f $f ]] || f=docs/$name.md      # legacy layout, before docs/canon/
  if [[ ! -f $f && -n ${lname:-} ]]; then   # legacy name, before the rename
    p=$lp; kind=$lkind; f=docs/canon/$lname.md; [[ -f $f ]] || f=docs/$lname.md
  fi
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
if ((cactus_too)); then
  if ! command -v cactus >/dev/null; then
    echo "sync.sh: cactus not on PATH; skipping --cactus" >&2
  else
    # records: .ai/cactus/qN-<id>-<slug>.md; the DB row decides if one qualifies
    recs=(); keys=()
    for f in .ai/cactus/q[0-9]*-*.md; do
      [[ -f $f ]] || continue
      recs+=("$f"); keys+=("$(basename "$f" | sed -E 's/^(q[0-9]+)-.*/\1/')")
    done
    rows=""
    if ((${#keys[@]})); then
      rows=$(cactus get "${keys[@]}" --json 2>/dev/null) || {   # one unknown key fails the batch
        rows=$(for k in "${keys[@]}"; do cactus get "$k" --json 2>/dev/null || true; done | jq -s 'add // []')
      }
    fi
    # key, question → answer for rows answered, not skipped, without a pr tag, not in thread denied (auto-posted blocked commands)
    qual=$(jq -r '.[] | select(.answer != null and (.answer.skipped | not) and .pr == null and .thread != "denied")
      | (((.answer.selected // []) | join(", ")) as $s | (.answer.text // "") as $t
         | if $s == "" then $t elif $t == "" then $s else "\($s) — \($t)" end) as $a
      | select($a != "")
      | [.key, ((.text + " → " + $a) | gsub("[\r\n\t]+"; " "))] | @tsv' <<<"${rows:-[]}")
    run kind q answer
    live=()
    for i in "${!keys[@]}"; do
      k=${keys[$i]}
      v=$(awk -F'\t' -v k="$k" '$1 == k { print substr($0, length(k) + 2) }' <<<"$qual")
      [[ -n $v ]] || continue
      run "$k" "$v" --kind answer --home "${recs[$i]}:1"; live+=("$k")
    done
    for k in $(existing 'q' | grep -E '^q[0-9]+$' || true); do
      [[ " ${live[*]-} " == *" $k "* ]] || run rm "$k"
    done
  fi
fi
echo "refs: $("$refs_bin" 2>/dev/null | wc -l | tr -d " ") total" >&2
