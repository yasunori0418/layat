#!/usr/bin/env bash
# テストコード ⇔ テストドキュメント（CASE）対応の契約テスト。go test も nix eval も呼ばない。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/test-doc-map.sh   # devShell / CI から直接
#   nix flake check ./dev                                        # checks.test-doc-map 経由
#
# 検証対象。番号は下の節見出しに対応する:
#   1.  順方向 — 全 CASE の target が実在するテスト資産を指す
#   2.  逆方向 — 全テスト資産に CASE がある（除外リストにあるものを除く）
#   3.  1:1 一意性 — 1 資産に 2 つ以上の CASE が張られていない
#   4.  データファイルの健全性 — 区分表が docs/test/ のディレクトリと一致し、
#       除外リストが実在する資産だけを挙げている
#   5.  静的リストの健全性 — test-inventory.sh の FLAKE_CHECKS が flake ファイルの
#       checks 定義と一致する（片側更新漏れの検出）
#   6.  自己検証 — §1〜§3 が呼ぶ judge_* と、その結果を pass / fault へ振り分ける
#       run_judge を合成フィクスチャへ当て、期待どおりに振る舞うことを確かめる

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

require_commands "test-doc-map.sh（nix develop '.?dir=dev#sara' から実行する）" yq git || exit 1
require_yq_go test-doc-map.sh || exit 1

cd "$(testdoc_repo_root)" || exit 1

inventory_sh=dev/scripts/test-inventory.sh
categories_tsv=dev/tests/test-categories.tsv
exclusions_tsv=dev/tests/test-doc-exclusions.tsv

for f in "$inventory_sh" "$categories_tsv" "$exclusions_tsv"; do
  if [ ! -f "$f" ]; then
    echo "test-doc-map.sh: $f が無い" >&2
    exit 1
  fi
done

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# --- 入力の収集 --------------------------------------------------------------

bash "$inventory_sh" --static > "$work/inventory" || {
  echo "test-doc-map.sh: $inventory_sh --static が失敗した" >&2
  exit 1
}
cut -f1 "$work/inventory" | LC_ALL=C sort > "$work/assets"

# 逆方向（§2）の走査対象。§4 の stale 除外検査も同じ変数を読む。
reverse_scan=$work/assets

if [ ! -s "$work/assets" ]; then
  echo "test-doc-map.sh: テスト資産を 1 件も列挙できなかった" >&2
  exit 1
fi

# CASE の <ファイルパス>\t<target> 表。frontmatter の target を yq で読む。
# yq の失敗（frontmatter の破損）と target が空は区別して報告する。
: > "$work/case-targets"
unreadable=0
while IFS= read -r file; do
  if ! target=$(yq --front-matter=extract '.target // ""' "$file" 2>/dev/null); then
    fault "入力: $file の frontmatter を yq で読めなかった（YAML の構文を確認する）"
    unreadable=1
    continue
  fi
  printf '%s\t%s\n' "$file" "$target" >> "$work/case-targets"
done < <(grep -rl '^type: test_case' docs/test/ | LC_ALL=C sort)

if [ ! -s "$work/case-targets" ]; then
  echo "test-doc-map.sh: CASE を 1 件も読めなかった" >&2
  exit 1
fi

if [ "$unreadable" -eq 0 ]; then
  pass "全 CASE の frontmatter を読める"
fi

read_tsv "$exclusions_tsv" | cut -f1 | LC_ALL=C sort > "$work/exclusions"

# --- 判定（§1〜§3 の本体。§6 が同じ関数を合成フィクスチャで叩く） --------------
#
# 各関数は診断行を stdout へ出し、違反があれば 1 を返す。呼び出し側が pass / fault へ振り分ける。

# 順方向: case_targets（<ファイル>\t<target>）の target が assets に在るか。
judge_forward() {
  local case_targets=$1 assets=$2
  local violated=0 file target
  while IFS=$'\t' read -r file target; do
    if [ -z "$target" ]; then
      echo "順方向: $file に target が無い（docs/model.yaml で required なので sara check も落ちる）"
      violated=1
      continue
    fi
    if ! grep -qxF "$target" "$assets"; then
      echo "順方向: $file の target '$target' に対応するテスト資産が無い（リネーム / 削除の追従漏れ）"
      violated=1
    fi
  done < "$case_targets"
  return "$violated"
}

