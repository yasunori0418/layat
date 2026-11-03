---
id: "SDR-9bec5976-67cd-407a-afa9-3d394a6204ca"
type: sara_adr
name: "欠陥 issue のクローズ時にテスト系 item の欠落を追補する"
status: 採用
issues:
  - "#583"
  - "#268"
  - "#203"
  - "#267"
origin: "Issue #268（epic #203 残ブロッカー）の判断（2026-08-09）で確定。ADR-0051 §3 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed"
---
# SDR-9bec5976-67cd-407a-afa9-3d394a6204ca: 欠陥 issue のクローズ時にテスト系 item の欠落を追補する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed, GitHub Issue #583, #268, #203, #267, `docs/agents/issue-tracker.md`
- 起点: Issue #268（epic #203 残ブロッカー）の判断（2026-08-09）で確定。ADR-0051 §3 を書き直した

## 背景

欠陥はグラフに置かず GitHub Issues で管理する（SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed）。グラフの
risk / test_condition は予防の体系だが、起きた欠陥から学ばなければ、ある時点の分析を写したまま古びていく。

## 決定

欠陥 issue をクローズする前に、その欠陥を事前に捕まえられたはずの risk / test_condition が欠けて
いなかったかを検討する。欠けていれば、その item を `docs/` に追補する。手順は
`docs/agents/issue-tracker.md` の「Defect issues」節に置く。

## 根拠

欠陥の記録はグラフの外に置くが、欠陥から得た教訓は規範としてグラフへ戻す必要がある。

## 棄却した代替案

### 追補を検討せずに欠陥 issue をクローズする

risk / test_condition が起きた欠陥から更新されず、ある時点の分析を写したまま古びていく。
