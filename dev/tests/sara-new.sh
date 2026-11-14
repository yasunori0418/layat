#!/usr/bin/env bash
# sara-new（sara init を包む item 起票ラッパー）の検証。
#
# 実行:
#   nix develop ./dev -c dev/tests/sara-new.sh   # devShell から直接
#   nix flake check ./dev                         # checks.sara-new 経由
set -uo pipefail

fail=0
pass() { printf 'ok   - %s\n' "$1"; }
fault() {
  printf 'FAIL - %s\n' "$1"
  fail=1
}

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# モデルは実リポジトリの docs/model.yaml。SARA_NEW_MODEL_YAML があればそれを、無ければ git ルート基準で引く。
# どちらでも解決できなければ失敗させる。
model_yaml="${SARA_NEW_MODEL_YAML:-}"
contract_root="$(git rev-parse --show-toplevel 2>/dev/null || printf '.')"
[[ -f "$model_yaml" ]] || model_yaml="$contract_root/docs/model.yaml"
if [[ ! -f "$model_yaml" ]]; then
  fault "docs/model.yaml を解決できない（model.yaml=$model_yaml）"
  exit 1
fi

# sara が動くリポジトリの最小形（sara.toml + model.yaml + 空の docs/）を作る。
# 節ごとに作り直す。
make_repo() {
  local root="$1"
  mkdir -p "$root/docs"
  cp "$model_yaml" "$root/docs/model.yaml"
  cat >"$root/sara.toml" <<'EOF'
model_schema = "docs/model.yaml"

[repositories]
paths = ["./docs"]
EOF
}

uuid_re='[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'

# --- 0. model.yaml の型 ⟷ prefix の一意性 -------------------------------------
#
# 2 つの型が同じ prefix を持たないことを確かめる。抽出は awk で行い、`- id:` の行数と組の数を突合する。
mapfile -t model_pairs < <(
  awk '
    /^item_types:/  { in_types = 1; next }
    /^relations:/   { in_types = 0; next }
    !in_types       { next }
    /^  - id: /     { type = $3; next }
    /^    prefix: / { if (type != "") { print type, $2; type = "" } }
  ' "$model_yaml"
)
model_type_lines="$(awk '
  /^item_types:/ { in_types = 1; next }
  /^relations:/  { in_types = 0; next }
  in_types && /^  - id: / { n++ }
  END { print n + 0 }
' "$model_yaml")"

if [[ "${#model_pairs[@]}" -eq "$model_type_lines" && "$model_type_lines" -gt 0 ]]; then
  pass "model.yaml の全型から prefix を抽出する（$model_type_lines 型）"
else
  fault "model.yaml の全型から prefix を抽出する（型 $model_type_lines 件に対し抽出 ${#model_pairs[@]} 件）"
fi

declare -A model_type_of_prefix=()
prefix_unique=1
for pair in "${model_pairs[@]}"; do
  model_type="${pair%% *}"
  model_p="${pair##* }"
  shared="${model_type_of_prefix[$model_p]:-}"
  if [[ -n "$shared" ]]; then
    fault "model.yaml の prefix $model_p が重複している（${model_type_of_prefix[$model_p]} と $model_type）"
    prefix_unique=0
  fi
  model_type_of_prefix["$model_p"]="$model_type"
done
if [[ "$prefix_unique" -eq 1 && "${#model_pairs[@]}" -gt 0 ]]; then
  pass "model.yaml の prefix が型ごとに一意（型 ⟷ prefix が 1:1）"
fi

# SUT が PATH に無ければ異常系の節が自明に緑になるため、ここで存在を確かめて落とす。
if ! command -v sara-new >/dev/null 2>&1; then
  fault "SUT（sara-new）が PATH に無い。devShell 経由で実行すること"
  exit 1
fi

# SUT を repo のルートで呼ぶ（sara.toml を既定パスで引かせるため）。
# 非ゼロ終了でも落とさず、stderr は fault メッセージ用に拾う。
stdout_capture=""
stderr_capture=""
status_capture=0
run_sara_new() {
  local root="$1"
  shift
  # 固定パスを毎回切って使い回す（直前の 1 回分だけを保持する）。
  local err_file="$work/stderr"
  : >"$err_file"
  stdout_capture="$(cd "$root" && "$@" 2>"$err_file")" && status_capture=0 || status_capture=$?
  stderr_capture="$(cat "$err_file" 2>/dev/null)"
}

