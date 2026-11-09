#!/usr/bin/env bash
# sara-gap（グラフ未カバーの 7 節と到達率の列挙コマンド）の検証。fixture に実物の docs/model.yaml を重ねる。
#
# 実行:
#   nix develop ./dev -c dev/tests/sara-gap.sh   # devShell から直接
#   nix flake check ./dev                         # checks.sara-gap 経由
set -uo pipefail

fail=0
pass() { printf 'ok   - %s\n' "$1"; }
fault() {
  printf 'FAIL - %s\n' "$1"
  fail=1
}

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# fixture はスクリプト自身からの相対で解決する。
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fixture="$script_dir/fixtures/sara-gap"

if [[ ! -d "$fixture/docs" ]]; then
  fault "fixture を解決できない（$fixture）"
  exit 1
fi

# モデルは実リポジトリの docs/model.yaml。SARA_GAP_MODEL_YAML があればそれを、無ければ git ルート基準で引く。
# どちらでも解決できなければ失敗させる。
model_yaml="${SARA_GAP_MODEL_YAML:-}"
contract_root="$(git rev-parse --show-toplevel 2>/dev/null || printf '.')"
[[ -f "$model_yaml" ]] || model_yaml="$contract_root/docs/model.yaml"
if [[ ! -f "$model_yaml" ]]; then
  fault "docs/model.yaml を解決できない（model.yaml=$model_yaml）"
  exit 1
fi

# fixture の複製に実物モデルを重ねて、実行可能な sara リポジトリの形を作る。
make_fixture() {
  mkdir -p "$1"
  cp -r "$fixture"/. "$1"
  cp "$model_yaml" "$1/model.yaml"
}

base="$work/base"
make_fixture "$base"

# base の期待値は unthreatened 2・unmitigated 2・uncovered 1・unfollowed 1・unsatisfied_quality 1・exempt 1。
# 鎖が最後までつながる要求は REQ-aaaa0001 だけで、DSG-dd000001 は risk の有無によらず列挙しない。

run_gap() {
  stdout_capture="$("$@" 2>/dev/null)" && status_capture=0 || status_capture=$?
}

# 値が一致すれば pass、違えば実際の値つきで fault にする。
assert_eq() {
  if [[ "$2" == "$3" ]]; then
    pass "$1"
  else
    fault "$1（期待: $2 実際: $3）"
  fi
}

# 文字列 $3 に正規表現 $2 が一致するかを確かめる。
assert_has() {
  if grep -Eq -- "$2" <<<"$3"; then
    pass "$1"
  else
    fault "$1（$2 が出力に無い）"
  fi
}

assert_lacks() {
  if ! grep -Eq -- "$2" <<<"$3"; then
    pass "$1"
  else
    fault "$1（$2 が出力に出ている）"
  fi
}

# text 出力 $2 から見出し $1 の節の本文だけを取り出す。
section_body() {
  local line inside=0
  while IFS= read -r line; do
    if [[ "$line" == "## "* ]]; then
      if [[ "$line" == "$1" ]]; then inside=1; else inside=0; fi
      continue
    fi
    if [[ "$inside" -eq 1 ]]; then printf '%s\n' "$line"; fi
  done <<<"$2"
}

h_unthreatened='## threatens も risk_exempt も無い requirement'
h_unmitigated='## mitigates されていない risk'
h_uncovered='## covers されていない test_condition'
h_unfollowed='## follows を持たない test_case'
h_unsatisfied='## infrastructure に satisfies されていない quality'
h_exempt='## risk_exempt を持つ requirement（情報）'
h_reach='## 到達率'
expected_headings="$(printf '%s\n' "$h_unthreatened" "$h_unmitigated" "$h_uncovered" "$h_unfollowed" "$h_unsatisfied" "$h_exempt" "$h_reach")"
design_mark='\(design 側に risk あり\)$'
full_id='^[A-Z]+-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'$'\t'
short_id='^[A-Z]+-[0-9a-f]{8}'$'\t'

# --- 1. 7 節の検出と exit 1 ---------------------------------------------------

run_gap env SARA_GAP_ROOT="$base" sara-gap
gap_out="$stdout_capture"

assert_eq "ギャップがあれば exit 1" 1 "$status_capture"
assert_eq "見出しを 7 つ、この順に出す" "$expected_headings" "$(grep '^## ' <<<"$gap_out")"

sec_unthreatened="$(section_body "$h_unthreatened" "$gap_out")"
sec_unfollowed="$(section_body "$h_unfollowed" "$gap_out")"
sec_exempt="$(section_body "$h_exempt" "$gap_out")"

