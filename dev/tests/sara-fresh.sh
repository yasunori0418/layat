#!/usr/bin/env bash
# sara-fresh（再検証待ちの item と人の確認が無い high の risk の列挙コマンド）の検証。
# item と sara.toml は一時ディレクトリに作り、実物の docs/model.yaml を重ねる。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/sara-fresh.sh   # devShell から直接
#   nix flake check ./dev                                       # checks.sara-fresh 経由
set -uo pipefail

fail=0
pass() { printf 'ok   - %s\n' "$1"; }
fault() {
  printf 'FAIL - %s\n' "$1"
  fail=1
}

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# モデルは実リポジトリの docs/model.yaml。SARA_FRESH_MODEL_YAML があればそれを、無ければ git ルート基準で引く。
model_yaml="${SARA_FRESH_MODEL_YAML:-}"
contract_root="$(git rev-parse --show-toplevel 2>/dev/null || printf '.')"
[[ -f "$model_yaml" ]] || model_yaml="$contract_root/docs/model.yaml"
if [[ ! -f "$model_yaml" ]]; then
  fault "docs/model.yaml を解決できない（model.yaml=$model_yaml）"
  exit 1
fi

run_fresh() {
  stdout_capture="$("$@" 2>/dev/null)" && status_capture=0 || status_capture=$?
}

assert_eq() {
  if [[ "$2" == "$3" ]]; then
    pass "$1"
  else
    fault "$1（期待: $2 実際: $3）"
  fi
}

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

# --- グラフの組み立て ----------------------------------------------------------

# <リポジトリ> <ファイル名> <ID> <型> <name> の item を書く。frontmatter の残り（status 以降）は標準入力から受ける。
item() {
  local repo=$1 file=$2 id=$3 type=$4 name=$5
  {
    printf -- '---\nid: "%s"\ntype: %s\nname: "%s"\n' "$id" "$type" "$name"
    cat
    printf -- '---\n# %s\n' "$name"
  } >"$repo/docs/$file.md"
}

# requirement の必須フィールド。
req_spec='specification: |
  The fixture SHALL keep this requirement.
specification_ja: |
  fixture はこの要求を保たなければならない。'

SOL="SOL-5011a001-0000-4000-8000-000000000001"
UC="UC-0ca11001-0000-4000-8000-000000000002"
UC_NEW="UC-0ca11002-0000-4000-8000-000000000003"
UC_DEP="UC-0ca11003-0000-4000-8000-000000000004"
REQ_STALE="REQ-aaaa0001-0000-4000-8000-000000000010"
REQ_FRESH="REQ-aaaa0002-0000-4000-8000-000000000011"
REQ_EQUAL="REQ-aaaa0003-0000-4000-8000-000000000012"
REQ_OFFSET="REQ-aaaa0004-0000-4000-8000-000000000013"
REQ_VER_AFTER="REQ-aaaa0005-0000-4000-8000-000000000014"
REQ_VER_BEFORE="REQ-aaaa0006-0000-4000-8000-000000000015"
REQ_DEP_UP="REQ-aaaa0007-0000-4000-8000-000000000016"
REQ_DEP_DOWN="REQ-aaaa0008-0000-4000-8000-000000000017"
REQ_TWO="REQ-aaaa0009-0000-4000-8000-000000000018"
REQ_PEER="REQ-aaaa000a-0000-4000-8000-000000000019"
SDR="SDR-dddd0001-0000-4000-8000-000000000020"
RISK_NONE="RISK-11110001-0000-4000-8000-000000000030"
RISK_AGENT="RISK-11110002-0000-4000-8000-000000000031"
RISK_HUMAN_BEFORE="RISK-11110003-0000-4000-8000-000000000032"
RISK_HUMAN_AFTER="RISK-11110004-0000-4000-8000-000000000033"
RISK_HUMAN_EQUAL="RISK-11110005-0000-4000-8000-000000000034"
RISK_MEDIUM="RISK-11110006-0000-4000-8000-000000000035"
RISK_HIGH_DEP="RISK-11110007-0000-4000-8000-000000000036"
RISK_OLD="RISK-11110008-0000-4000-8000-000000000037"
TC_OLD="TC-33330001-0000-4000-8000-000000000040"
CASE_OLD="CASE-55550001-0000-4000-8000-000000000050"
DSG_OLD="DSG-dd000001-0000-4000-8000-000000000051"
TP="TP-77770001-0000-4000-8000-000000000052"
SOL_NEW="SOL-5011a002-0000-4000-8000-000000000053"
UC_OLD="UC-0ca11004-0000-4000-8000-000000000054"