# 逆方向: scan_target の各要素が covered か exclusions のどちらかに在るか。
# 第 1 引数は「走査対象」、第 2・第 3 は「そこに在れば違反を消す集合」。
# 第 4 引数は診断メッセージに載せる除外リストのパス（案内文にのみ使う）。
judge_reverse() {
  local scan_target=$1 covered=$2 exclusions=$3 exclusions_hint=${4:-$3}
  local violated=0 asset

  # 免除集合（covered / exclusions）が走査対象の部分集合であることを確かめる。
  # 違反でも early return せず、未カバー資産の走査を続ける。
  local stray
  stray=$(LC_ALL=C comm -23 <(LC_ALL=C sort -u "$covered" "$exclusions") \
    <(LC_ALL=C sort -u "$scan_target"))
  if [ -n "$stray" ]; then
    echo "逆方向: 免除集合が走査対象の部分集合でない（引数配線の誤り、または CASE / 除外リストが走査対象へ追従していない。走査対象に無い: $(printf '%s' "$stray" | tr '\n' ' ')）"
    violated=1
  fi

  while IFS= read -r asset; do
    if grep -qxF "$asset" "$covered"; then
      continue
    fi
    if grep -qxF "$asset" "$exclusions"; then
      continue
    fi
    echo "逆方向: テスト資産 '$asset' に CASE が無い（CASE を起こすか $exclusions_hint へ除外理由付きで追加する）"
    violated=1
  done < "$scan_target"
  return "$violated"
}

# 1:1: case_targets の target に重複が無いか。
judge_unique() {
  local case_targets=$1
  local dups target owners
  dups=$(cut -f2 "$case_targets" | grep -v '^$' | LC_ALL=C sort | uniq -d)
  if [ -z "$dups" ]; then
    return 0
  fi
  while IFS= read -r target; do
    owners=$(awk -F'\t' -v t="$target" '$2 == t { printf "%s ", $1 }' "$case_targets")
    echo "1:1 違反: '$target' に複数の CASE が張られている（$owners）"
  done <<< "$dups"
  return 1
}

# 判定関数を実データへ当て、診断行を fault へ流す。§6 がこの関数自体も検証する。
run_judge() {
  local ok_message=$1
  shift
  local diagnostics emitted=0 line
  if diagnostics=$("$@"); then
    pass "$ok_message"
    return 0
  fi
  while IFS= read -r line; do
    if [ -n "$line" ]; then
      fault "$line"
      emitted=1
    fi
  done <<< "$diagnostics"
  # 非ゼロで診断行が無いときも必ず fault を 1 件出す。
  if [ "$emitted" -eq 0 ]; then
    fault "$ok_message — 判定が違反を返したが診断行が空（判定側の実装漏れ）"
  fi
  return 1
}

# --- 1. 順方向（CASE の target が実在する） ----------------------------------

run_judge "全 CASE の target が実在するテスト資産を指す" \
  judge_forward "$work/case-targets" "$work/assets"

# --- 2. 逆方向（全テスト資産に CASE がある） ---------------------------------

cut -f2 "$work/case-targets" | grep -v '^$' | LC_ALL=C sort -u > "$work/covered"

run_judge "全テスト資産に CASE がある（除外リストを除く）" \
  judge_reverse "$reverse_scan" "$work/covered" "$work/exclusions" "$exclusions_tsv"

# --- 3. 1:1 一意性 -----------------------------------------------------------

# 順方向の 1:1 は target が単一 text であることで型が担保する。ここは逆方向の重複だけを見る。
run_judge "1 テスト資産に張られた CASE は高々 1 件（1:1）" \
  judge_unique "$work/case-targets"

# --- 4. データファイルの健全性 -----------------------------------------------

# 区分表 ⟷ docs/test/ のディレクトリ。
read_tsv "$categories_tsv" | cut -f1 | LC_ALL=C sort > "$work/categories"
find docs/test -mindepth 1 -maxdepth 1 -type d 2>/dev/null |
  sed 's|.*/||' |
  LC_ALL=C sort > "$work/case-dirs"

if diff -q "$work/categories" "$work/case-dirs" >/dev/null 2>&1; then
  pass "区分表が docs/test/ のディレクトリと一致する"
else
  only_table=$(comm -23 "$work/categories" "$work/case-dirs" | tr '\n' ' ')
  only_dirs=$(comm -13 "$work/categories" "$work/case-dirs" | tr '\n' ' ')
  fault "区分表と docs/test/ が不一致（表のみ: ${only_table:-なし}/ ディレクトリのみ: ${only_dirs:-なし}）"