assert_has "risk の無い REQ を節 1 に出す" '^REQ-bbbb0002' "$sec_unthreatened"
assert_has "mitigates されていない RISK を節 2 に出す" '^RISK-22220002' "$(section_body "$h_unmitigated" "$gap_out")"
assert_has "covers されていない TC を節 3 に出す" '^TC-44440002' "$(section_body "$h_uncovered" "$gap_out")"
assert_has "follows の無い CASE を節 4 に出す" '^CASE-55550001' "$sec_unfollowed"
assert_has "INF に satisfies されない QA を節 5 に出す" '^QA-88880002' "$(section_body "$h_unsatisfied" "$gap_out")"

for nowant in REQ-aaaa0001 RISK-11110001 TC-33330001 CASE-66660002 QA-88880001; do
  assert_lacks "張り先を持つ $nowant を列挙しない" "^$nowant" "$gap_out"
done

assert_lacks "design を列挙しない" 'DSG-dd000001' "$gap_out"
assert_lacks "risk_exempt を持つ REQ を節 1 に出さない" 'REQ-cccc0003' "$sec_unthreatened"
assert_has "risk_exempt を持つ REQ を節 6 に出す" '^REQ-cccc0003' "$sec_exempt"
assert_has "design 側にだけ risk がある REQ に印を付ける" "^REQ-dddd0004.*$design_mark" "$sec_unthreatened"
assert_lacks "design 側に risk が無い REQ には印を付けない" "^REQ-bbbb0002.*$design_mark" "$sec_unthreatened"

# --- 2. text の行形式 ----------------------------------------------------------

req_line="$(grep '^REQ-bbbb0002' <<<"$gap_out")"
assert_eq "行が <フル ID>\\t<name>\\t<docs/ 前置の file> の形" \
  $'REQ-bbbb0002-0000-4000-8000-000000000004\trisk が無い要求\tdocs/req-gap.md' "$req_line"

# 節 7 の前までの行だけを見る。節 7 の行は ID を持たない。
item_lines="$(grep -Ev '^(## .*|なし|)$' <<<"${gap_out%%"$h_reach"*}")"
assert_eq "節 1〜6 の行は 8 件" 8 "$(grep -c '' <<<"$item_lines")"
assert_eq "節 1〜6 の行の先頭がすべてフル ID" 8 "$(grep -Ec "$full_id" <<<"$item_lines")"
assert_eq "8 文字省略形を出さない" 0 "$(grep -Ec "$short_id" <<<"$gap_out")"

sec_reach="$(section_body "$h_reach" "$gap_out")"
assert_has "節 7 に requirement の到達率を出す" $'^requirement\t1/3\t33\\.3%$' "$sec_reach"
assert_has "節 7 に test_plan の到達率を出す" $'^test_plan\t1/1\t100%$' "$sec_reach"
assert_has "節 7 に quality の到達率を出す" $'^quality\t1/2\t50%$' "$sec_reach"

# --- 3. --json -----------------------------------------------------------------

run_gap env SARA_GAP_ROOT="$base" sara-gap --json
json_out="$stdout_capture"

assert_eq "--json でも exit code は同じ（ギャップあり = 1）" 1 "$status_capture"
assert_eq "--json は 7 キーのオブジェクトを返す" \
  "exempt,reach,uncovered,unfollowed,unmitigated,unsatisfied_quality,unthreatened" \
  "$(jq -r 'keys | join(",")' <<<"$json_out" 2>/dev/null)"
assert_eq "--json の件数が 2,2,1,1,1,1" "2,2,1,1,1,1" \
  "$(jq -r '[.unthreatened, .unmitigated, .uncovered, .unfollowed, .unsatisfied_quality, .exempt | length] | join(",")' <<<"$json_out" 2>/dev/null)"
assert_eq "--json の行のキーは id / name / file だけ" '["file,id,name"]' \
  "$(jq -c '[.unthreatened, .unmitigated, .uncovered, .unfollowed, .unsatisfied_quality, .exempt | .[] | keys | join(",")] | unique' <<<"$json_out" 2>/dev/null)"
assert_eq "--json の行の id はフル ID で file は docs/ 前置" "true" \
  "$(jq -r '.unsatisfied_quality[0] | (.id == "QA-88880002-0000-4000-8000-000000000010") and (.name != "") and (.file | startswith("docs/"))' <<<"$json_out" 2>/dev/null)"
assert_eq "exempt に risk_exempt を持つ REQ が入る" "REQ-cccc0003-0000-4000-8000-00000000000b" \
  "$(jq -r '.exempt[0].id' <<<"$json_out" 2>/dev/null)"
