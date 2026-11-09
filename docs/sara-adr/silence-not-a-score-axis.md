---
id: "SDR-7178b787-cf37-4753-9cea-12a0d65df410"
type: sara_adr
name: "沈黙性を risk のスコア軸にせず本文に書く"
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-303"
    resource: "https://github.com/yasunori0418/layat/issues/303"
    title: "#303"
  - id: "issue-292"
    resource: "https://github.com/yasunori0418/layat/issues/292"
    title: "#292"
origin: "Issue #303 の着手前 grilling（2026-08-12）で確定。ADR-0052 §1c と題名の決定を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-0fd284d6-60dc-4c04-a413-3d7926ae10ac"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# 沈黙性を risk のスコア軸にせず本文に書く

- ステータス: stable
- 日付: 2026-11-03
- 関連: [risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md), GitHub Issue #583, #303, #292, `docs/agents/sara-graph.md`
- 起点: Issue #303 の着手前 grilling（2026-08-12）で確定。ADR-0052 §1c と題名の決定を書き直した

## 背景

[risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md) と決め、`impact` から沈黙の不整合を外した。
気づけるかどうかは脅威の性質として重要な情報で、捨てるには惜しい。これをどこに残すかを決める必要があった。

## 決定

沈黙性（気づけるかどうか）は risk のスコアの軸にせず、各 risk の本文に散文で書く。detectability を
`impact` と並ぶ第 2 の軸としては持たない。

## 根拠

軸として持つかどうかは、情報の価値ではなく、その軸で risk の集合を分けられるかで決まる。layat では沈黙する
脅威がドメイン全体に広がっており、detectability の軸も `impact` と同じ偏りを再現するだけになる。

散文に残せば、情報は失われない。失われるのはスコアの軸としての地位だけである。

## 棄却した代替案

### detectability を第 2 の軸として追加する

`docs/model.yaml` へのフィールド追加、全 risk への値の追記、`level` の導出表を 9 セルから 27 セルへ広げる
変更が要る。それでも得られる序列は今と変わらない。

沈黙する脅威と見える脅威が同じくらいの数に分かれるほどドメインの性質が変われば、この判断を見直す余地がある。