field() { printf '%s\n' "$stdout_capture" | sed -n "s/^$1:[[:space:]]*//p"; }

# 起票の必須入力（--description と generated の actor）。§11 だけが外して異常系を確かめる。
desc=(--description "起票テスト")
# section を持つ型（requirement）の起票に渡す節の値。§12 は省略の拒否を確かめるために外す。
# 偽の sara の節は仮ファイルに section を書かないので渡さない。
sec=(--section cli)
generated_by="claude-code/sara-new-test"
export SARA_NEW_GENERATED_BY="$generated_by"

# --- 1. 起票と rename ---------------------------------------------------------

repo="$work/basic"
make_repo "$repo"
run_sara_new "$repo" sara-new requirement lock-ordering docs/requirements -- "${sec[@]}" "${desc[@]}"

created="$(field file)"
created_id="$(field id)"

if [[ "$status_capture" -eq 0 ]]; then
  pass "起票に成功して exit 0"
else
  fault "起票に成功して exit 0（実際: exit=$status_capture 出力: $stdout_capture stderr: $stderr_capture）"
fi

if [[ "$created_id" =~ ^REQ-${uuid_re}$ ]]; then
  pass "採番 ID が <PREFIX>-<フル UUIDv4> 形式"
else
  fault "採番 ID が <PREFIX>-<フル UUIDv4> 形式（実際: $created_id）"
fi

want_file="docs/requirements/lock-ordering.md"

if [[ "$created" == "$want_file" ]]; then
  pass "ファイル名が <slug>.md で、出力の file: と一致する"
else
  fault "ファイル名が <slug>.md（期待: $want_file 実際: $created）"
fi

if [[ -n "$created" && -f "$repo/$created" ]]; then
  pass "その名前のファイルが実在する"
else
  fault "その名前のファイルが実在する（$repo/$created が無い）"
fi

# frontmatter の id が、出力の id: と一致する。
if grep -q "id: \"$created_id\"" "$repo/$created" 2>/dev/null; then
  pass "frontmatter の id が出力の id: と一致する"
else
  fault "frontmatter の id が出力の id: と一致する（id=$created_id）"
fi

# --name を渡さない既定経路で、item の name が仮ファイル名ではなく slug 由来になる。
actual_name="$(sed -n 's/^name: "\(.*\)"$/\1/p' "$repo/$created" 2>/dev/null)"
if [[ "$actual_name" == "lock-ordering" ]]; then
  pass "--name 未指定でも name が slug 由来になる（仮ファイル名が漏れない）"
else
  fault "--name 未指定でも name が slug 由来になる（期待: lock-ordering 実際: $actual_name）"
fi

# 一時ファイル・一時ディレクトリが残っていない。件数と名指しの両方で見る。
leftovers="$(find "$repo/docs/requirements" -name '*.md' -type f | wc -l)"
if [[ "$leftovers" -eq 1 ]]; then
  pass "生成物は 1 ファイルだけ（一時ファイルを残さない）"
else
  fault "生成物は 1 ファイルだけ（実際: $leftovers 件）"
fi

tmp_left="$(find "$repo/docs" -name '.sara-new-*' | wc -l)"
if [[ "$tmp_left" -eq 0 ]]; then
  pass "一時ファイル・一時ディレクトリ（.sara-new-*）を残さない"
else
  fault "一時ファイル・一時ディレクトリ（.sara-new-*）を残さない（実際: $tmp_left 件）"
fi

# --- 2. 出力形式 --------------------------------------------------------------
#
# 採番結果を機械的に拾える 2 行で出す。

# id: と file: を個別に数える。
id_lines="$(printf '%s\n' "$stdout_capture" | grep -c '^id: ')"
file_lines="$(printf '%s\n' "$stdout_capture" | grep -c '^file: ')"
if [[ "$id_lines" -eq 1 && "$file_lines" -eq 1 ]]; then
  pass "id: / file: をそれぞれ 1 行ずつ出力する"
else
  fault "id: / file: をそれぞれ 1 行ずつ出力する（id: $id_lines 行 file: $file_lines 行）"
fi

# --- 3. sara init へのオプション透過 ------------------------------------------
#
# `--` 以降を sara init へ透過する。

