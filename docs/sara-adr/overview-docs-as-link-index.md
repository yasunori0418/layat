---
id: "SDR-536848bb-8901-4e73-8c66-861de3f91bcd"
type: sara_adr
name: "概要文書は item へのリンク集に縮退させる"
status: stable
sources:
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §1 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-5ea31001-96f8-4324-a40b-647cdfa5cc9d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# SDR-536848bb-8901-4e73-8c66-861de3f91bcd: 概要文書は item へのリンク集に縮退させる

- ステータス: stable
- 日付: 2026-11-01
- 関連: [ドキュメントは sara のナレッジグラフとして item に分割する](docs-as-sara-knowledge-graph.md), GitHub Issue #203, `docs/concept.md`, `docs/design.md`, `docs/spec.md`
- 起点: epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §1 を書き直した

## 背景

[ドキュメントは sara のナレッジグラフとして item に分割する](docs-as-sara-knowledge-graph.md) と決めたので、`docs/spec.md`・
`docs/design.md`・`docs/concept.md` の規範的な内容は item へ移る。元の 3 文書をどう扱うかを決める
必要があった。

## 決定

3 文書は捨てず、全体像と item へのリンク集に縮退させる。規範的な内容は item が持ち、3 文書は通読の
起点として残す。各文書の書き方の規約（散文の行数上限など）は、それぞれの冒頭に置く。

## 根拠

数十の item を `sara query` で辿ることは、通読の代わりにならない。全体像を上から読める起点が要る。

規範を item だけに置けば、3 文書と item のあいだで内容が食い違わない。3 文書は索引に徹するので、
実装への追従は item を直すことで済む。

## 棄却した代替案

### 3 文書を廃止し、item だけにする

全体像を通読する手段が無くなる。新しく読む人が、どこから読めばよいか分からない。

### 3 文書に規範を残し、item と二重に持つ

同じ規範が 2 か所にあると、片方だけが更新されて食い違う。
