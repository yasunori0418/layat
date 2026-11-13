#!/usr/bin/env bash
# sara のグラフから OKF の index.md を生成する。docs/・型ごとのディレクトリ・docs/test/ とその区分に 1 本ずつ置く。
#
#   okf-index.sh           # index.md を書き出す
#   okf-index.sh --check   # 既存の index.md と比べ、食い違うパスを出して exit 1
#
# exit code は 0 = 生成・一致 / 1 = 食い違いあり / 2 = 引数の誤り・sara check の失敗・表に無い section の値。

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

categories_tsv="$(cd "$(dirname "$0")" && pwd)/../tests/test-categories.tsv"
cd "$(testdoc_repo_root)"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 型ごとのディレクトリ: <docs/ からのディレクトリ>|<型>|<見出し>|<docs/index.md に載せる説明>
TYPE_DIRS=(
  "solution|solution|解決策（\`solution\`）|layat が解く課題と、その解き方の全体像"
  "use-cases|use_case|ユースケース（\`use_case\`）|利用者が layat をどう使うかの場面"
  "requirements|requirement|要件（\`requirement\`）|layat が満たす振る舞いの規範"
  "design|design|設計（\`design\`）|要件をどう実現するかの設計判断"
  "quality|quality|品質（\`quality\`）|開発とリリースの品質を保つ規範"
  "test-plan|test_plan|テスト計画（\`test_plan\`）|何をどの層のテストで確かめるかの方針"
  "infrastructure|infrastructure|基盤（\`infrastructure\`）|CI・配布・ドキュメントサイトなどの基盤の構成"
  "risks|risk|リスク（\`risk\`）|要件や設計が崩れたときに起きる脅威"
  "adr|adr|ADR（\`adr\`）|layat の設計判断の記録"
  "sara-adr|sara_adr|SDR（\`sara_adr\`）|ドキュメントグラフの運用についての判断の記録"
  "ubiquitous-language|ubiquitous_term|用語（\`ubiquitous_term\`）|layat の用語の定義と避ける綴り"
)
# index.md を section の節に分ける型の節: <型>|<section の値>|<## の見出し>。節はこの順に並べる。
SECTIONS=(
  "requirement|architecture|アーキテクチャ概要"
  "requirement|lib-api|lib API"
  "requirement|lib-mkmanifest|lib API / \`lib.mkManifest\`"
  "requirement|lib-markers|lib API / マーカー構築子"
  "requirement|lib-mkapplycommand|lib API / \`lib.mkApplyCommand\`"
  "requirement|lib-input-validation|lib API / 入力の確認"
  "requirement|cli|CLI 仕様"
  "requirement|cli-entrypoint|CLI 仕様 / entrypoint の発見・アドレッシング"
  "requirement|cli-subcommands|CLI 仕様 / サブコマンド体系"
  "requirement|cli-output|CLI 仕様 / 出力ストリームと終了コード"
  "requirement|cli-json|CLI 仕様 / outturn 準拠の \`--json\`"
  "requirement|cli-flow|CLI 仕様 / 実行フロー"
  "requirement|cli-reproducibility|CLI 仕様 / 再現性スタンス"
  "requirement|entries-schema|entries スキーマ仕様"
  "requirement|manifest-schema|manifest.json スキーマ"
  "requirement|placement-symlink|配置動作仕様 / symlink モード"
  "requirement|placement-rollback|配置動作仕様 / 途中失敗時の巻き戻し"
  "requirement|placement-copy|配置動作仕様 / copy モード・out-of-store・recopy / reset"
  "requirement|generations|世代管理仕様"
  "requirement|root-resolution|root の解決"
  "requirement|modules|モジュールオプション仕様・モジュール別動作仕様"
  "requirement|errors|エラー仕様"
  "requirement|dependencies|依存関係"
  "design|project-layout|プロジェクト構成"
  "design|layers|レイヤー構成"
  "design|flake-outputs|flake.nix outputs 設計"
  "design|core-logic|コアロジック設計（lib データ生成 + Go エンジン）"
  "design|module-integration|モジュール統合設計"
  "design|test-strategy|テスト戦略"
  "ubiquitous_term|core|Core placement abstraction"
  "ubiquitous_term|input-output|Placement input and output"
  "ubiquitous_term|modes|Placement modes"
  "ubiquitous_term|kinds|Placement kinds"
  "ubiquitous_term|state|State management"
  "ubiquitous_term|errors|Error handling"
)
TEST_TITLE="テスト条件とテストケース"
TEST_DESC="機能の区分ごとのテスト条件（\`test_condition\`）とテストケース（\`test_case\`）"

if ! sara check --format json > "$work/graph.json" 2> "$work/graph.err"; then
  echo "okf-index.sh: sara check が失敗した" >&2
  cat "$work/graph.err" >&2
  exit 2
fi

