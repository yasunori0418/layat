#!/usr/bin/env bash
# sara init で item を起票し、`<YYYYMMDD>-<フル UUID>-<slug>.md` へ改名する。
#
#   sara-new <型> <slug> <配置ディレクトリ> [-- <sara init のオプション>...]
#
# ID は sara init が採番する。ADR は対象外（`sara init adr <パス>` を直接使う）。

set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: sara-new <type> <slug> <dir> [-- <sara init options>...]

  type   sara の型名（requirement / test_case / test-case …）。ADR は対象外
  slug   ファイル名に使う短い識別子（英小文字・数字・ハイフン）
  dir    配置ディレクトリ（カレントからの相対。無ければ作る）
  --     以降を sara init へそのまま渡す（--name / --specification …）

例:
  sara-new requirement lock-ordering docs/requirements
  sara-new test-case engine-lock docs/test/atomicity -- --name "engine の lock"

出力（機械可読の 2 行）:
  id:   採番された正式 ID
  file: 起票したファイルのパス

注: ADR は連番を維持するため対象外（sara init adr <パス> を直接使う）
EOF
}

# help は引数個数に先立って処理する（`sara-new --help extra` も help になる）。
case "${1-}" in
  -h | --help)
    usage
    exit 0
    ;;
esac

if [ "$#" -lt 3 ]; then
  usage
  exit 2
fi

type=$1
slug=$2
dir=$3
shift 3

# `--` 以降は sara init への透過オプション。区切りが無ければ余分な引数を拒否する。
if [ "$#" -gt 0 ]; then
  if [ "$1" != "--" ]; then
    usage
    exit 2
  fi
  shift
fi

# ADR は連番を維持するため拒否する。
case "$type" in
  adr | ADR)
    echo "sara-new: ADR は連番を維持する（sara init adr <パス> を直接使う）" >&2
    exit 2
    ;;
esac

# slug はファイル名へそのまま入るため、起票の前に文字種を検査する。
case "$slug" in
  "" | *[!a-z0-9-]*)
    echo "sara-new: slug は英小文字・数字・ハイフンのみ（実際: '$slug'）" >&2
    exit 2
    ;;
esac

# sara init のサブコマンド名はハイフン区切り（test-case）なので、型名の `_` を `-` へ寄せる。
init_subcommand=$(printf '%s' "$type" | tr '_' '-')

mkdir -p "$dir"

# sara 呼び出しの seam（契約テストが差し替える）。
sara_cmd=${SARA_NEW_SARA:-sara}

# 採番前の仮ファイル。sara init は --name が無いと stem から name を導出するため stem を slug にする。
# 並列起動の衝突はプロセス ID を持つ隠しディレクトリで避ける。
tmp_dir="$dir/.sara-new-$$"
tmp_file="$tmp_dir/$slug.md"
cleanup() { rm -rf "$tmp_dir"; }
trap cleanup EXIT

mkdir -p "$tmp_dir"

# 装飾（色・絵文字）を落として ID 行を安定させる。stderr は素通しにする。
init_out=$("$sara_cmd" --no-color --no-emoji init "$init_subcommand" "$tmp_file" "$@")

# `  ID:   <ID>` 行から採番結果を採る。読めなければ rename せず落とす。
id=$(printf '%s\n' "$init_out" | sed -n 's/^[[:space:]]*ID:[[:space:]]*\([^[:space:]]*\).*/\1/p' | head -n 1)

if [ -z "$id" ]; then
  echo "sara-new: sara init の出力から ID を読めなかった（sara の出力形式が変わった可能性）" >&2
  printf '%s\n' "$init_out" >&2
  exit 1
fi

# ファイル名は正式 ID から `<PREFIX>-` を落とした UUID 部を使う。
uuid=${id#*-}

# 剥がした残りが UUID の形か検査する。prefix がハイフンを含む型はここで落ちる。
case "$uuid" in
  [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;;
  *)
    echo "sara-new: 採番 ID から UUID 部を取り出せなかった: $id" >&2
    echo "sara-new: prefix がハイフンを含む型は本ラッパーの対象外（ID 形式を確認すること）" >&2
    exit 1
    ;;
esac

target="$dir/$(date +%Y%m%d)-$uuid-$slug.md"

# 起票済み item を上書きしない。
if [ -e "$target" ]; then
  echo "sara-new: 出力先が既に存在する: $target" >&2
  exit 1
fi

mv "$tmp_file" "$target"

printf 'id: %s\n' "$id"
printf 'file: %s\n' "$target"