repo3="$work/passthru"
make_repo "$repo3"
run_sara_new "$repo3" sara-new requirement passthru docs/requirements -- "${sec[@]}" --name "透過テスト" "${desc[@]}"
passthru_file="$(field file)"
if [[ "$status_capture" -eq 0 ]] && grep -q 'name: "透過テスト"' "$repo3/$passthru_file" 2>/dev/null; then
  pass "-- 以降のオプションを sara init へ透過する"
else
  fault "-- 以降のオプションを sara init へ透過する（exit=$status_capture file=$passthru_file）"
fi

# --- 4. 配置ディレクトリの自動作成と型名の表記ゆれ ----------------------------
#
# 配置ディレクトリを作り、型名のアンダースコア表記（test_case）も受ける。

repo4="$work/mkdir"
make_repo "$repo4"
run_sara_new "$repo4" sara-new test_case first-case docs/test/new-area -- "${desc[@]}"
mkdir_file="$(field file)"
if [[ "$status_capture" -eq 0 && -f "$repo4/$mkdir_file" && "$mkdir_file" == docs/test/new-area/* ]]; then
  pass "存在しない配置ディレクトリを作って起票する（型名はアンダースコア表記）"
else
  fault "存在しない配置ディレクトリを作って起票する（exit=$status_capture file=$mkdir_file）"
fi

# ハイフン表記（sara init のサブコマンド名そのもの）も同じ型として通る。
run_sara_new "$repo4" sara-new test-case second-case docs/test/new-area -- "${desc[@]}"
hyphen_file="$(field file)"
hyphen_id="$(field id)"
if [[ "$status_capture" -eq 0 && "$hyphen_id" == CASE-* ]]; then
  pass "型名のハイフン表記も同じ型として通る（test-case → CASE）"
else
  fault "型名のハイフン表記も同じ型として通る（exit=$status_capture id=$hyphen_id file=$hyphen_file）"
fi

# --- 5. slug の検査 -----------------------------------------------------------
#
# 検査は起票の前に行い、失敗時にファイルを残さない。

# 無効側の同値クラスは不正文字ごとに分ける。index は生成物の index.md と同じ名前になるので拒否する。
repo5="$work/slug"
make_repo "$repo5"
for bad in "" "has space" "../escape" "a/b" "UPPER" "under_score" "dot.ted" "index"; do
  run_sara_new "$repo5" sara-new requirement "$bad" docs/requirements -- "${sec[@]}" "${desc[@]}"
  bad_status="$status_capture"
  bad_files="$(find "$repo5/docs" -name '*.md' -type f | wc -l)"
  if [[ "$bad_status" -eq 2 && "$bad_files" -eq 0 ]]; then
    pass "不正な slug '$bad' を exit 2 で拒否し、ファイルを残さない"
  else
    fault "不正な slug '$bad' を exit 2 で拒否し、ファイルを残さない（exit=$bad_status 残 $bad_files 件）"
  fi
done

# 有効側（英小数字とハイフン）は通し、その slug がファイル名へ入ることまで見る。
run_sara_new "$repo5" sara-new requirement a1-b2 docs/requirements -- "${sec[@]}" "${desc[@]}"
valid_slug_file="$(field file)"
if [[ "$status_capture" -eq 0 && "$valid_slug_file" == docs/requirements/a1-b2.md ]]; then
  pass "英小文字・数字・ハイフンの slug を受理し、ファイル名へ入れる（境界の有効側）"
else
  fault "英小文字・数字・ハイフンの slug を受理し、ファイル名へ入れる（exit=$status_capture file=$valid_slug_file）"
fi

# 予約は index の完全一致だけで、index で始まる slug は通す。
run_sara_new "$repo5" sara-new requirement index-1 docs/requirements -- "${sec[@]}" "${desc[@]}"
index_prefix_file="$(field file)"
if [[ "$status_capture" -eq 0 && "$index_prefix_file" == docs/requirements/index-1.md ]]; then
  pass "index で始まる slug は受理する（予約の境界の有効側）"
else
  fault "index で始まる slug は受理する（exit=$status_capture file=$index_prefix_file）"
fi

# --- 6. ADR の型 ---------------------------------------------------------------
#
# adr と sara_adr は UUID で起票する。

repo6="$work/adr"
make_repo "$repo6"
for spec in "adr ADR docs/adr adr-decision" "sara_adr SDR docs/sara-adr underscore-decision" "sara-adr SDR docs/sara-adr hyphen-decision"; do
  read -r adr_type adr_prefix adr_dir adr_slug <<<"$spec"
  run_sara_new "$repo6" sara-new "$adr_type" "$adr_slug" "$adr_dir" -- --status 採用 "${desc[@]}"
  adr_id="$(field id)"
  adr_file="$(field file)"
  if [[ "$status_capture" -eq 0 && "$adr_id" =~ ^${adr_prefix}-${uuid_re}$ \
    && "$adr_file" == "$adr_dir/$adr_slug.md" && -f "$repo6/$adr_file" ]]; then
    pass "$adr_type を ${adr_prefix}-<フル UUIDv4> で起票する"
  else
    fault "$adr_type を ${adr_prefix}-<フル UUIDv4> で起票する（exit=$status_capture id=$adr_id file=$adr_file stderr: $stderr_capture）"
  fi
done

# --- 7. sara init の失敗を伝播する --------------------------------------------
#
# sara 呼び出しは seam（SARA_NEW_SARA）経由で差し替える。

fake_sara="$work/fake-sara"
# shebang はサンドボックスに /usr/bin/env が無いため実行中の bash の絶対パスを埋め込む。
# 偽 sara は SUT の呼び出し形と仮ファイルの stem も検査し、失敗する前に仮ファイルを作る。
{
  printf '#!%s\n' "$BASH"
  cat <<'FAKE'
# 引数契約: sara-new は必ず `--no-color --no-emoji init <型> <仮パス> [透過分...]`
# の順で呼ぶ。違えば偽 sara 自身が非ゼロで落ちて、テスト側の期待コードと食い違う。
if [ "$1" != "--no-color" ] || [ "$2" != "--no-emoji" ] || [ "$3" != "init" ]; then
  echo "fake sara: 想定外の呼ばれ方: $*" >&2
  exit 90
fi
if [ "$4" != "$SARA_NEW_FAKE_WANT_TYPE" ]; then
  echo "fake sara: 型が想定と違う（期待 $SARA_NEW_FAKE_WANT_TYPE 実際 $4）" >&2
  exit 91
fi

# 仮ファイル（第 5 引数）を実際に作る。
tmp_path=$5
if [ -z "$tmp_path" ]; then
  echo "fake sara: 仮ファイルのパスが渡っていない" >&2
  exit 92
fi

# stem が slug と一致すること（name の由来なので規約の一部）。
tmp_stem=${tmp_path##*/}
tmp_stem=${tmp_stem%.md}
if [ "$tmp_stem" != "$SARA_NEW_FAKE_WANT_STEM" ]; then
  echo "fake sara: 仮ファイルの stem が想定と違う（期待 $SARA_NEW_FAKE_WANT_STEM 実際 $tmp_stem）" >&2
  exit 94
fi

: >"$tmp_path"

case "${SARA_NEW_FAKE_MODE:-}" in
  fail)
    echo "fake sara: boom" >&2
    exit 3
    ;;
  no-id)
    echo "[OK] Created something with Requirement template"
    exit 0
    ;;
  id)
    printf '  ID:   %s\n' "$SARA_NEW_FAKE_ID"
    exit 0
    ;;
  unclosed-id)
    # frontmatter を開いて閉じない仮ファイルを作り、ID だけは正しく返す。
    printf -- '---\nid: "%s"\n' "$SARA_NEW_FAKE_ID" >"$tmp_path"
    printf '  ID:   %s\n' "$SARA_NEW_FAKE_ID"
    exit 0
    ;;
  race)
    # 正しい仮ファイルを作り、その間に別の起票が同じ出力先を置いた状況を作る。
    printf -- '---\nid: "%s"\n---\n' "$SARA_NEW_FAKE_ID" >"$tmp_path"
    printf '先着\n' >"$SARA_NEW_FAKE_RACE_TARGET"
    printf '  ID:   %s\n' "$SARA_NEW_FAKE_ID"
    exit 0
    ;;
  *)
    # モード指定漏れを成功扱いにしない。
    echo "fake sara: 未知のモード: ${SARA_NEW_FAKE_MODE:-（未設定）}" >&2
    exit 93
    ;;
