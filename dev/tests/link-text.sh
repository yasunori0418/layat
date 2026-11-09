#!/usr/bin/env bash
# リンク文字列・frontmatter の name・H1 の裸のタグと、リンク前後の name の繰り返しが無いことを確かめる契約テスト。
# 文書とコードに、許可外の場所のフル ID と 8 文字省略形が無いことも確かめる。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/link-text.sh               # devShell / CI から直接
#   nix develop '.?dir=dev#sara' -c dev/tests/link-text.sh --self-test   # 自己テストだけ
#   nix flake check ./dev                                                 # checks.link-text 経由

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

require_commands "link-text.sh（nix develop '.?dir=dev#sara' から実行する）" git awk find sort cut diff grep || exit 1

case "${1-}" in
  "") self_test_only=0 ;;
  --self-test) self_test_only=1 ;;
  *)
    echo "usage: link-text.sh [--self-test]" >&2
    exit 2
    ;;
esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 検出を「<種別>\t<path>:<line>: <内容>」で標準出力へ出す。種別は tag（裸のタグ）と dup（name の繰り返し）。
scan() {
  LC_ALL=C awk '
    # コードスパンを同じバイト数の \001 で塗り、元の行と位置を揃える。
    function mask(s, out, i, n, run, j, k, stop, m) {
      out = ""; i = 1; n = length(s)
      while (i <= n) {
        if (substr(s, i, 1) != "`") { out = out substr(s, i, 1); i++; continue }
        run = 0
        while (substr(s, i + run, 1) == "`") run++
        stop = 0; j = i + run
        while (j <= n) {
          if (substr(s, j, 1) != "`") { j++; continue }
          k = 0
          while (substr(s, j + k, 1) == "`") k++
          if (k == run) { stop = j; break }
          j += k
        }
        if (!stop) { out = out substr(s, i, run); i += run; continue }
        for (m = i; m < stop + run; m++) out = out "\001"
        i = stop + run
      }
      return out
    }
    function unescape(s) { gsub(/\\\[/, "[", s); gsub(/\\\]/, "]", s); return s }
    function report(kind, detail) { printf "%s\t%s:%d: %s\n", kind, FILENAME, FNR, detail }
    function check_tag(masked, shown) { if (masked ~ /<[A-Za-z][^<> ]*>/) report("tag", shown) }
    function ends_with(s, t) { return length(s) >= length(t) && substr(s, length(s) - length(t) + 1) == t }
    function check_links(line, masked, off, m, start, stop, text, x, before, after) {
      off = 0
      while (match(substr(masked, off + 1), /\[(\\.|[^][\\])*\]\([^)]*\)/)) {
        start = off + RSTART
        stop = start + RLENGTH - 1
        m = substr(masked, start, RLENGTH)
        sub(/\]\([^)]*\)$/, "", m)
        text = substr(line, start + 1, length(m) - 1)
        check_tag(substr(m, 2), text)
        x = unescape(text)
        before = unescape(substr(line, 1, start - 1))
        after = unescape(substr(line, stop + 1))
        if (x != "" && (index(after, "「" x "」") == 1 || index(after, "（" x "）") == 1 ||
            (ends_with(before, x "（") && index(after, "）") == 1)))
          report("dup", text)
        off = stop
      }
    }
    FNR == 1 { in_fm = ($0 == "---"); fence_char = ""; if (in_fm) next }
    in_fm {
      if ($0 == "---") in_fm = 0
      else if ($0 ~ /^name:/) check_tag(mask($0), $0)
      next
    }
    fence_char != "" {
      t = $0; sub(/^[ \t]+/, "", t); sub(/[ \t]+$/, "", t)
      if (length(t) >= fence_len) { c = t; gsub(fence_char, "", c); if (c == "") fence_char = "" }
      next
    }
    # 情報文字列にバッククォートを含む行はフェンスではなくコードスパンとして扱う。
    match($0, /^[ \t]*(```+|~~~+)/) {
      t = substr($0, RSTART, RLENGTH); sub(/^[ \t]+/, "", t)
      if (!(substr(t, 1, 1) == "`" && index(substr($0, RLENGTH + 1), "`"))) {
        fence_char = substr(t, 1, 1); fence_len = length(t); next
      }
    }
    {
      masked = mask($0)
      if ($0 ~ /^# /) check_tag(masked, $0)
      check_links($0, masked)
    }
  ' "$@"
}

# CLAUDE.md と dev/skills/ が ID 規約の例として使う ID。
example_id=REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4

# 検出を「<種別>\t<path>:<line>: <ID>」で標準出力へ出す。種別は full（許可外のフル ID）と short（8 文字省略形）。
scan_ids() {
  LC_ALL=C awk -v example="$example_id" '
    function hex(n, s) { s = ""; while (n-- > 0) s = s "[0-9a-f]"; return s }
    function report(kind, id) { printf "%s\t%s:%d: %s\n", kind, FILENAME, FNR, id }
    BEGIN {
      prefix = "(SOL|UC|REQ|DSG|QA|TP|INF|RISK|TC|CASE|ADR|SDR|UBQ)"
      full = prefix "-" hex(8) "-" hex(4) "-" hex(4) "-" hex(4) "-" hex(12)
      short = "[^A-Za-z0-9-]" prefix "-" hex(8) "[^0-9a-f-]"
      id_line = "^id:[ \t]*\"?" full "\"?[ \t]*$"
      relation_line = "^[ \t]*-[ \t]+\"?" full "\"?[ \t]*$"
    }
    FNR == 1 {
      in_fm = (FILENAME ~ /\.md$/ && $0 == "---")
      if (in_fm) next
    }
    in_fm && $0 == "---" { in_fm = 0; next }
    in_fm && ($0 ~ id_line || $0 ~ relation_line) { next }
    {
      line = $0
      while (match(line, full)) {
        id = substr(line, RSTART, RLENGTH)
        if (!(id == example && (FILENAME == "CLAUDE.md" || FILENAME ~ /^dev\/skills\//))) report("full", id)
        line = substr(line, 1, RSTART - 1) " " substr(line, RSTART + RLENGTH)
      }
      line = " " line " "
      while (match(line, short)) {
        report("short", substr(line, RSTART + 1, RLENGTH - 2))
        line = substr(line, RSTART + RLENGTH - 1)
      }
    }
  ' "$@"
}

# --- 自己テスト -----------------------------------------------------------------

fixture="$work/fixture"
mkdir -p "$fixture"

cat > "$fixture/detect.md" <<'EOF'
---
name: "タグ <state> 入りの name"
---
# タグ <link-farm> 入りの H1

[タグ <name> 入りのリンク](a.md)
[丸括弧の繰り返し](b.md)（丸括弧の繰り返し）
[鉤括弧の繰り返し](c.md)「鉤括弧の繰り返し」
前置の繰り返し（[前置の繰り返し](d.md)）
[角括弧 \[x\] の繰り返し](e.md)「角括弧 [x] の繰り返し」
[`code` の繰り返し](f.md)（`code` の繰り返し）
1 行に [一つ目の <a>](g.md) と [二つ目の <b>](h.md) がある
[`code` と裸の <c> の混在](i.md)
閉じない ` がある行の [タグ <d> 入り](j.md)
``` `span` の行の [タグ <e> 入り](k.md)
[前の行でフェンスに入らない <f>](l.md)
[大文字の <Name> 入りのリンク](m.md)
[後続もエスケープ \[y\] の繰り返し](n.md)「後続もエスケープ \[y\] の繰り返し」
前置のエスケープ \[z\] の繰り返し（[前置のエスケープ \[z\] の繰り返し](o.md)）
```
[閉じたフェンスの中の <g>](p.md)
```
[閉じたフェンスの後の <h>](q.md)
EOF

# 閉じないフェンスで終わるファイルの次のファイルも走査できることを確かめる。
cat > "$fixture/detect-unclosed.md" <<'EOF'
本文
```
EOF

cat > "$fixture/detect-nofm.md" <<'EOF'
# frontmatter の無いファイルの <title>

[タグ <name> 入りのリンク](a.md)
EOF

cat > "$work/detect.expected" <<'EOF'
dup	detect.md:10
dup	detect.md:11
dup	detect.md:18
dup	detect.md:19
dup	detect.md:7
dup	detect.md:8
dup	detect.md:9
tag	detect-nofm.md:1
tag	detect-nofm.md:3
tag	detect.md:12
tag	detect.md:12
tag	detect.md:13
tag	detect.md:14
tag	detect.md:15
tag	detect.md:16
tag	detect.md:17
tag	detect.md:2
tag	detect.md:23
tag	detect.md:4
tag	detect.md:6
EOF

cat > "$fixture/clean.md" <<'EOF'
---
name: "コードスパンの `<state>` は許す"
description: "<frontmatter> と [タグ <name> 入り](a.md) の name 以外のキーは見ない"
---
# コードスパンの `<name>` は許す

## 二段目の見出しの <name> は見ない

[コードスパンの `<name>` は許す](a.md)
[非 ASCII の <鍵> は許す](b.md)
[a < b と c > d は語でない](c.md)
[空白入りの <a b> は語でない](c.md)
[リンク](d.md)「別の語」
[リンク](d.md)（別の語）
別の語（[リンク](d.md)）
リンクの外の <name> は見ない
`[タグ <name> 入り](e.md)「タグ <name> 入り」` はコードスパンの中
`` [二重の <name> のスパン](e.md) `` もコードスパンの中
`` ` [内側の <name>](e.md) `` は長さの同じ閉じまでコードスパン
[リンク](d.md) 「リンク」は直後でない
[リンク](d.md)「リンクの別名」は完全一致でない
[リンク](d.md)（リンクの別名）は完全一致でない
リンク（[リンク](d.md)、補足）は直後が閉じ括弧でない

```bash
[コードブロックの <name>](e.md)（コードブロックの <name>）
# コードブロックの <name>
```

~~~
# チルダのコードブロックの <name>
~~~

````md
```
# 長いフェンスの中の短いフェンスの <name>
````

- リスト

  ```
  [字下げしたコードブロックの <name>](e.md)
  ```
EOF

(cd "$fixture" && scan detect-unclosed.md detect-nofm.md detect.md clean.md) > "$work/self.out"
self_rc=$?
if [ "$self_rc" -ne 0 ]; then
  fault "自己テスト: 走査が異常終了した（exit $self_rc）"
fi

cut -d: -f1,2 "$work/self.out" | LC_ALL=C sort > "$work/self.detected"
if diff <(grep -v $'\tclean\.md:' "$work/self.detected") "$work/detect.expected" > "$work/self.diff"; then
  pass "自己テスト: 検出すべき例を全件検出する"
else
  fault "自己テスト: 検出すべき例の結果が期待と食い違う"
  cat "$work/self.diff"
fi

if grep $'\tclean\.md:' "$work/self.out" > "$work/self.fp"; then
  fault "自己テスト: 検出してはいけない例を検出した"
  cat "$work/self.fp"
else
  pass "自己テスト: 検出してはいけない例を 0 件にする"
fi

ids="$work/ids"
mkdir -p "$ids/dev/skills" "$ids/dev/other"

cat > "$ids/detect.md" <<'EOF'
---
id: "REQ-11111111-2222-4333-8444-555555555555"
description: "REQ-11111111-2222-4333-8444-555555555556 を説明文に書く"
origin: "→ REQ-abcdef01 を来歴に書く"
derives_from:
  - "UC-11111111-2222-4333-8444-555555555555"
---
# REQ-11111111-2222-4333-8444-555555555555: 自身の H1

# REQ-11111111-2222-4333-8444-555555555557: 二つ目の H1

地の文 REQ-11111111-2222-4333-8444-555555555558 と `TP-11111111-2222-4333-8444-555555555559`
[DSG-11111111-2222-4333-8444-55555555555a](../x.md) と [x](../REQ-11111111-2222-4333-8444-55555555555b.md)
- REQ-11111111-2222-4333-8444-55555555555c の説明
- "RISK-11111111-2222-4333-8444-55555555555d"

日本語に隣接する省略形（→ REQ-abcdef01）
```
CASE-abcdef01 と ADR-12345678
```
EOF

cat > "$ids/detect.sh" <<'EOF'
# REQ-11111111-2222-4333-8444-555555555555: シェルのコメント
id: "REQ-11111111-2222-4333-8444-555555555555"
  - "REQ-11111111-2222-4333-8444-555555555555"
echo "→ TP-abcdef01"
EOF

cat > "$ids/nofm.md" <<'EOF'
# REQ-11111111-2222-4333-8444-555555555555: frontmatter の無い文書の H1
REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4 は例示の場所の外
EOF

cat > "$ids/CLAUDE.md" <<'EOF'
`REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` は規約の例
`REQ-11111111-2222-4333-8444-555555555555` は例ではない
EOF

cat > "$ids/dev/skills/example.md" <<'EOF'
`REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` は規約の例
EOF

cat > "$ids/dev/other/example.md" <<'EOF'
`REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` は dev/skills の外
EOF

cat > "$ids/clean.md" <<'EOF'
---
id: REQ-11111111-2222-4333-8444-555555555555
satisfies:
  - REQ-11111111-2222-4333-8444-555555555556
---
# name だけの H1

REQ-xxxxxxxx と ADR-0012 と ADR-NNNN と TC-999 と REQ-ABCDEF01
XREQ-abcdef01 と REQ-abcdef01-foo と REQ-abcdef012 と abcdef01
EOF

LC_ALL=C sort > "$work/ids.expected" <<'EOF'
full	detect.md:3
short	detect.md:4
full	detect.md:8
full	detect.md:10
full	detect.md:12
full	detect.md:12
full	detect.md:13
full	detect.md:13
full	detect.md:14
full	detect.md:15
short	detect.md:17
short	detect.md:19
short	detect.md:19
full	detect.sh:1
full	detect.sh:2
full	detect.sh:3
short	detect.sh:4
full	nofm.md:1
full	nofm.md:2
full	CLAUDE.md:2
full	dev/other/example.md:1
EOF

(cd "$ids" && scan_ids detect.md detect.sh nofm.md CLAUDE.md dev/skills/example.md dev/other/example.md clean.md) > "$work/ids.out"
ids_rc=$?
if [ "$ids_rc" -ne 0 ]; then
  fault "自己テスト: ID の走査が異常終了した（exit $ids_rc）"
fi

if diff <(cut -d: -f1,2 "$work/ids.out" | LC_ALL=C sort) "$work/ids.expected" > "$work/ids.diff"; then
  pass "自己テスト: 許可外の ID だけを検出する"
else
  fault "自己テスト: ID の検出が期待と食い違う"
  cat "$work/ids.diff"
fi

if [ "$fail" -ne 0 ]; then
  echo
  printf 'link-text: 自己テストに失敗したため本番の走査へ進まない\n' >&2
  exit 1
fi

if [ "$self_test_only" -eq 1 ]; then
  echo
  printf 'link-text: 自己テスト通過\n'
  exit 0
fi

# --- 対象ファイル -------------------------------------------------------------

cd "$(testdoc_repo_root)" || exit 1

# flake check のサンドボックスは git 管理外なので find へ倒す。
if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
  git ls-files -z -- 'docs/*.md' README.md README.ja.md CLAUDE.md
else
  find docs -name '*.md' -print0
  printf '%s\0' README.md README.ja.md CLAUDE.md
fi | LC_ALL=C sort -z > "$work/targets"

mapfile -d '' -t targets < "$work/targets"
if [ "${#targets[@]}" -eq 0 ]; then
  echo "link-text.sh: 対象ファイルが 1 件も無い" >&2
  exit 1
fi
for f in "${targets[@]}"; do
  if [ ! -f "$f" ]; then
    echo "link-text.sh: 対象ファイルが無い: $f" >&2
    exit 1
  fi
done

# --- 検出 -------------------------------------------------------------------

scan "${targets[@]}" > "$work/scan.out"
scan_rc=$?
if [ "$scan_rc" -ne 0 ]; then
  echo "link-text.sh: 走査が異常終了した（exit $scan_rc）" >&2
  exit 1
fi

id_roots=(docs internal cmd lib modules tests templates .github dev/scripts dev/skills)
id_files=(README.md README.ja.md CLAUDE.md sara.toml flake.nix example.envrc example.worktreeinclude example.worktreeskip)
if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
  git ls-files -z -- "${id_roots[@]}" "${id_files[@]}"
else
  find "${id_roots[@]}" -type f -print0
  printf '%s\0' "${id_files[@]}"
fi | LC_ALL=C sort -z > "$work/id-targets"

mapfile -d '' -t id_targets < "$work/id-targets"
for f in "${id_targets[@]}"; do
  if [ ! -f "$f" ]; then
    echo "link-text.sh: 対象ファイルが無い: $f" >&2
    exit 1
  fi
done

scan_ids "${id_targets[@]}" >> "$work/scan.out"
ids_rc=$?
if [ "$ids_rc" -ne 0 ]; then
  echo "link-text.sh: ID の走査が異常終了した（exit $ids_rc）" >&2
  exit 1
fi

report_kind() {
  local kind=$1 label=$2 count
  count=$(grep -c "^$kind"$'\t' "$work/scan.out")
  if [ "$count" -eq 0 ]; then
    pass "${label}が無い"
    return
  fi
  fault "${label}が ${count} 件ある"
  grep "^$kind"$'\t' "$work/scan.out" | cut -f2-
}

report_kind tag "リンク文字列・name・H1 の裸のタグ"
report_kind dup "リンク前後の name の繰り返し"
report_kind full "許可外の場所のフル ID"
report_kind short "8 文字省略形"

if [ -s "$work/scan.out" ] && [ "$fail" -eq 0 ]; then
  fault "走査の検出がどの種別にも分類されない"
  cat "$work/scan.out"
fi

# --- 結果 -------------------------------------------------------------------

echo
if [ "$fail" -eq 0 ]; then
  printf 'link-text: 全アサーション通過（表記 %d 件・ID %d 件）\n' "${#targets[@]}" "${#id_targets[@]}"
  exit 0
fi

printf 'link-text: 失敗あり\n' >&2
exit 1
