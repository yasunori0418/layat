#!/usr/bin/env bash
# sara のグラフから OKF の index.md を生成する。docs/・型ごとのディレクトリ・docs/test/ とその区分に 1 本ずつ置く。
#
#   okf-index.sh           # index.md を書き出す
#   okf-index.sh --check   # 既存の index.md と比べ、食い違うパスを出して exit 1
#
# exit code は 0 = 生成・一致 / 1 = 食い違いあり / 2 = 引数の誤り・sara check の失敗。

set -euo pipefail

usage() {
  echo "usage: okf-index.sh [--check]" >&2
}

mode=write
case "${1-}" in
  "") ;;
  --check) mode=check ;;
  -h | --help)
    usage
    exit 0
    ;;
  *)
    usage
    exit 2
    ;;
esac

# shellcheck source=dev/scripts/lib-testdoc.sh
. "$(dirname "$0")/lib-testdoc.sh"

require_commands okf-index.sh sara jq awk cmp || exit 2

cd "$(testdoc_repo_root)"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 型ごとのディレクトリ: <docs/ からのディレクトリ>|<型>|<見出し>|<docs/index.md に載せる説明>
TYPE_DIRS=(
  "solution|solution|解決策（solution）|layat が解く課題と、その解き方の全体像"
  "use-cases|use_case|ユースケース（use_case）|利用者が layat をどう使うかの場面"
  "requirements|requirement|要求（requirement）|layat が満たす振る舞いの規範"
  "design|design|設計（design）|要求をどう実現するかの設計判断"
  "quality|quality|品質（quality）|開発とリリースの品質を保つ規範"
  "test-plan|test_plan|テスト計画（test_plan）|何をどの層のテストで確かめるかの方針"
  "infrastructure|infrastructure|基盤（infrastructure）|CI・配布・ドキュメントサイトなどの基盤の構成"
  "risks|risk|リスク（risk）|要求や設計が崩れたときに起きる脅威"
  "adr|adr|ADR（adr）|layat の設計判断の記録"
  "sara-adr|sara_adr|SDR（sara_adr）|ドキュメントグラフの運用についての判断の記録"
  "ubiquitous-language|ubiquitous_term|用語（ubiquitous_term）|layat の用語の定義と避ける綴り"
)
TEST_TITLE="テスト条件とテストケース"
TEST_DESC="機能の区分ごとのテスト条件（test_condition）とテストケース（test_case）"

if ! sara check --format json > "$work/graph.json" 2> "$work/graph.err"; then
  echo "okf-index.sh: sara check が失敗した" >&2
  cat "$work/graph.err" >&2
  exit 2
fi

