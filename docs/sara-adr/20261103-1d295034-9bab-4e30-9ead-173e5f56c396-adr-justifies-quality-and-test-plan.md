---
id: "SDR-1d295034-9bab-4e30-9ead-173e5f56c396"
type: sara_adr
name: "ADR の justifies 対象を specification を持つ全型に広げる"
status: 採用
issues:
  - "#583"
  - "#237"
  - "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §5 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "ADR-0049"
references:
  - "SDR-34916a83-0cff-4524-a8ff-5a3195af52c3"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
---
# SDR-1d295034-9bab-4e30-9ead-173e5f56c396: ADR の justifies 対象を specification を持つ全型に広げる

- ステータス: 採用
- 日付: 2026-11-03
- 関連: SDR-34916a83-0cff-4524-a8ff-5a3195af52c3, SDR-2457aa00-fab7-4c99-af53-2e6643c21c79, GitHub Issue #583, #237, #203, `docs/model.yaml`
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §5 を書き直した

## 背景

ADR 系の型の `justifies` は、`requirement` / `design` / `infrastructure` だけを対象にしていた。quality と
test_plan を設けた（SDR-34916a83-0cff-4524-a8ff-5a3195af52c3）ことで、規範の文を持つ型が増えた。

## 決定

ADR 系の型の `justifies` の対象を `requirement` / `design` / `infrastructure` / `quality` / `test_plan` にする。
`specification` を持つ型には、どれにも ADR から接続できる。

## 根拠

ADR は「なぜその規範にしたか」を残し、規範そのものは `justifies` の先の item が持つ。品質方針やテスト計画の
規範も決定から生まれるので、その決定を記録した ADR から辿れる必要がある。

## 棄却した代替案

### 対象を `requirement` / `design` / `infrastructure` のまま据え置く

quality と test_plan の規範を決める ADR が、その item へ接続できない。