# item: <型>\t<docs/ からのディレクトリ>\t<ファイル名>\t<name>\t<description>\t<status>\t<リポジトリ相対パス>\t<section>\t<name_ja>
jq -r '.items[]
  | (.file | sub("^\\./"; "")) as $path
  | [ .type,
      ($path | sub("^docs/"; "") | sub("/[^/]*$"; "")),
      ($path | sub("^.*/"; "")),
      .name,
      ((.description // "") | gsub("\\s*\\n\\s*"; " ")),
      (.fields.status // ""),
      $path,
      (.fields.section // ""),
      (.fields.name_ja // "") ]
  | @tsv' "$work/graph.json" > "$work/items.tsv"

# 節: <型>\t<section の値>\t<見出し>
printf '%s\n' "${SECTIONS[@]}" | tr '|' '\t' > "$work/sections.tsv"

# 節に分ける型の item で、section の値が表に無いものがあれば何も書かずに止める。
unknown=$(awk -F '\t' '
  NR == FNR { known[$1 "\t" $2] = 1; sectioned[$1] = 1; next }
  ($1 in sectioned) && !(($1 "\t" $8) in known) { print $7 ": section \"" $8 "\"" }
' "$work/sections.tsv" "$work/items.tsv")
if [ -n "$unknown" ]; then
  echo "okf-index.sh: section の値が表（SECTIONS）に無い item がある:" >&2
  printf '%s\n' "$unknown" >&2
  exit 2
fi

# 本文の段落: <リポジトリ相対パス>\t<H1 の次に来る最初の段落>
# 段落は見出し・箇条書き・表・引用・HTML・コードフェンスの中身・字下げ行を除いた最初の散文の塊。
# 行は、改行の両側が ASCII のときだけ空白でつなぎ、日本語の文字が接するときは空白なしでつなぐ。
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
    if (para == "") para = line
    else if (para ~ /[ -~]$/ && line ~ /^[ -~]/) para = para " " line
    else para = para line
  }
  END { flush() }
' > "$work/paras.tsv"

# エントリ: <型>\t<docs/ からのディレクトリ>\t<ファイル名>\t<index.md の 1 行>\t<section>
# リンク文字列は name とし、name_ja を持つ用語は「name（name_ja）」にする。
# リンク文字列に化けないよう角括弧だけをエスケープし、説明はそのまま載せる。
awk -F '\t' '
  function esc(s) { gsub(/\[/, "\\[", s); gsub(/\]/, "\\]", s); return s }
  NR == FNR { para[$1] = $2; next }
  {
    desc = ($5 != "") ? $5 : para[$7]
    label = ($1 == "ubiquitous_term" && $9 != "") ? $4 "（" $9 "）" : $4
    line = "* [" esc(label) "](./" $3 ")"
    if (desc != "") line = line " - " desc
    if (($1 == "adr" || $1 == "sara_adr") && $6 != "") line = line " (" $6 ")"
    print $1 "\t" $2 "\t" $3 "\t" line "\t" $8
  }
' "$work/paras.tsv" "$work/items.tsv" | LC_ALL=C sort -t $'\t' -k3,3 > "$work/entries.tsv"

gen="$work/gen"

# 指定のディレクトリ・型のエントリをファイル名順に出す。
entries_of() {
  awk -F '\t' -v dir="$1" -v type="$2" '$2 == dir && $1 == type { print $4 }' "$work/entries.tsv"
}

# 指定のディレクトリ・型のエントリを、表の節ごとに ## の見出しの下へ並べる。エントリの無い節は出さない。
sectioned_entries_of() {
  local dir=$1 type=$2 body="" t value heading entries
  while IFS=$'\t' read -r t value heading; do
    [ "$t" = "$type" ] || continue
    entries=$(awk -F '\t' -v dir="$dir" -v type="$type" -v value="$value" \
      '$2 == dir && $1 == type && $5 == value { print $4 }' "$work/entries.tsv")
    [ -n "$entries" ] || continue
    [ -n "$body" ] && body+=$'\n\n'
    body+="## $heading"$'\n\n'"$entries"
  done < "$work/sections.tsv"
  printf '%s' "$body"
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
  if cut -f1 "$work/sections.tsv" | grep -qxF "$type"; then
    write_index "docs/$dir/index.md" "$title" "$(sectioned_entries_of "$dir" "$type")"
  else
    write_index "docs/$dir/index.md" "$title" "$(entries_of "$dir" "$type")"
  fi
  top+="* [$title](./$dir/index.md) - $desc"$'\n'
done
top+="* [$TEST_TITLE](./test/index.md) - $TEST_DESC"

# docs/test/ の区分と並び順は dev/tests/test-categories.tsv に従う。
categories=$(read_tsv "$categories_tsv" | cut -f1)
test_list=""
for cat in $categories; do
  tc=$(entries_of "test/$cat" test_condition)
  cs=$(entries_of "test/$cat" test_case)
  body="## テスト条件（\`test_condition\`）"
  [ -n "$tc" ] && body+=$'\n\n'"$tc"
  body+=$'\n\n'"## テストケース（\`test_case\`）"
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

# 生成対象から外れたのに残っている index.md も食い違いとして出す。再生成では消えないので削除を促す。
extra=0
while IFS= read -r path; do
  if ! grep -qxF "$path" "$work/targets"; then
    [ "$extra" -eq 0 ] && echo "okf-index.sh: 生成対象に無い index.md が残っている（削除する）:" >&2
    printf '%s\n' "$path"
    extra=1
  fi
done < <(find docs -name index.md -not -path 'docs/agents/*' -not -path 'docs/dev/*' | LC_ALL=C sort)

[ "$stale" -eq 0 ] && [ "$extra" -eq 0 ]