assert_eq "unfollowed に follows の無い CASE だけが入る" "CASE-55550001-0000-4000-8000-00000000000a" \
  "$(jq -r '[.unfollowed[].id] | join(",")' <<<"$json_out" 2>/dev/null)"
assert_eq "reach の 3 値が期待どおり" "true" \
  "$(jq '.reach == {requirement: {total: 3, reached: 1, ratio: (1 / 3)}, test_plan: {total: 1, reached: 1, ratio: 1}, quality: {total: 2, reached: 1, ratio: 0.5}}' <<<"$json_out" 2>/dev/null)"

# --- 4. ギャップを埋めた派生 fixture ------------------------------------------
#
# 免除の REQ と risk の無い DSG-dd000001 は残し、どちらも exit に効かないことを確かめる。

fill_gaps() {
  make_fixture "$1"
  rm "$1"/docs/{req-gap,risk-gap,tc-gap,qa-gap,req-design-only,risk-on-design}.md
  sed -i '/REQ-dddd0004/d' "$1/docs/dsg-gap.md"
  sed -i 's|^covers:$|follows:\n  - "TP-77770001-0000-4000-8000-00000000000e"\ncovers:|' "$1/docs/case.md"
}

clean="$work/clean"
fill_gaps "$clean"

run_gap env SARA_GAP_ROOT="$clean" sara-gap
clean_out="$stdout_capture"
assert_eq "ギャップが無ければ exit 0（exempt は exit に影響しない）" 0 "$status_capture"
assert_eq "見出しは埋めても 7 つ" 7 "$(grep -c '^## ' <<<"$clean_out")"
assert_eq "0 件の節 1〜5 は「なし」と出す" 5 "$(grep -c '^なし$' <<<"$clean_out")"
assert_has "exempt は埋めても節 6 に出る" '^REQ-cccc0003' "$(section_body "$h_exempt" "$clean_out")"
assert_lacks "risk の無い design を列挙しない" 'DSG-dd000001' "$clean_out"

run_gap env SARA_GAP_ROOT="$clean" sara-gap --json
assert_eq "--json でもギャップ 5 配列が空で exit 0" "0:0" \
  "$status_capture:$(jq -r '[.unthreatened, .unmitigated, .uncovered, .unfollowed, .unsatisfied_quality | length] | add' <<<"$stdout_capture" 2>/dev/null)"
assert_eq "--json の reach がすべて 1" "1,1,1" \
  "$(jq -r '[.reach.requirement, .reach.test_plan, .reach.quality | .ratio] | join(",")' <<<"$stdout_capture" 2>/dev/null)"

# 節 1〜5 が空でも、follows されない TP があれば到達率だけで exit 1 になる。
reach_only="$work/reach-only"
fill_gaps "$reach_only"
cat >"$reach_only/docs/tp-unfollowed.md" <<'MD'
---
id: "TP-77770002-0000-4000-8000-000000000013"
type: test_plan
name: "follows されないテスト計画"
status: stable
derives_from:
  - "SOL-5011a001-0000-4000-8000-000000000001"
specification: |
  The test process SHALL leave this test plan unfollowed.
specification_ja: |
  テストプロセスはこのテスト計画を follows されない状態に保たなければならない。
---
# TP-77770002: follows されないテスト計画
MD

run_gap env SARA_GAP_ROOT="$reach_only" sara-gap
reach_only_out="$stdout_capture"
assert_eq "到達率だけが 1 未満でも exit 1" 1 "$status_capture"
assert_eq "到達率だけが 1 未満のとき節 1〜5 は空" 5 "$(grep -c '^なし$' <<<"$reach_only_out")"

# CASE が無くなると TC までつながっていても requirement は到達にならない。
no_case="$work/no-case"
fill_gaps "$no_case"
rm "$no_case/docs/case.md" "$no_case/docs/case-followed.md"

run_gap env SARA_GAP_ROOT="$no_case" sara-gap --json
assert_eq "鎖が CASE まで届かない REQ は reached に数えない" "true" \
  "$(jq '.reach.requirement == {total: 1, reached: 0, ratio: 0}' <<<"$stdout_capture" 2>/dev/null)"

# --- 5. sara check が失敗するグラフでは exit 2 ---------------------------------
#
# broken reference（実在しない risk への mitigates）を混ぜ、ギャップ一覧を出さないことを確かめる。

