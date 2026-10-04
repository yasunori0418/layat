# テストドキュメント対応のスクリプト群が共有する処理。source 専用で、呼び出し側は自分の位置から解決する:
#
#   . "$(dirname "$0")/lib-testdoc.sh"          # dev/scripts/ から
#   . "$(dirname "$0")/../scripts/lib-testdoc.sh"  # dev/tests/ から

# データファイル（TSV）の実データだけを読む。行頭 # のコメントと空行を落とす。
read_tsv() { grep -v '^[[:space:]]*#' "$1" | grep -v '^[[:space:]]*$'; }

# 走査基点をリポジトリルートへ解決する。git 管理外ではカレントを返す。
testdoc_repo_root() { git rev-parse --show-toplevel 2>/dev/null || printf '.'; }

# yq が mikefarah/yq（Go 実装・v4）であることを確かめる。python-yq は別構文で eval が黙って空を返す。
# 第 1 引数は診断メッセージ用のスクリプト名。
require_yq_go() {
  local caller=${1:-yq}
  if ! yq --version 2>&1 | grep -q 'mikefarah\|version v4'; then
    echo "$caller: mikefarah/yq v4 が要る（実際: $(yq --version 2>&1)）" >&2
    return 1
  fi
}

# 必須コマンドの存在確認。第 1 引数は診断メッセージ用のスクリプト名、以降が必須コマンド。
require_commands() {
  local caller=$1
  shift
  local cmd missing=0
  for cmd in "$@"; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      echo "$caller: $cmd が要る" >&2
      missing=1
    fi
  done
  return "$missing"
}
