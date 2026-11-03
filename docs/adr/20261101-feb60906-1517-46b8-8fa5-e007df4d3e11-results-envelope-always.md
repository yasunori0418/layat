---
id: "ADR-feb60906-1517-46b8-8fa5-e007df4d3e11"
type: adr
name: "--json は常に results[] エンベロープを出す"
status: 採用
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §2 を書き直した"
justifies:
  - "REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead"
  - "REQ-5c2e64c3-09a7-4ae8-b60c-4f1ccabce4fd"
  - "REQ-2353259f-5878-452a-8e11-3445de69abc2"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
---
# ADR-feb60906-1517-46b8-8fa5-e007df4d3e11: --json は常に results[] エンベロープを出す

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, GitHub Issue #552, yasunori0418/outturn#1, outturn ADR-0013
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §2 を書き直した

## 背景

layat には、単一の config を扱うコマンドと、`--all` で複数の config を扱うコマンドがある。
エンベロープの容器は、扱う主体が 1 つか複数かで形を分けることも、1 つの形にそろえることもできた。

## 決定

`--json` は stdout にエンベロープを 1 文書だけ書く。トップレベルは常に `results[]`（`SubjectResult` の
配列）とし、単数の `result` をトップに置かない。実行形態を区別する判別子フィールドも持たない。
`specVersion` / `tool` / `command` はトップに 1 度だけ置く。

フィールドの一覧は REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead に、出す時機は
REQ-5c2e64c3-09a7-4ae8-b60c-4f1ccabce4fd に、stdout の専有は REQ-2353259f-5878-452a-8e11-3445de69abc2 に置く。

## 根拠

件数や起動の形で出力の形が変わると、消費側は形を予測できない。outturn も当初は起動の性質で決まる
`mode` を持っていた。容器を常時 `results[]` にそろえると、`mode` は要素数や subject の有無と同じ
情報を二重に持つだけになり、outturn ADR-0013 で廃止された。消費側が依存してよいのは、一様な
`results[]` の形だけになる。

## 棄却した代替案

### 単一実行ではトップに `result` を置き、`--all` だけ `results[]` にする

消費側が 2 つの形を読み分ける必要がある。

### `mode` 判別子で形を区別する

要素数と subject から分かる情報を二重に持つ。