# 上流が新しい下流・上流が古い下流・verified の効き方・justifies・deprecated・high の risk の各側を 1 つのグラフに置く。
make_graph() {
  local repo=$1
  mkdir -p "$repo/docs"
  cp "$model_yaml" "$repo/model.yaml"
  cat >"$repo/sara.toml" <<'TOML'
model_schema = "model.yaml"

[repositories]
paths = ["./docs"]

[validation]
strict_mode = true
TOML

  item "$repo" sol "$SOL" solution "根" <<'YAML'
status: stable
generated:
  by: "human:fixture"
  at: "2026-01-01T00:00:00+09:00"
YAML
  item "$repo" uc "$UC" use_case "上流のユースケース" <<YAML
status: stable
refines:
  - "$SOL"
generated:
  by: "human:fixture"
  at: "2026-03-01T00:00:00+09:00"
YAML
  item "$repo" uc-new "$UC_NEW" use_case "もう 1 つの新しいユースケース" <<YAML
status: stable
refines:
  - "$SOL"
generated:
  by: "human:fixture"
  at: "2026-03-05T00:00:00+09:00"
YAML
  item "$repo" uc-dep "$UC_DEP" use_case "廃止したユースケース" <<YAML
status: deprecated
refines:
  - "$SOL"
generated:
  by: "human:fixture"
  at: "2026-06-01T00:00:00+09:00"
YAML

  # 上流（2026-03-01）より古い。
  item "$repo" req-stale "$REQ_STALE" requirement "上流より古い要求" <<YAML
status: stable
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-01T00:00:00+09:00"
YAML
  # 上流より新しい。
  item "$repo" req-fresh "$REQ_FRESH" requirement "上流より新しい要求" <<YAML
status: stable
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-04-01T00:00:00+09:00"
YAML
  # 上流と同じ時刻。
  item "$repo" req-equal "$REQ_EQUAL" requirement "上流と同じ時刻の要求" <<YAML
status: stable
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-03-01T00:00:00+09:00"
YAML
  # 文字列では上流（2026-03-01T00:00+09:00 = 2026-02-28T15:00Z）より古く見えるが、時刻では新しい。
  item "$repo" req-offset "$REQ_OFFSET" requirement "オフセット違いで新しい要求" <<YAML
status: stable
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-28T16:00:00Z"
YAML
  # generated.at 以降の verified（actor は問わない）で基準時刻が上流より新しくなる。
  item "$repo" req-ver-after "$REQ_VER_AFTER" requirement "確認で基準時刻が進んだ要求" <<YAML
status: stable
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-01T00:00:00+09:00"
verified:
  - by: "claude-code/fixture"
    at: "2026-03-15T00:00:00+09:00"
YAML
  # generated.at より前の verified は基準時刻に効かない。
  item "$repo" req-ver-before "$REQ_VER_BEFORE" requirement "生成より前の確認しか無い要求" <<YAML
status: stable
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-01T00:00:00+09:00"
verified:
  - by: "human:fixture"
    at: "2026-01-15T00:00:00+09:00"
YAML
  # deprecated の上流だけを指す。
  item "$repo" req-dep-up "$REQ_DEP_UP" requirement "廃止した上流を指す要求" <<YAML
status: stable
derives_from:
  - "$UC_DEP"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-01T00:00:00+09:00"
YAML
  # 自身が deprecated。
  item "$repo" req-dep-down "$REQ_DEP_DOWN" requirement "廃止した古い要求" <<YAML
status: deprecated
derives_from:
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-01T00:00:00+09:00"
YAML
  # 新しい上流を 2 つ持つ。
  item "$repo" req-two "$REQ_TWO" requirement "新しい上流を 2 つ持つ要求" <<YAML
status: stable
derives_from:
  - "$UC_NEW"
  - "$UC"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-02-01T00:00:00+09:00"
YAML
  # depends_on で指す要求（2026-04-01）より古いが、peer の relation は辿らない。derives_from の上流よりは新しい。
  item "$repo" req-peer "$REQ_PEER" requirement "依存先より古い要求" <<YAML
status: stable
derives_from:
  - "$UC"
depends_on:
  - "$REQ_FRESH"
$req_spec
generated:
  by: "claude-code/fixture"
  at: "2026-03-20T00:00:00+09:00"
YAML
  # justifies で指す要求（2026-04-01）より古いが、justifies は辿らない。
  item "$repo" sdr "$SDR" sara_adr "古い決定の記録" <<YAML
status: stable
justifies:
  - "$REQ_FRESH"
generated:
  by: "claude-code/fixture"
  at: "2026-01-10T00:00:00+09:00"
YAML

  # risk() で作る risk はどれも上流（2026-04-01）より新しく、節 1 には出ない。
  risk "$repo" risk-none "$RISK_NONE" "確認の無い high の risk" high stable ""
  risk "$repo" risk-agent "$RISK_AGENT" "エージェントの確認だけの high の risk" high stable '
verified:
  - by: "claude-code/fixture"
    at: "2026-05-02T00:00:00+09:00"'
  risk "$repo" risk-human-before "$RISK_HUMAN_BEFORE" "生成より前に人が確認した high の risk" high stable '
verified:
  - by: "human:fixture"
    at: "2026-04-20T00:00:00+09:00"'
  risk "$repo" risk-human-after "$RISK_HUMAN_AFTER" "生成の後に人が確認した high の risk" high stable '
verified:
  - by: "claude-code/fixture"
    at: "2026-05-02T00:00:00+09:00"
  - by: "human:fixture"
    at: "2026-05-03T00:00:00+09:00"'
  risk "$repo" risk-human-equal "$RISK_HUMAN_EQUAL" "生成と同じ時刻に人が確認した high の risk" high stable '
verified:
  - by: "human:fixture"
    at: "2026-04-30T16:00:00Z"'
  risk "$repo" risk-medium "$RISK_MEDIUM" "確認の無い medium の risk" medium stable ""
  risk "$repo" risk-high-dep "$RISK_HIGH_DEP" "廃止した確認の無い high の risk" high deprecated ""

  # derives_from 以外の relation で新しい上流を指す下流。
  item "$repo" risk-old "$RISK_OLD" risk "要求より古い medium の risk" <<YAML
status: stable
threatens:
  - "$REQ_FRESH"
likelihood: medium
impact: medium
level: medium
generated:
  by: "claude-code/fixture"
  at: "2026-03-20T00:00:00+09:00"
YAML
  item "$repo" tc-old "$TC_OLD" test_condition "risk より古いテスト条件" <<YAML
status: stable
mitigates:
  - "$RISK_NONE"
generated:
  by: "claude-code/fixture"
  at: "2026-04-15T00:00:00+09:00"
YAML
  item "$repo" dsg-old "$DSG_OLD" design "要求より古い設計" <<YAML
status: stable
satisfies:
  - "$REQ_FRESH"
generated:
  by: "claude-code/fixture"
  at: "2026-03-25T00:00:00+09:00"
YAML
  item "$repo" tp "$TP" test_plan "新しいテスト計画" <<YAML
status: stable
derives_from:
  - "$SOL"
specification: |
  The test process SHALL keep this test plan.
specification_ja: |
  テストプロセスはこのテスト計画を保たなければならない。
generated:
  by: "claude-code/fixture"
  at: "2026-05-10T00:00:00+09:00"
YAML
  # covers の先は 2026-04-15、follows の先は 2026-05-10。
  item "$repo" case-old "$CASE_OLD" test_case "テスト条件とテスト計画より古いテストケース" <<YAML
status: stable
resource: "tests/fixture/dummy_test.go"
covers:
  - "$TC_OLD"
follows:
  - "$TP"
generated:
  by: "claude-code/fixture"
  at: "2026-04-01T00:00:00+09:00"
YAML
  item "$repo" sol-new "$SOL_NEW" solution "新しい根" <<'YAML'
status: stable
generated:
  by: "human:fixture"
  at: "2026-06-01T00:00:00+09:00"
YAML
  item "$repo" uc-old "$UC_OLD" use_case "根より古いユースケース" <<YAML
status: stable
refines:
  - "$SOL_NEW"
generated:
  by: "human:fixture"
  at: "2026-02-01T00:00:00+09:00"
YAML
}

