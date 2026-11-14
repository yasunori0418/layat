#!/usr/bin/env bash
# OKF の index.md（dev/scripts/okf-index.sh の生成物）の契約テスト。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/okf-index.sh   # devShell / CI から直接
#   nix flake check ./dev                                     # checks.okf-index 経由

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

require_commands "okf-index.sh（nix develop '.?dir=dev#sara' から実行する）" sara jq git || exit 1

script_dir=$(cd "$(dirname "$0")" && pwd)
generator="$script_dir/../scripts/okf-index.sh"
repo_root=$(cd "$(testdoc_repo_root)" && pwd)

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# --- (1) コミット済みの index.md が生成結果と一致する -------------------------

if (cd "$repo_root" && bash "$generator" --check) > "$work/check.out" 2>&1; then
  pass "コミット済みの index.md が生成結果と一致する"
else
  fault "(1) index.md が生成結果と食い違う（dev/scripts/okf-index.sh で再生成する）"
  sed 's/^/    /' "$work/check.out"
fi

# --- 小さな入力の組み立て ------------------------------------------------------

# 実物の docs/model.yaml を重ね、上流を持たない item とテスト条件 1 本だけのグラフを作る。
fixture="$work/fixture"
mkdir -p "$fixture/docs/adr" "$fixture/docs/sara-adr" "$fixture/docs/ubiquitous-language" "$fixture/docs/requirements" "$fixture/docs/test/copy"
cp "$repo_root/docs/model.yaml" "$fixture/docs/model.yaml"
cat > "$fixture/sara.toml" <<'EOF'
model_schema = "docs/model.yaml"

[repositories]
paths = ["./docs"]

[validation]
strict_mode = false
EOF

# description フィールドを持つ ADR。本文の段落より description を優先する。
cat > "$fixture/docs/adr/with-field.md" <<'EOF'
---
id: "ADR-11111111-1111-4111-8111-111111111111"
type: adr
name: "フィールドを持つ決定"
description: "フィールドの要約"
status: stable
---
# ADR-11111111-1111-4111-8111-111111111111: フィールドを持つ決定

本文の段落は使わない。
EOF

# description が無く、本文の箇条書き・見出し・コードフェンスの後に段落がある ADR。段落は箇条書きの行で終わる。
cat > "$fixture/docs/adr/from-paragraph.md" <<'EOF'
---
id: "ADR-22222222-2222-4222-8222-222222222222"
type: adr
name: "results[] を返す"
status: deprecated
---
# ADR-22222222-2222-4222-8222-222222222222: results[] を返す

- ステータス: deprecated
- 日付: 2026-01-01

## 背景

```bash
echo 段落ではない
```

最初の段落は
`results[]`
の後に English
words を
挟む。
- 箇条書きの行は段落に含めない

次の段落は使わない。
EOF

# description も段落も無い SDR。
cat > "$fixture/docs/sara-adr/no-description.md" <<'EOF'
---
id: "SDR-33333333-3333-4333-8333-333333333333"
type: sara_adr
name: "説明の無い決定"
status: stable
---
# SDR-33333333-3333-4333-8333-333333333333: 説明の無い決定

- ステータス: stable

## 背景

- 箇条書きだけ

| 表 | 段落ではない |

> 引用は段落ではない

<details>HTML は段落ではない</details>

    字下げ行は段落ではない

~~~
フェンスの中は段落ではない
~~~
EOF

# ADR 以外の item。H1 の直後に段落があり、status を持っても添え字は付かない。
cat > "$fixture/docs/ubiquitous-language/term.md" <<'EOF'
---
id: "UBQ-44444444-4444-4444-8444-444444444444"
type: ubiquitous_term
name: "term"
section: core
status: stable
---
# UBQ-44444444-4444-4444-8444-444444444444: term

The term
definition.
EOF

# section の並びがファイル名の並びと逆になる語。節は表の順、節の中はファイル名順に並ぶ。z core は name_ja を持つ。
cat > "$fixture/docs/ubiquitous-language/a-mode.md" <<'EOF'
---
id: "UBQ-66666666-6666-4666-8666-666666666666"
type: ubiquitous_term
name: "a mode"
section: modes
status: stable
---
# UBQ-66666666-6666-4666-8666-666666666666: a mode

A mode.
EOF
cat > "$fixture/docs/ubiquitous-language/z-core.md" <<'EOF'
---
id: "UBQ-77777777-7777-4777-8777-777777777777"
type: ubiquitous_term
name: "z core"
name_ja: "中核[語]"
section: core
status: stable
---
# UBQ-77777777-7777-4777-8777-777777777777: z core

A core term.
EOF

# ubiquitous_term と同じ値 errors を持つ requirement。見出しは型ごとの表から引く。
cat > "$fixture/docs/requirements/error.md" <<'EOF'
---
id: "REQ-88888888-8888-4888-8888-888888888888"
type: requirement
name: "エラーの要件"
section: errors
status: stable
specification: |
  The fixture SHALL keep this requirement.
