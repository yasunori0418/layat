---
id: "CASE-0cd1bf07-dc62-4666-8dd4-c61f0be096c0"
type: test_case
name: "input_errors_test.go — 実行前に分かる入力の誤りの E_INPUT と 1 行の文と載る層"
description: "cmd/layat/input_errors_test.go が、実行を始める前に分かる 11 種類の入力の誤りが E_INPUT で 1 行になることと、name を指定した apply とそうでない apply でエラーの載る層を確かめる"
resource: "cmd/layat/input_errors_test.go"
status: stable
covers:
  - "TC-34f4b492-1f96-4f2f-b51f-7f87ccc232de"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T06:35:51+09:00"
---
# input_errors_test.go — 実行前に分かる入力の誤りの E_INPUT と 1 行の文と載る層

## 対象

`cmd/layat/input_errors_test.go`

## 検証内容

- 11 種類の入力の誤りを、apply・gitignore・list-generations・init のコマンドを実際に実行して
  起こす。`--yes` の無い拒否は、reset と prune が共有する `confirmPolicy` で起こす。どの経路でも
  `classifyError` の code が `E_INPUT` で、エラーの文に改行が無いこと
- `apply foo --jobs 2` と、root の flag を付けた `apply foo` のエンベロープで、`results[0]` の subject が `foo`、
  `results[0].errors[0].code` が `E_INPUT` で、トップレベルの `errors[]` が空であること
- `apply --all foo` のエンベロープで、`results[]` が空で、トップレベルの `errors[]` の code が
  `E_INPUT` であること
