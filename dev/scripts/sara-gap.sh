#!/usr/bin/env bash
# sara ドキュメントグラフの未カバーを 3 段で列挙する。
#
#   sara-gap [--json]
#
# exit code は 0 = ギャップなし / 1 = ギャップあり / 2 = sara check 失敗・JSON 形状異常。

set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: sara-gap [--json]

sara ドキュメントグラフの未カバーを 3 段で列挙する:
  1. threatens されていない requirement / design
  2. mitigates されていない risk
  3. covers されていない test_condition

exit code:
  0  ギャップなし
  1  ギャップあり
  2  sara check の失敗・JSON 形状の異常
EOF
}

json_out=0
case "${1-}" in
  -h | --help)
    usage
    exit 0
    ;;
  --json) json_out=1 ;;
  "") ;;
  *)
    usage
    exit 2
    ;;
esac
if [ "$#" -gt 1 ]; then
  usage
  exit 2
fi

# sara.toml のあるリポジトリルートで sara を実行する。
# SARA_GAP_ROOT はサンドボックスの契約テストが fixture を指すための seam。
repo_root=${SARA_GAP_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || printf '.')}
cd "$repo_root" || {
  echo "sara-gap: リポジトリルートへ移動できない: $repo_root" >&2
  exit 2
}

# sara 呼び出しの seam（テストが JSON 形状異常を決定論的に再現するため）。
sara_cmd=${SARA_GAP_SARA:-sara}

# グラフが invalid の間はギャップを算出しない（壊れたグラフから作る一覧は
# 信用できないため。stderr は sara の診断をそのまま通す）。
if ! graph=$("$sara_cmd" check --format json); then
  echo "sara-gap: sara check が失敗した（先にグラフを valid にすること）" >&2
  exit 2
fi

# .items 配列と valid を確認する。sara のバージョン更新で JSON 形状が変わった
# とき、黙って空の結果（＝ギャップなし）を返す事故を防ぐガード。
if ! printf '%s' "$graph" | jq -e '(.items | type) == "array"' >/dev/null; then
  echo "sara-gap: sara check の JSON に .items 配列が無い（sara の出力形状が変わった可能性）" >&2
  exit 2
fi
if ! printf '%s' "$graph" | jq -e '.valid == true' >/dev/null; then
  echo "sara-gap: sara check が invalid を報告した（先にグラフを valid にすること）" >&2
  exit 2
fi

# 宣言辺（relations.<型> の ID 配列）から張り先の集合を作り、各段の未カバーを
# 逆引きする 1 パス。relations を持たない item もあるので欠落は空配列に倒す。
# file はリポジトリルート相対で ./ が前置されるので剥がす。
gaps=$(printf '%s' "$graph" | jq '
  def targets(t): [.items[] | (.relations[t]? // [])[]] | unique;
  def row: {id, name, file: (.file | ltrimstr("./"))};
  targets("threatens") as $threatened
  | targets("mitigates") as $mitigated
  | targets("covers") as $covered
  | {
      unthreatened: ([.items[]
        | select(.type == "requirement" or .type == "design")
        | select(.id as $i | ($threatened | index($i)) | not) | row] | sort_by(.id)),
      unmitigated: ([.items[]
        | select(.type == "risk")
        | select(.id as $i | ($mitigated | index($i)) | not) | row] | sort_by(.id)),
      uncovered: ([.items[]
        | select(.type == "test_condition")
        | select(.id as $i | ($covered | index($i)) | not) | row] | sort_by(.id))
    }')

if [ "$json_out" -eq 1 ]; then
  printf '%s\n' "$gaps"
else
  printf '%s' "$gaps" | jq -r '
    def section(title; rows):
      ["## " + title]
      + (if (rows | length) == 0 then ["なし"] else [rows[] | "\(.id)\t\(.name)\t\(.file)"] end);
    section("threatens されていない requirement / design"; .unthreatened)
    + [""]
    + section("mitigates されていない risk"; .unmitigated)
    + [""]
    + section("covers されていない test_condition"; .uncovered)
    | .[]'
fi

total=$(printf '%s' "$gaps" | jq '[.[] | length] | add')
if [ "$total" -gt 0 ]; then
  exit 1
fi
