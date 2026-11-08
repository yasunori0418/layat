#!/usr/bin/env bash
# 用語索引（docs/glossary.md）⇔ 用語 item（ubiquitous_term）の 1:1 を確かめる契約テスト。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/glossary-index.sh   # devShell / CI から直接
#   nix flake check ./dev                                          # checks.glossary-index 経由

# 1 回の実行で全失敗を報告するため -e は使わない。
set -uo pipefail

fail=0
pass() { printf 'ok   - %s\n' "$1"; }
fault() {
  printf 'FAIL - %s\n' "$1"
  fail=1
}

# shellcheck source=dev/scripts/lib-testdoc.sh
. "$(dirname "$0")/../scripts/lib-testdoc.sh"

require_commands "glossary-index.sh（nix develop '.?dir=dev#sara' から実行する）" sara jq git || exit 1

cd "$(testdoc_repo_root)" || exit 1

index=docs/glossary.md
if [ ! -f "$index" ]; then
  echo "glossary-index.sh: $index が無い" >&2
  exit 1
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# --- 入力の収集 --------------------------------------------------------------

# item 側: <ID>\t<リポジトリルート相対のファイルパス>
if ! sara check --format json > "$work/graph.json" 2> "$work/graph.err"; then
  echo "glossary-index.sh: sara check が失敗した" >&2
  cat "$work/graph.err" >&2
  exit 1
fi
jq -r '.items[] | select(.type == "ubiquitous_term") | [.id, (.file | sub("^\\./"; ""))] | @tsv' \
  "$work/graph.json" | LC_ALL=C sort > "$work/items"

if [ ! -s "$work/items" ]; then
  echo "glossary-index.sh: ubiquitous_term の item を 1 件も読めなかった" >&2
  exit 1
fi

frontmatter_id() {
  awk 'NR == 1 && $0 != "---" { exit }
    NR > 1 && $0 == "---" { exit }
    NR > 1 && /^id:/ { sub(/^id:[[:space:]]*/, ""); gsub(/"/, ""); print; exit }' "$1"
}

# 索引側: <ID>\t<リポジトリルート相対のリンク先>。索引の各行（`- [` で始まる行）のリンクは docs/ からの相対で、
# ID はリンク先の frontmatter から読む。
bad_link=0
: > "$work/links-unsorted"
while IFS= read -r target; do
  path="docs/$target"
  if [ ! -f "$path" ]; then
    fault "(3) リンク先のファイルが無い: $path"
    bad_link=1
    continue
  fi
  id=$(frontmatter_id "$path")
  if [ -z "$id" ]; then
    fault "(3) リンク先の frontmatter に id が無い: $path"
    bad_link=1
    continue
  fi
  printf '%s\t%s\n' "$id" "$path" >> "$work/links-unsorted"
done < <(sed -nE 's/^- \[(\\.|[^]\\])*\]\(([^)]*)\).*$/\2/p' "$index")
LC_ALL=C sort "$work/links-unsorted" > "$work/links"

if [ ! -s "$work/links" ]; then
  fault "入力: $index に用語 item へのリンクが 1 件も無い"
fi

cut -f1 "$work/items" > "$work/item-ids"
cut -f1 "$work/links" > "$work/link-ids"

# --- (1) 全 item が索引に 1 回ずつ載る ----------------------------------------

missing=$(LC_ALL=C comm -23 "$work/item-ids" <(LC_ALL=C sort -u "$work/link-ids"))
dup=$(uniq -d "$work/link-ids")
if [ -z "$missing" ] && [ -z "$dup" ]; then
  pass "全 ubiquitous_term が索引に 1 回ずつ載る"
else
  while IFS= read -r id; do
    [ -n "$id" ] && fault "(1) 索引に載っていない item: $id"
  done <<< "$missing"
  while IFS= read -r id; do
    [ -n "$id" ] && fault "(1) 索引に 2 回以上載っている ID: $id"
  done <<< "$dup"
fi

# --- (2) 索引の ID が全部実在する ---------------------------------------------

unknown=$(LC_ALL=C comm -13 "$work/item-ids" <(LC_ALL=C sort -u "$work/link-ids"))
if [ -z "$unknown" ]; then
  pass "索引の UBQ- ID が全部 item として実在する"
else
  while IFS= read -r id; do
    [ -n "$id" ] && fault "(2) item が実在しない ID: $id"
  done <<< "$unknown"
fi

# --- (3) リンク先のファイルが実在し、その ID の item である ------------------

while IFS=$'\t' read -r id path; do
  expected=$(awk -F '\t' -v id="$id" '$1 == id { print $2 }' "$work/items")
  if [ -n "$expected" ] && [ "$expected" != "$path" ]; then
    fault "(3) リンク先が別の item のファイル: $id → $path（正: $expected）"
    bad_link=1
  fi
done < "$work/links"
if [ "$bad_link" -eq 0 ] && [ -s "$work/links" ]; then
  pass "索引のリンク先が全部実在し、その ID の item のファイルを指す"
fi

# --- 結果 -------------------------------------------------------------------

echo
if [ "$fail" -eq 0 ]; then
  printf 'glossary-index: 全アサーション通過（item %d 件 / 索引の行 %d 件）\n' \
    "$(wc -l < "$work/items")" "$(wc -l < "$work/links")"
  exit 0
fi

printf 'glossary-index: 失敗あり\n' >&2
exit 1
