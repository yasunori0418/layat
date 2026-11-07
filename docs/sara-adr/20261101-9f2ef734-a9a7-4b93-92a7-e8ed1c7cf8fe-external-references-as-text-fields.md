---
id: "SDR-9f2ef734-a9a7-4b93-92a7-e8ed1c7cf8fe"
type: sara_adr
name: "item にできない外部参照は text フィールドへ逃がす"
status: stable
sources:
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §2 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:32+09:00"
---
# SDR-9f2ef734-a9a7-4b93-92a7-e8ed1c7cf8fe: item にできない外部参照は text フィールドへ逃がす

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #203, `docs/adr/README.md`
- 起点: epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §2 を書き直した

## 背景

ADR は GitHub の Issue や、決定の起点になったセッションを参照する。これらは sara の item ではないので、
relation では表せない。relation の参照先は実在する item でなければならず、実在しない ID は破損した
参照として検出される。

## 決定

item にできない外部参照は text のフィールドに書く。ADR では、Issue 番号を `issues`（text のリスト）に、
起点を `origin`（text）に置く。どの本文の行をどのフィールドへ写すかは `docs/adr/README.md` に置く。

## 根拠

外部参照を relation から外せば、破損した参照の確認が item どうしの参照だけを見るようになり、誤検出が
起きない。text として frontmatter に残すので、Issue 番号で ADR を探す手段も失われない。

## 棄却した代替案

### Issue を item として取り込む

Issue の内容は GitHub にあり、取り込むと二重に管理することになる。

### 外部参照を本文だけに書く

frontmatter に無いと、機械的に一覧を取れない。
