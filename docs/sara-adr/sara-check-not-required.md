---
id: "SDR-6cc54665-c239-4235-92a8-f73e26c5a5ef"
type: sara_adr
name: "sara check を required status check にしない"
section: verification
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-255"
    resource: "https://github.com/yasunori0418/layat/issues/255"
    title: "#255"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の移行完了（Batch 2・2026-08-08）を受けた残論点判断（Issue #255）で確定。ADR-0050 §2 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626"
  - "ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511"
  - "ADR-8035e7c6-7dc5-4aa8-9a37-403998fa7f9b"
  - "ADR-cb010758-7887-470a-ba61-19b2e99f0053"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# sara check を required status check にしない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [sara check を strict モードで回す](sara-check-strict-mode.md), [main へのマージにテストの成功を必須にする](../adr/tests-required-for-merge.md), [main への直接 push を ruleset で禁じる](../adr/no-direct-push-to-main.md), [変更検出で nix を回さない PR の必須の確認を skip で通す](../adr/change-detection-gates-ci-jobs.md), GitHub Issue #583, #255, #203
- 起点: epic #203 の移行完了（Batch 2・2026-08-08）を受けた残論点判断（Issue #255）で確定。ADR-0050 §2 を書き直した

## 背景

[sara check を strict モードで回す](sara-check-strict-mode.md) と決めたことで、CI の `sara` ジョブは
張り忘れで失敗するようになった。このジョブの失敗でマージを止めるかどうかを決める必要があった。

## 決定

CI の `sara` ジョブを required status check に登録しない。`main` の ruleset と DoD は変えない。
`sara` ジョブは張り忘れで赤くなるが、マージは止めない。

## 根拠

検出は機械に任せ、直し方の判断は人に残す。張り忘れの直し方が自明でない場合（受け皿の item を新しく
起こす必要があるなど）、CI が赤いままレビューで扱う余地を残しておきたい。

## 棄却した代替案

### `sara` ジョブを required status check にする

張り忘れの直し方が決まるまで、関係の無い変更まで含めてマージが止まる。
