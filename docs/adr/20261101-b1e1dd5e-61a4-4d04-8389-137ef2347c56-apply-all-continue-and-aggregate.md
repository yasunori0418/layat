---
id: "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
type: adr
name: "apply --all は失敗しても残りを続け、最後に集約して非 0 で終える"
status: 採用
issues:
  - "#550"
origin: "ドキュメント計画段階のレビュー（2026-06-13）で確定。ADR-0013 §3 を書き直した"
justifies:
  - "REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af"
supersedes:
  - "ADR-0013"
references:
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
  - "ADR-521765d7-7e31-42f4-b548-979185a7586a"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-74bf4a1f-bb74-44a7-a3cb-3b3e87c24f95"
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
  - "ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995"
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-af37a636-a798-4c03-8fa4-e73ed6375a92"
---
# ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56: apply --all は失敗しても残りを続け、最後に集約して非 0 で終える

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30, ADR-521765d7-7e31-42f4-b548-979185a7586a, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-86a77232-dc99-4c04-bd6d-2127410e8828, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-74bf4a1f-bb74-44a7-a3cb-3b3e87c24f95, ADR-2119433e-de69-4447-85a5-575d58246007, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-45380022-d959-4b22-aa05-5993ac1f3c87, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c, ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1, ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3, ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995, ADR-c4ad9987-850e-4834-bbea-d8de3b073548, ADR-6d2327f7-f513-42c3-89f8-808467db1914, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, ADR-af37a636-a798-4c03-8fa4-e73ed6375a92, GitHub Issue #550
- 起点: ドキュメント計画段階のレビュー（2026-06-13）で確定。ADR-0013 §3 を書き直した

## 背景

`apply --all` は複数の config を 1 回で適用する。そのうち 1 つが失敗したとき、残りを適用するか、
全体を止めるか、全体を巻き戻すかを決める必要がある。

## 決定

`apply --all` は、失敗した config を飛ばして残りの適用を続け、最後に成功と失敗をまとめて表示する。
1 つでも失敗があれば非 0 で終える。`--all` 全体を 1 つの atomic な単位にはしない。

集約と終了の規範は REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af に置く。

## 根拠

各 config は独立した profile を持ち、config ごとに atomic に適用される（ADR-fe035038-4415-4a1e-aa5c-509fdfce9955）。1 つの失敗が
他の config の状態を壊すことはないので、残りを止める理由がない。非 0 で終えれば、集約の結果を
CI の合否判定にも使える。

config は並列に適用する（ADR-e7a57f51-41da-460c-8f51-c84e94643f3d）が、継続と集約の振る舞いは並列でも同じである。

## 棄却した代替案

### 最初の失敗で止める（fail-fast）

config は互いに独立なのに、1 つの失敗で残りが止まる。適用済みと未適用の config が中途半端に
混在し、どこまで進んだかを利用者が追う必要が出る。

### `--all` 全体を atomic にする（全部成功か、全部巻き戻すか）

project mode は世代を公開せず rollback を持たないので、全体の巻き戻しが意味論として成り立たない。
実装も過剰になる。
