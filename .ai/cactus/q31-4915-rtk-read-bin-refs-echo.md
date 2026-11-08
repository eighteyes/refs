# q31 — rtk read bin/refs && echo ---SYNC && rtk read skills/harvest/bin/sync.sh

status: answered
act: run
kind: confirm
thread: denied
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-08T07:35:31.545381+00:00

## Context

Read launcher script and sync.sh
(denied by permission mode)

## Options

- approve
- deny

## Answer

approve
answered at: 2026-11-08T07:36:12.494814+00:00

## Result

exit: 0

```
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
— exit 0 —
```

log: /var/folders/vy/jl7086td5nv4_6ms16c6g1rr0000gn/T/cactus-q31-3576993452.log
