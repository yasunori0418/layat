---
id: "ADR-14cf38ad-23d3-4d25-b14e-3e5da5ac7826"
type: adr
name: "manifest に identity を足しても schemaVersion を据え置く"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-498"
    resource: "https://github.com/yasunori0418/layat/issues/498"
    title: "#498"
origin: "Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §5b を書き直した"
justifies:
  - "REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f"
  - "REQ-97c1e088-a17e-46d9-a9a1-83d1757d0f7d"
references:
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-0db061d1-fa30-436d-91d7-c0cec8804c52"
  - "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-14cf38ad-23d3-4d25-b14e-3e5da5ac7826: manifest に identity を足しても schemaVersion を据え置く

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-0db061d1-fa30-436d-91d7-c0cec8804c52, ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e, GitHub Issue #580, #498
- 起点: Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §5b を書き直した

## 背景

出所の判定に使う `identity` は、manifest のトップレベルの任意フィールドとして足した。manifest は
`schemaVersion` を持ち、v1 に固定している（ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd）。
形を変えたときに `schemaVersion` を上げるかを決める必要がある。

## 決定

manifest に任意フィールド `identity` を足しても、`schemaVersion` は 1 のまま据え置く。`identity`
付きの manifest を、未知のフィールドを拒む旧い CLI が読めないことは、pre-1.0
の既知の事項として受け入れる。

## 根拠

`identity` を持たない manifest の形は変わらず、旧い manifest は新しい CLI でそのまま読める。
互換が崩れるのは `identity` を宣言した利用者が旧い CLI を使う場合だけで、その範囲は pre-1.0
の非互換として扱える。

## 棄却した代替案

### `schemaVersion` を 2 に上げる

`identity` を使わない利用者の manifest まで新しい版になり、旧い CLI との互換を全員が失う。
