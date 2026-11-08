---
id: "ADR-1848079b-7011-4653-b7cd-c279205262c4"
type: adr
name: "manifest.json の entry は target を識別子にして name を持たない"
status: stable
origin: "ADR-0014 §3b を書き直した"
justifies:
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
references:
  - "ADR-c494e86e-ceb9-4380-b983-4e09553b01d6"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-ee31b8d0-f7c5-442f-ae90-dcd1dd16b7a0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-1848079b-7011-4653-b7cd-c279205262c4: manifest.json の entry は target を識別子にして name を持たない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [entries を target をキーにした attrset にする](entries-keyed-by-target.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), [manifest 内の target の衝突を評価時に検出する](manifest-target-conflict-eval-error.md), `docs/spec.md`
- 起点: ADR-0014 §3b を書き直した

## 背景

`manifest.json` の `entries[]` は、Nix 側の entry に合わせて `name` を記録していた。entries を
target をキーにした attrset にすると、Nix 側から `name` フィールドが無くなる。engine に渡す
`manifest.json` の entry を何で識別するかを決め直す必要がある。

## 決定

`manifest.json` の `entries[]` は `name` を持たず、`target` を entry の識別子とする。属性キーも
記録しない。記録するフィールドの一覧は [manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する](../requirements/manifest-entries-array.md) に置く。

## 根拠

engine が配置・比較・表示に使う値は、実際の配置先である target だけで足りる。識別子を 1 つに
しておけば、Nix 側と engine 側で entry の同一性の定義が食い違わない。変えた時点では `manifest.json`
の版はリリース前で、利用者の manifest を移行する必要も無かった。

## 棄却した代替案

### 表示用のラベルとして属性キーを記録する

属性キーと target の 2 つの識別子が並び、上書きした entry では食い違う。engine がどちらで比較
するかという問いが manifest の読み手に残る。
