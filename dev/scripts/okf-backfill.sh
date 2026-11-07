#!/usr/bin/env bash
# docs/ の item に、無ければ generated（git の履歴から）と status: stable を足す。冪等。
#
#   dev/scripts/okf-backfill.sh --default-by <actor>

set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: okf-backfill.sh --default-by <actor>

  --default-by  本文を最後に変えたコミットに Co-Authored-By が無いときの generated.by
                （例: human:<GitHub ID> / claude-code/<モデル ID>）

docs/ 配下の item（先頭の frontmatter に type: を持つ .md）を書き換える:
  generated  無ければ frontmatter の末尾に {by, at} を足す。本文（frontmatter の外）を
             最後に変えたコミットの Co-Authored-By の表示名から by を、author date から at を取る
  status     adr / sara_adr 以外の型で無ければ、name: の直後に status: stable を足す
EOF
}

default_by=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --default-by)
      [ "$#" -ge 2 ] || {
        usage
        exit 2
      }
      default_by=$2
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

case "$default_by" in
  "")
    usage
    exit 2
    ;;
  *'"'* | *\\* | *$'\n'*)
    echo "okf-backfill.sh: --default-by に二重引用符・バックスラッシュ・改行は使えない（実際: '$default_by'）" >&2
    exit 2
    ;;
esac

cd "$(git rev-parse --show-toplevel)"

# 標準入力の本文（frontmatter の外）を出す。frontmatter が無ければ全体を本文とみなす。
body_of() {
  awk '
    NR == 1 && $0 == "---" { fm = 1; next }
    NR == 1 { body = 1 }
    fm && $0 == "---" { fm = 0; body = 1; next }
    body { print }
  '
}

# Co-Authored-By の値（`Claude Opus 5 (1M context) <…>`）を actor（`claude-code/claude-opus-5`）にする。
actor_of() {
  local name=${1%%<*}
  name=$(printf '%s' "$name" | sed -E 's/\([^)]*\)//g' | tr '[:upper:]' '[:lower:]' |
    sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')
  printf 'claude-code/%s' "$name"
}

# 本文を最後に変えたコミットから generated の by と at を求め、`<by>\t<at>` を出す。
# 改名を跨いで辿るため、各コミットの旧版は 1 つ古いログ項目のパスで読む。
generated_of() {
  local file=$1 log hashes=() paths=() i h trailer
  log=$(git log --follow -M --format='%x01%H' --name-only -- "$file")
  while IFS= read -r line; do
    case "$line" in
      $'\x01'*) hashes+=("${line#$'\x01'}") ;;
      "") ;;
      *) paths+=("$line") ;;
    esac
  done <<<"$log"
  if [ "${#hashes[@]}" -eq 0 ] || [ "${#hashes[@]}" -ne "${#paths[@]}" ]; then
    echo "okf-backfill.sh: $file の履歴を git log から読めなかった（未コミットのファイルは対象外）" >&2
    return 1
  fi
  for ((i = 0; i < ${#hashes[@]}; i++)); do
    h=${hashes[i]}
    if ((i + 1 < ${#hashes[@]})) &&
      [ "$(git show "$h:${paths[i]}" | body_of)" = "$(git show "$h^:${paths[i + 1]}" 2>/dev/null | body_of)" ]; then
      continue
    fi
    trailer=$(git log -1 --format='%(trailers:key=Co-Authored-By,valueonly,separator=%x0a)' "$h" | sed -n '1p')
    if [ -n "$trailer" ]; then
      printf '%s\t%s\n' "$(actor_of "$trailer")" "$(git log -1 --format=%aI "$h")"
    else
      printf '%s\t%s\n' "$default_by" "$(git log -1 --format=%aI "$h")"
    fi
    return 0
  done
  echo "okf-backfill.sh: $file の本文を変えたコミットが見つからなかった" >&2
  return 1
}

# frontmatter から `<type>\t<generated の有無>\t<status の有無>` を出す。item でなければ何も出さない。
inspect() {
  awk '
    FNR == 1 { if ($0 != "---") exit; next }
    $0 == "---" { if (type != "") printf "%s\t%d\t%d\n", type, gen, st; exit }
    /^type:/ { type = $2; gsub(/"/, "", type) }
    /^generated:/ { gen = 1 }
    /^status:/ { st = 1 }
  ' "$1"
}

added_generated=0
added_status=0
while IFS= read -r file; do
  meta=$(inspect "$file")
  [ -n "$meta" ] || continue
  IFS=$'\t' read -r type has_generated has_status <<<"$meta"
  need_status=0
  if [ "$has_status" -eq 0 ] && [ "$type" != adr ] && [ "$type" != sara_adr ]; then
    need_status=1
  fi
  by='' at=''
  if [ "$has_generated" -eq 0 ]; then
    generated=$(generated_of "$file") || exit 1
    IFS=$'\t' read -r by at <<<"$generated"
  fi
  [ -n "$by" ] || [ "$need_status" -eq 1 ] || continue

  tmp=$(mktemp)
  awk -v by="$by" -v at="$at" -v need_status="$need_status" '
    NR == 1 { fm = 1; print; next }
    fm && $0 == "---" {
      if (by != "") { print "generated:"; print "  by: \"" by "\""; print "  at: \"" at "\"" }
      fm = 0
    }
    { print }
    fm && need_status && /^name:/ { print "status: stable"; need_status = 0 }
    END { if (need_status) exit 1 }
  ' "$file" >"$tmp" || {
    echo "okf-backfill.sh: $file の frontmatter に name: が無く、status を足せなかった" >&2
    rm -f "$tmp"
    exit 1
  }
  cat "$tmp" >"$file"
  rm -f "$tmp"
  if [ -n "$by" ]; then added_generated=$((added_generated + 1)); fi
  if [ "$need_status" -eq 1 ]; then added_status=$((added_status + 1)); fi
done < <(find docs -name '*.md' -type f | LC_ALL=C sort)

echo "okf-backfill.sh: generated を $added_generated 件、status を $added_status 件足した" >&2