# <リポジトリ> <ファイル名> <ID> <name> <level> <status> <verified の YAML>。generated.at は 2026-05-01T01:00:00+09:00。
risk() {
  local repo=$1 file=$2 id=$3 name=$4 level=$5 status=$6 verified=$7
  item "$repo" "$file" "$id" risk "$name" <<YAML
status: $status
threatens:
  - "$REQ_FRESH"
likelihood: $level
impact: $level
level: $level
generated:
  by: "claude-code/fixture"
  at: "2026-05-01T01:00:00+09:00"$verified
YAML
}

h_stale='## 再検証待ち（上流の generated.at が基準時刻より新しい item）'
h_risk='## human: の verified が無い high の risk'

base="$work/base"
make_graph "$base"

# --- 1. 節 1: 再検証待ち --------------------------------------------------------

run_fresh env SARA_FRESH_ROOT="$base" sara-fresh
out="$stdout_capture"

assert_eq "列挙があれば exit 1" 1 "$status_capture"
assert_eq "見出しを 2 つ、この順に出す" "$(printf '%s\n' "$h_stale" "$h_risk")" "$(grep '^## ' <<<"$out")"

sec_stale="$(section_body "$h_stale" "$out")"
assert_has "上流が新しい下流を挙げる" "^$REQ_STALE"$'\t' "$sec_stale"
assert_lacks "上流が古い下流を挙げない" "^$REQ_FRESH" "$sec_stale"
assert_lacks "上流と同じ時刻の下流を挙げない" "^$REQ_EQUAL" "$sec_stale"
assert_lacks "UTC オフセットを揃えて比べる（文字列順では古く見える下流を挙げない）" "^$REQ_OFFSET" "$sec_stale"
assert_lacks "generated.at 以降の verified で基準時刻が進んだ下流を挙げない" "^$REQ_VER_AFTER" "$sec_stale"
assert_has "generated.at より前の verified は基準時刻に効かない" "^$REQ_VER_BEFORE"$'\t' "$sec_stale"
assert_lacks "justifies の辺は挙げない" "^$SDR" "$sec_stale"
assert_lacks "peer の relation（depends_on）の辺は挙げない" "^$REQ_PEER" "$sec_stale"
assert_lacks "deprecated の上流しか指さない下流を挙げない" "^$REQ_DEP_UP" "$sec_stale"
assert_lacks "deprecated の下流を挙げない" "^$REQ_DEP_DOWN" "$sec_stale"
assert_lacks "上流より新しい根とユースケースを挙げない" "^($SOL|$UC)" "$sec_stale"
assert_has "threatens で新しい上流を指す下流を挙げる" "^$RISK_OLD"$'\t' "$sec_stale"
assert_has "mitigates で新しい上流を指す下流を挙げる" "^$TC_OLD"$'\t' "$sec_stale"
assert_has "satisfies で新しい上流を指す下流を挙げる" "^$DSG_OLD"$'\t' "$sec_stale"
assert_has "covers と follows で新しい上流を指す下流を挙げる" "^$CASE_OLD"$'\t' "$sec_stale"
assert_has "refines で新しい上流を指す下流を挙げる" "^$UC_OLD"$'\t' "$sec_stale"
assert_eq "節 1 は 8 件" 8 "$(grep -c '' <<<"$sec_stale")"

