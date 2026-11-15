---
id: "ADR-50e3689b-c0ee-4109-a6e6-5b8bbeacabf9"
type: adr
name: "golangci-lint に wrapcheck を加え、包まずに返すヘルパーは理由つきの nolint で外す"
description: "golangci-lint の既定のリンタに errorlint・forbidigo・nilerr・errname・nolintlint・wrapcheck を足してエラー処理の規約からの逸脱を検出し、包まずに返すヘルパーは関数の直前に理由つきの nolint を書いて外す"
section: dev-build
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "[golangci-lint でエラー処理の規約を検出する](golangci-lint-enforces-error-conventions.md) を書き直した（Issue #677 の仕上げ、2026-11-15）"
justifies:
  - "QA-a92341b9-c873-406e-8b70-a64f56d8a7d6"
supersedes:
  - "ADR-949badd9-027a-48f6-bdd1-75c30ab8e213"
references:
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
  - "ADR-6f0383d9-0176-47c4-a9da-1a41e1fd3747"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-16T00:00:36+09:00"
---
# golangci-lint に wrapcheck を加え、包まずに返すヘルパーは理由つきの nolint で外す

- ステータス: stable
- 日付: 2026-11-15
- 関連: GitHub Issue #677, [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), [Go の静的解析を CGO_ENABLED=0 で行う](go-checks-without-cgo.md)
- 起点: [golangci-lint でエラー処理の規約を検出する](golangci-lint-enforces-error-conventions.md) を書き直した（Issue #677 の仕上げ、2026-11-15）

## 背景

エラーは、op と対象が分かるいちばん下の関数で 1 回だけ包む
（[エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする](../design/error-wrap-once-at-origin.md)）。
golangci-lint は既定の 5 つのリンタ（errcheck・govet・ineffassign・staticcheck・unused）に、
errorlint・forbidigo・nilerr・errname・nolintlint を加えて実行していた。どのリンタも外部のパッケージの
エラーを包まずに返す箇所を指摘しない。`os` や `encoding/json` のエラーを包まずに返すと、利用者には
何をしようとして失敗したのかが分からない文が届く。この形の箇所が新しく入っても、レビューで気づくほかに
止める手段が無かった。

## 決定

リポジトリ直下の `.golangci.yml` で、既定の 5 つに次のリンタを加える。

- errorlint: sentinel の `==`・`!=` による比較と、型アサーションによるエラーの判別を検出する
- forbidigo: `os.IsNotExist`・`os.IsExist`・`os.IsPermission` を禁じ、`errors.Is` と
  `fs.ErrNotExist`・`fs.ErrExist`・`fs.ErrPermission` の組で書かせる。テストのコードにも当てる
- nilerr: エラーを受け取ったのに nil を返す箇所を検出する
- errname: `Err` で始まらない sentinel の変数名と、`Error` で終わらないエラーの型名を検出する
- nolintlint: `//nolint` に理由とリンタ名を必須にし、どの指摘も外していない `//nolint` を禁じる
- wrapcheck: 外部のパッケージから受け取ったエラーを包まずに返す箇所を検出する

wrapcheck は layat 自身のモジュール（`github.com/yasunori0418/layat/*`）から受け取ったエラーを対象から
外す。テストのコードと、テスト用の偽の generation の実装（`internal/generation/fake.go`）にも当てない。
呼ぶ側が包む前提で外部のエラーをそのまま返すヘルパーには、関数の直前に
`//nolint:wrapcheck // <どの関数が包むか>` を書いて外す。

意図して規約から外れる箇所は、どのリンタでも `//nolint:<リンタ名> // <理由>` を書いて外す。flake の
check は `--config` でこの設定ファイルを渡して実行する。

## 根拠

wrapcheck を有効にした時点の指摘は 28 件だった。直す必要があったのは、`--json` の envelope の encode と
書き込みの失敗の 2 件と、item の id を導けない失敗の 1 件の計 3 件である。残りの 25 件は、呼ぶ側が
対象を添えて包む下のヘルパーだった。偽陽性が多いリンタを入れると `//nolint` が増える。それでも入れるのは、
素の `os` や `json` のエラーが新しく入るのを止める手段がほかに無いからである。ヘルパーの関数の直前に
書く `//nolint` は関数の全体の指摘を外すので、外す箇所は関数の単位で 6 つと 1 行の関数 3 つに収まる。
理由に包む関数を書くので、包む側が変わったときに `//nolint` も見直す目印になる。

nolintlint で理由とリンタ名を必須にすれば、外した箇所ごとに理由がコードに残り、リンタ名の無い
`//nolint` が別のリンタの指摘まで外すことも無くなる。設定ファイルを check に `--config` で渡すのは、
check の sandbox へ持ち込む Go のソースに設定ファイルが入らないからである。渡さないと既定のリンタだけで
実行され、違反があっても check が成功する。ソースの側に含めると、設定を変えるたびに layat のビルドも
やり直しになる。

## 棄却した代替案

### err113 を入れる

`fmt.Errorf` で作る動的なエラーを全部指摘する。発生源で文脈を加えて wrap する規約と両立しない。

### errcheck の `check-blank` を有効にする

`_ =` で捨てたエラーを指摘するが、導入した時点の指摘の大半は lock の解放のように捨ててよい箇所で、
理由付きの `//nolint` が並ぶだけになる。

### revive のエラーのルールや gocritic の ruleguard を入れる

revive のエラーの文字列と命名のルールは staticcheck と重なる。ruleguard は `go.mod` へ依存を加える必要が
あるうえ、エラーの文字列の形を見るルールに取りこぼしが出た。

### wrapcheck を入れない

`//nolint` は増えないが、外部のパッケージのエラーを包まずに返す箇所が新しく入っても検出できない。
レビューで気づかなければ、何に失敗したのかが分からない文が利用者まで届く。
