---
id: "SDR-d5304533-00ae-423b-a459-7c67c1cf093c"
type: sara_adr
name: "提案を棄却した決定は status 不採用 で記録する"
status: deprecated
sources:
  - id: "issue-499"
    resource: "https://github.com/yasunori0418/layat/issues/499"
    title: "#499"
origin: "Issue #499 の grilling（2026-11-01）で確定。ADR-0069 の付随決定と ADR-0048 §2 の status enum を書き直した"
justifies:
  - "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T22:51:53+09:00"
---
# SDR-d5304533-00ae-423b-a459-7c67c1cf093c: 提案を棄却した決定は status 不採用 で記録する

> **廃止（2026-11-07）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → [status を OKF の 3 値にし全ての型に持たせる](status-okf-values.md)

- ステータス: deprecated
- 日付: 2026-11-01
- 関連: GitHub Issue #499, `docs/agents/adr-rules.md`
- 起点: Issue #499 の grilling（2026-11-01）で確定。ADR-0069 の付随決定と ADR-0048 §2 の status enum を書き直した

## 背景

ADR の `status` は、実データの表記に合わせて日本語の値を持つ enum にしている。組み込みの英語の値
（`proposed` / `accepted` / `deprecated`）は使っていない。

提案を検討して採らないと決めた結果も、再提案のときの判断基準として ADR に残したい。`採用` の ADR に
「採らない」と書くと、ステータスだけを見た読者は何かを採用した記録だと受け取る。

## 決定

提案を棄却した決定を記録する ADR は `status: 不採用` とする。`不採用` は「採らない」という決定が
有効であることを表し、ADR 自体は有効なまま残る。

ADR の `status` は日本語の値を持つ enum とし、値は `提案` / `採用` / `不採用` / `廃止` の 4 つ。
各値の意味は `docs/agents/adr-rules.md` の「ステータス」節に置く。

## 根拠

棄却をステータスで表せば、ステータスだけを見て「やらないと決めた記録」だと分かる。

`廃止` とは意味が違う。`廃止` は後継に置き換えられて失効した ADR で、`不採用` はいまも有効な
棄却の決定である。値を分けておけば、棄却の記録を失効したものと取り違えない。

## 棄却した代替案

### `却下` を加える

`不採用` と意味が重なる。似た値が 2 つあると、どちらを使うかを書き手が毎回判断することになる。

### 棄却を `採用` のまま本文で表す

ステータスと決定の向きが食い違い、ステータスでの絞り込みが使えなくなる。
