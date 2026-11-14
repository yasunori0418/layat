---
id: "SDR-3fe7901b-5cc0-4677-8761-fcdd4118450a"
type: sara_adr
name: "deprecated の item は index.md の末尾の節にまとめる"
description: "index.md は status が deprecated の item を主題の並びと section の節から外し、末尾の「有効でない item」の節にファイル名順でまとめて出す"
status: stable
sources:
  - id: "issue-743"
    resource: "https://github.com/yasunori0418/layat/issues/743"
    title: "#743"
origin: "epic #743 の grilling（2026-11-14）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
references:
  - "SDR-52dccd7e-666c-46f8-8a8d-1a50dd0f721b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T16:13:50+09:00"
---
# deprecated の item は index.md の末尾の節にまとめる

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #743, `dev/scripts/okf-index.sh`
- 起点: epic #743 の grilling（2026-11-14）で確定

## 背景

[OKF の index.md を sara check の出力から生成してコミットする](okf-index-generation.md) で生成する
index.md は、`status` が `deprecated` の item を有効な item と同じ並びに置いている。deprecated の
item は `requirement` に 1 件、`design` に 1 件、`adr` に 5 件、`sara_adr` に 14 件ある。行末に
ステータスが出るのは `adr` と `sara_adr` だけで、ほかの型では行を見ても有効かどうか分からない。

deprecated の item には、後の決定に置き換えられた決定のほかに、棄却した提案（`adr` 4 件・
`requirement` 1 件・`design` 1 件）も含まれる（→ [棄却した提案の要件と設計は提案の形の題で deprecated にして残す](rejected-proposals-deprecated.md)）。
`section` の節に分ける型では、deprecated の item も主題の `section` を持つので、有効な item と
同じ節に並ぶ。

## 決定

index.md は、`status` が `deprecated` の item を主題の並びから外し、末尾の節にまとめて出す。

- 型ごとのディレクトリの index.md は、deprecated の item を主題の並びに出さず、末尾の
  ``## 有効でない item（`deprecated`）`` の下にファイル名順で出す。主題ごとの小見出しは付けない。
  `section` の節に分ける型でも、分けない型でも同じにする
- `docs/test/<区分>/index.md` は、テスト条件の節とテストケースの節から deprecated の item を外し、
  2 つの節の後ろに末尾の節を 1 つ置く。`docs/index.md` と `docs/test/index.md` は item を並べない
  ので、対象にしない
- deprecated の item が 1 件も無い index.md には、末尾の節の見出しを出さない。`status` が `draft` の
  item は主題の並びに残す。`adr` と `sara_adr` の行末のステータス表示は変えない
- 末尾の節を「廃止」と呼ばない。`deprecated` は置き換えられた決定のほかに、棄却した提案も
  含むためである
- `section` の enum に deprecated を表す値は足さない。deprecated の item も主題の `section` を持つ

## 根拠

主題の並びに有効な item だけを残せば、いまの規範を探す読み手は行末の表示を確かめずに済む。
deprecated の item は後継の決定や棄却した理由を辿るときに読むので、一覧から消さずに末尾へ残す。
item の一覧は生成した index.md だけが持つので、一覧から外すと、その item へ辿れなくなる。

振り分けは、生成スクリプトが frontmatter の `status` を見て行う。`status` は item が既に持つ値なので、
item を deprecated にするときに書き換える値は 1 つで済む。末尾の節に入るのは 1 本の index.md あたり
多くて 14 件で、主題ごとの小見出しを付けると 1〜2 件の節が並ぶ。

`draft` の item は置き換えられても棄却されてもいないので、主題の並びに残す。

## 棄却した代替案

### `section` の enum に deprecated を表す値を足す

`status` と `section` の 2 か所で同じことを表すことになる。item を deprecated にするたびに
`section` も書き換える必要があり、片方だけ直したときの食い違いが残る。元の主題も失われる。

### 末尾の節の中を主題ごとの小見出しに分ける

件数が少なく、小見出しの多くが 1〜2 件の節になる。

### 主題の節に残し、行末のステータス表示で見分ける

行末のステータス表示は `adr` と `sara_adr` にしか出ず、ほかの型では見分けられない。2 型でも、
有効な item を探すたびに行末を確かめる必要が残る。
