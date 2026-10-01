---
id: "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
type: requirement
name: "内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する"
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The CLI SHALL disclose the commands that its manifest generator runs internally (the
  nix commands for the default nix generator), for instance through `layat --help`, so
  that the user can run them selectively by hand; the disclosure under `--debug` SHALL
  also name the generator. Switching to an
  arbitrary generation and garbage-collecting generations SHALL be done with the standard
  `nix-env` / `nix-collect-garbage` against the profile path, rather than with dedicated
  layat subcommands.
specification_ja: |
  CLI は `layat --help` 等で manifest 生成器が内部実行するコマンド（既定の nix 生成器では
  nix コマンド）を開示し、ユーザーが選択的に手で実行できるようにしなければならない。
  `--debug` での開示には生成器名も含めなければならない。任意世代への切替・世代の GC は layat 専用の
  サブコマンドではなく、標準の `nix-env` / `nix-collect-garbage` を profile パスに対して
  使って行わなければならない。
---
# REQ-4ffda99a-7062-4c00-915f-70b525cb215b: 内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する

## 仕様

- 透明性: `layat --help` 等で manifest 生成器が内部実行するコマンド（nix 生成器では nix
  コマンド）を開示し、ユーザーが選択的に手で実行できる。開示の主体は生成器で、`--debug`
  のときは生成器名と内部コマンドを出す。
- 任意世代への切替・世代の GC は標準の `nix-env` / `nix-collect-garbage` を profile パスに
  対して使う。

`--debug` による nix コマンドの開示（冗長度と直交させる分離）は REQ-0a123b89-0399-4f76-b988-56a5f7e0becf の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の透明性・任意世代切替の箇条書き 2 項。

決定の実体は ADR-0007 §3「透明性」（`layat --help` 等で内部実行する nix コマンドを開示し、
ユーザーが選択的に手で実行できるようにする）。開示の主体を生成器にしたのは
ADR-0055「manifest 生成器の契約」§8。
