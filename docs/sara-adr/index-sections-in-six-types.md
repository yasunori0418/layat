---
id: "SDR-4e92d0b5-7674-4956-854d-dd5da9cf42bb"
type: sara_adr
name: "index.md の節は 6 つの型で item の section から生成する"
description: "requirement・design・ubiquitous_term・adr・sara_adr・risk の index.md は、item が持つ必須の enum フィールド section ごとに節を分けて生成し、値は item に手で書く"
status: stable
sources:
  - id: "issue-743"
    resource: "https://github.com/yasunori0418/layat/issues/743"
    title: "#743"
origin: "epic #743 の grilling（2026-11-14）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
supersedes:
  - "SDR-fcae2485-b7f5-4717-9c3a-4e585161efec"
references:
  - "SDR-52dccd7e-666c-46f8-8a8d-1a50dd0f721b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T16:42:24+09:00"
---
# index.md の節は 6 つの型で item の section から生成する

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #743, `docs/model.yaml`, `dev/scripts/okf-index.sh`, `dev/scripts/sara-new.sh`
- 起点: epic #743 の grilling（2026-11-14）で確定

## 背景

[index.md の節は item の section から生成する](index-sections-from-item-section.md) は、`requirement`・
`design`・`ubiquitous_term` の index.md を `section` の節に分けた。`adr`（272 件）・`sara_adr`（64 件）・
`risk`（51 件）の index.md は、まだ item をファイル名順の 1 本で並べており、主題ごとに読めない（#743）。

3 型には、節立ての元になる手書きの文書が無い。`adr` と `sara_adr` の主題は relation からは決まらない。
`justifies` は複数の型の item を指し、指す相手の節は主題と一致しないことがある。`risk` の区分も
`test_condition` の置き場所からは決まらない。1 本の `risk` を確かめる `test_condition` が、複数の区分に
分かれていることがある。

元の SDR の根拠には、`sara-new` が `--section` を省くと enum の先頭の値を入れて起票を終える、とある。
節に分ける型が 6 つに増えると、値を渡し忘れた item が気づかないうちに先頭の節へ入る。

## 決定

`requirement`・`design`・`ubiquitous_term`・`adr`・`sara_adr`・`risk` の 6 型の index.md を、item の
`section` ごとの節に分けて生成する。

- 6 型に必須の enum フィールド `section` を持たせる。取りうる値の集合は `docs/model.yaml` の enum が持つ
- `risk` を除く 5 型の見出しの文言と節の並び順は、`dev/scripts/okf-index.sh` の中の表が持つ。`risk` の
  見出しと並び順の持ち方は別の決定に分ける
- item の `section` の値が表に無ければ、スクリプトはその値を標準エラーに出し、index.md を書かずに
  exit 2 で止まる
- `adr`・`sara_adr`・`risk` の値は item の frontmatter に手で書き、relation からは導かない
- `sara-new` は `section` を持つ型で `--section` の省略を拒否する
- ほかの型の index.md と `docs/index.md` は節に分けない

## 根拠

節を item 自身が持てば、item を足すときにその 1 ファイルで節まで決まり、一覧は生成で揃う。必須の
enum にすると、`section` の無い frontmatter と綴りの誤った値を `sara check` が止める。

`adr`・`sara_adr`・`risk` の値を手で書くのは、背景のとおり relation から主題が一意に決まらないから
である。導く規則を作っても、規則で決まらない item には人が値を決めることになる。

`sara-new` が省略を拒否すれば、起票のたびに節を選ぶことになり、気づかないうちに先頭の節へ入る item が出ない。

見出しの文言と並び順をスクリプトの表に置く根拠は元の SDR と同じで、sara の enum が値の集合しか持てず、
見出しの文言を表せないことにある。

## 棄却した代替案

### `adr` と `sara_adr` の節を relation から導く

`justifies` の指す item の型や節から主題を決めると、複数の型を指す ADR の節が一意に決まらない。
決まらない item を例外として手で書くなら、全件を手で書くほうが規則が 1 つで済む。

### 節に分ける型を元の 3 型のまま据え置く

`adr` は 272 件あり、ファイル名順の 1 本の並びでは関連する決定を探せない。

### `sara-new` で `--section` の省略を許し、enum の先頭の値を入れる

値を選び忘れた item が先頭の節へ入り、index.md を読むまで気付けない。
