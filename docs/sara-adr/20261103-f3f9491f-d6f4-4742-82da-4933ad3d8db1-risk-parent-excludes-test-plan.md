---
id: "SDR-f3f9491f-d6f4-4742-82da-4933ad3d8db1"
type: sara_adr
name: "risk の親に test_plan を含めない"
status: 採用
issues:
  - "#583"
  - "#237"
  - "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §2b を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "ADR-0049"
references:
  - "SDR-34916a83-0cff-4524-a8ff-5a3195af52c3"
---
# SDR-f3f9491f-d6f4-4742-82da-4933ad3d8db1: risk の親に test_plan を含めない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: SDR-34916a83-0cff-4524-a8ff-5a3195af52c3, GitHub Issue #583, #237, #203, `docs/model.yaml`, `docs/agents/sara-graph.md`
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §2b を書き直した

## 背景

test_plan を設けた（SDR-34916a83-0cff-4524-a8ff-5a3195af52c3）ことで、risk の親に test_plan を加える
かどうかを決める必要があった。テスト計画はどの risk を扱うかを述べるので、両者には関係がある。

## 決定

risk の `parent_types` は `requirement` と `design` だけにし、test_plan を含めない。

## 根拠

risk の親は「そのリスクが何に対して発生しているか」を表す。リスクが発生するのは要求や設計に対してで、
テスト計画はリスクへの対処の側にある。

テスト計画がどの risk を扱うかは、下流の test_condition から `mitigates` で risk へ張る辺で既に辿れる。
親としても張ると同じ関係を二重に持つ。

## 棄却した代替案

### risk の親に test_plan を加える

リスクの発生元と対処を同じ親の辺で表すことになり、`threatens` の意味がぼやける。

### risk と test_plan の間に peer relation を新設する

テスト計画との紐付けは test_condition 経由で足りている。不足が見えないうちに relation を足さない。