esac
FAKE
} >"$fake_sara"
chmod +x "$fake_sara"

repo7="$work/initfail"
make_repo "$repo7"
run_sara_new "$repo7" env SARA_NEW_SARA="$fake_sara" SARA_NEW_FAKE_MODE=fail \
  SARA_NEW_FAKE_WANT_TYPE=requirement SARA_NEW_FAKE_WANT_STEM=boom \
  sara-new requirement boom docs/requirements -- "${desc[@]}"
initfail_status="$status_capture"
initfail_files="$(find "$repo7/docs" -name '*.md' -type f | wc -l)"
initfail_tmp="$(find "$repo7/docs" -name '.sara-new-*' | wc -l)"
# 期待コードは具体値で押さえる。SUT は偽 sara の exit 3 をそのまま伝播する。
if [[ "$initfail_status" -eq 3 ]]; then
  pass "sara init の失敗（exit 3）をそのまま伝播する"
else
  fault "sara init の失敗（exit 3）をそのまま伝播する（実際: exit=$initfail_status stderr: $stderr_capture）"
fi
if [[ "$initfail_files" -eq 0 && "$initfail_tmp" -eq 0 ]]; then
  pass "sara init 失敗時に一時ファイル・一時ディレクトリを残さない"
else
  fault "sara init 失敗時に一時ファイル・一時ディレクトリを残さない（md $initfail_files 件 / tmp $initfail_tmp 件）"
