---
id: "ADR-d8758f04-ff16-4339-8ea5-5d53b9daf7ee"
type: adr
name: "`layat:` はエラーの文字列に持たせず、stderr へ書く場所で付ける"
description: "エラーと警告の文字列は layat: を持たず、cmd/layat が stderr へ書く 1 箇所で行の先頭に付け、--json の message には付けない"
section: cli-output
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）。[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](engine-errors-wrapped-once-at-origin.md) のエラーの文字列に layat: を持たせる決定を書き直した"
justifies:
  - "DSG-dae4180f-ffe0-4cbd-8bee-0bda331c8043"
  - "DSG-a45f92c5-89de-487f-8c93-60d3b509bad0"
supersedes:
  - "ADR-5228fec0-09ad-494f-8c8d-d887fc009999"
references:
  - "ADR-9e328362-2e90-441a-8eca-40081ce71b1b"
  - "ADR-0ad0321a-97f6-4db9-b46c-c9b01c2c28b9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:00:46+09:00"
---
# `layat:` はエラーの文字列に持たせず、stderr へ書く場所で付ける

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #677, [エラーはエンベロープに構造化し stderr のテキストと併存させる](structured-errors-with-stderr.md), [エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では再 wrap しない](errors-wrapped-once-at-origin.md)
- 起点: Issue #677 の grilling（2026-11-14）。[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](engine-errors-wrapped-once-at-origin.md) のエラーの文字列に layat: を持たせる決定を書き直した

## 背景

エラーの文字列は `layat:` で始めていた。別のエラーに埋め込まれたり `--all` の行にまとめられたり
すると、1 行の中に `layat:` が 2 度現れた。警告の先頭は失敗や通知と同じ `layat:` で、行の先頭では
見分けられなかった。`--json` の `errors[].message` にも `layat:` が入っていた。

## 決定

- エラーと警告の文字列は `layat:` を持たない。`internal/` は `layat:` を書かない
- `cmd/layat/` が stderr へ書く 1 箇所で、行の先頭に付ける。失敗は `layat: <エラー>`、警告は
  `layat: warning: <文>`、通知は `layat: <文>` とする
- `--json` の message には付けない

整形の細目は [stderr へ書く行は cmd の 1 箇所で `layat:` を付け、単発の失敗・`--all` の失敗行・status の行・警告が同じ整形を通る](../design/stderr-lines-prefixed-in-one-place.md) に置く。

## 根拠

`layat:` は、その行を layat が出したことを端末で示すためのもので、エラーの中身ではなく出力の性質である。
付ける場所が 1 つなら、文字列がどう埋め込まれても重ならない。エラーはエンベロープと stderr の
テキストの両方に出る（→ [エラーはエンベロープに構造化し stderr のテキストと併存させる](structured-errors-with-stderr.md)）。
エンベロープの message は機械が読む値なので、出力の性質である `layat:` を持たない方がよい。警告に
`warning:` を足せば、失敗や通知と行の先頭で見分けられる。

## 棄却した代替案

### stderr の行からも `layat:` を外す

nix など子プロセスの出力と同じ stderr に混ざったとき、どの行を layat が出したのかを見分けられなくなる。

### エラーの文字列に持たせたまま、重なりを個別に剥がす

埋め込むたびに剥がす処理が要り、1 箇所でも漏れると重なる。`--json` の message からも別に剥がす
必要が残る。
