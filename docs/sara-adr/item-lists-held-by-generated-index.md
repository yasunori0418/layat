---
id: "SDR-958fd2ec-836d-41ae-b67f-4d316ddabf4b"
type: sara_adr
name: "item の一覧は生成した index.md だけが持つ"
description: "型ごとの item の一覧は okf-index.sh が生成する index.md だけが持ち、手書きの文書は一覧を持たず、item を足したときは index.md を再生成するだけにする"
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
  at: "2026-11-14T16:43:01+09:00"
---
# item の一覧は生成した index.md だけが持つ

- ステータス: stable
- 日付: 2026-11-13
- 関連: GitHub Issue #719, `dev/scripts/okf-index.sh`
- 起点: epic #719 の grilling（2026-11-13）で確定

## 背景

[概要文書は item へのリンク集に縮退させる](overview-docs-as-link-index.md) で、`docs/spec.md`・
`docs/design.md`・`docs/concept.md` は全体像と item へのリンク集になった。その後
[OKF の index.md を sara check の出力から生成してコミットする](okf-index-generation.md) で、型ごとの
index.md を生成するようになった。3 文書のリンク集は index.md と同じ一覧を二重に持つ（#719）。

手書きのリンク集は item を足すたびに直す必要がある。index.md との食い違いを検出する手段も無い。

## 決定

型ごとの item の一覧は、`dev/scripts/okf-index.sh` が生成する index.md だけが持つ。この決定に沿って、次のとおりにする。

- 手書きの文書は item の一覧を持たない。手書きの文書から item へ張るリンクは、散文が必要とする
  ものに限る
- item を足したときは index.md を再生成するだけにし、手書きの文書へリンクを足さない
- 主題ごとの節立ては、item の `section` から index.md の節として生成する
  （→ [index.md の節は 6 つの型で item の section から生成する](index-sections-in-six-types.md)）

## 根拠

一覧を生成物だけに置く根拠は、item を足したときに直す手書きの文書が無くなることにある。生成物と item の食い違いは
`dev/tests/okf-index.sh` が検出するので、手書きのリンク集で起きていた直し忘れは起きない。

リンク集を残した理由は、全体像を通読する起点が要ることだった。起点に要るのは全体像の散文で、
item の一覧ではない。一覧は index.md から辿れる。

## 棄却した代替案

### 手書きのリンク集と index.md の両方を持ち、食い違いを契約テストで検出する

ずれは検出できるが、item を足すたびに 2 か所を直す手間は減らない。

### index.md を手で書き、生成をやめる

item の `name` や `description` を変えるたびに一覧を直すことになり、手書きのリンク集と同じ直し忘れが
起きる。