assert_eq "行が <フル ID>\\t<name>\\t<file>\\t<新しい上流の ID> の形" \
  "$REQ_STALE"$'\t'"上流より古い要求"$'\t'"docs/req-stale.md"$'\t'"$UC" \
  "$(grep "^$REQ_STALE" <<<"$sec_stale")"
assert_eq "新しい上流が複数なら ID 順にカンマ区切りで並べる" \
  "$REQ_TWO"$'\t'"新しい上流を 2 つ持つ要求"$'\t'"docs/req-two.md"$'\t'"$UC,$UC_NEW" \
  "$(grep "^$REQ_TWO" <<<"$sec_stale")"

# --- 2. human: の verified が無い high の risk ----------------------------------

sec_risk="$(section_body "$h_risk" "$out")"
assert_has "verified の無い high の risk を挙げる" "^$RISK_NONE"$'\t' "$sec_risk"
assert_has "human: 以外の actor の verified だけの high の risk を挙げる" "^$RISK_AGENT"$'\t' "$sec_risk"
assert_has "generated.at より前の human: の verified だけの high の risk を挙げる" "^$RISK_HUMAN_BEFORE"$'\t' "$sec_risk"
assert_lacks "generated.at より後の human: の verified を持つ high の risk を挙げない" "^$RISK_HUMAN_AFTER" "$sec_risk"
assert_lacks "generated.at と同じ時刻の human: の verified を持つ high の risk を挙げない" "^$RISK_HUMAN_EQUAL" "$sec_risk"
assert_lacks "medium の risk を挙げない" "^$RISK_MEDIUM" "$sec_risk"
assert_lacks "deprecated の high の risk を挙げない" "^$RISK_HIGH_DEP" "$sec_risk"
assert_eq "high の risk の節の行は <フル ID>\\t<name>\\t<file> の形" \
  "$RISK_NONE"$'\t'"確認の無い high の risk"$'\t'"docs/risk-none.md" \
  "$(grep "^$RISK_NONE" <<<"$sec_risk")"

