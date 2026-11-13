---
id: "SDR-fcae2485-b7f5-4717-9c3a-4e585161efec"
type: sara_adr
name: "index.md の節は item の section から生成する"
description: "requirement・design・ubiquitous_term の index.md は、item が持つ必須の enum フィールド section ごとに節を分けて生成し、見出しと並び順は生成スクリプトの表に置く"
status: stable
sources:
  - id: "issue-719"
    resource: "https://github.com/yasunori0418/layat/issues/719"
    title: "#719"
origin: "epic #719 の grilling（2026-11-13）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
references:
  - "SDR-52dccd7e-666c-46f8-8a8d-1a50dd0f721b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T14:18:35+09:00"
---
# index.md の節は item の section から生成する

- ステータス: stable
- 日付: 2026-11-13
- 関連: GitHub Issue #719, `docs/model.yaml`, `dev/scripts/okf-index.sh`
- 起点: epic #719 の grilling（2026-11-13）で確定

## 背景

[OKF の index.md を sara check の出力から生成してコミットする](okf-index-generation.md) で生成する
index.md は、型ごとの item をファイル名順に 1 本で並べている。`requirement` は 152 件、`design` は
28 件、`ubiquitous_term` は 30 件あり、1 本の並びでは主題ごとに読めない。

主題ごとの節立ては `docs/spec.md`・`docs/design.md`・`docs/glossary.md` が手で持っている。3 文書は
節の下に item へのリンクを並べており、index.md と同じ一覧を二重に持つ（#719）。手書きの一覧は
item を足すたびに直す必要があり、食い違いを検出する手段も無い。

## 決定

`requirement`・`design`・`ubiquitous_term` の index.md を、item の `section` ごとの節に分けて生成する。

- 3 型に必須の enum フィールド `section` を足す。取りうる値は `docs/model.yaml` の enum が持つ
- `dev/scripts/okf-index.sh` は、3 型の index.md を `section` ごとの `##` の節に分けて出す。見出しの
  文言と節の並び順はスクリプトの中の表に置き、節の中はファイル名順に並べる。item の無い節は出さない
- item の `section` の値が表に無ければ、スクリプトはその値を標準エラーに出し、index.md を書かずに
  exit 2 で止まる
- ほかの型の index.md と `docs/index.md` は節に分けずに出す
- 各 item の値は 3 文書の節立てを写して決める

## 根拠

節を item 自身が持てば、item を足すときにその 1 ファイルで節まで決まり、一覧は生成で揃う。必須の
enum にすると、`section` の無い frontmatter と綴りの誤った値を `sara check` が止める。`sara-new` は
`--section` を省くと enum の先頭の値を入れて起票を終えるので、起票では値を明示して渡す。

見出しの文言と並び順は、モデルの enum に置かずにスクリプトの表に置く。sara の enum は値の集合しか持てず、
見出しの文言を表せないためである。enum に値を足して表を直し忘れても、スクリプトが exit 2 で止まる
ので気付ける。

## 棄却した代替案

### 節立てを概要文書に手で持ち続ける

一覧の二重持ちが残り、index.md との食い違いを検出できない。

### `section` の値を見出しの文言そのものにする

「lib API / `lib.mkManifest`」のような見出しを値にすると、frontmatter に記号と空白を含む値が
並ぶ。見出しの文言を直すたびに、その節の item をすべて書き換えることにもなる。

### 節ごとにディレクトリを分ける

item の移動で、item 同士の相対リンクと出典のパスが大量に変わる。節を変えるたびにファイルを
動かすことにもなる。
