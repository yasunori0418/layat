---
id: "TC-b4262641-776c-427d-8331-1a94b29fa568"
type: test_condition
name: "--help の開示ブロックがサブコマンドごとに実際の nix 呼び出しと一致する"
mitigates:
  - "RISK-81cadaa3-3ed2-4cb0-bc74-72ff3afa9cce"
---
# TC-b4262641-776c-427d-8331-1a94b29fa568: --help の開示ブロックがサブコマンドごとに実際の nix 呼び出しと一致する

## テスト条件

`layat --help` の開示ブロックが、各サブコマンドを実際に実行したときの nix 呼び出しと一致する
ことを検証する。

- ブロックに載るラベル（`init <template>` / `apply <name>` / `apply --all` / `gitignore <name>` /
  `reset / rollback / list-generations`）の集合が、テストが実行するサブコマンドの集合と一致する。
  ブロックからの行の削除と、テストが実行しないラベルの追加のどちらでも検出する
- 各ラベルの行に並ぶ nix コマンドが、そのサブコマンドの実際の呼び出し列（サブコマンド、
  installable、フラグ。`--out-link` / `--no-link --print-out-paths` / `--apply` / `--json` /
  `--raw` を含む）と順序まで一致する。実際の値は placeholder（`<ep>` / `<system>` / `<name>` /
  `<profileDir>` / `<ref>` / `<template>` / `'<rootKind map>'`）へ正規化して比べる
- nix を実行しないサブコマンド（prune / status）は、profile と記録がある状態で実行しても
  nix を一度も呼ばずに成功し、`--help` にその旨の 1 行がある

優先度は mitigates 先の RISK-81cadaa3-3ed2-4cb0-bc74-72ff3afa9cce の level（low）に従う。
