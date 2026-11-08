---
id: "ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511"
type: adr
name: "main へのマージにテストの成功を必須にする"
status: stable
origin: "ADR-0030 §1a（main の ruleset で required status check を課す）を書き直した"
justifies:
  - "INF-8b97573f-d4d6-4abf-85e2-d859afbd96c6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511: main へのマージにテストの成功を必須にする

- ステータス: stable
- 日付: 2026-11-03
- 起点: ADR-0030 §1a（main の ruleset で required status check を課す）を書き直した

## 背景

PR では flake check の matrix と E2E が実行される。しかしそれだけでは、テストが失敗していても main へ
マージできてしまう。「PR 経由でマージし、テストを通す」という運用ルールは、技術的には担保されて
いなかった。

## 決定

main の ruleset で、CI が PR で出す flake check の 3 つの check と `e2e` を required status check に
する。bypass を許さず、owner も PR とテストを通さなければマージできない。ブランチの up-to-date は
強制せず、レビューの承認も要求しない。

必須にする check 名と ruleset の設定は [マージの条件（main の ruleset と required status checks）](../infrastructure/merge-gate.md) に置く。

## 根拠

運用ルールを技術的な強制に置き換えると、誤ったマージを構造的に防げる。個人で運用していて PR が
多くないうちは、up-to-date を強制して CI を再実行させるコストの方が大きい。マージ条件はテストの
成功だけにし、承認者の居ない個人運用で承認を待たせない。

## 棄却した代替案

### 運用ルールだけで済ませる

強制力が無く、誤ったマージを防げない。

### 旧来の branch protection を使う

ruleset はその後継で、表現力が高く、GitHub も推奨している。

### 成否を集約するジョブを 1 つだけ必須にする

必須にする check 名は 1 つで済むが、集約ジョブの実装が要る。matrix の個々の成否も ruleset から
見えにくくなる。
