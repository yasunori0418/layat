---
id: "SDR-5cddbc1a-b9bc-46ae-8272-c567153643f9"
type: sara_adr
name: "specification の必須の強度を SHALL 系の綴りに固定する"
description: "specification の必須の強度は SHALL 系だけで書き、specification_ja は同じ強度の規範助動詞で終える。"
section: authoring
status: stable
sources:
  - id: "issue-649"
    resource: "https://github.com/yasunori0418/layat/issues/649"
    title: "#649"
origin: "Issue #649 の grilling で確定"
justifies:
  - "QA-1f2fac5e-cfc3-46d7-89eb-34a9be7a07bf"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:12:36+09:00"
---
# specification の必須の強度を SHALL 系の綴りに固定する

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #649, `docs/agents/sara-graph.md`
- 起点: Issue #649 の grilling で確定

## 背景

`requirement`・`quality`・`test_plan` は、規範を英語の `specification` と日本語の
`specification_ja` の 2 つの欄に持つ。sara は型を問わず `specification` に RFC2119 キーワードが
あるかを確かめるが、どの綴りかは問わず、`specification_ja` は見ない。綴りと強度の対応は
`docs/agents/sara-graph.md` の規約として運用してきたが、決定の記録は無かった。

## 決定

- `specification` は必須の強度を SHALL / SHALL NOT だけで書き、MUST / MUST NOT を書かない。
  SHOULD / SHOULD NOT / MAY はそれぞれの強度のまま使う
- `specification_ja` で規範を述べる文は、同じ強度の規範助動詞で終える。背景・根拠・例を述べる文は
  平叙文のままにする
- 強度の対応は SHALL を「しなければならない」、SHALL NOT を「してはならない」、SHOULD を
  「すべきである」、SHOULD NOT を「すべきでない」、MAY を「してもよい」に固定する
- これらはレビューで守る。検査が要るときは散文の linter（textlint など）に置き、自前のスクリプトは書かない

## 根拠

2 つの欄はどちらも規範で、どちらも他方の訳注ではない。綴りと強度の対応を固定すれば、2 つの欄を
1 つの規範として読める。

## 棄却した代替案

### MUST と SHALL の混在を許す

RFC2119 では同義だが、同じ強度を 2 つの綴りで書くと、item 群は 2 つが違う強度であるかのように読める。

### `specification_ja` を平叙文で書く

平叙文（「〜する」）の規範は、実装への要求ではなく今の実装の説明として読めてしまう。
