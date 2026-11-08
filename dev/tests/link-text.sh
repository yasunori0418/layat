#!/usr/bin/env bash
# リンク文字列・frontmatter の name・H1 の裸のタグと、リンク前後の name の繰り返しが無いことを確かめる契約テスト。
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

if [ -s "$work/scan.out" ] && [ "$fail" -eq 0 ]; then
  fault "走査の検出が tag / dup のどちらにも分類されない"
  cat "$work/scan.out"
fi

# --- 結果 -------------------------------------------------------------------

echo
if [ "$fail" -eq 0 ]; then
  printf 'link-text: 全アサーション通過（対象ファイル %d 件）\n' "${#targets[@]}"
  exit 0
fi

printf 'link-text: 失敗あり\n' >&2
exit 1
