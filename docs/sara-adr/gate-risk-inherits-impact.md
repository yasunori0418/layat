---
id: "SDR-c2989d44-2ba4-4a42-9f4f-35871405be65"
type: sara_adr
name: "自動検証の形骸化は守備対象の impact を継承する"
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
origin: "Issue #303 の着手前 grilling（2026-08-12）で確定。ADR-0052 §2 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-0fd284d6-60dc-4c04-a413-3d7926ae10ac"
  - "SDR-1a02013e-c57b-4079-9c72-8e9fe99e499a"
  - "SDR-7178b787-cf37-4753-9cea-12a0d65df410"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:25:57+09:00"
---
# SDR-c2989d44-2ba4-4a42-9f4f-35871405be65: 自動検証の形骸化は守備対象の impact を継承する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md), [複数の失敗方向を持つ risk は最も重い方向で採点する](risk-heaviest-direction.md), [沈黙性を risk のスコア軸にせず本文に書く](silence-not-a-score-axis.md), GitHub Issue #583, #303, #292, `docs/agents/sara-graph.md`
- 起点: Issue #303 の着手前 grilling（2026-08-12）で確定。ADR-0052 §2 を書き直した

## 背景

自動検証が気づかれないまま形骸化する risk（テストは緑なのに何も確認していない類）は、その自動検証自身が開発の中で
完結する。`impact` を回復可能性で採点し（[risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md)）、沈黙性を軸から
外す（[沈黙性を risk のスコア軸にせず本文に書く](silence-not-a-score-axis.md)）と、この類型はそのまま読めば low になる。

## 決定

自動検証の形骸化を表す risk の `impact` は、その自動検証が捕まえるはずの regression の `impact` を継承する。
自動検証が開発の中で完結することを理由に low へ下げない。1 つの自動検証が複数の regression を守るときは、
最も重い守備対象を継承する（[複数の失敗方向を持つ risk は最も重い方向で採点する](risk-heaviest-direction.md)）。

継承元は item の本文に名指しし、できるだけその regression を持つ risk の ID で書く。手順は
`docs/agents/sara-graph.md` に置く。

## 根拠

テストの検出力が失われた影響は、守っていたはずの脅威が無防備になる形でしか現れない。重さは自動検証では
なく守備対象で決まる。

継承先の値は継承元の値に従う。継承元を名指ししておかないと、継承元を採点し直したときに追従すべき item を
辿れない。

## 棄却した代替案

### 自動検証の形骸化を一律に low とする

開発の中で完結するという見かけで採点すると、守備対象が実データの破壊でも low になり、テストの劣化が
見過ごされる。

### 沈黙するので一律に high とする

沈黙を軸にしない決定と矛盾する。守備対象が軽い自動検証まで high に並び、序列が付かなくなる。