specification_ja: |
  fixture はこの要求を保たなければならない。
---
# REQ-88888888-8888-4888-8888-888888888888: エラーの要件

エラーの説明。
EOF

# docs/test/ の区分に置くテスト条件。
cat > "$fixture/docs/test/copy/condition.md" <<'EOF'
---
id: "TC-55555555-5555-4555-8555-555555555555"
type: test_condition
name: "条件"
status: stable
---
# TC-55555555-5555-4555-8555-555555555555: 条件

条件の説明。
EOF

if ! (cd "$fixture" && bash "$generator") > "$work/gen.out" 2>&1; then
  fault "入力: 小さな入力から index.md を生成できなかった"
  sed 's/^/    /' "$work/gen.out"
  exit 1
fi

# 期待する行が index.md に 1 行そのまま含まれることを確かめる。
expect_line() {
  local label=$1 file=$2 line=$3
  if grep -qxF -- "$line" "$fixture/$file"; then
    pass "$label"
  else
    fault "$label（$file に次の行が無い: $line）"
    sed 's/^/    /' "$fixture/$file"
  fi
}

# --- (2) description の 3 通りの代用と status の添え字 -------------------------

expect_line "description フィールドを使い、ADR の status を添える" docs/adr/index.md \
  '* [フィールドを持つ決定](./with-field.md) - フィールドの要約 (stable)'
expect_line "description が無ければ H1 の次の最初の段落を使い、日本語が接する改行は空白なしでつなぐ" docs/adr/index.md \
  '* [results\[\] を返す](./from-paragraph.md) - 最初の段落は`results[]`の後に English words を挟む。 (deprecated)'
expect_line "表・引用・HTML・字下げ行・~~~ のフェンスは段落に数えず、段落が無ければ ' - ' ごと省く" docs/sara-adr/index.md \
  '* [説明の無い決定](./no-description.md) (stable)'
expect_line "ADR 以外の item には status の添え字を付けず、英数字の行は空白でつなぐ" docs/ubiquitous-language/index.md \
  '* [term](./term.md) - The term definition.'

# --- (3) docs/test/ の区分 ------------------------------------------------------

want_copy=$(cat <<'EOF'
# copy

## テスト条件（`test_condition`）

* [条件](./condition.md) - 条件の説明。

## テストケース（`test_case`）
EOF
)
if [ "$(cat "$fixture/docs/test/copy/index.md")" = "$want_copy" ]; then
  pass "区分の index.md は型ごとの見出しの下にエントリを並べる"
else
  fault "(3) docs/test/copy/index.md の形が食い違う"
  sed 's/^/    /' "$fixture/docs/test/copy/index.md"
fi

want_cats=$(read_tsv "$script_dir/test-categories.tsv" | cut -f1 | tr '\n' ' ')
got_cats=$(grep -oE '^\* \[[^]]*\]' "$fixture/docs/test/index.md" | sed -E 's/^\* \[(.*)\]$/\1/' | tr '\n' ' ')
if [ "$got_cats" = "$want_cats" ]; then
  pass "docs/test/index.md は区分を test-categories.tsv の順に並べる"
else
  fault "(3) docs/test/index.md の区分の並びが test-categories.tsv と食い違う（実際: $got_cats）"
fi

# --- (4) 並びはファイル名順 ----------------------------------------------------

# deprecated の from-paragraph.md は末尾の節へ移るので、写しに主題の並びへ残る ADR を 1 件足して並びを見る。
ordered="$work/ordered"
cp -r "$fixture" "$ordered"
sed 's/8111-111111111111"$/8111-111111111112"/' "$fixture/docs/adr/with-field.md" > "$ordered/docs/adr/another.md"
(cd "$ordered" && bash "$generator") > /dev/null 2>&1
order=$(sed '/^## 有効でない item/q' "$ordered/docs/adr/index.md" | grep -oE '\(\./[^)]*\)' | tr -d '()' | tr '\n' ' ')
if [ "$order" = "./another.md ./with-field.md " ]; then
  pass "エントリはファイル名順に並ぶ"
else
  fault "(4) エントリがファイル名順に並ばない（実際: $order）"
fi

# --- (5) requirement・design・ubiquitous_term は section の節に分ける ----------

want_terms=$(cat <<'EOF'
# 用語（`ubiquitous_term`）

## Core placement abstraction

* [term](./term.md) - The term definition.
* [z core（中核\[語\]）](./z-core.md) - A core term.

## Placement modes

* [a mode](./a-mode.md) - A mode.
EOF
)
if [ "$(cat "$fixture/docs/ubiquitous-language/index.md")" = "$want_terms" ]; then
  pass "節は表の順に並び、節の中はファイル名順に並び、item の無い節は出さない"
