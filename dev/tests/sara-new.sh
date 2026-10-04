#!/usr/bin/env bash
# sara-new（sara init を包む item 起票ラッパー）の検証。
#
# 実行:
#   nix develop ./dev -c dev/tests/sara-new.sh   # devShell から直接
#   nix flake check ./dev                         # checks.sara-new 経由
#
# 検証対象。番号は下の節見出しに対応する:
#   0.  docs/model.yaml の型 ⟷ prefix が 1:1（SUT ではなくモデル側の不変条件）
#   1.  item を起票し `<YYYYMMDD>-<フル UUID>-<slug>.md` へ rename する。
#       frontmatter の id と、ファイル名の UUID 部が一致する
#   1b. --name 未指定の既定経路でも name が slug 由来になる（仮ファイル名が漏れない）
#   1c. 一時ファイル・一時ディレクトリ（.sara-new-*）を残さない
#   2.  採番 ID とファイルパスを機械可読な 2 行（id: / file:）で出力する
#   3.  sara init へオプションを透過する（-- 以降）
#   4.  配置ディレクトリを作る（無ければ mkdir -p）
#   4b. 型名はアンダースコア表記（model.yaml・規約文書）とハイフン表記
#       （sara init のサブコマンド名）の両方を受ける
#   5.  slug の検査（空・不正文字は exit 2 で、ファイルを残さない）
#   5b. 英小文字・数字・ハイフンの slug を受理し、ファイル名へ入れる（境界の有効側）
#   6.  ADR は連番維持のため exit 2 で拒否する
#   7.  sara init の失敗（exit 3）をそのまま伝播し、一時ファイルを残さない（seam で再現）
#   8.  sara init の出力から ID を読めなければ exit 1 で落ち、一時ファイルを残さない（seam）
#   8b. 採番 ID から UUID 部を取り出せなければ exit 1 で落ちる（prefix がハイフンを
#       含む型。正常系では踏まない経路なので seam で押さえる）
#   9.  出力先が既存なら上書きせず exit 1（起票済み item を潰さない）
#   10. 引数の異常系（引数不足 = 2 / -- 区切り無しの余分引数 = 2 / --help = 0）
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
  if [[ -n "${model_type_of_prefix[$model_p]:-}" ]]; then
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

# --- 1. 起票と rename ---------------------------------------------------------

repo="$work/basic"
make_repo "$repo"
# 日付は SUT 実行の前後で採る（日付の跨ぎに備える）。
date_before="$(date +%Y%m%d)"
run_sara_new "$repo" sara-new requirement lock-ordering docs/requirements
date_after="$(date +%Y%m%d)"

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

uuid="${created_id#REQ-}"
want_before="docs/requirements/$date_before-$uuid-lock-ordering.md"
want_after="docs/requirements/$date_after-$uuid-lock-ordering.md"

if [[ "$created" == "$want_before" || "$created" == "$want_after" ]]; then
  pass "ファイル名が <YYYYMMDD>-<フル UUID>-<slug>.md で、出力の file: と一致する"
else
  fault "ファイル名が <YYYYMMDD>-<フル UUID>-<slug>.md（期待: $want_before 実際: $created）"
fi

if [[ -n "$created" && -f "$repo/$created" ]]; then
  pass "その名前のファイルが実在する"
else
  fault "その名前のファイルが実在する（$repo/$created が無い）"
fi

# frontmatter の id と、ファイル名に埋めた UUID が一致する。
if grep -q "id: \"$created_id\"" "$repo/$created" 2>/dev/null; then
  pass "frontmatter の id とファイル名の UUID が同じ item を指す"