fi

# --- 8. ID を読めなければ失敗する ---------------------------------------------
#
# sara の出力から ID を読めなければ rename せず落ちる。

repo8="$work/noid"
make_repo "$repo8"
run_sara_new "$repo8" env SARA_NEW_SARA="$fake_sara" SARA_NEW_FAKE_MODE=no-id \
  SARA_NEW_FAKE_WANT_TYPE=requirement SARA_NEW_FAKE_WANT_STEM=noid \
  sara-new requirement noid docs/requirements -- "${desc[@]}"
noid_status="$status_capture"
noid_files="$(find "$repo8/docs" -name '*.md' -type f | wc -l)"
noid_tmp="$(find "$repo8/docs" -name '.sara-new-*' | wc -l)"
# ID を読めない経路は SUT 自身の判断で落ちるので exit 1（具体値で押さえる）。
if [[ "$noid_status" -eq 1 ]]; then
  pass "sara init の出力から ID を読めなければ exit 1 で落ちる"
else
  fault "sara init の出力から ID を読めなければ exit 1 で落ちる（実際: exit=$noid_status stderr: $stderr_capture）"
fi
if [[ "$noid_files" -eq 0 && "$noid_tmp" -eq 0 ]]; then
  pass "ID 読み取り失敗時に一時ファイル・一時ディレクトリを残さない"
else
  fault "ID 読み取り失敗時に一時ファイル・一時ディレクトリを残さない（md $noid_files 件 / tmp $noid_tmp 件）"
fi

# --- 8c. frontmatter が無い・閉じていなければ失敗する -------------------------
#
# generated を足す先（閉じの `---`）が無ければ exit 1 で落ちる。id モードは空の仮ファイル、
# unclosed-id モードは開きの `---` だけを持つ仮ファイルを作る。

for nofm_mode in id unclosed-id; do
  repo8c="$work/nofm-$nofm_mode"
  make_repo "$repo8c"
  run_sara_new "$repo8c" env SARA_NEW_SARA="$fake_sara" SARA_NEW_FAKE_MODE="$nofm_mode" \
    SARA_NEW_FAKE_ID='REQ-11111111-2222-4333-8444-555555555555' \
    SARA_NEW_FAKE_WANT_TYPE=requirement SARA_NEW_FAKE_WANT_STEM=nofm \
    sara-new requirement nofm docs/requirements -- "${desc[@]}"
  nofm_files="$(find "$repo8c/docs" -name '*.md' -type f | wc -l)"
  nofm_tmp="$(find "$repo8c/docs" -name '.sara-new-*' | wc -l)"
  if [[ "$status_capture" -eq 1 && "$stderr_capture" == *"frontmatter の閉じ"* ]]; then
    pass "frontmatter の閉じが無ければ exit 1 で落ち、理由を出す（$nofm_mode）"
  else
    fault "frontmatter の閉じが無ければ exit 1 で落ち、理由を出す（$nofm_mode。実際: exit=$status_capture stderr: $stderr_capture）"
  fi
  if [[ "$nofm_files" -eq 0 && "$nofm_tmp" -eq 0 ]]; then
    pass "frontmatter の閉じが無いときにファイル・一時ディレクトリを残さない（$nofm_mode）"
  else
    fault "frontmatter の閉じが無いときにファイル・一時ディレクトリを残さない（$nofm_mode。md $nofm_files 件 / tmp $nofm_tmp 件）"
  fi