# --- 3. --json -----------------------------------------------------------------

run_fresh env SARA_FRESH_ROOT="$base" sara-fresh --json
json="$stdout_capture"

assert_eq "--json でも exit code は同じ（列挙あり = 1）" 1 "$status_capture"
assert_eq "--json は stale と unverified_high_risk のキーを持つ" "stale,unverified_high_risk" \
  "$(jq -r 'keys | join(",")' <<<"$json" 2>/dev/null)"
assert_eq "stale の要素のキーは file / id / name / upstreams" '["file,id,name,upstreams"]' \
  "$(jq -c '[.stale[] | keys | join(",")] | unique' <<<"$json" 2>/dev/null)"
assert_eq "upstreams の要素のキーは at / id / relation" '["at,id,relation"]' \
  "$(jq -c '[.stale[].upstreams[] | keys | join(",")] | unique' <<<"$json" 2>/dev/null)"
assert_eq "unverified_high_risk の要素のキーは file / id / name" '["file,id,name"]' \
  "$(jq -c '[.unverified_high_risk[] | keys | join(",")] | unique' <<<"$json" 2>/dev/null)"
assert_eq "stale の ID は text と同じ 8 件を ID 順に並べる" "$CASE_OLD,$DSG_OLD,$REQ_STALE,$REQ_VER_BEFORE,$REQ_TWO,$RISK_OLD,$TC_OLD,$UC_OLD" \
  "$(jq -r '[.stale[].id] | join(",")' <<<"$json" 2>/dev/null)"
assert_eq "upstreams は上流の ID・relation・generated.at の元の文字列を持つ" \
  "[{\"id\":\"$UC\",\"relation\":\"derives_from\",\"at\":\"2026-03-01T00:00:00+09:00\"}]" \
  "$(jq -c --arg id "$REQ_STALE" '.stale[] | select(.id == $id) | .upstreams' <<<"$json" 2>/dev/null)"
