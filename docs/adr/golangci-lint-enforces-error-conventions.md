---
id: "ADR-949badd9-027a-48f6-bdd1-75c30ab8e213"
type: adr
name: "golangci-lint でエラー処理の規約を検出する"
description: "golangci-lint の既定のリンタに errorlint・forbidigo・nilerr・errname・nolintlint を足してエラー処理の規約からの逸脱を検出し、nolint には理由とリンタ名を必須にする"
section: dev-build
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）"
justifies:
  - "QA-a92341b9-c873-406e-8b70-a64f56d8a7d6"
references:
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
  - "ADR-6f0383d9-0176-47c4-a9da-1a41e1fd3747"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:19:56+09:00"
---
# golangci-lint でエラー処理の規約を検出する

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #677, [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), [Go の静的解析を CGO_ENABLED=0 で行う](go-checks-without-cgo.md)
- 起点: Issue #677 の grilling（2026-11-14）

## 背景

golangci-lint の設定ファイルは無く、golangci-lint は既定の 5 つのリンタ（errcheck・govet・ineffassign・
staticcheck・unused）だけで走っていた。エラーは `errors.Is`・`errors.As` で判別する
（[エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする](../design/error-wrap-once-at-origin.md)）が、
`os.IsNotExist` による判別や sentinel の `!=` による比較が残っていても、既定のリンタでは検出されない。
`os.IsNotExist` は wrap されたエラーを辿らないので、wrap を足すと判別が気づかないうちに外れる。

## 決定

リポジトリ直下に `.golangci.yml` を置き、既定の 5 つに次のリンタを足す。

- errorlint: sentinel の `==`・`!=` による比較と、型アサーションによるエラーの判別を検出する
- forbidigo: `os.IsNotExist`・`os.IsExist`・`os.IsPermission` を禁じ、`errors.Is` と
  `fs.ErrNotExist`・`fs.ErrExist`・`fs.ErrPermission` の組で書かせる。テストのコードにも当てる
- nilerr: エラーを受け取ったのに nil を返す箇所を検出する
- errname: sentinel の変数名が `Err` で始まり、エラーの型名が `Error` で終わることを検査する
- nolintlint: `//nolint` に理由とリンタ名を必須にし、効いていない `//nolint` を禁じる

意図して規約から外れる箇所は `//nolint:<リンタ名> // <理由>` を書いて外す。flake の check は
`--config` でこの設定ファイルを渡して実行する。

## 根拠

選んだリンタが導入した時点で出した指摘は、機械的に直せるものと、意図した 1 箇所だけだった。偽陽性が
多いリンタを入れると `//nolint` が増え、規約から外れた箇所と偽陽性の区別が付かなくなる。nolintlint で
理由とリンタ名を必須にすれば、外した箇所ごとに理由がコードに残り、リンタ名の無い `//nolint` が別の
リンタの指摘まで外すことも無くなる。設定ファイルを check に `--config` で渡すのは、check の sandbox へ
持ち込む Go のソースに設定ファイルが入らないからである。渡さないと既定のリンタだけで実行され、
違反があっても check が成功する。ソースの側に足すと、設定を変えるたびに layat のビルドもやり直しになる。

## 棄却した代替案

### err113 を入れる

`fmt.Errorf` で作る動的なエラーを全部指摘する。発生源で文脈を足して wrap する規約と両立しない。

### errcheck の `check-blank` を有効にする

`_ =` で捨てたエラーを指摘するが、導入した時点の指摘の大半は lock の解放のように捨ててよい箇所で、
理由付きの `//nolint` が並ぶだけになる。

### revive のエラーのルールや gocritic の ruleguard を入れる

revive のエラーの文字列と命名のルールは staticcheck と重なる。ruleguard は `go.mod` へ依存を足す必要が
あるうえ、エラーの文字列の形を見るルールに取りこぼしが出た。
