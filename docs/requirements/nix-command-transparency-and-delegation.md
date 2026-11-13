---
id: "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
type: requirement
name: "内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する"
section: cli-subcommands
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The CLI SHALL disclose, for instance through `layat --help`, the commands that the
  manifest generator runs internally (the nix commands for the default nix generator), so
  that the user can run them selectively by hand. Under `--debug`, the generator itself
  SHALL write each internal command it runs, prefixed with the generator name. Switching to an
  arbitrary generation and garbage-collecting generations SHALL be done with the standard
  commands of the generation backend (`nix-env` / `nix-collect-garbage` for the default
  nixprofile backend) against the profile path, rather than with dedicated layat
  subcommands.
specification_ja: |
  CLI は `layat --help` 等で manifest 生成器が内部実行するコマンド（既定の nix 生成器では
  nix コマンド）を開示し、ユーザーが選択的に手で実行できるようにしなければならない。
  `--debug` のときは、生成器自身が実行する内部コマンドを生成器名付きで書かなければならない。任意世代への切替・世代の GC は layat 専用の
  サブコマンドではなく、世代バックエンドの標準コマンド（既定の nixprofile では `nix-env` /
  `nix-collect-garbage`）を profile パスに対して使って行わなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:04:55+09:00"
---
# 内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する

## 仕様

- 透明性: `layat --help` 等で manifest 生成器が内部実行するコマンド（nix 生成器では nix
  コマンド）を開示し、ユーザーが選択的に手で実行できる。`--help` の文面は CLI が持ち、
  `--debug` の逐次開示は生成器が担う（実行する内部コマンドを生成器名付きで書く）。失敗時の要約の組み立ては
  [生成器は診断を CLI が渡す writer へ素通しし、失敗を構造化して返し、表示は CLI が決める](generator-diagnostics-contract.md) の担当。
- 任意世代への切替・世代の GC は世代バックエンドの標準コマンド（既定の nixprofile では
  `nix-env` / `nix-collect-garbage`）を profile パスに対して使う。engine 自身は `nix-env` を
  叩かない（[engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md)）。

`--debug` を冗長度（`-v`）と直交させる分離そのものは [冗長度は -v、デバッグは --debug に分離し --json と直交させる](verbose-and-debug-separation.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の透明性・任意世代切替の箇条書き 2 項。

決定の実体は [内部コマンドの開示で透明性を保つ](../adr/internal-command-disclosure.md)（`layat --help` 等で内部実行する nix コマンドを開示し、
ユーザーが選択的に手で実行できるようにする）。`--help` の文面を CLI、`--debug` の
逐次開示を生成器に分けたのは [内部コマンドの開示を CLI と生成器で分担する](../adr/internal-command-disclosure-split.md)。世代の切替・GC の委譲先を
世代バックエンドの標準コマンドとしたのは [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)。
