---
id: "TC-a273033c-3693-44c0-93dd-c6bb646a5bfd"
type: test_condition
name: "fmt.Errorf と errors.New の書式に failed to の形が無い"
description: "internal/ と cmd/layat/ の非テストの Go で、fmt.Errorf と errors.New の第 1 引数の文字列リテラルに failed to の形が無いことを確かめる"
status: stable
mitigates:
  - "RISK-5a82656f-d1fa-4f71-8037-1daf307bfc95"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-16T00:16:41+09:00"
---
# fmt.Errorf と errors.New の書式に failed to の形が無い

## 条件

`internal/` と `cmd/layat/` の非テストの Go を構文木として読み、`fmt.Errorf` と `errors.New` の
呼び出しのうち、第 1 引数が文字列リテラルのものを全部見る。その文字列が次の 3 つの形のどれかに
当たれば、エラーの書式が失敗の種類ごとの書式から外れているとして失敗させる。

- `failed to ` で始まる
- ` failed: ` を含む
- ` failed` で終わる

許可リストは持たない。新しい失敗を加えるときに既存の書き方を写しても、この形は検査で止まる。

見るのはエラーを作る 2 つの呼び出しだけである。`fmt.Sprintf` で組み立てる文字列、`exitError` の
msg、失敗の行の見出しは、書式が同じ形をしていても対象にしない。テストのコードも見ない。

検出が実際に働くことも確かめる。`internal/engine/` に `fmt.Errorf("failed to x: %w", err)` を
1 つ加えた状態で検査が失敗し、元に戻すと成功することを、テストの外で一度確かめておく。
