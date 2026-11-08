---
id: "SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed"
type: sara_adr
name: "欠陥はグラフに置かず GitHub Issues で管理する"
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-268"
    resource: "https://github.com/yasunori0418/layat/issues/268"
    title: "#268"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
  - id: "issue-267"
    resource: "https://github.com/yasunori0418/layat/issues/267"
    title: "#267"
origin: "Issue #268（epic #203 残ブロッカー）の判断（2026-08-09）で確定。ADR-0051 §1・§2 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb"
  - "SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed: 欠陥はグラフに置かず GitHub Issues で管理する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root は solution と ADR 系の型だけにする](roots-solution-and-adr-types-only.md), [sara check を strict モードで回す](sara-check-strict-mode.md), [model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](model-yaml-full-replacement.md), GitHub Issue #583, #268, #203, #267, `docs/model.yaml`, `docs/agents/issue-tracker.md`
- 起点: Issue #268（epic #203 残ブロッカー）の判断（2026-08-09）で確定。ADR-0051 §1・§2 を書き直した

## 背景

グラフには欠陥を表す defect 型があった。defect は自分から upstream relation を張れず、親側の test_case が
下流向きの `reveals` で接続する唯一の型だった。そのため次の例外を抱えていた。

- defect item は作った時点で必ず接続漏れになる
- [sara check を strict モードで回す](sara-check-strict-mode.md) ので、defect を起こすと即座に
  exit 1 になる
- 型の関係図を機械生成するとき、上流の辺だけを描く規約から `reveals` だけが外れる

## 決定

欠陥はグラフに置かず、GitHub Issues（`bug` label）で管理する。`docs/model.yaml` は defect 型と
`reveals` / `is_revealed_by` を持たず、テスト系の型は risk → test_condition → test_case で終わる。

トレーサビリティは issue から docs への片方向にする。欠陥 issue の本文に、発見元の CASE と遡る先の
REQ / DSG を書く。docs から issue へは張らない。運用の細目は `docs/agents/issue-tracker.md` の
「Defect issues」節に置く。

## 根拠

docs のグラフは、要求・設計からリスクとテストの観点を前倒しで管理する予防の体系で、持つのは規範
（あるべき姿）である。欠陥は規範ではなく、起きた出来事の記録になる。出来事の記録・状態の遷移・修正作業の
追跡は issue tracker の領分である。

defect の例外はこの食い違いの症状だった。型を外すと「張り忘れ = 接続漏れの warning」が例外なく全型で成り立ち
（[root は solution と ADR 系の型だけにする](roots-solution-and-adr-types-only.md)）、strict とも無条件に両立する。

## 棄却した代替案

### defect に `is_revealed_by` を足して上流へ張れるようにする

接続漏れと strict の衝突は解けるが、出来事の記録をグラフに持ち込む食い違いは残る。issue との二重管理も
解消しない。

### issue を正として、defect item を接続用のスタブとして残す

item の役目が issue への転送だけになる。欠陥が出るたびに docs へ PR を出す手間と二重管理が残る。

### strict を解除する

defect 1 型のために、全型の機械的な担保を弱めることになる。
