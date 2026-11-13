---
id: "SDR-2c5c87e3-f502-4b26-aca3-7617053281fb"
type: sara_adr
name: "手書きの概要は各ディレクトリの README.md に置く"
description: "item から導けない全体像の散文は、その主題の item を置くディレクトリの README.md に置き、README.md は frontmatter と item の一覧を持たず、同じディレクトリの index.md を 1 本だけ指す"
status: stable
sources:
  - id: "issue-719"
    resource: "https://github.com/yasunori0418/layat/issues/719"
    title: "#719"
origin: "epic #719 の grilling（2026-11-13）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
supersedes:
  - "SDR-536848bb-8901-4e73-8c66-861de3f91bcd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T14:35:49+09:00"
---
# 手書きの概要は各ディレクトリの README.md に置く

- ステータス: stable
- 日付: 2026-11-13
- 関連: [概要文書は item へのリンク集に縮退させる](overview-docs-as-link-index.md), [item の一覧は生成した index.md だけが持つ](item-lists-held-by-generated-index.md), GitHub Issue #719, `docs/README.md`, `docs/requirements/README.md`, `docs/design/README.md`
- 起点: epic #719 の grilling（2026-11-13）で確定

## 背景

[item の一覧は生成した index.md だけが持つ](item-lists-held-by-generated-index.md) と決めたので、
`docs/concept.md`・`docs/spec.md`・`docs/design.md` から item の一覧が外れる。残るのは、設計の哲学・
既存ツールとの比較・north-star・設計目標の表など、item から導けない全体像の散文である。この散文を
どこに置くかを決める必要があった。

## 決定

手書きの概要はその主題の item を置くディレクトリの README.md に置く。

- `docs/concept.md` を `docs/README.md` に、`docs/spec.md` を `docs/requirements/README.md` に、
  `docs/design.md` を `docs/design/README.md` に移す
- README.md は frontmatter を持たず、item の一覧も持たない。冒頭で同じディレクトリの index.md を
  1 本だけ指す
- 散文の中のリンクは、散文が必要とするもの（根拠の ADR など）に限る
- `docs/index.md` から README.md へのリンクは足さない
- 書き方の規約（散文の行数上限など）は、各 README.md の冒頭に置く

## 根拠

GitHub はディレクトリを開くと README.md を表示する。ディレクトリを辿った読み手は、全体像の散文と
index.md への入口に同じ場所で出会える。

frontmatter を持たない README.md は `sara` が読み飛ばすので、item でない文書がグラフに混ざらない。
散文は item から導けないので生成できず、手で書くしかない。一覧を持たなければ、item を足しても
README.md を直す必要は無い。

## 棄却した代替案

### 3 文書を `docs/` 直下に残し、一覧だけを外す

文書の名前と置き場所が型のディレクトリと対応しない。読み手は全体像と一覧を別々の場所で探すことに
なる。

### 散文も item にする

哲学・比較・north-star は規範にあたらず、受け皿になる型も無い。item にすると、規範を持たない item が
グラフに混ざる。

### 散文を index.md の冒頭に埋め込んで生成する

生成スクリプトが手書きの散文を取り込む仕組みを持つことになり、生成物と手書きの文書の境界が
曖昧になる。