else
  fault "(5) docs/ubiquitous-language/index.md の節分けが食い違う"
  sed 's/^/    /' "$fixture/docs/ubiquitous-language/index.md"
fi

want_reqs=$(cat <<'EOF'
# 要件（`requirement`）

## エラー仕様

* [エラーの要件](./error.md) - エラーの説明。
EOF
)
if [ "$(cat "$fixture/docs/requirements/index.md")" = "$want_reqs" ]; then
  pass "見出しは型ごとの表から引く（ubiquitous_term と同じ値 errors でも requirement の見出しになる）"
else
  fault "(5) docs/requirements/index.md の節分けが食い違う"
  sed 's/^/    /' "$fixture/docs/requirements/index.md"
fi

# --- (6) 用語のリンク文字列に name_ja を添える --------------------------------

expect_line "name_ja を持つ語はリンク文字列を「name（name_ja）」にし、角括弧をエスケープする" docs/ubiquitous-language/index.md \
  '* [z core（中核\[語\]）](./z-core.md) - A core term.'
expect_line "name_ja を持たない語はリンク文字列を name だけにする" docs/ubiquitous-language/index.md \
  '* [term](./term.md) - The term definition.'

# --- (7) deprecated の item は末尾の節にまとめる ------------------------------

want_adrs=$(cat <<'EOF'
# ADR（`adr`）

* [フィールドを持つ決定](./with-field.md) - フィールドの要約 (stable)

## 有効でない item（`deprecated`）

* [results\[\] を返す](./from-paragraph.md) - 最初の段落は`results[]`の後に English words を挟む。 (deprecated)
EOF
)
if [ "$(cat "$fixture/docs/adr/index.md")" = "$want_adrs" ]; then
  pass "節に分けない型では、deprecated の item を主題の並びから外して末尾の節にだけ出す"
else
  fault "(7) docs/adr/index.md の deprecated の item の置き場所が食い違う"
  sed 's/^/    /' "$fixture/docs/adr/index.md"
fi

headed=""
for f in docs/requirements/index.md docs/sara-adr/index.md docs/ubiquitous-language/index.md docs/test/copy/index.md; do
  grep -qxF '## 有効でない item（`deprecated`）' "$fixture/$f" && headed+=" $f"
done
if [ -z "$headed" ]; then
  pass "deprecated の item が無い index.md には末尾の節の見出しを出さない"
else
  fault "(7) deprecated の item が無いのに末尾の節の見出しが出た:$headed"
fi

# 写しに、節に分ける型の deprecated と draft の要件、区分の deprecated のテスト条件とテストケースを足す。
deprecated="$work/deprecated"
cp -r "$fixture" "$deprecated"
sed 's/8888-888888888888/8888-888888888890/; s/^status: stable$/status: deprecated/; s/エラーの要件/古い要件/; s/エラーの説明。/古い説明。/' \
  "$fixture/docs/requirements/error.md" > "$deprecated/docs/requirements/a-old.md"
sed 's/8888-888888888888/8888-888888888891/; s/^status: stable$/status: draft/; s/エラーの要件/下書きの要件/; s/エラーの説明。/下書きの説明。/' \
  "$fixture/docs/requirements/error.md" > "$deprecated/docs/requirements/z-draft.md"
sed 's/8555-555555555555/8555-555555555556/; s/^status: stable$/status: deprecated/; s/条件/古い条件/g' \
  "$fixture/docs/test/copy/condition.md" > "$deprecated/docs/test/copy/old-condition.md"
cat > "$deprecated/docs/test/copy/a-case.md" <<'EOF'
---
id: "CASE-99999999-9999-4999-8999-999999999999"
type: test_case
name: "古いケース"
resource: "internal/engine/copy_test.go"
status: deprecated
---
# CASE-99999999-9999-4999-8999-999999999999: 古いケース

ケースの説明。
EOF
if ! (cd "$deprecated" && bash "$generator") > "$work/deprecated.out" 2>&1; then
  fault "(7) deprecated の item を足した入力から index.md を生成できなかった"
  sed 's/^/    /' "$work/deprecated.out"
fi

want_dep_reqs=$(cat <<'EOF'
# 要件（`requirement`）

## エラー仕様

* [エラーの要件](./error.md) - エラーの説明。
* [下書きの要件](./z-draft.md) - 下書きの説明。

## 有効でない item（`deprecated`）

* [古い要件](./a-old.md) - 古い説明。
EOF
)
if [ "$(cat "$deprecated/docs/requirements/index.md")" = "$want_dep_reqs" ]; then
  pass "節に分ける型では、deprecated の item を section の節から外し、小見出しを付けずに末尾の節にだけ出す"
else
  fault "(7) docs/requirements/index.md の deprecated の item の置き場所が食い違う"
  sed 's/^/    /' "$deprecated/docs/requirements/index.md"