# item: <型>\t<docs/ からのディレクトリ>\t<ファイル名>\t<name>\t<description>\t<status>\t<リポジトリ相対パス>
jq -r '.items[]
  | (.file | sub("^\\./"; "")) as $path
  | [ .type,
      ($path | sub("^docs/"; "") | sub("/[^/]*$"; "")),
      ($path | sub("^.*/"; "")),
      .name,
      ((.description // "") | gsub("\\s*\\n\\s*"; " ")),
      (.fields.status // ""),
      $path ]
  | @tsv' "$work/graph.json" > "$work/items.tsv"

# 本文の段落: <リポジトリ相対パス>\t<H1 の次に来る最初の段落>
# 段落は見出し・箇条書き・表・引用・HTML・コードフェンスの中身・字下げ行を除いた最初の散文の塊。
cut -f7 "$work/items.tsv" | xargs -r awk '
  function flush() { if (!done && para != "") { print file "\t" para; done = 1 } }
  FNR == 1 { flush(); file = FILENAME; fm = 0; h1 = 0; fence = 0; para = ""; done = 0 }
  done { next }
  FNR == 1 && /^---[[:space:]]*$/ { fm = 1; next }
  fm == 1 { if (/^---[[:space:]]*$/) fm = 2; next }
  !h1 { if (/^# /) h1 = 1; next }
  /^[[:space:]]*(```|~~~)/ { flush(); fence = !fence; next }
  fence { next }
  /^[[:space:]]*$/ { flush(); next }
  para == "" && (/^[[:space:]]/ || /^#/ || /^([-*+]|[0-9]+\.)[[:space:]]/ || /^[|>]/ || /^</) { next }
  para != "" && (/^#/ || /^([-*+]|[0-9]+\.)[[:space:]]/ || /^[|>]/) { flush(); next }
  {
    line = $0
    sub(/^[[:space:]]+/, "", line)
    sub(/[[:space:]]+$/, "", line)
    para = (para == "") ? line : para " " line
  }
  END { flush() }
' > "$work/paras.tsv"

# エントリ: <型>\t<docs/ からのディレクトリ>\t<ファイル名>\t<index.md の 1 行>
awk -F '\t' '
  function esc(s) { gsub(/\[/, "\\[", s); gsub(/\]/, "\\]", s); return s }
  NR == FNR { para[$1] = $2; next }
  {
    desc = ($5 != "") ? $5 : para[$7]
    line = "* [" esc($4) "](./" $3 ")"
    if (desc != "") line = line " - " esc(desc)
    if (($1 == "adr" || $1 == "sara_adr") && $6 != "") line = line " (" $6 ")"
    print $1 "\t" $2 "\t" $3 "\t" line
  }
' "$work/paras.tsv" "$work/items.tsv" | LC_ALL=C sort -t $'\t' -k3,3 > "$work/entries.tsv"

gen="$work/gen"

# 指定のディレクトリ・型のエントリをファイル名順に出す。
entries_of() {
  awk -F '\t' -v dir="$1" -v type="$2" '$2 == dir && $1 == type { print $4 }' "$work/entries.tsv"
}

# 見出しとエントリから index.md を 1 本書く。エントリが無ければ見出しだけにする。
write_index() {
  local out="$gen/$1" title=$2 body=$3
  mkdir -p "$(dirname "$out")"
  {
    printf '# %s\n' "$title"
    [ -n "$body" ] && printf '\n%s\n' "$body"
  } > "$out"
  printf '%s\n' "$1" >> "$work/targets"
}

: > "$work/targets"
top=""
for row in "${TYPE_DIRS[@]}"; do
  IFS='|' read -r dir type title desc <<< "$row"
  write_index "docs/$dir/index.md" "$title" "$(entries_of "$dir" "$type")"
  top+="* [$title](./$dir/index.md) - $desc"$'\n'
done
top+="* [$TEST_TITLE](./test/index.md) - $TEST_DESC"

# docs/test/ の区分は item の置き場所から導く。
categories=$(cut -f2 "$work/entries.tsv" | awk -F / '$1 == "test" && NF == 2 { print $2 }' | LC_ALL=C sort -u)
test_list=""
for cat in $categories; do
  tc=$(entries_of "test/$cat" test_condition)
  cs=$(entries_of "test/$cat" test_case)
  body="## テスト条件（test_condition）"
  [ -n "$tc" ] && body+=$'\n\n'"$tc"
  body+=$'\n\n'"## テストケース（test_case）"
  [ -n "$cs" ] && body+=$'\n\n'"$cs"
  write_index "docs/test/$cat/index.md" "$cat" "$body"
  test_list+="* [$cat](./$cat/index.md)"$'\n'
done
write_index "docs/test/index.md" "$TEST_TITLE" "${test_list%$'\n'}"
write_index "docs/index.md" "layat のドキュメント" "$top"

if [ "$mode" = write ]; then
  while IFS= read -r path; do
    mkdir -p "$(dirname "$path")"
    cp "$gen/$path" "$path"
  done < "$work/targets"
  exit 0
fi

stale=0
while IFS= read -r path; do
  if ! cmp -s "$gen/$path" "$path"; then
    [ "$stale" -eq 0 ] && echo "okf-index.sh: 生成結果と食い違う index.md がある（okf-index.sh で再生成する）:" >&2
    printf '%s\n' "$path"
    stale=1
  fi
done < "$work/targets"
exit "$stale"