else
  fault "frontmatter の id とファイル名の UUID が同じ item を指す（id=$created_id）"
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
run_sara_new "$repo3" sara-new requirement passthru docs/requirements -- --name "透過テスト"
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
run_sara_new "$repo4" sara-new test_case first-case docs/test/new-area
mkdir_file="$(field file)"
if [[ "$status_capture" -eq 0 && -f "$repo4/$mkdir_file" && "$mkdir_file" == docs/test/new-area/* ]]; then
  pass "存在しない配置ディレクトリを作って起票する（型名はアンダースコア表記）"
else
  fault "存在しない配置ディレクトリを作って起票する（exit=$status_capture file=$mkdir_file）"
fi

# ハイフン表記（sara init のサブコマンド名そのもの）も同じ型として通る。
run_sara_new "$repo4" sara-new test-case second-case docs/test/new-area
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

# 無効側の同値クラスは不正文字ごとに分ける。
repo5="$work/slug"
make_repo "$repo5"
for bad in "" "has space" "../escape" "a/b" "UPPER" "under_score" "dot.ted"; do
  run_sara_new "$repo5" sara-new requirement "$bad" docs/requirements
  bad_status="$status_capture"
  bad_files="$(find "$repo5/docs" -name '*.md' -type f | wc -l)"
  if [[ "$bad_status" -eq 2 && "$bad_files" -eq 0 ]]; then
    pass "不正な slug '$bad' を exit 2 で拒否し、ファイルを残さない"
  else
    fault "不正な slug '$bad' を exit 2 で拒否し、ファイルを残さない（exit=$bad_status 残 $bad_files 件）"
  fi
done

# 有効側（英小数字とハイフン）は通し、その slug がファイル名へ入ることまで見る。
run_sara_new "$repo5" sara-new requirement a1-b2 docs/requirements
valid_slug_file="$(field file)"
if [[ "$status_capture" -eq 0 && "$valid_slug_file" == *-a1-b2.md ]]; then
  pass "英小文字・数字・ハイフンの slug を受理し、ファイル名へ入れる（境界の有効側）"
else
  fault "英小文字・数字・ハイフンの slug を受理し、ファイル名へ入れる（exit=$status_capture file=$valid_slug_file）"
fi

# --- 6. ADR は連番維持のため拒否する ------------------------------------------
#
# ADR だけ id_format が {prefix}-{seq:04} で、`sara init adr` を直接使う。

repo6="$work/adr"
make_repo "$repo6"
run_sara_new "$repo6" sara-new adr some-decision docs/adr
adr_status="$status_capture"
adr_files="$(find "$repo6/docs" -name '*.md' -type f | wc -l)"
if [[ "$adr_status" -eq 2 ]]; then
  pass "ADR は exit 2 で拒否する（連番維持）"
else
  fault "ADR は exit 2 で拒否する（実際: exit=$adr_status）"
fi
if [[ "$adr_files" -eq 0 ]]; then
  pass "ADR 拒否時はファイルを作らない"
else
  fault "ADR 拒否時はファイルを作らない（実際: $adr_files 件）"
fi

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
  sara-new requirement boom docs/requirements
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
  sara-new requirement noid docs/requirements
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

# --- 8b. UUID 部を取り出せなければ失敗する ------------------------------------
#
# prefix がハイフンを含む型（TEST-CASE 等）では UUID 部が `CASE-<uuid>` になり、形の検査で落ちる。

repo8b="$work/badprefix"
make_repo "$repo8b"
run_sara_new "$repo8b" env SARA_NEW_SARA="$fake_sara" SARA_NEW_FAKE_MODE=id \
  SARA_NEW_FAKE_ID='TEST-CASE-11111111-2222-4333-8444-555555555555' \
  SARA_NEW_FAKE_WANT_TYPE=requirement SARA_NEW_FAKE_WANT_STEM=badprefix \
  sara-new requirement badprefix docs/requirements
badprefix_status="$status_capture"
badprefix_files="$(find "$repo8b/docs" -name '*.md' -type f | wc -l)"
badprefix_tmp="$(find "$repo8b/docs" -name '.sara-new-*' | wc -l)"
if [[ "$badprefix_status" -eq 1 ]]; then
  pass "UUID 部を取り出せない ID なら exit 1 で落ちる"
else
  fault "UUID 部を取り出せない ID なら exit 1 で落ちる（実際: exit=$badprefix_status stderr: $stderr_capture）"
fi
if [[ "$badprefix_files" -eq 0 && "$badprefix_tmp" -eq 0 ]]; then
  pass "UUID 部の検査で落ちたときにファイル・一時ディレクトリを残さない"
else
  fault "UUID 部の検査で落ちたときにファイル・一時ディレクトリを残さない（md $badprefix_files 件 / tmp $badprefix_tmp 件）"
fi

# --- 9. 既存ファイルを上書きしない --------------------------------------------
#
# rename 先が既存なら上書きせず落ちる。

repo9="$work/clobber"
make_repo "$repo9"
run_sara_new "$repo9" sara-new requirement dup docs/requirements
first_file="$(field file)"
if [[ -z "$first_file" ]]; then
  fault "§9 の前提（1 件目の起票）が失敗した（$stdout_capture）"
else
  # 1 件目と同じ ID を返す偽 sara（id モード）で rename 先を衝突させる。
  first_id="$(sed -n 's/^id: "\(.*\)"$/\1/p' "$repo9/$first_file")"
  run_sara_new "$repo9" env SARA_NEW_SARA="$fake_sara" SARA_NEW_FAKE_MODE=id \
    SARA_NEW_FAKE_ID="$first_id" SARA_NEW_FAKE_WANT_TYPE=requirement \
    SARA_NEW_FAKE_WANT_STEM=dup \
    sara-new requirement dup docs/requirements
  clobber_status="$status_capture"
  clobber_files="$(find "$repo9/docs/requirements" -name '*.md' -type f | wc -l)"
  clobber_tmp="$(find "$repo9/docs" -name '.sara-new-*' | wc -l)"

  # 1 件目と 2 件目の間で日付が変わると衝突が成立しないため、判定を見送る。
  first_date="${first_file##*/}"
  first_date="${first_date%%-*}"
  if [[ "$first_date" != "$(date +%Y%m%d)" ]]; then
    pass "§9 は日付が跨いだため判定を見送る（衝突の前提が崩れる）"
  elif [[ "$clobber_status" -eq 1 ]]; then
    pass "rename 先が既存なら exit 1 で拒否する"
    if [[ "$clobber_files" -eq 1 && "$clobber_tmp" -eq 0 ]]; then
      pass "既存ファイルを上書きせず、一時ファイル・一時ディレクトリも残さない"
    else
      fault "既存ファイルを上書きせず、一時ファイル・一時ディレクトリも残さない（md $clobber_files 件 / tmp $clobber_tmp 件）"
    fi
  else
    fault "rename 先が既存なら exit 1 で拒否する（実際: exit=$clobber_status stderr: $stderr_capture）"
  fi
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

run_sara_new "$repo10" sara-new --help
if [[ "$status_capture" -eq 0 ]]; then
  pass "--help は exit 0"
else
  fault "--help は exit 0（実際: exit=$status_capture）"
fi

exit "$fail"
