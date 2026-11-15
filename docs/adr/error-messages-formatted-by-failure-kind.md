---
id: "ADR-162abde5-7224-4c85-b56d-48751e0d3dc5"
type: adr
name: "エラーの文字列は `cannot <op>` を保ち、原因が対象を持たないときだけ対象を括弧で添える"
description: "エラーの文字列は失敗の種類ごとの書式で書き、I/O と parse の失敗は cannot で始めて、対象の括弧は原因のエラーが対象を持たないときと原因に無い情報を足すときだけ書く"
section: dev-build
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）。[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](engine-errors-wrapped-once-at-origin.md) の文字列の書式の決定を書き直した"
justifies:
  - "DSG-a45f92c5-89de-487f-8c93-60d3b509bad0"
supersedes:
  - "ADR-5228fec0-09ad-494f-8c8d-d887fc009999"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:00:46+09:00"
---
# エラーの文字列は `cannot <op>` を保ち、原因が対象を持たないときだけ対象を括弧で添える

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #677
- 起点: Issue #677 の grilling（2026-11-14）。[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](engine-errors-wrapped-once-at-origin.md) の文字列の書式の決定を書き直した

## 背景

エラーの文字列の書式は `cannot <op> (<target>): %w` の 1 つだけが決まっていた。原因が `*fs.PathError`
のように同じパスを出す失敗では、括弧の対象と原因のパスが 2 度並ぶ。子プロセスの失敗・検証の失敗・
前提が成り立たないとき・入力の誤り・sentinel はこの書式に当てはまらず、箇所ごとに `failed to <op>` や
`<op> failed` などの形が混在していた。

## 決定

エラーの文字列は失敗の種類ごとに次の書式で書く。

- I/O と parse の失敗は `cannot <op>[ (<target>)]: %w` とする。括弧の対象は、原因のエラーが対象を
  持たないときと、原因に無い情報を足すときだけ書く
- 子プロセスの失敗は `cannot run <コマンド> (<target>): %w`、検証の失敗は
  `invalid <検証したもの> (<target>): <詳細>` とする
- 前提が成り立たないときは成り立たない事柄を書いて対象を括弧で添え、入力の誤りは 1 行で書く。
  sentinel は対象も案内文も持たない短い句にする
- `failed to` の形は使わない

書式の細目と、どの呼び出しが原因にパスを出すかの一覧は [エラーの文字列は失敗の種類ごとの書式で書き、括弧の対象は原因が対象を持たないときだけ添える](../design/error-message-format-by-failure-kind.md) に置く。

## 根拠

`cannot` で始める形は、既存のエラーの文字列の多くとそれを見るテストがすでに使っており、失敗である
ことが文の先頭で分かる。括弧の対象を原因に合わせて省けば、同じパスが 2 度並ぶことは無くなる。
原因がパスを出すかは呼び出しごとに決まっていて lint では見られないので、一覧を設計に置き、書き手が
引けるようにする。種類ごとに書式を分けるのは、元のエラーを持たない失敗や対象を持たない sentinel に
`cannot <op>` を当てると文が成り立たないからである。

## 棄却した代替案

### Go の慣習の `<op> <target>: %w` に寄せる

`*fs.PathError` の文字列と形が揃うが、既存の `cannot` で始まる文字列と、それを見るテストを全部
書き換えることになる。得られるのは見た目の統一だけで、読み手が受け取る情報は増えない。

### 括弧の対象を常に書く

原因がパスを出す失敗で、同じパスが 2 度並ぶ。