fi

main_part=$(sed '/^## 有効でない item/q' "$deprecated/docs/requirements/index.md")
if grep -qxF '* [下書きの要件](./z-draft.md) - 下書きの説明。' <<< "$main_part" &&
  [ "$(grep -cF '(./z-draft.md)' "$deprecated/docs/requirements/index.md")" -eq 1 ]; then
  pass "draft の item は主題の並びに残す"
else
  fault "(7) draft の item が主題の並びに無い"
  sed 's/^/    /' "$deprecated/docs/requirements/index.md"
fi

want_dep_copy=$(cat <<'EOF'
# copy

## テスト条件（`test_condition`）

* [条件](./condition.md) - 条件の説明。

## テストケース（`test_case`）

## 有効でない item（`deprecated`）

* [古いケース](./a-case.md) - ケースの説明。
* [古い条件](./old-condition.md) - 古い条件の説明。
EOF
)
if [ "$(cat "$deprecated/docs/test/copy/index.md")" = "$want_dep_copy" ]; then
  pass "区分の index.md は、deprecated のテスト条件とテストケースを 2 つの節の後ろの 1 つの節にファイル名順でまとめる"
else
  fault "(7) docs/test/copy/index.md の deprecated の item の置き場所が食い違う"
  sed 's/^/    /' "$deprecated/docs/test/copy/index.md"
fi

# --- (8) 表に無い section の値で exit 2 ----------------------------------------

# model.yaml の写しの requirement の enum に値を 1 つ足し、sara check は通るが生成器の表に無い値を作る。
unknown="$work/unknown"
cp -r "$fixture" "$unknown"
sed -i 's/^            - dependencies$/&\n            - zz-unlisted-section/' "$unknown/docs/model.yaml"
sed 's/^section: errors$/section: zz-unlisted-section/; s/8888-888888888888/8888-888888888889/' \
  "$fixture/docs/requirements/error.md" > "$unknown/docs/requirements/unlisted.md"
cp "$unknown/docs/requirements/index.md" "$work/unknown-before.md"
(cd "$unknown" && bash "$generator") > "$work/unknown.out" 2> "$work/unknown.err"
rc=$?
if [ "$rc" -eq 2 ] && grep -qF zz-unlisted-section "$work/unknown.err" &&
  grep -qF '表（SECTIONS）に無い' "$work/unknown.err" && ! grep -qF 'sara check が失敗した' "$work/unknown.err" &&
  cmp -s "$unknown/docs/requirements/index.md" "$work/unknown-before.md"; then
  pass "表に無い section の値は標準エラーにその値を出し、index.md を書かずに exit 2 で止まる"
else
  fault "(8) 表に無い section の値で exit 2 にならない（exit $rc）"
  sed 's/^/    /' "$work/unknown.err"
fi

# --- (9) --check は食い違うパスを出して exit 1 を返し、index.md を書き換えない --

# --check を走らせ、exit 1 で出力が期待するパスだけであり、案内が期待する方だけであることを確かめる。
expect_stale() {
  local label=$1 want=$2 hint=$3 other=$4 rc
  (cd "$fixture" && bash "$generator" --check) > "$work/stale.out" 2> "$work/stale.err"
  rc=$?
  if [ "$rc" -eq 1 ] && [ "$(cat "$work/stale.out")" = "$want" ] &&
    grep -qF "$hint" "$work/stale.err" && ! grep -qF "$other" "$work/stale.err"; then
    pass "$label"
  else
    fault "$label（exit $rc・出力: $(tr '\n' ' ' < "$work/stale.out")）"
    sed 's/^/    /' "$work/stale.err"
  fi
}

printf '手で書き換えた\n' >> "$fixture/docs/adr/index.md"
cp "$fixture/docs/adr/index.md" "$work/edited.md"
expect_stale "--check は内容の食い違う index.md のパスを出し、再生成を促して exit 1 を返す" \
  docs/adr/index.md '再生成する' '削除する'
if cmp -s "$fixture/docs/adr/index.md" "$work/edited.md"; then
  pass "--check は既存の index.md を書き換えない"
else
  fault "(9) --check が既存の index.md を書き換えた"
fi

# 書き出しで内容の食い違いを解消してから、生成対象に無い index.md だけを残す。
(cd "$fixture" && bash "$generator") > /dev/null 2>&1
mkdir -p "$fixture/docs/test/removed" "$fixture/docs/agents"
printf '# removed\n' > "$fixture/docs/test/removed/index.md"
printf '# agents\n' > "$fixture/docs/agents/index.md"
expect_stale "--check は生成対象に無い index.md のパスを出し、削除を促して exit 1 を返す（docs/agents/ は見ない）" \
  docs/test/removed/index.md '削除する' '再生成する'

exit "$fail"
