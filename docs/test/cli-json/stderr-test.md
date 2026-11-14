---
id: "CASE-9ab9c055-4929-4ca8-aef3-11dd816ddd15"
type: test_case
name: "stderr_test.go — stderr の行の形・errors.Join の分け方・失敗の詳細の出口・layat: を含むリテラルの所在"
description: "cmd/layat/stderr_test.go が、stderr へ書く行の形と、layat: を含む文字列リテラルが出力のファイルの外に無いことを確かめる"
resource: "cmd/layat/stderr_test.go"
status: stable
covers:
  - "TC-cccd086a-5fb7-48d7-96ca-f4126e268396"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:41:09+09:00"
---
# stderr_test.go — stderr の行の形・errors.Join の分け方・失敗の詳細の出口・layat: を含むリテラルの所在

## 対象

`cmd/layat/stderr_test.go`

## 検証内容

- **行の形** — 通知が `layat: <文>`、警告が `layat: warning: <文>`、失敗の行が
  `layat: <見出し>: <エラー>`、`--debug` の開示が `layat: + nix <引数>` になること。
  schemaVersion の食い違いの案内が空行に続けて出ること
- **単発の失敗** — `printError` が、1 つのエラーを `layat: ` で始まる 1 行で書き、要素の中の改行には
  接頭辞を付けないこと。`errors.Join` で束ねたエラーは要素ごとに `layat: ` で始まる行へ分け、
  複数の `%w` を持つ `fmt.Errorf` は 1 つの文として書くこと。generator の失敗は要約の行にだけ
  接頭辞を付け、案内の行には付けないこと
- **失敗の行と errors.Join** — `printFailure` が、最初の要素を見出しの行に続け、残りの要素を
  それぞれ `layat: ` で始まる行にすること
- **失敗の詳細の出口** — 偽の nix と Fake の世代 backend で apply と rollback を実際に駆動し、
  conflict の列挙と undo の報告が `layat: <文>` で出て `layat: warning:` で始まらないこと。
  apply では通常の engine の警告が `layat: warning: ` で始まること
- **リテラルの所在** — `go/parser` で `internal/` と `cmd/layat/` の非テストの `.go` を読み、
  `layat: ` を含む文字列リテラルが `stderr.go` の外に無いこと
