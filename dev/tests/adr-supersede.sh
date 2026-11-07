#!/usr/bin/env bash
# supersedes が指す ADR が status: deprecated かを frontmatter だけから確かめる。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/adr-supersede.sh
set -uo pipefail

fail=0
pass() { printf 'ok   - %s\n' "$1"; }
fault() {
  printf 'FAIL - %s\n' "$1"
  fail=1
}

docs_dir="$(dirname "$0")/../../docs"
if [[ ! -d "$docs_dir" ]]; then
  fault "docs ディレクトリが無い（$docs_dir）"
  exit 1
fi

# frontmatter（先頭の --- から次の --- まで）だけを読み、ADR 型の item を
# `item <id> <status>` と `edge <id> <supersedes 先>` の行で出す。
mapfile -t records < <(
  find "$docs_dir" -name '*.md' -type f -exec awk '
    function unquote(v) { gsub(/^[[:space:]]+|[[:space:]]+$/, "", v); gsub(/^["\047]|["\047]$/, "", v); return v }
    function flush() {
      if (is_adr && id != "") {
        print "item", id, (status == "" ? "-" : status)
        for (i = 1; i <= n; i++) print "edge", id, sup[i]
      }
    }
    FNR == 1 { in_fm = ($0 == "---"); id = ""; status = ""; is_adr = 0; key = ""; n = 0; next }
    !in_fm { next }
    $0 == "---" { flush(); in_fm = 0; next }
    /^[A-Za-z_]+:/ {
      key = $0; sub(/:.*/, "", key)
      val = $0; sub(/^[A-Za-z_]+:/, "", val); val = unquote(val)
      if (key == "id") id = val
      else if (key == "status") status = val
      else if (key == "type") is_adr = (val == "adr" || val == "sara_adr")
      next
    }
    key == "supersedes" && /^[[:space:]]+-[[:space:]]/ { v = $0; sub(/^[[:space:]]+-/, "", v); sup[++n] = unquote(v) }
  ' {} +
)

declare -A status_of=()
edges=()
for rec in "${records[@]}"; do
  read -r kind from to <<<"$rec"
  if [[ "$kind" == item ]]; then
    status_of["$from"]="$to"
  else
    edges+=("$from $to")
  fi
done

if [[ "${#status_of[@]}" -gt 0 ]]; then
  pass "ADR を frontmatter から読む（${#status_of[@]} 本・supersedes ${#edges[@]} 辺）"
else
  fault "ADR を frontmatter から読む（$docs_dir に ADR が見つからない）"
fi

# supersedes で指された ADR は deprecated である。
bad_targets=0
for edge in "${edges[@]}"; do
  read -r from to <<<"$edge"
  if [[ "${status_of[$to]:-}" != deprecated ]]; then
    fault "$from が supersedes で指す $to が status: deprecated でない（実際: ${status_of[$to]:-ADR が見つからない}）"
    bad_targets=1
  fi
done
[[ "$bad_targets" -eq 0 ]] && pass "supersedes で指された ADR は全て status: deprecated"

exit "$fail"
