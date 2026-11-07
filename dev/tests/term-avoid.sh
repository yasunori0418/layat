#!/usr/bin/env bash
# 用語 item（ubiquitous_term）の avoid にある綴りが docs・README・CLAUDE.md の散文に無いことを確かめる契約テスト。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/term-avoid.sh   # devShell / CI から直接
#   nix flake check ./dev                                      # checks.term-avoid 経由

set -uo pipefail

pass() { printf 'ok   - %s\n' "$1"; }
fault() { printf 'FAIL - %s\n' "$1"; }

# shellcheck source=dev/scripts/lib-testdoc.sh
. "$(dirname "$0")/../scripts/lib-testdoc.sh"

require_commands "term-avoid.sh（nix develop '.?dir=dev#sara' から実行する）" sara jq git textlint find grep || exit 1

cd "$(testdoc_repo_root)" || exit 1

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# --- 辞書の生成 ---------------------------------------------------------------

if ! sara check --format json > "$work/graph.json" 2> "$work/graph.err"; then
  echo "term-avoid.sh: sara check が失敗した" >&2
  cat "$work/graph.err" >&2
  exit 1
fi
if ! jq -e '(.items | type) == "array"' "$work/graph.json" > /dev/null; then
  echo "term-avoid.sh: sara check の JSON に .items 配列が無い（sara の出力形状が変わった可能性）" >&2
  exit 1
fi

# ASCII の英数字・空白・ハイフンだけの綴りは単語境界付きの正規表現にし、行をまたぐ空白にも当てる。
jq '[.items[]
    | select(.type == "ubiquitous_term")
    | .name as $name
    | (.fields.avoid // [])[]
    | {
        expected: $name,
        pattern: (if test("^[A-Za-z0-9 -]+$")
          then "/\\b" + (gsub("(?<c>[.*+?^${}()|\\[\\]\\\\/])"; "\\\(.c)") | gsub(" +"; "\\s+")) + "\\b/"
          else . end)
      }]' "$work/graph.json" > "$work/rules.json" || {
  echo "term-avoid.sh: avoid の綴りから辞書を作れなかった" >&2
  exit 1
}

spellings=$(jq length "$work/rules.json")
if [ "$spellings" -eq 0 ]; then
  echo "term-avoid.sh: ubiquitous_term の avoid に綴りが 1 件も無い" >&2
  exit 1
fi

# YAML は JSON の上位集合なので、prh.yml は JSON で書く。
if ! jq '{version: 1, rules: .}' "$work/rules.json" > "$work/prh.yml" ||
  ! jq -n --arg prh "$work/prh.yml" '{rules: {prh: {rulePaths: [$prh]}}}' > "$work/.textlintrc.json"; then
  echo "term-avoid.sh: prh.yml / .textlintrc.json を生成できなかった" >&2
  exit 1
fi

# --- 対象ファイル -------------------------------------------------------------

{
  find docs -name '*.md' \
    -not -path 'docs/adr/*' -not -path 'docs/sara-adr/*' -not -path 'docs/ubiquitous-language/*'
  printf '%s\n' README.md README.ja.md CLAUDE.md
} | LC_ALL=C sort > "$work/targets"

mapfile -t targets < "$work/targets"
for f in "${targets[@]}"; do
  if [ ! -f "$f" ]; then
    echo "term-avoid.sh: 対象ファイルが無い: $f" >&2
    exit 1
  fi
done

# --- 検出 -------------------------------------------------------------------

textlint --config "$work/.textlintrc.json" -f unix "${targets[@]}" > "$work/lint.out" 2> "$work/lint.err"
rc=$?
# textlint は rule の読み込み失敗でも exit 1 を返すので、prh の検出行があるときだけ検出として扱う。
if [ "$rc" -eq 0 ]; then
  pass "avoid の綴りが対象の散文に無い"
  echo
  printf 'term-avoid: 全アサーション通過（対象ファイル %d 件 / 綴り %d 件）\n' "${#targets[@]}" "$spellings"
  exit 0
fi
if [ "$rc" -eq 1 ] && grep -q '\[Error/prh\]$' "$work/lint.out"; then
  fault "avoid の綴りが対象の散文にある"
  cat "$work/lint.out"
  echo
  printf 'term-avoid: 失敗あり\n' >&2
  exit 1
fi
echo "term-avoid.sh: textlint が異常終了した（exit $rc）" >&2
cat "$work/lint.out" "$work/lint.err" >&2
exit 1
