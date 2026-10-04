#!/usr/bin/env bash
# テスト資産を列挙する。--static はファイル走査だけ、--full は go test と nix eval を呼ぶ。
#
#   dev/scripts/test-inventory.sh --static   # ファイル粒度（契約テストが使う）
#   dev/scripts/test-inventory.sh --full     # テスト名粒度（対応表生成が使う）
#
# 出力（TSV・資産識別子でソート）:
#   --static: <資産識別子>\t<種別>
#   --full:   <資産識別子>\t<種別>\t<テスト名>
#             テスト名を持たない種別（e2e / namaka / flake-check）は 3 列目が空。
#
# 資産識別子は CASE frontmatter の target と同じ表記:
#   - 実在するファイル → リポジトリ相対パス（namaka だけは末尾スラッシュのディレクトリ）
#   - flake check      → checks.<name> 形式（dev flake は dev:checks.<name>）

set -uo pipefail

# shellcheck source=dev/scripts/lib-testdoc.sh
. "$(dirname "$0")/lib-testdoc.sh"

usage() {
  cat >&2 <<'EOF'
usage: test-inventory.sh (--static | --full | --module-generated-checks)

  --static  ファイル粒度で列挙する（fd / glob のみ。go test も nix eval も呼ばない）
  --full    テスト名粒度で列挙する（go test -json + nix eval を呼ぶため重い）
  --module-generated-checks
            flakeModule 由来（flake ファイルに定義行を持たない）check だけを列挙する。
            契約テストが flake ⟷ 静的リストの grep 突合から除くために使う

出力は TSV。詳細はスクリプト冒頭のコメントを参照。
EOF
}

# 引数個数は mode 判定より先に見る。
if [ "$#" -ne 1 ]; then
  usage
  exit 2
fi

case "$1" in
  -h | --help)
    usage
    exit 0
    ;;
  --static) mode=static ;;
  --full) mode=full ;;
  --module-generated-checks) mode=module-generated-checks ;;
  *)
    usage
    exit 2
    ;;
esac

# 走査はリポジトリルート基準で行う。
cd "$(testdoc_repo_root)" || exit 1

# flake check の静的リスト（`nix eval` を避けるため列挙を持つ）。
# flake.nix / dev/flake.nix の checks 定義との突合は dev/tests/test-doc-map.sh の §5 が行う。
FLAKE_CHECKS=(
  checks.go-vet
  checks.golangci-lint
  checks.hm-module
  checks.namaka
  checks.nix-unit
  checks.layat
  checks.treefmt
  dev:checks.risk-matrix
  dev:checks.sara-gap
  dev:checks.sara-new
  dev:checks.test-doc-map
)

# flakeModule 由来（flake ファイルに `checks.<name> =` の定義行が無い）check。
# nix-unit は nix-unit の flakeModule、treefmt は treefmt-nix の flakeModule が生む。
# 契約テストの §5 はこの 2 件を grep 突合から除く。
MODULE_GENERATED_CHECKS=(
  checks.nix-unit
  checks.treefmt
)

# --- 列挙（ファイル粒度） ----------------------------------------------------

# sara devShell には fd が無いため find を使う。

list_go_files() {
  find cmd internal -name '*_test.go' -type f 2>/dev/null | sed 's|^\./||'
}

list_nix_unit_files() {
  find tests/nix-unit -maxdepth 1 -name '*.nix' -type f 2>/dev/null | sed 's|^\./||'
}

# namaka の CASE 粒度はディレクトリ（末尾スラッシュ）。`_` 始まりの内部ディレクトリは除く。
list_namaka_dirs() {
  find tests/namaka -mindepth 1 -maxdepth 1 -type d -not -name '_*' 2>/dev/null |
    sed 's|^\./||' |
    sed 's|$|/|'
}

list_e2e_files() {
  find tests/e2e/scenarios -name '*.sh' -type f 2>/dev/null | sed 's|^\./||'
}

emit_static() {
  list_go_files | sed 's|$|\tgo|'
  list_nix_unit_files | sed 's|$|\tnix-unit|'
  list_namaka_dirs | sed 's|$|\tnamaka|'
  list_e2e_files | sed 's|$|\te2e|'
  printf '%s\tflake-check\n' "${FLAKE_CHECKS[@]}"
}

if [ "$mode" = module-generated-checks ]; then
  printf '%s\n' "${MODULE_GENERATED_CHECKS[@]}" | LC_ALL=C sort
  exit 0
