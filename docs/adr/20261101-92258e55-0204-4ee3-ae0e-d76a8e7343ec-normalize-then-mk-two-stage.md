---
id: "ADR-92258e55-0204-4ee3-ae0e-d76a8e7343ec"
type: adr
name: "入力の確認を normalizeManifest に切り出す"
status: stable
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §4（2 段分割）を書き直した"
justifies:
  - "REQ-b232ec98-af3b-41f3-a050-29d417322002"
  - "TP-36e90d5d-4524-4294-bc72-ee263bb02782"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:30:55+09:00"
---
# ADR-92258e55-0204-4ee3-ae0e-d76a8e7343ec: 入力の確認を normalizeManifest に切り出す

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §4（2 段分割）を書き直した

## 背景

`mkManifest` は入力を確認・正規化し、その結果から `manifest.json` と symlink farm を含む
derivation を組む。確認と正規化の結果を Nix の評価テストで確かめたいが、derivation どうしを
比べて不変条件を確かめるのは難しい。

## 決定

入力の確認・既定値の適用・marker から clean enum への変換を、純粋なデータ関数
`normalizeManifest` に切り出す。`mkManifest` は `normalizeManifest` の出力から derivation を
組むだけにする。

2 つの関数の役割分担は REQ-b232ec98-af3b-41f3-a050-29d417322002 に、評価テストの方針は
TP-36e90d5d-4524-4294-bc72-ee263bb02782 に置く。

## 根拠

確認と正規化の結果が素の attrset になるので、nix-unit と namaka でそのまま確かめられる。
エラーの系統も、型とメッセージを指定して回帰テストにできる。lib をテストできる純粋なデータ生成器に
保つという方針を、関数の形で直接担保する。

## 棄却した代替案

### `mkManifest` を分けずに derivation だけを返す

確認と正規化のロジックが derivation の中に閉じ、nix-unit で不変条件を確かめにくい。