done

# --- 9. 同じ slug の再起票を拒否する -------------------------------------------
#
# 同じディレクトリに同じ slug で 2 回起票すると、2 回目は既存ファイルを上書きせず落ちる。

repo9="$work/clobber"
make_repo "$repo9"
run_sara_new "$repo9" sara-new requirement dup docs/requirements -- "${sec[@]}" "${desc[@]}"
first_file="$(field file)"
if [[ -z "$first_file" ]]; then
  fault "§9 の前提（1 件目の起票）が失敗した（$stdout_capture）"
else
  first_content="$(cat "$repo9/$first_file")"
  run_sara_new "$repo9" sara-new requirement dup docs/requirements -- "${sec[@]}" "${desc[@]}"
  clobber_status="$status_capture"
  clobber_files="$(find "$repo9/docs/requirements" -name '*.md' -type f | wc -l)"
  clobber_tmp="$(find "$repo9/docs" -name '.sara-new-*' | wc -l)"

  if [[ "$clobber_status" -eq 1 ]]; then
    pass "同じ slug の 2 回目の起票は exit 1 で拒否する"
  else
    fault "同じ slug の 2 回目の起票は exit 1 で拒否する（実際: exit=$clobber_status stderr: $stderr_capture）"
  fi
  if [[ "$clobber_files" -eq 1 && "$clobber_tmp" -eq 0 && "$(cat "$repo9/$first_file")" == "$first_content" ]]; then
    pass "既存ファイルを上書きせず、一時ファイル・一時ディレクトリも残さない"
  else
    fault "既存ファイルを上書きせず、一時ファイル・一時ディレクトリも残さない（md $clobber_files 件 / tmp $clobber_tmp 件）"
  fi
fi

# 起票の最中に同じ出力先が現れても、先に置かれたファイルを上書きせず落ちる。
repo9b="$work/race"
make_repo "$repo9b"
run_sara_new "$repo9b" env SARA_NEW_SARA="$fake_sara" SARA_NEW_FAKE_MODE=race \
  SARA_NEW_FAKE_ID='REQ-11111111-2222-4333-8444-555555555555' \
  SARA_NEW_FAKE_RACE_TARGET=docs/requirements/race.md \
  SARA_NEW_FAKE_WANT_TYPE=requirement SARA_NEW_FAKE_WANT_STEM=race \
  sara-new requirement race docs/requirements -- "${desc[@]}"
race_status="$status_capture"
race_content="$(cat "$repo9b/docs/requirements/race.md" 2>/dev/null)"
race_tmp="$(find "$repo9b/docs" -name '.sara-new-*' | wc -l)"
if [[ "$race_status" -eq 1 && "$race_content" == "先着" && "$race_tmp" -eq 0 ]]; then
  pass "起票の最中に現れた出力先を上書きせず exit 1 で落ち、一時ディレクトリも残さない"
else
  fault "起票の最中に現れた出力先を上書きせず exit 1 で落ちる（exit=$race_status 内容: $race_content tmp $race_tmp 件）"
fi

# --- 10. 引数の異常系 ---------------------------------------------------------

repo10="$work/args"
make_repo "$repo10"

run_sara_new "$repo10" sara-new
if [[ "$status_capture" -eq 2 ]]; then
  pass "引数なしは exit 2"
else
  fault "引数なしは exit 2（実際: exit=$status_capture）"
fi

# 境界は 2/3 の間（型・slug・dir の 3 個が必須）。無効側を押さえる。
run_sara_new "$repo10" sara-new requirement only-slug
if [[ "$status_capture" -eq 2 ]]; then
  pass "配置ディレクトリを省くと exit 2（境界の無効側）"
else
  fault "配置ディレクトリを省くと exit 2（実際: exit=$status_capture）"