assert_eq "upstreams の relation は宣言した relation の名前" "covers,follows,satisfies,threatens,mitigates,refines" \
  "$(jq -r '[.stale[] | select(.id | test("^(CASE|DSG|RISK|TC|UC)-")) | .upstreams[].relation] | join(",")' <<<"$json" 2>/dev/null)"
assert_eq "unverified_high_risk の ID は 3 件を ID 順に並べる" "$RISK_NONE,$RISK_AGENT,$RISK_HUMAN_BEFORE" \
  "$(jq -r '[.unverified_high_risk[].id] | join(",")' <<<"$json" 2>/dev/null)"
assert_eq "--json の file は docs/ 前置" "docs/risk-none.md" \
  "$(jq -r '.unverified_high_risk[0].file' <<<"$json" 2>/dev/null)"

# --- 4. 列挙が空なら exit 0 ----------------------------------------------------

clean="$work/clean"
make_graph "$clean"
rm "$clean"/docs/{req-stale,req-ver-before,req-two,risk-old,tc-old,case-old,dsg-old,uc-old,risk-none,risk-agent,risk-human-before}.md

run_fresh env SARA_FRESH_ROOT="$clean" sara-fresh
assert_eq "列挙が空なら exit 0" 0 "$status_capture"
assert_eq "空の節は「なし」と出す" 2 "$(grep -c '^なし$' <<<"$stdout_capture")"

run_fresh env SARA_FRESH_ROOT="$clean" sara-fresh --json
assert_eq "--json でも列挙が空なら exit 0 で 2 配列が空" '0:{"stale":[],"unverified_high_risk":[]}' \
  "$status_capture:$(jq -c '.' <<<"$stdout_capture" 2>/dev/null)"

# --- 5. exit 2 -----------------------------------------------------------------

expect_exit2() {
  local label=$1
  shift
  run_fresh "$@"
  if [[ "$status_capture" -eq 2 && -z "$stdout_capture" ]]; then
    pass "$label は一覧を出さずに exit 2"
  else
    fault "$label は一覧を出さずに exit 2（実際: exit=$status_capture stdout=$stdout_capture）"
  fi
}

broken="$work/broken"
make_graph "$broken"
item "$broken" broken TC-99990009-0000-4000-8000-00000000000b test_condition "壊れた参照を持つテスト条件" <<'YAML'
status: stable
mitigates:
  - "RISK-deadbeef-0000-4000-8000-00000000000c"
generated:
  by: "claude-code/fixture"
  at: "2026-05-01T00:00:00+09:00"
YAML
expect_exit2 "sara check が失敗するグラフ" env SARA_FRESH_ROOT="$broken" sara-fresh

no_generated="$work/no-generated"
make_graph "$no_generated"
sed -i '/^generated:$/,/^  at:/d' "$no_generated/docs/req-fresh.md"
expect_exit2 "generated.at を持たない item" env SARA_FRESH_ROOT="$no_generated" sara-fresh

bad_at="$work/bad-at"
make_graph "$bad_at"
sed -i 's/^  at: "2026-04-01T00:00:00+09:00"$/  at: "2026-04-01 00:00"/' "$bad_at/docs/req-fresh.md"
expect_exit2 "時刻として読めない at" env SARA_FRESH_ROOT="$bad_at" sara-fresh

not_list="$work/not-list"
make_graph "$not_list"
sed -i '/^verified:$/,$d' "$not_list/docs/req-ver-after.md"
printf 'verified: "human:fixture"\n---\n# 確認で基準時刻が進んだ要求\n' >>"$not_list/docs/req-ver-after.md"
expect_exit2 "verified がリストでない item" env SARA_FRESH_ROOT="$not_list" sara-fresh