fi

# 区分表の 2 列目（説明）が空でないこと。
empty_desc=$(read_tsv "$categories_tsv" | awk -F'\t' 'NF < 2 || $2 == "" { print $1 }' | tr '\n' ' ')
if [ -z "$empty_desc" ]; then
  pass "区分表の全行が説明を持つ"
else
  fault "区分表に説明の無い行がある（$empty_desc）"
fi

# 除外リストが実在する資産だけを挙げていること。
stale_exclusion=0
while IFS= read -r asset; do
  if ! grep -qxF "$asset" "$reverse_scan"; then
    fault "除外リストの '$asset' は列挙されるテスト資産に無い（stale な除外）"
    stale_exclusion=1
  fi
done < "$work/exclusions"

if [ "$stale_exclusion" -eq 0 ]; then
  pass "除外リストが実在するテスト資産だけを挙げている"
fi

# 除外リストの 2 列目（理由）が空でないこと。
empty_reason=$(read_tsv "$exclusions_tsv" | awk -F'\t' 'NF < 2 || $2 == "" { print $1 }' | tr '\n' ' ')
if [ -z "$empty_reason" ]; then
  pass "除外リストの全行が理由を持つ"
else
  fault "除外リストに理由の無い行がある（$empty_reason）"
fi

# CASE と除外リストが重複していないこと。
both=$(LC_ALL=C comm -12 "$work/covered" "$work/exclusions" | tr '\n' ' ')
if [ -z "$both" ]; then
  pass "除外リストと CASE の target が排他である"
else
  fault "除外されているのに CASE がある（$both）"
fi

# --- 5. 静的リストの健全性（FLAKE_CHECKS ⟷ flake ファイル） ------------------

# test-inventory.sh の静的リストと、flake ファイルの `checks.<name> =` 定義行を集合比較する。

# 静的リストの flake check 識別子。冒頭で保存した --static の出力から採る。
awk -F'\t' '$2 == "flake-check" { print $1 }' "$work/inventory" |
  LC_ALL=C sort > "$work/listed-checks"

# flakeModule 生成分（flake ファイルに定義行が無い）は grep 突合の対象外。
bash "$inventory_sh" --module-generated-checks | LC_ALL=C sort > "$work/module-checks" || {
  echo "test-doc-map.sh: $inventory_sh --module-generated-checks が失敗した" >&2
  exit 1
}

# flake ファイル側の定義。`checks.<name> =` の行から名前を採る。
extract_checks() {
  grep -oE '^[[:space:]]*checks\.[A-Za-z0-9_-]+[[:space:]]*=' "$1" |
    sed -E 's/^[[:space:]]*checks\.([A-Za-z0-9_-]+)[[:space:]]*=.*/\1/'
}

{
  extract_checks flake.nix | sed 's|^|checks.|'
  extract_checks dev/flake.nix | sed 's|^|dev:checks.|'
} | LC_ALL=C sort -u > "$work/defined-checks"

if [ ! -s "$work/defined-checks" ]; then
  fault "静的リスト: flake ファイルから checks 定義を 1 件も抽出できなかった（grep の前提が崩れた）"
else
  # 静的リスト（module 生成分を除く）⟷ flake の定義。
  LC_ALL=C comm -23 "$work/listed-checks" "$work/module-checks" > "$work/listed-hand"

  # module 生成分は only_flake から除く（診断は下の wrongly_module 側が出す）。
  LC_ALL=C comm -13 "$work/listed-hand" "$work/defined-checks" |
    LC_ALL=C comm -23 - "$work/module-checks" > "$work/only-flake"

  only_flake=$(tr '\n' ' ' < "$work/only-flake")
  only_list=$(LC_ALL=C comm -23 "$work/listed-hand" "$work/defined-checks" | tr '\n' ' ')

  if [ -z "$only_flake" ] && [ -z "$only_list" ]; then
    pass "FLAKE_CHECKS が flake ファイルの checks 定義と一致する"
  fi
  if [ -n "$only_flake" ]; then
    fault "静的リスト: flake に在るが FLAKE_CHECKS に無い（$only_flake）— $inventory_sh へ追加する"
  fi
  if [ -n "$only_list" ]; then
    fault "静的リスト: FLAKE_CHECKS に在るが flake の定義に無い（$only_list）— 消えた check の残骸か、flakeModule 生成分なら --module-generated-checks 側へ移す"
  fi

  # MODULE_GENERATED_CHECKS ⊆ FLAKE_CHECKS。
  not_listed=$(LC_ALL=C comm -13 "$work/listed-checks" "$work/module-checks" | tr '\n' ' ')
  if [ -z "$not_listed" ]; then
    pass "flakeModule 生成分が FLAKE_CHECKS にも載っている"
  else
    fault "静的リスト: flakeModule 生成分として挙げているが FLAKE_CHECKS に無い（$not_listed）— 列挙から落ちる"
  fi

  # module 生成分は逆に flake ファイルへ定義行を持たないはず（持つなら手書き側の管理へ移す）。
  wrongly_module=$(LC_ALL=C comm -12 "$work/module-checks" "$work/defined-checks" | tr '\n' ' ')
  if [ -z "$wrongly_module" ]; then
    pass "flakeModule 生成分が flake ファイルに定義行を持たない"
  else
    fault "静的リスト: flakeModule 生成分として除外しているが flake に定義行がある（$wrongly_module）"
  fi