fi

# `--` 区切り無しの余分な引数は受け付けない。
run_sara_new "$repo10" sara-new requirement stray-arg docs/requirements extra
stray_status="$status_capture"
stray_files="$(find "$repo10/docs" -name '*.md' -type f | wc -l)"
if [[ "$stray_status" -eq 2 && "$stray_files" -eq 0 ]]; then
  pass "-- 区切り無しの余分な引数は exit 2 で拒否し、ファイルを残さない"
else
  fault "-- 区切り無しの余分な引数は exit 2 で拒否し、ファイルを残さない（exit=$stray_status 残 $stray_files 件）"
fi

run_sara_new "$repo10" env -u SARA_NEW_GENERATED_BY sara-new --help
if [[ "$status_capture" -eq 0 ]]; then
  pass "--help は exit 0（SARA_NEW_GENERATED_BY が無くても）"
else
  fault "--help は exit 0（実際: exit=$status_capture）"
fi

# --- 11. --description と generated ------------------------------------------
#
# 必須入力が欠けたら起票の前に exit 2 で拒否し、配置ディレクトリも作らない。

repo11="$work/okf"
make_repo "$repo11"

# 拒否の判定。exit 2・ファイルも配置ディレクトリも残さない・stderr が理由を名指しする。
expect_rejected() {
  local label=$1 dir=$2 reason=$3
  local made_dir=0
  [[ -e "$repo11/$dir" ]] && made_dir=1
  if [[ "$status_capture" -eq 2 && "$made_dir" -eq 0 && "$stderr_capture" == *"$reason"* ]]; then
    pass "$label"
  else
    fault "$label（exit=$status_capture 配置ディレクトリ作成=$made_dir stderr: $stderr_capture）"
  fi
}

run_sara_new "$repo11" sara-new requirement no-desc docs/no-desc
expect_rejected "-- が無く --description も無ければ exit 2 で拒否する" docs/no-desc description

run_sara_new "$repo11" sara-new requirement no-desc docs/no-desc-2 -- "${sec[@]}" --name "要約なし"
expect_rejected "-- 以降に --description が無ければ exit 2 で拒否する" docs/no-desc-2 description

run_sara_new "$repo11" env -u SARA_NEW_GENERATED_BY \
  sara-new requirement no-actor docs/no-actor -- "${sec[@]}" "${desc[@]}"
expect_rejected "SARA_NEW_GENERATED_BY が無ければ exit 2 で拒否する" docs/no-actor SARA_NEW_GENERATED_BY

run_sara_new "$repo11" env SARA_NEW_GENERATED_BY= \
  sara-new requirement empty-actor docs/empty-actor -- "${sec[@]}" "${desc[@]}"
expect_rejected "SARA_NEW_GENERATED_BY が空なら exit 2 で拒否する" docs/empty-actor SARA_NEW_GENERATED_BY

run_sara_new "$repo11" env 'SARA_NEW_GENERATED_BY=claude-code/"quoted"' \
  sara-new requirement quoted-actor docs/quoted-actor -- "${sec[@]}" "${desc[@]}"
expect_rejected "SARA_NEW_GENERATED_BY が二重引用符を含めば exit 2 で拒否する" docs/quoted-actor SARA_NEW_GENERATED_BY

# --description の別表記（sara init が受ける -d / --description=…）も指定として通る。
spelled_n=0
for spelling in "-d|起票テスト" "--description=起票テスト"; do
  if [[ "$spelling" == *"|"* ]]; then
    spelled=("${spelling%%|*}" "${spelling#*|}")
  else
    spelled=("$spelling")
  fi
  spelled_n=$((spelled_n + 1))
  run_sara_new "$repo11" sara-new requirement "spelled-$spelled_n" docs/requirements -- "${sec[@]}" "${spelled[@]}"
  spelled_file="$(field file)"
  if [[ "$status_capture" -eq 0 ]] && grep -q '^description: "起票テスト"$' "$repo11/$spelled_file" 2>/dev/null; then
    pass "--description の別表記 ${spelled[0]} を受け、description が入る"
  else
    fault "--description の別表記 ${spelled[0]} を受け、description が入る（exit=$status_capture file=$spelled_file stderr: $stderr_capture）"
  fi
done

