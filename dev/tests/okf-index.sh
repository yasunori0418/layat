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
mkdir -p "$fixture/docs/adr" "$fixture/docs/sara-adr" "$fixture/docs/ubiquitous-language" "$fixture/docs/test/copy"
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

order=$(grep -oE '\(\./[^)]*\)' "$fixture/docs/adr/index.md" | tr -d '()' | tr '\n' ' ')
if [ "$order" = "./from-paragraph.md ./with-field.md " ]; then
  pass "エントリはファイル名順に並ぶ"
else
  fault "(4) エントリがファイル名順に並ばない（実際: $order）"
fi

# --- (5) --check は食い違うパスを出して exit 1 を返し、index.md を書き換えない --

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
  fault "(5) --check が既存の index.md を書き換えた"
fi

# 書き出しで内容の食い違いを解消してから、生成対象に無い index.md だけを残す。
(cd "$fixture" && bash "$generator") > /dev/null 2>&1
mkdir -p "$fixture/docs/test/removed" "$fixture/docs/agents"
printf '# removed\n' > "$fixture/docs/test/removed/index.md"
printf '# agents\n' > "$fixture/docs/agents/index.md"
expect_stale "--check は生成対象に無い index.md のパスを出し、削除を促して exit 1 を返す（docs/agents/ は見ない）" \
  docs/test/removed/index.md '削除する' '再生成する'

exit "$fail"
