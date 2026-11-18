---
id: "TC-7f30ff0f-a7bd-4a4b-ade9-7dc619207803"
type: test_condition
name: "outturn-validate が確かめない MUST を layat のテストで確かめる"
description: "outturn-validate が確かめない MUST（workingDirectory が実行時の作業ディレクトリと同じ絶対パスであること、取得できない実行がエンベロープを出さないこと）を Go テストと E2E で確かめる"
status: stable
mitigates:
  - "RISK-7b73d79d-9ac0-4f2a-88e8-9e03a565d46f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-18T23:30:06+09:00"
---
# outturn-validate が確かめない MUST を layat のテストで確かめる

## 条件

outturn の MUST のうち、適合の確認が確かめないものを、layat のテストで確かめる。

- 値の一致: 一時ディレクトリを指す symlink のパスへ移ってから `--json` の実行を通しで行い、`workingDirectory` が絶対パスで、
  テストが同じ時点で取得した作業ディレクトリと同じ文字列（symlink のパスのまま）であること
- 取得できない実行: 作業ディレクトリを取得できない `--json` の実行が、サブコマンドの処理に入らず、
  stdout に 1 バイトも書かず、エラーの文が `cannot get the working directory: ` で始まり、終了コード 1 で
  終わること
- 観測しない実行: `--json` を付けない実行が作業ディレクトリを観測しないこと
- 実 nix での形: E2E の `--json` の出力の全てで、`workingDirectory` が文字列で `/` で始まること

値の一致は Go テストだけで確かめ、E2E は絶対パスの形だけを確かめる。E2E の作業ディレクトリはシナリオごとに
変わるので、値を比べる基準を持たないためである。
