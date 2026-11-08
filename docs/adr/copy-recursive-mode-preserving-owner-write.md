---
id: "ADR-f27830a5-4d13-4f61-be9e-3f04240ff322"
type: adr
name: "copy は再帰コピーで mode を保存し owner-write にする"
status: stable
origin: "ADR-0011 §7（再帰コピー）と ADR-0016 §1 を書き直した"
justifies:
  - "REQ-84e3c717-adf5-4ff3-b0db-d039b82ef19c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:02:52+09:00"
---
# ADR-f27830a5-4d13-4f61-be9e-3f04240ff322: copy は再帰コピーで mode を保存し owner-write にする

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/concept.md`, `docs/spec.md`, `docs/design.md`
- 起点: ADR-0011 §7（再帰コピー）と ADR-0016 §1 を書き直した

## 背景

copy は store のファイルやディレクトリを target へ複製する。store のファイルは read-only
（0444 / 0555）なので、mode をそのまま写すと複製先も read-only になり、利用者が編集できない。
copy は編集するために選ぶ method なので、mode の扱いを決める必要があった。

## 決定

copy は engine が自前の再帰コピー（`filepath.WalkDir` と `io.Copy` と `os.Chmod`）で行い、元の
mode を保存したうえで owner の書き込みビットを付ける（例: 0444 → 0644、0555 → 0755）。実行ビットや
group / other のビットはそのまま保つ。規範の細目は REQ-84e3c717-adf5-4ff3-b0db-d039b82ef19c に置く。

## 根拠

read-only のまま複製すると、利用者は copy のたびに手で `chmod` しなければならない。owner-write
だけを足せば、実行ビットなどの相対的な構造を失わずに編集できる状態になる。

標準の `os.CopyFS` は mode を保存せず、symlink も扱えないので、この要件を満たさない。

## 棄却した代替案

### mode を厳密に保存する（read-only のまま）

store の mode に忠実だが、copy を選んだのに編集できない。

### copy に `writable` オプションを足す

柔軟だが entry のスキーマが膨らむ。copy では書き込めることがほぼ常に望ましく、選べる利点が薄い。

### `os.CopyFS` を使う

実装は短くなるが、mode を保存せず symlink を扱えない。