bad_verified_at="$work/bad-verified-at"
make_graph "$bad_verified_at"
sed -i 's/^    at: "2026-03-15T00:00:00+09:00"$/    at: "soon"/' "$bad_verified_at/docs/req-ver-after.md"
expect_exit2 "verified の at を時刻として読めない item" env SARA_FRESH_ROOT="$bad_verified_at" sara-fresh
assert_has "verified の at を読めない item のファイル名を stderr に挙げる" '^  \./docs/req-ver-after\.md: verified の at' \
  "$(env SARA_FRESH_ROOT="$bad_verified_at" sara-fresh 2>&1 >/dev/null)"

# 偽の sara で JSON の形の異常を再現する。shebang はサンドボックスに /usr/bin/env が無いため実行中の bash の絶対パスを埋め込む。
fake_sara() {
  {
    printf '#!%s\n' "$BASH"
    printf "printf %%s '%s'\n" "$2"
  } >"$1"
  chmod +x "$1"
}

# sara check が受け付けても yq が読めない frontmatter は、偽の sara で item として渡して再現する。
bad_yaml="$work/bad-yaml"
mkdir -p "$bad_yaml/docs"
printf -- '---\nverified: [\n---\n# 壊れた frontmatter\n' >"$bad_yaml/docs/bad.md"
fake_sara "$work/fake-bad-yaml" '{"valid": true, "items": [{"id": "SOL-5011a001-0000-4000-8000-000000000001", "type": "solution", "name": "x", "file": "./docs/bad.md", "relations": {}, "fields": {"status": "stable"}}]}'
expect_exit2 "frontmatter を YAML として読めない item" env SARA_FRESH_ROOT="$bad_yaml" SARA_FRESH_SARA="$work/fake-bad-yaml" sara-fresh
yaml_err="$(env SARA_FRESH_ROOT="$bad_yaml" SARA_FRESH_SARA="$work/fake-bad-yaml" sara-fresh 2>&1 >/dev/null)"
assert_has "YAML として読めない item は stderr でその旨を知らせる" 'YAML として読めない' "$yaml_err"
assert_has "YAML として読めない item の名前を stderr に挙げる" '^  \./docs/bad\.md$' "$yaml_err"

fake_sara "$work/fake-noitems" '{}'
expect_exit2 ".items 配列が無い JSON" env SARA_FRESH_ROOT="$base" SARA_FRESH_SARA="$work/fake-noitems" sara-fresh

fake_sara "$work/fake-invalid" '{"items": [], "valid": false}'
expect_exit2 "exit 0 でも valid: false の JSON" env SARA_FRESH_ROOT="$base" SARA_FRESH_SARA="$work/fake-invalid" sara-fresh

fake_sara "$work/fake-empty" '{"items": [], "valid": true}'
run_fresh env SARA_FRESH_ROOT="$base" SARA_FRESH_SARA="$work/fake-empty" sara-fresh --json
assert_eq "item が 0 件なら exit 0" '0:{"stale":[],"unverified_high_risk":[]}' \
  "$status_capture:$(jq -c '.' <<<"$stdout_capture" 2>/dev/null)"

# --- 6. 走査先をリポジトリルート基準で解決する ---------------------------------

gitrepo="$work/gitrepo"
make_graph "$gitrepo"
git -C "$gitrepo" init -q
mkdir -p "$gitrepo/sub"
sub_status=0
(cd "$gitrepo/sub" && sara-fresh >/dev/null 2>&1) || sub_status=$?
assert_eq "サブディレクトリから叩いても git ルートのグラフを対象にする（列挙あり = exit 1）" 1 "$sub_status"

# --- 7. 引数 -------------------------------------------------------------------

run_fresh sara-fresh --help
assert_eq "--help は exit 0" 0 "$status_capture"

usage_status=0
usage_err="$(sara-fresh --bogus 2>&1 >/dev/null)" || usage_status=$?
assert_eq "未知の引数は exit 2" 2 "$usage_status"
assert_has "未知の引数では usage を出す" '^usage: sara-fresh' "$usage_err"

run_fresh sara-fresh --json extra
assert_eq "引数過多は exit 2" 2 "$status_capture"

exit "$fail"
