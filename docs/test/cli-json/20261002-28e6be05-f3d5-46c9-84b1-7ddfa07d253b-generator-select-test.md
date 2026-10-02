---
id: "CASE-28e6be05-f3d5-46c9-84b1-7ddfa07d253b"
type: test_case
name: "generator_select_test.go — 選択の優先順位・strict な設定ファイル・--manifest 経路の環境非依存"
target: "cmd/layat/generator_select_test.go"
covers:
  - "TC-8377d290-1f47-48f3-844c-bd51f257e25e"
---
# CASE-28e6be05-f3d5-46c9-84b1-7ddfa07d253b: generator_select_test.go — 選択の優先順位・strict な設定ファイル・--manifest 経路の環境非依存

## 対象

`cmd/layat/generator_select_test.go`

## 検証内容

- **優先順位** — 解決関数をテーブル駆動で呼び、5 段の各段が下の段に勝つこと、値の決まった段
  より下の壊れた設定ファイルを読まないこと
- **指定なし** — 空の環境変数と `generator` キーの無い設定ファイルが次の段へ進むこと、
  ディレクトリ未指定でその段を飛ばすこと
- **入力不正** — 未知の値（フラグ・環境変数・各設定ファイル）・未知キー・パース失敗が
  出どころを含む 1 行の `inputError` になること
- **読めない設定ファイル** — `layat.toml` がディレクトリのとき `inputError` にならず `E_IO` に
  分類されること

nix は起動しない。
