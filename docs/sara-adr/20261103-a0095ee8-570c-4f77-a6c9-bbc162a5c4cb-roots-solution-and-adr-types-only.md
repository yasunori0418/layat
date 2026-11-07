---
id: "SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb"
type: sara_adr
name: "root は solution と ADR 系の型だけにする"
status: deprecated
issues:
  - "#583"
  - "#237"
  - "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §1b を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-c15921a9-7aea-407c-a34b-7465783b1af1"
  - "SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:29:00+09:00"
---
# SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb: root は solution と ADR 系の型だけにする

> **廃止（2026-11-07）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → SDR-d3f88093-dc98-4703-af45-df5b27527ffa

- ステータス: deprecated
- 日付: 2026-11-03
- 関連: SDR-c15921a9-7aea-407c-a34b-7465783b1af1, SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed, SDR-2457aa00-fab7-4c99-af53-2e6643c21c79, GitHub Issue #583, #237, #203, `docs/model.yaml`
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §1b を書き直した

## 背景

sara の接続漏れの判定は、item が `direction: upstream` の relation を実際に 1 本でも張っているかだけで
決まる。型単位・item 単位で接続漏れの warning を抑止する設定は無い。

親を持たない型（root）の item は、構造上いつも接続漏れになる。root の型が増えるほど、warning の中で
張り忘れではない常在の項目の割合が増え、本物の張り忘れが埋もれる。

## 決定

親を持たない型は solution と ADR 系の型だけにする。それ以外の型は
`parent_types` を持ち、各 item は upstream relation を 1 本以上張る。これで「張り忘れ = 接続漏れの
warning」が全型で成り立つ。型と relation の定義は `docs/model.yaml` に置く。

## 根拠

solution はグラフの根で、親を持たないのが定義である。ADR 系の型は規範の外にある決定の記録で、
`justifies` で規範の item へ接続する。どちらも上流を持たないことに理由がある。

それ以外の型に親が無いのは、受け皿となる型の欠落か張り忘れのどちらかである。どちらも接続漏れの warning
として見えるほうがよい。

## 棄却した代替案

### 接続漏れの warning を型単位で抑止する

sara に抑止の設定が無い。仮にあっても、抑止した型では本物の張り忘れも見えなくなる。

### 受け皿の無い item を root の型に置いたままにする

その型の item が構造上ずっと接続漏れになり、warning が張り忘れの信号としてはたらかなくなる。