# §1 で起票した item の frontmatter の末尾 3 行が generated ブロックになっている。
mapfile -t fm_tail < <(awk 'NR == 1 { next } /^---$/ { exit } { print }' "$repo/$created" 2>/dev/null | tail -n 3)
iso_re='^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[+-][0-9]{2}:[0-9]{2}$'
gen_at="$(printf '%s' "${fm_tail[2]:-}" | sed -n 's/^  at: "\(.*\)"$/\1/p')"
if [[ "${fm_tail[0]:-}" == "generated:" && "${fm_tail[1]:-}" == "  by: \"$generated_by\"" && -n "$gen_at" ]]; then
  pass "frontmatter の末尾に generated.by（SARA_NEW_GENERATED_BY の値）と generated.at を付ける"
else
  fault "frontmatter の末尾に generated.by と generated.at を付ける（実際: ${fm_tail[*]:-なし}）"
fi
if [[ "$gen_at" =~ $iso_re ]]; then
  pass "generated.at が UTC オフセット付きの ISO 8601（秒まで）"
else
  fault "generated.at が UTC オフセット付きの ISO 8601（秒まで）（実際: ${gen_at:-空}）"
fi

# generated を足した frontmatter を sara がそのまま読める。
if (cd "$repo" && sara --no-color --no-emoji check >/dev/null 2>&1); then
  pass "generated を付けた item を sara check が受け付ける"
else
  fault "generated を付けた item を sara check が受け付ける（$(cd "$repo" && sara --no-color --no-emoji check 2>&1 | tail -n 5)）"
fi

# --- 12. section を持つ型の --section ------------------------------------------
#
# sara init は --section が無くても enum の先頭の値を入れるため、sara-new が省略を拒否する。

repo12="$work/section"
make_repo "$repo12"

run_sara_new "$repo12" sara-new requirement no-section docs/requirements -- "${desc[@]}"
nosec_files="$(find "$repo12/docs" -name '*.md' -type f | wc -l)"
nosec_tmp="$(find "$repo12/docs" -name '.sara-new-*' | wc -l)"
if [[ "$status_capture" -eq 2 && "$stderr_capture" == *"--section"* && "$stderr_capture" == *"docs/model.yaml"* ]]; then
  pass "section を持つ型を --section なしで起票すると exit 2 で拒否し、理由を出す"
else
  fault "section を持つ型を --section なしで起票すると exit 2 で拒否し、理由を出す（exit=$status_capture stderr: $stderr_capture）"
fi
if [[ "$nosec_files" -eq 0 && "$nosec_tmp" -eq 0 ]]; then
  pass "--section の省略を拒否したとき、ファイルも一時ディレクトリも残さない"
else
  fault "--section の省略を拒否したとき、ファイルも一時ディレクトリも残さない（md $nosec_files 件 / tmp $nosec_tmp 件）"
fi

# --section <値> と --section=<値> のどちらの綴りでも起票でき、その値が入る。
sec_n=0
for spelling in "--section|cli" "--section=cli"; do
  if [[ "$spelling" == *"|"* ]]; then
    sec_spelled=("${spelling%%|*}" "${spelling#*|}")
  else
    sec_spelled=("$spelling")
  fi
  sec_n=$((sec_n + 1))
  run_sara_new "$repo12" sara-new requirement "with-section-$sec_n" docs/requirements -- "${sec_spelled[@]}" "${desc[@]}"
  sec_file="$(field file)"
  if [[ "$status_capture" -eq 0 ]] && grep -q '^section: cli$' "$repo12/$sec_file" 2>/dev/null; then
    pass "--section の綴り '$spelling' で起票でき、section が入る"
  else
    fault "--section の綴り '$spelling' で起票でき、section が入る（exit=$status_capture file=$sec_file stderr: $stderr_capture）"
  fi
done

# section を持たない型は --section なしで起票できる。
run_sara_new "$repo12" sara-new test_case no-section-case docs/test/section -- "${desc[@]}"
nosec_case_file="$(field file)"
if [[ "$status_capture" -eq 0 && -f "$repo12/$nosec_case_file" ]]; then
  pass "section を持たない型（test_case）は --section なしで起票できる"
else
  fault "section を持たない型（test_case）は --section なしで起票できる（exit=$status_capture file=$nosec_case_file stderr: $stderr_capture）"
fi

exit "$fail"