fi

# --- 6. 自己検証（判定関数が両側へ倒れるか） ---------------------------------
#
# §1〜§3 が呼ぶ judge_* を合成フィクスチャ（$work/self/ 配下）へ当て、
# 違反なしで真・違反ありで偽の両側へ倒れることを確かめる。

self=$work/self
mkdir -p "$self"

# 違反のないフィクスチャ。空集合は空行も含まない空ファイルにする。
printf 'a_test.go\nb_test.go\n' > "$self/assets-ok"
printf 'CASE-a.md\ta_test.go\nCASE-b.md\tb_test.go\n' > "$self/case-targets-ok"
printf 'a_test.go\nb_test.go\n' > "$self/covered-ok"
: > "$self/exclusions-empty"

# 違反のあるフィクスチャ。判定内の分岐ごとに 1 つずつ用意する。
# 順方向 (a): 実在しない target を指す。
printf 'CASE-a.md\tgone_test.go\n' > "$self/case-targets-dangling"
# 順方向 (b): target が空。空 target 分岐だけを踏ませるため、assets 側に空行を含める。
printf 'CASE-b.md\t\n' > "$self/case-targets-empty"
printf '\na_test.go\n' > "$self/assets-with-empty"
# 逆方向: covered にも exclusions にも無い資産（c_test.go）。免除集合は走査対象の部分集合に保つ。
printf 'a_test.go\nc_test.go\n' > "$self/assets-uncovered"
printf 'a_test.go\n' > "$self/covered-subset"
# 逆方向: 除外リストだけで消える資産（除外経路の単独検証用）。
printf 'c_test.go\n' > "$self/assets-c-only"
# 1:1: 同じ target を 2 CASE が張る。
printf 'CASE-a.md\ta_test.go\nCASE-b.md\ta_test.go\n' > "$self/case-targets-dup"

self_fail=0

# 期待どおりに真（違反なし）を返すか。
expect_judge_pass() {
  local label=$1
  shift
  if ! "$@" >/dev/null; then
    fault "自己検証: $label — 違反のないフィクスチャで偽を返した（判定が過剰）"
    self_fail=1
  fi
}

# 期待どおりに偽（違反あり）を返し、診断行を出すか。
expect_judge_fail() {
  local label=$1
  shift
  local diagnostics
  if diagnostics=$("$@"); then
    fault "自己検証: $label — 違反のあるフィクスチャで真を返した（判定が常に真になる退行）"
    self_fail=1
  elif [ -z "$diagnostics" ]; then
    fault "自己検証: $label — 偽を返したが診断行が空（失敗の原因が報告されない）"
    self_fail=1
  fi
}

expect_judge_pass "judge_forward" \
  judge_forward "$self/case-targets-ok" "$self/assets-ok"
expect_judge_fail "judge_forward（実在しない target）" \
  judge_forward "$self/case-targets-dangling" "$self/assets-ok"
# 空 target 分岐は assets-with-empty（空行を含む）と組ませて単独に踏ませる。
expect_judge_fail "judge_forward（target が空）" \
  judge_forward "$self/case-targets-empty" "$self/assets-with-empty"

# 逆方向で違反を消す 2 経路（covered / exclusions）をそれぞれ単独で確かめる。
expect_judge_pass "judge_reverse（covered 経路）" \
  judge_reverse "$self/assets-ok" "$self/covered-ok" "$self/exclusions-empty"
