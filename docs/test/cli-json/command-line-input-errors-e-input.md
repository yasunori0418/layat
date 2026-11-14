---
id: "TC-34f4b492-1f96-4f2f-b51f-7f87ccc232de"
type: test_condition
name: "実行前に誤りと分かる入力が E_INPUT で 1 行になり、name を指定した実行では subject の層に載る"
description: "実行を始める前に誤りと分かる入力が --json で E_INPUT に分類されて 1 行の文になり、name を指定した apply では results[0].errors[] に、subject が決まる前の誤りではトップレベルの errors[] に載ることを確かめる"
status: stable
mitigates:
  - "RISK-bd1e4ada-4a0d-4672-96cf-028f488c0540"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T06:35:51+09:00"
---
# 実行前に誤りと分かる入力が E_INPUT で 1 行になり、name を指定した実行では subject の層に載る

## 条件

実行を始める前に、コマンドライン・環境変数・設定ファイル・stdin が TTY かどうかだけで誤りと分かる
入力が、`--json` で `E_INPUT` に分類されることを確かめる。対象は次の 11 種類（gitignore と
list-generations の項目は 2 種類ずつに数える）で、どれも人間向けの文が改行を含まない 1 行で
あることもあわせて見る。

- apply の `<name>` と `--all` の併用
- apply の `--jobs` を `--all` なしで指定
- root の flag（`--project-root` / `--home-root` / `--system-root`）の複数指定
- root の flag を `--all` なしで指定
- `--jobs` の負の値
- gitignore と list-generations の `<name>` と `--all` の併用
- gitignore と list-generations で `<name>` も `--all` も無い
- init の未知のテンプレート名
- 非対話の実行で `--yes` の無い破壊的な操作

分類が `E_LAYAT_FAILED` に戻ると、消費者は入力を直すべき失敗と再試行してよい失敗を区別できない。

エラーの載る層も条件に含む。`apply <name>` は subject が決まってから入力を確かめるので、
`apply foo --jobs 2` のエラーは `results[0].errors[]` に載り、トップレベルの `errors[]` は空になる。
`apply --all foo` は subject が決まる前に拒否されるので、トップレベルの `errors[]` に載る。
