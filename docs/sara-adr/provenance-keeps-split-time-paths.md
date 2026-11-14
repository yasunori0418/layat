---
id: "SDR-10e78124-1b8a-4a06-bbc6-1d42e58eea27"
type: sara_adr
name: "item の出典は分割した時点のパスのまま残す"
description: "item の `## 出典` にある分割した時点の文書のパスは書き換えず、出典の読み方は `docs/model.md` の 1 か所に置く"
section: authoring
status: stable
sources:
  - id: "issue-719"
    resource: "https://github.com/yasunori0418/layat/issues/719"
    title: "#719"
origin: "epic #719 の grilling（2026-11-13）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T16:29:14+09:00"
---
# item の出典は分割した時点のパスのまま残す

- ステータス: stable
- 日付: 2026-11-13
- 関連: GitHub Issue #719, `docs/model.md`
- 起点: epic #719 の grilling（2026-11-13）で確定

## 背景

item の `## 出典` は epic #203 で文書を item に分割した時点に、原文のどこにあったかを記録している。
[手書きの概要は各ディレクトリの README.md に置く](overview-prose-in-directory-readme.md) で 3 文書を
移し、`docs/glossary.md` も廃止するので、出典が名指しする `docs/spec.md`・`docs/design.md`・
`docs/concept.md`・`docs/glossary.md` は現在のリポジトリから無くなる。

これらの旧パスは 166 件の item の出典にある。旧パスを書き換えるかを決める必要があった。

## 決定

item の `## 出典` にある旧パスは書き換えない。

- 出典は分割した時点の原文の位置の記録として読む。item から現行の仕様へ辿るには relation
  （`sara query`）を使う
- 出典の読み方と、git の履歴から原文を読む手順は、`docs/model.md` の「item の `## 出典` の読み方」の
  節に置く。README.md の書き方の規約は、その節を指す 1 行だけを持つ

## 根拠

出典が記録するのは、分割した時点に原文のどこにあったかであり、現在の置き場所ではない。これが旧パスを残す根拠である。新しい
パスへ書き換えても、移した README.md には一覧も節の本文も残っていないので、指す先に記述の無い
出典になる。旧パスのままなら、`git show` で分割した時点の原文を開ける。

読み方を `docs/model.md` の 1 か所に置けば、README.md の 3 本に同じ手順を重ねて持たずに済む。

## 棄却した代替案

### 出典の旧パスを新しいパスへ書き換える

166 件の item の本文が変わり、`generated` の更新も要る。書き換えた先には分割した時点の記述が無いので、
出典の役目を果たさない。

### 出典の節を削る

原文へ戻る手がかりが無くなり、item の由来を後から確かめられない。