printf 'c_test.go\n' > "$self/exclusions-c"
: > "$self/covered-empty"
expect_judge_pass "judge_reverse（除外リスト経路）" \
  judge_reverse "$self/assets-c-only" "$self/covered-empty" "$self/exclusions-c"
expect_judge_fail "judge_reverse（どちらにも無い）" \
  judge_reverse "$self/assets-uncovered" "$self/covered-subset" "$self/exclusions-empty"
# 部分集合契約。走査対象に無い要素を免除集合が含むと落ち、診断が配線の誤りを指す。
expect_judge_fail "judge_reverse（走査対象と免除集合の取り違え）" \
  judge_reverse "$self/covered-ok" "$self/assets-uncovered" "$self/exclusions-c"

wiring_diagnostic=$(judge_reverse "$self/covered-ok" "$self/assets-uncovered" "$self/exclusions-c")
if ! printf '%s' "$wiring_diagnostic" | grep -q '引数配線の誤り'; then
  fault "自己検証: judge_reverse（部分集合契約）— 配線の誤りを指す診断が出ない（実際: ${wiring_diagnostic:-空}）"
  self_fail=1
fi

expect_judge_pass "judge_unique" \
  judge_unique "$self/case-targets-ok"
expect_judge_fail "judge_unique" \
  judge_unique "$self/case-targets-dup"

# run_judge 自体の検証。pass / fault を数える版へ一時的に差し替えて、呼ばれ方を観測する。
judge_always_ok() { return 0; }
judge_always_violates() { echo "合成の違反診断"; return 1; }
judge_violates_silently() { return 1; }

probe_run_judge() {
  local pass_count=0 fault_count=0
  # shellcheck disable=SC2317  # 差し替え後に run_judge から呼ばれる
  pass() { pass_count=$((pass_count + 1)); }
  # shellcheck disable=SC2317
  fault() { fault_count=$((fault_count + 1)); }

  run_judge "probe" "$@" >/dev/null 2>&1
  local status=$?

  printf '%d\t%d\t%d\n' "$pass_count" "$fault_count" "$status"
}

# probe_run_judge は pass / fault を差し替えるため、必ずコマンド置換（サブシェル）で呼ぶ。
probe_ok=$(probe_run_judge judge_always_ok)
probe_violation=$(probe_run_judge judge_always_violates)
probe_silent=$(probe_run_judge judge_violates_silently)

check_probe() {
  local label=$1 actual=$2 want_pass=$3 want_fault_min=$4 want_status=$5
  local got_pass got_fault got_status
  IFS=$'\t' read -r got_pass got_fault got_status <<< "$actual"
  if [ "$got_pass" != "$want_pass" ] ||
    [ "$got_fault" -lt "$want_fault_min" ] ||
    [ "$got_status" != "$want_status" ]; then
    fault "自己検証: run_judge（$label）— pass=$got_pass fault=$got_fault exit=$got_status（期待: pass=$want_pass fault>=$want_fault_min exit=$want_status）"
    self_fail=1
  fi
}

# 違反なし → pass 1 回・fault 0 回・exit 0。
check_probe "違反なし" "$probe_ok" 1 0 0
# 違反あり → pass 0 回・fault 1 回以上・exit 1。
check_probe "違反あり" "$probe_violation" 0 1 1
# 違反ありだが診断行が空 → 黙って通さず fault へ倒す。
check_probe "違反ありで診断が空" "$probe_silent" 0 1 1

# 実データ側の入力の非空性。assets / covered が空だと §1〜§3 が空虚に真になる。
for name in assets covered case-targets; do
  if [ ! -s "$work/$name" ]; then
    fault "自己検証: $name が空（§1〜§3 が空虚に真になる）"
    self_fail=1
  fi
done

if [ "$self_fail" -eq 0 ]; then
  pass "自己検証: judge_* と run_judge が合成フィクスチャで期待どおり振る舞う"
fi

# --- 結果 -------------------------------------------------------------------

echo
if [ "$fail" -eq 0 ]; then
  printf 'test-doc-map: 全アサーション通過（CASE %d 件 / テスト資産 %d 件 / 除外 %d 件）\n' \
    "$(wc -l < "$work/case-targets")" "$(wc -l < "$work/assets")" "$(wc -l < "$work/exclusions")"
  exit 0
fi

printf 'test-doc-map: 失敗あり\n' >&2
exit 1