fi

if [ "$mode" = static ]; then
  emit_static | LC_ALL=C sort
  exit 0
fi

# --- 列挙（テスト名粒度・--full のみ） ---------------------------------------

require_commands "test-inventory.sh --full（nix develop ./dev から実行する）" go nix jq || exit 1

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Go: `go test -json` の run イベントからテスト名を採る（サブテスト込み）。
# テストの合否は見ない。
go test -json ./... 2>/dev/null > "$work/go-json"

jq -r 'select(.Action == "run" and .Test != null) | .Test' "$work/go-json" |
  LC_ALL=C sort -u > "$work/go-names"

if [ ! -s "$work/go-names" ]; then
  echo "test-inventory.sh: go test -json からテスト名を採れなかった" >&2
  exit 1
fi

# テストを 1 件も走らせずに fail したパッケージ（ビルド失敗・TestMain / init の異常終了）があれば落とす。
jq -r 'select(.Action == "run" and .Package != null) | .Package' "$work/go-json" |
  LC_ALL=C sort -u > "$work/go-ran-packages"
jq -r 'select(.Action == "fail" and .Package != null and .Test == null) | .Package' "$work/go-json" |
  LC_ALL=C sort -u > "$work/go-failed-packages"

unbuilt=$(LC_ALL=C comm -13 "$work/go-ran-packages" "$work/go-failed-packages")
if [ -n "$unbuilt" ]; then
  echo "test-inventory.sh: テストを 1 件も走らせずに失敗したパッケージがある" >&2
  echo "  （ビルド失敗、または TestMain / init の異常終了）:" >&2
  printf '%s\n' "$unbuilt" | sed 's/^/  /' >&2
  echo "test-inventory.sh: 列挙が不完全なので中断する（go test <パッケージ> で原因を確認する）" >&2
  exit 1
fi

# テスト名 → ファイルの帰属表を `^func Test` の grep で作る。同名関数は両ファイルへ出す。
: > "$work/func-map"
while IFS= read -r file; do
  grep -oE '^func (Test[A-Za-z0-9_]*)' "$file" |
    sed 's/^func //' |
    while IFS= read -r fn; do
      printf '%s\t%s\n' "$fn" "$file" >> "$work/func-map"
    done
done < <(list_go_files)

# サブテスト（Parent/Sub）は親のファイルへ寄せる。
while IFS= read -r name; do
  toplevel=${name%%/*}
  files=$(awk -F'\t' -v fn="$toplevel" '$1 == fn { print $2 }' "$work/func-map")
  if [ -z "$files" ]; then
    # 帰属先不明は警告して飛ばす。
    echo "test-inventory.sh: 警告: テスト $name の定義ファイルを特定できなかった" >&2
    continue
  fi
  while IFS= read -r file; do
    printf '%s\tgo\t%s\n' "$file" "$name"
  done <<< "$files"
done < "$work/go-names" > "$work/go-rows"

# nix-unit: 各ファイルを `{ lib, layat }` で直接 import して attrNames を採る。
# getFlake のため --impure が要る。
: > "$work/nix-unit-rows"
while IFS= read -r file; do
  names=$(nix eval --impure --json --expr "
    let
      flake = builtins.getFlake (builtins.toString ./.);
      lib = flake.inputs.nixpkgs.lib;
      layat = import ./lib;
    in
    builtins.attrNames (import ./$file { inherit lib layat; })
  " 2>/dev/null | jq -r '.[]')
  if [ -z "$names" ]; then
    echo "test-inventory.sh: 警告: $file から nix-unit のテスト名を採れなかった" >&2
    printf '%s\tnix-unit\t\n' "$file" >> "$work/nix-unit-rows"
    continue
  fi
  while IFS= read -r name; do
    printf '%s\tnix-unit\t%s\n' "$file" "$name" >> "$work/nix-unit-rows"
  done <<< "$names"
done < <(list_nix_unit_files)

# e2e / namaka / flake check は static と同粒度（テスト名の内訳を持たない）。
{
  cat "$work/go-rows"
  cat "$work/nix-unit-rows"
  list_namaka_dirs | sed 's|$|\tnamaka\t|'
  list_e2e_files | sed 's|$|\te2e\t|'
  printf '%s\tflake-check\t\n' "${FLAKE_CHECKS[@]}"
} | LC_ALL=C sort
