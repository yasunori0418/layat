---
id: "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
type: adr
name: "空の entries を正当な全クリアとして許可する"
status: 採用
origin: "ADR-0018 の後に行った docs の横断検証で残っていた細目。ADR-0019 §4 を書き直した"
justifies:
  - "REQ-fc64de4c-c82b-419c-8706-07d8d97daa37"
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
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385: 空の entries を正当な全クリアとして許可する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30, ADR-521765d7-7e31-42f4-b548-979185a7586a, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-86a77232-dc99-4c04-bd6d-2127410e8828, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-74bf4a1f-bb74-44a7-a3cb-3b3e87c24f95, ADR-2119433e-de69-4447-85a5-575d58246007, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-45380022-d959-4b22-aa05-5993ac1f3c87, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c, ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1, ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3, ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995, ADR-c4ad9987-850e-4834-bbea-d8de3b073548, ADR-6d2327f7-f513-42c3-89f8-808467db1914, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, `docs/spec.md`
- 起点: ADR-0018 の後に行った docs の横断検証で残っていた細目。ADR-0019 §4 を書き直した

## 背景

`entries = {}` の config を apply すると、前の世代が置いた symlink が全て古くなったものとして除去され、
その config の配置が全て消える。これを正当な操作として扱うか、誤りとして止めるかを決める必要が
あった。

## 決定

空の `entries` は、その config の配置を全て撤去する正当な表現として許可する。評価時のエラーにも
警告にもしない。

規範は REQ-fc64de4c-c82b-419c-8706-07d8d97daa37 に置く。

## 根拠

古くなった配置の除去は保守的で、layat が置いた symlink だけを消し、実ファイルには触れない（ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c）。
全て撤去しても利用者のデータは失われないので、空を特別に扱う理由が無い。

entry を 1 つ以上必須にすると、config の配置を消すという正当な手段が無くなる。

## 棄却した代替案

### 空の `entries` に警告を出すか、評価時のエラーにする

誤った全クリアに気付く助けにはなるが、意図した撤去を妨げるか、毎回の雑音になる。
