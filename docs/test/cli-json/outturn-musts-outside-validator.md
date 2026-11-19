---
id: "TC-7f30ff0f-a7bd-4a4b-ade9-7dc619207803"
type: test_condition
name: "outturn-validate が確かめない MUST を layat のテストで確かめる"
description: "outturn-validate が確かめない MUST（workingDirectory が実行時の作業ディレクトリと同じ絶対パスであること、取得できない実行がエンベロープを出さないこと、暗黙の入力の解決結果を info に記録すること）を Go テストと E2E で確かめる"
status: stable
mitigates:
  - "RISK-7b73d79d-9ac0-4f2a-88e8-9e03a565d46f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-19T00:08:45+09:00"
---
# outturn-validate が確かめない MUST を layat のテストで確かめる

## 条件

outturn の MUST のうち、適合の確認が確かめないものを、layat のテストで確かめる。

- 値の一致: 一時ディレクトリを指す symlink のパスへ移ってから `--json` を付けた実コマンドを最後まで実行し、`workingDirectory` が絶対パスで、
  テストが同じ時点で取得した作業ディレクトリと同じ文字列（symlink のパスのまま）であること
- 取得できない実行: 作業ディレクトリを取得できない `--json` の実行が、サブコマンドの処理に入らず、
  stdout に 1 バイトも書かず、エラーの文が `cannot get the working directory: ` で始まり、終了コード 1 で
  終わること（`help` と `completion` を除く）
- 観測しない実行: `--json` を付けない実行が、`workingDirectory` に出すための観測をしないこと
- `help` と `completion`: 作業ディレクトリを取得できない状態で、`help --json` と `completion bash --json` が
  作業ディレクトリを読まず、本文を出して成功すること
- 実 nix での形: E2E の `--json` の出力の全てで、`workingDirectory` が文字列で `/` で始まること
- 暗黙の入力の記録: `-f`・`--generator`・`--root` を省いた実行で、トップの `info` に探索で見つけた
  `entrypoint` と採用した `generator` が、`result.info` に解決した `root` があること。3 つとも与えた実行では、
  どちらの階層にも `info` が無いこと。`--manifest` の実行に `entrypoint` と `generator` が無いこと。
  apply（`--all` と `--dryrun` を含む）・reset・rollback・list-generations の単発・gitignore のそれぞれで確かめ、
  `--all` では `entrypoint` と `generator` がトップに 1 度だけ、`root` が主体ごとにあること。探索に失敗した
  実行のトップの `info` が `generator` だけであること
- 暗黙の入力の実 nix での形: E2E で、`-f` を省いた apply の `entrypoint` と `root` が `/` で始まる文字列で、
  `generator` が文字列であること。`--manifest` の apply の `info` に `entrypoint` と `generator` が無いこと

値の一致は Go テストだけで確かめ、E2E は絶対パスの形だけを確かめる。暗黙の入力の値も同じで、E2E は形だけを確かめる。E2E の作業ディレクトリはシナリオごとに
変わるので、値を比べる基準を持たないためである。
