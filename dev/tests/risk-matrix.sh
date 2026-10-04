#!/usr/bin/env bash
# docs/risks/*.md の level が likelihood × impact のマトリクス（dev/tests/risk-matrix.tsv）と一致するかの契約テスト。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/risk-matrix.sh   # devShell / CI から直接
#   nix flake check ./dev                                       # checks.risk-matrix 経由
#
# enum の妥当性は sara check が見る。ここは 3 フィールドの関係だけを見て、値が引けなければ FAIL にする。

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

require_commands "risk-matrix.sh（nix develop '.?dir=dev#sara' から実行する）" yq || exit 1
require_yq_go risk-matrix.sh || exit 1

# 走査基点をリポジトリルートへ解決する。git 管理外ではカレントを返す。
repo_root=$(git rev-parse --show-toplevel 2>/dev/null || printf '.')

# 走査対象 docs/risks の在り処は 2 経路ある:
#   1. RISK_DOCS_DIR（checks.risk-matrix のサンドボックス。作業ツリーが無いので
#      nix が store path を渡す）
#   2. git のリポジトリルート基準（`nix develop` からの直接実行・CI の sara job）
risks_dir="${RISK_DOCS_DIR:-}"
[[ -d "$risks_dir" ]] || risks_dir="$repo_root/docs/risks"

if [[ ! -d "$risks_dir" ]]; then
  echo "risk-matrix.sh: risk item のディレクトリを解決できない（$risks_dir）" >&2
  exit 1
fi

# --- level 導出マトリクスを正本の TSV から読む -------------------------------

# TSV はこのスクリプトと同じディレクトリから読む。
matrix_tsv="$(dirname "$0")/risk-matrix.tsv"

if [[ ! -f "$matrix_tsv" ]]; then
  echo "risk-matrix.sh: マトリクスの正本が無い（$matrix_tsv）" >&2
  exit 1
fi

declare -A LEVEL_MATRIX=()
tsv_rows=0
tsv_ok=1
while IFS=$'\t' read -r likelihood impact level; do
  tsv_rows=$((tsv_rows + 1))
  if [[ -z "$likelihood" || -z "$impact" || -z "$level" ]]; then
    fault "risk-matrix.tsv の $tsv_rows 行目に空の列がある（likelihood='$likelihood' impact='$impact' level='$level'）"
    tsv_ok=0
    continue
  fi
  key="$likelihood:$impact"
  if [[ -n "${LEVEL_MATRIX[$key]:-}" ]]; then
    fault "risk-matrix.tsv で $key が 2 回現れる（既出: ${LEVEL_MATRIX[$key]}・再出: $level）"
    tsv_ok=0
    continue
  fi
  LEVEL_MATRIX["$key"]="$level"
done < <(read_tsv "$matrix_tsv")

# 3 値（high / medium / low）の直積 9 セルが揃っていることを確かめる。
if [[ "${#LEVEL_MATRIX[@]}" -ne 9 ]]; then
  fault "risk-matrix.tsv のセルが 9 件ではない（実際: ${#LEVEL_MATRIX[@]} 件・データ行 $tsv_rows 行）"
  tsv_ok=0
fi

# マトリクスを引けない状態で走査を回しても全件 FAIL するだけで診断の役に立たない。
if [[ "$tsv_ok" -eq 0 ]]; then
  exit 1
fi

pass "risk-matrix.tsv から ${#LEVEL_MATRIX[@]} セルのマトリクスを読んだ"

# --- 全 risk item の level が導出と一致する -----------------------------------

# ファイル一覧は改行区切りで受ける（ファイル名は空白・改行を含まない）。
mapfile -t risk_files < <(find "$risks_dir" -maxdepth 1 -type f -name '*.md' | LC_ALL=C sort)

# 対象 0 件を緑にしない。
if [[ "${#risk_files[@]}" -eq 0 ]]; then
  fault "docs/risks に risk item が 1 件も無い（走査先: $risks_dir）"
  exit "$fail"
fi

# 1 ファイル 1 回の yq で 3 フィールドを採る。yq の失敗（YAML の破損）と値が空は区別して報告する。
checked=0
for f in "${risk_files[@]}"; do
  name="$(basename "$f")"

  if ! fields="$(yq --front-matter=extract -r \
    '[(.likelihood // ""), (.impact // ""), (.level // "")] | @tsv' "$f" 2>/dev/null)"; then
    fault "$name: frontmatter を yq で読めなかった（YAML の構文を確認する）"
    continue
  fi

  IFS=$'\t' read -r likelihood impact level <<<"$fields"

  if [[ -z "$likelihood" || -z "$impact" || -z "$level" ]]; then
    fault "$name: frontmatter に 3 フィールドが揃っていない（likelihood='$likelihood' impact='$impact' level='$level'）"
    continue
  fi

  want="${LEVEL_MATRIX[$likelihood:$impact]:-}"
  if [[ -z "$want" ]]; then
    # enum 外の値を確かめられるよう生値を引用符で囲んで出す。
    fault "$name: likelihood='$likelihood' impact='$impact' に対応するマトリクスのセルが無い（enum 外の値）"
    continue
  fi

  if [[ "$level" != "$want" ]]; then
    fault "$name: level が導出と食い違う（likelihood=$likelihood impact=$impact → $want、実際: $level）"
    continue
  fi

  checked=$((checked + 1))
done

if [[ "$checked" -eq "${#risk_files[@]}" ]]; then
  pass "risk $checked 件の level が likelihood × impact の導出と一致する"
fi

exit "$fail"
