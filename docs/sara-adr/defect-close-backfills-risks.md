---
id: "SDR-9bec5976-67cd-407a-afa9-3d394a6204ca"
type: sara_adr
name: "欠陥 issue のクローズ時にテスト系 item の欠落を追補する"
section: risk-test-defect
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
origin: "Issue #268（epic #203 残ブロッカー）の判断（2026-08-09）で確定。ADR-0051 §3 を書き直した"
justifies:
  - "QA-87b7776a-9ece-42ac-9a56-78daacb42217"
references:
  - "SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# 欠陥 issue のクローズ時にテスト系 item の欠落を追補する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [欠陥はグラフに置かず GitHub Issues で管理する](defects-in-github-issues.md), [model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](model-yaml-full-replacement.md), GitHub Issue #583, #268, #203, #267, `docs/agents/issue-tracker.md`
- 起点: Issue #268（epic #203 残ブロッカー）の判断（2026-08-09）で確定。ADR-0051 §3 を書き直した

## 背景

[欠陥はグラフに置かず GitHub Issues で管理する](defects-in-github-issues.md)。グラフの
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
