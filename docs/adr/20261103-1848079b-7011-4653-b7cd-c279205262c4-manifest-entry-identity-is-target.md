---
id: "ADR-1848079b-7011-4653-b7cd-c279205262c4"
type: adr
name: "manifest.json の entry は target を識別子にして name を持たない"
status: 採用
origin: "ADR-0014 §3b を書き直した"
justifies:
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
references:
  - "ADR-c494e86e-ceb9-4380-b983-4e09553b01d6"
---
# ADR-1848079b-7011-4653-b7cd-c279205262c4: manifest.json の entry は target を識別子にして name を持たない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-c494e86e-ceb9-4380-b983-4e09553b01d6, `docs/spec.md`
- 起点: ADR-0014 §3b を書き直した

## 背景

`manifest.json` の `entries[]` は、Nix 側の entry に合わせて `name` を記録していた。entries を
target をキーにした attrset にすると、Nix 側から `name` フィールドが無くなる。engine に渡す
`manifest.json` の entry を何で識別するかを決め直す必要がある。

## 決定

`manifest.json` の `entries[]` は `name` を持たず、`target` を entry の識別子とする。属性キーも
記録しない。記録するフィールドの一覧は REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336 に置く。

## 根拠

engine が配置・比較・表示に使う値は、実際の配置先である target だけで足りる。識別子を 1 つに
しておけば、Nix 側と engine 側で entry の同一性の定義が食い違わない。変えた時点では `manifest.json`
の版はリリース前で、利用者の manifest を移行する必要も無かった。

## 棄却した代替案

### 表示用のラベルとして属性キーを記録する

属性キーと target の 2 つの識別子が並び、上書きした entry では食い違う。engine がどちらで比較
するかという問いが manifest の読み手に残る。
