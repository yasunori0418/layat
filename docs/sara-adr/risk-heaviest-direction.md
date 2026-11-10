---
id: "SDR-1a02013e-c57b-4079-9c72-8e9fe99e499a"
type: sara_adr
name: "複数の失敗方向を持つ risk は最も重い方向で採点する"
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
origin: "Issue #303 の着手前 grilling（2026-08-12）で確定。ADR-0052 §1b を書き直した"
justifies:
  - "QA-b14e0a14-b9c9-4cc6-9b7a-0ce95d30d289"
references:
  - "SDR-0fd284d6-60dc-4c04-a413-3d7926ae10ac"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# 複数の失敗方向を持つ risk は最も重い方向で採点する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md), GitHub Issue #583, #303, #292, `docs/agents/sara-graph.md`
- 起点: Issue #303 の着手前 grilling（2026-08-12）で確定。ADR-0052 §1b を書き直した

## 背景

[risk の impact を検出後の回復可能性で採点する](risk-impact-as-recoverability.md) と決めた。1 つの
risk が複数の失敗の仕方を持ち、方向によって回復できる度合いが違うとき、どの方向で採点するかを決める
必要があった。

## 決定

複数の失敗方向を持つ risk は、最も重い方向で `impact` を採る。ある方向を判定から外してよいのは、別の
risk がその方向を自分の射程として明示していることを確かめ、その risk を名指しできるときだけにする。

## 根拠

平均や最も起こりやすい方向で採ると、その risk を立てた理由である最も重い失敗方向が採点から消える。

委譲の先を確かめずに方向を外すと、受け取る側の射程が実はその方向に届いていないことがある。そうなると
どちらの item も計上せず、重い方向がグラフ全体から抜け落ちる。

## 棄却した代替案

### 最も起こりやすい方向で採る

その risk を立てた理由である最も重い失敗方向が採点から消える。平均で採っても同じことが起きる。

### 別の risk が扱っていそうな方向は確かめずに外す

受け取る側の射程がその方向に届いていなければ、重い方向をどの item も計上しない。
