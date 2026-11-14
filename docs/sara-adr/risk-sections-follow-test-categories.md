---
id: "SDR-5cc4a493-1c96-4703-a2b1-19e43bf95a30"
type: sara_adr
name: "risk の節はテストの区分と同じ語彙にする"
description: "risk の section の値・節の並び順・見出しは dev/tests/test-categories.tsv のテストの区分と同じにし、enum と区分の一致を契約テストが確かめる"
section: doc-layout
status: stable
sources:
  - id: "issue-743"
    resource: "https://github.com/yasunori0418/layat/issues/743"
    title: "#743"
origin: "epic #743 の grilling（2026-11-14）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
references:
  - "SDR-4e92d0b5-7674-4956-854d-dd5da9cf42bb"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T17:10:55+09:00"
---
# risk の節はテストの区分と同じ語彙にする

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #743, `docs/model.yaml`, `dev/scripts/okf-index.sh`, `dev/tests/test-categories.tsv`, `dev/tests/okf-index.sh`
- 起点: epic #743 の grilling（2026-11-14）で確定

## 背景

[index.md の節は 6 つの型で item の section から生成する](index-sections-in-six-types.md) は、`risk` の
index.md を `section` の節に分けると決めた。ただし、見出しと並び順の持ち方は別の決定に回した。

`risk` の脅威は、テスト系 item を通じてテストの区分へつながる。区分の一覧と並び順は
`dev/tests/test-categories.tsv` が持ち、`docs/test/` のディレクトリと対応表もこれに従う。`risk` に別の
語彙を立てると、同じ脅威をテスト側と `risk` 側の 2 つの語彙で分類することになる。

## 決定

`risk` の `section` にはテストの区分と同じ語彙を使う。

- `risk` の `section` の値はテストの区分（`dev/tests/test-categories.tsv`）と同じにする
- 節の一覧と並び順は、`dev/scripts/okf-index.sh` がこのファイルから読む。見出しには区分の名前を
  そのまま使う。スクリプトの節の表（`SECTIONS`）には `risk` の行を書かない
- `docs/model.yaml` の enum と区分が値も順も一致することは、契約テスト `dev/tests/okf-index.sh` が確かめる
- `risk` の値は、name が述べる不具合を確かめている `test_condition` の区分を手で書く。複数の区分に
  またがる `risk` も分解しない

## 根拠

区分の一覧を持つファイル 1 つから読めば、区分を増減しても `risk` の節は追従する。スクリプトに別の表を置く
必要も無い。sara の enum はファイルを参照できないので、`docs/model.yaml` には値を写して持つ。enum の
値と区分のずれは契約テストが検出する。

区分の名前を見出しに使えば、`docs/test/<区分>/` と index.md の節が同じ名前で対応する。`risk` から
テスト系 item へ辿るとき、名前の読み替えが要らない。

値を手で書く根拠は背景に挙げた SDR と同じである。`risk` を確かめる `test_condition` が複数の区分にあると、
relation から区分は一意に決まらない。

## 棄却した代替案

### スクリプトの節の表に `risk` の行を書く

ほかの 5 型と同じ形になる。しかし区分を増減するたびに、`dev/tests/test-categories.tsv` とスクリプトの
表の 2 か所を揃えなければならない。

### `risk` に主題の語彙を別に立てる

テストの区分と別の分類が増える。`risk` から `test_condition` へ辿るとき、2 つの語彙を突き合わせることになる。

### 複数の区分にまたがる `risk` を区分ごとに分解する

節分けのために、1 つの脅威が複数の item に割れる。`threatens` と `mitigates` も分けて張り直さなければ
ならない。
