---
id: "ADR-be5865a8-000a-491f-8438-6308076ea727"
type: adr
name: "rollback の戻り先は世代リンクから読む"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §3c を書き直した"
justifies:
  - "REQ-0e341430-17f0-498b-9439-65491652163a"
references:
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-a91585ad-971a-4558-9f25-b7fcd6e57671"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-be5865a8-000a-491f-8438-6308076ea727: rollback の戻り先は世代リンクから読む

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-a91585ad-971a-4558-9f25-b7fcd6e57671, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §3c を書き直した

## 背景

rollback は、離れる世代 N を比較の基準、戻る世代 N-1 を target
とした配置として行う（ADR-d97154cd-a2e5-4853-a682-800ff9435c89）。
比較の基準は記録を優先して読むようになった。target も記録から読めるのかを決める必要がある。

## 決定

rollback の target（N-1 の manifest）は、引き続き世代リンクから読む。
記録は比較の基準の読み出しにだけ使う。

## 根拠

記録が持つのは最後に置いた manifest 1 つだけで、それより前の世代の manifest を持たない。N-1
の内容を知っているのは世代だけである。

## 棄却した代替案

### 記録に過去の manifest を履歴として持たせ、target もそこから読む

記録が世代の管理を二重に抱えることになり、世代管理を layat の外へ出す方向と逆行する。
