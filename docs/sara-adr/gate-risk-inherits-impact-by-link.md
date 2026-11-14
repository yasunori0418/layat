---
id: "SDR-432a0189-5838-4b6f-93da-fd0f6067ddda"
type: sara_adr
name: "自動検証の形骸化は守備対象の impact を継承し、継承元をリンクで名指しする"
description: "自動検証の形骸化を表す risk の `impact` は守備対象の regression から継承し、継承元はその regression を持つ risk への Markdown リンクで名指しする。"
section: risk-test-defect
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
  - id: "issue-692"
    resource: "https://github.com/yasunori0418/layat/issues/692"
    title: "epic: 散文リンク移行の後始末 — 表示崩れ・重複表記・文の破綻・frontmatter のフル ID"
origin: "epic #692 の grilling（2026-11-08）で確定。[自動検証の形骸化は守備対象の impact を継承する](gate-risk-inherits-impact.md) を書き直した"
justifies:
  - "QA-b14e0a14-b9c9-4cc6-9b7a-0ce95d30d289"
supersedes:
  - "SDR-c2989d44-2ba4-4a42-9f4f-35871405be65"
references:
  - "SDR-0fd284d6-60dc-4c04-a413-3d7926ae10ac"
  - "SDR-1a02013e-c57b-4079-9c72-8e9fe99e499a"
  - "SDR-7178b787-cf37-4753-9cea-12a0d65df410"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:06:00+09:00"
---
# 自動検証の形骸化は守備対象の impact を継承し、継承元をリンクで名指しする

- ステータス: stable
- 日付: 2026-11-08
- 関連: [risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md), [複数の失敗方向を持つ risk は最も重い方向で採点する](risk-heaviest-direction.md), [沈黙性を risk のスコア軸にせず本文に書く](silence-not-a-score-axis.md), GitHub Issue #583, #303, #292, #692, `docs/agents/sara-graph.md`
- 起点: epic #692 の grilling（2026-11-08）で確定。[自動検証の形骸化は守備対象の impact を継承する](gate-risk-inherits-impact.md) を書き直した

## 背景

自動検証が気づかれないまま形骸化する risk（テストは緑なのに何も確認していない類）は、その自動検証自身が開発の中で
完結する。[risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md) と
[沈黙性を risk のスコア軸にせず本文に書く](silence-not-a-score-axis.md) の 2 つの決定の下では、この類型はそのまま読めば low になる。

## 決定

自動検証の形骸化を表す risk の `impact` は、その自動検証が捕まえるはずの regression の `impact` を継承する。
自動検証が開発の中で完結することを理由に low へ下げない。1 つの自動検証が複数の regression を守るときは、
[複数の失敗方向を持つ risk は最も重い方向で採点する](risk-heaviest-direction.md) に従い、最も重い守備対象を継承する。

継承元は item の本文で、その regression を持つ risk への Markdown リンクで名指しする。手順は
`docs/agents/sara-graph.md` に置く。

## 根拠

テストの検出力が失われた影響は、守っていたはずの脅威が無防備になる形でしか現れない。重さは自動検証では
なく守備対象で決まる。

継承先の値は継承元の値に従う。継承元を名指ししておかないと、継承元を採点し直したときに追従すべき item を
辿れない。リンクなら、読み手は継承元を名前で読め、その item へそのまま移動できる。

## 棄却した代替案

### 自動検証の形骸化を一律に low とする

開発の中で完結するという見かけで採点すると、守備対象が実データの破壊でも low になり、テストの劣化が
見過ごされる。

### 沈黙するので一律に high とする

沈黙を軸にしない決定と矛盾する。守備対象が軽い自動検証まで high に並び、序列が付かなくなる。
