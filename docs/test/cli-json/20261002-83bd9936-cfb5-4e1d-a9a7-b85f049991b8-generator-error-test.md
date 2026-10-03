---
id: "CASE-83bd9936-cfb5-4e1d-a9a7-b85f049991b8"
type: test_case
name: "generator_error_test.go — 生成器の失敗の人間向け / --json / --debug の 3 形"
target: "cmd/layat/generator_error_test.go"
covers:
  - "TC-4a39ca86-e866-41ec-93d2-663f8e6c1ce0"
---
# CASE-83bd9936-cfb5-4e1d-a9a7-b85f049991b8: generator_error_test.go — 生成器の失敗の人間向け / --json / --debug の 3 形

## 対象

`cmd/layat/generator_error_test.go`（`printGeneratorError` と `generatorErrorMessage`）

## 検証内容

- **人間向け** — 要約 1 行の後に `Guidance` が続き、生の診断（`Stderr`）を含まないこと。
  `Guidance` が空なら要約だけになること
- **`--debug`** — 要約行の末尾に `(generator: nix)` が付き、その後に `Guidance` が続くこと
- **`--json`** — `errors[].message` が要約・改行・`Stderr` の連結で、`--debug` の有無で
  変わらないこと。`Stderr` が空なら要約だけになること

`generator.Error` を直接組み立てて与え、nix のプロセスは起動しない。