broken="$work/broken"
make_fixture "$broken"
cat >"$broken/docs/broken.md" <<'MD'
---
id: "TC-99990009-0000-4000-8000-00000000000b"
type: test_condition
name: "壊れた参照を持つテスト条件"
mitigates:
  - "RISK-deadbeef-0000-4000-8000-00000000000c"
---
# TC-99990009: 壊れた参照を持つテスト条件
MD

run_gap env SARA_GAP_ROOT="$broken" sara-gap
if [[ "$status_capture" -eq 2 ]]; then
  pass "sara check が失敗するグラフでは exit 2"
else
  fault "sara check が失敗するグラフでは exit 2（実際: exit=$status_capture）"
fi
if [[ -z "$stdout_capture" ]]; then
  pass "sara check 失敗時はギャップ一覧を stdout に出さない"
else
  fault "sara check 失敗時はギャップ一覧を stdout に出さない（実際: $stdout_capture）"
fi

# --- 6. JSON 形状異常では exit 2（seam）----------------------------------------
#
# SARA_GAP_SARA seam で偽 sara を差し込み、JSON 形状の異常を再現する。
# shebang はサンドボックスに /usr/bin/env が無いため実行中の bash の絶対パスを埋め込む。

fake_noitems="$work/fake-sara-noitems"
{
  printf '#!%s\n' "$BASH"
  printf 'printf %%s "{}"\n'
} >"$fake_noitems"
chmod +x "$fake_noitems"

run_gap env SARA_GAP_ROOT="$base" SARA_GAP_SARA="$fake_noitems" sara-gap
if [[ "$status_capture" -eq 2 ]]; then
  pass ".items 配列が無い JSON では exit 2"
else
  fault ".items 配列が無い JSON では exit 2（実際: exit=$status_capture）"
fi

fake_invalid="$work/fake-sara-invalid"
{
  printf '#!%s\n' "$BASH"
  printf "printf %%s '{\"items\": [], \"valid\": false}'\n"
} >"$fake_invalid"
chmod +x "$fake_invalid"

run_gap env SARA_GAP_ROOT="$base" SARA_GAP_SARA="$fake_invalid" sara-gap
if [[ "$status_capture" -eq 2 ]]; then
  pass "exit 0 でも valid: false の JSON では exit 2"
else
  fault "exit 0 でも valid: false の JSON では exit 2（実際: exit=$status_capture）"
fi

# item が 0 件のグラフは全 reach の total が 0 で、ratio は 1 として exit 0 にする。
fake_empty="$work/fake-sara-empty"
{
  printf '#!%s\n' "$BASH"
  printf "printf %%s '{\"items\": [], \"valid\": true}'\n"
} >"$fake_empty"
chmod +x "$fake_empty"

run_gap env SARA_GAP_ROOT="$base" SARA_GAP_SARA="$fake_empty" sara-gap --json
assert_eq "item が 0 件なら reach の ratio は 1 で exit 0" "0:1,1,1" \
  "$status_capture:$(jq -r '[.reach.requirement, .reach.test_plan, .reach.quality | .ratio] | join(",")' <<<"$stdout_capture" 2>/dev/null)"

# --- 7. 走査先をリポジトリルート基準で解決する ---------------------------------
#
# SARA_GAP_ROOT 無しでサブディレクトリから実行しても git ルートの sara.toml / docs を対象にする。

gitrepo="$work/gitrepo"
make_fixture "$gitrepo"
git -C "$gitrepo" init -q
mkdir -p "$gitrepo/sub"
sub_status=0
(cd "$gitrepo/sub" && sara-gap >/dev/null 2>&1) || sub_status=$?
if [[ "$sub_status" -eq 1 ]]; then
  pass "サブディレクトリから叩いても git ルートのグラフを対象にする（ギャップ検出 = exit 1）"
else
  fault "サブディレクトリから叩いても git ルートのグラフを対象にする（実際: exit=$sub_status）"
fi

# --- 8. 引数の異常系 ------------------------------------------------------------

run_gap sara-gap --help
if [[ "$status_capture" -eq 0 ]]; then
  pass "--help は exit 0"
else
  fault "--help は exit 0（実際: exit=$status_capture）"
fi

run_gap sara-gap --bogus
if [[ "$status_capture" -eq 2 ]]; then
  pass "未知の引数は exit 2"
else
  fault "未知の引数は exit 2（実際: exit=$status_capture）"
fi

run_gap sara-gap --json extra
if [[ "$status_capture" -eq 2 ]]; then
  pass "引数過多は exit 2"
else
  fault "引数過多は exit 2（実際: exit=$status_capture）"
fi

exit "$fail"
