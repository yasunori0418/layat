---
id: "ADR-9bdf2604-c7b3-468a-b1e4-2478ef37477c"
type: adr
name: "rootKind に system を足しても schemaVersion を上げない"
status: 採用
origin: "次期マイルストーン計画の grilling（2026-07-04）。ADR-0036 §1a を書き直した"
justifies:
  - "REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb"
supersedes:
  - "ADR-0036"
references:
  - "ADR-0036"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
---
# ADR-9bdf2604-c7b3-468a-b1e4-2478ef37477c: rootKind に system を足しても schemaVersion を上げない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0036, ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。ADR-0036 §1a を書き直した

## 背景

`systemRoot` は root の marker として公開していたが、system mode を実装するまでは `normalizeManifest` が
評価時に拒否していた。system mode を実装すると、`manifest.json` の `root.rootKind` に新しい値 `"system"` が
現れる。`manifest.json` のスキーマは `schemaVersion = 1` で、版を上げるかどうかを決める必要があった。

## 決定

`rootKind` の値に `"system"` を足し、`schemaVersion` は 1 のまま据え置く。`rootKind` の値の一覧は
REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb に置く。

## 根拠

変更は列挙値を 1 つ足すだけで、既存のフィールドの意味も形も変わらない。古い CLI は知らない `rootKind` を
拒否するので、古い CLI に新しい manifest を渡しても、既存の確認で安全に止まる。版を上げなくても、
誤った配置は起きない。

## 棄却した代替案

### `schemaVersion` を 2 に上げる

古い CLI は新しい manifest を版の違いで拒否できるが、`rootKind` の確認で既に拒否できている。版を上げると
古い版を読む互換の扱いまで決める必要が生じ、列挙値 1 つの追加に見合わない。
