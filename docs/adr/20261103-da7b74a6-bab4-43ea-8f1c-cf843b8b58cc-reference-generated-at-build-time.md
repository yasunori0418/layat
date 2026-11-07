---
id: "ADR-da7b74a6-bab4-43ea-8f1c-cf843b8b58cc"
type: adr
name: "リファレンスをビルド時に生成する"
status: stable
sources:
  - id: "issue-57"
    resource: "https://github.com/yasunori0418/layat/issues/57"
    title: "#57"
origin: "Issue #57 のサイト設計。ADR-0037 §3 を書き直した"
justifies:
  - "INF-0865477b-21b2-4feb-9c06-1a882a4e595a"
references:
  - "ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:34:16+09:00"
---
# ADR-da7b74a6-bab4-43ea-8f1c-cf843b8b58cc: リファレンスをビルド時に生成する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4, GitHub Issue #57
- 起点: Issue #57 のサイト設計。ADR-0037 §3 を書き直した

## 背景

ドキュメントサイトには、手書きの usage ガイドのほかに API リファレンスを載せる
（ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4）。リファレンスの出どころは Nix の `lib/` と Go のソースの
2 つである。これをソースからどう作り、生成物をどう扱うかが決まっていなかった。

## 決定

API リファレンスはソースのコメントから抽出して作る。Nix の公開 API には RFC-145 形式の doc-comment を
付けて nixdoc で、Go は godoc のコメントから gomarkdoc で生成する。生成した markdown はリポジトリに
コミットせず、CI とローカルのビルドで毎回生成する。

抽出の対象と、Nix のコメントの使い分けは INF-0865477b-21b2-4feb-9c06-1a882a4e595a に置く。

## 根拠

生成物をコミットすると、生成し忘れや手での編集が紛れ込む余地が残る。ビルドのたびに生成すれば、
ソースとリファレンスがずれないことが仕組みで保証される。

## 棄却した代替案

### 生成した markdown をコミットする

配備の設定は簡単になるが、生成物とソースのずれを CI の確認と人の運用で防ぎ続けることになり、
その負担がずっと残る。
