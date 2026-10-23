---
id: "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
type: requirement
name: "NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す"
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Where `layat.enable` is true and `layat.configs` holds at least one config, the NixOS
  module SHALL define `system.activationScripts.layat` with `deps = [ "etc" "users" ]` and
  `supportsDryActivation = true`, and the nix-darwin module SHALL append to
  `system.activationScripts.postActivation.text`. Each SHALL kick the engine with a single
  `layat apply --all --manifest <name>=<link-farm> …` invocation, each link-farm being built
  by `mkManifest` at module evaluation time with the root `systemRoot`, or
  `systemRoot { subdir = <subdir>; }` for a config whose `subdir` is not null. The order of
  the pairs and the `--jobs`, `--recopy` and `--backup` flags SHALL follow the same rules as
  the home-manager module, and where no config is present `layat` SHALL NOT be invoked.
  The NixOS activation SHALL prepend the `bin` of `config.nix.package` to `PATH`, and SHALL
  add `--dryrun` when, and only when, `NIXOS_ACTION` is `dry-activate`. The nix-darwin
  activation SHALL NOT branch on a dry run. The activation SHALL NOT perform `nix eval` or
  `nix build`, and an engine error SHALL make the activation exit non-zero.
specification_ja: |
  `layat.enable` が true で `layat.configs` に config が 1 つ以上あるとき、NixOS モジュールは
  `system.activationScripts.layat` を `deps = [ "etc" "users" ]`・`supportsDryActivation = true`
  で定義しなければならず、nix-darwin モジュールは `system.activationScripts.postActivation.text`
  へ追記しなければならない。いずれも `layat apply --all --manifest <name>=<link-farm> …` の
  1 回の起動で engine を kick しなければならない。各 link-farm はモジュール評価時に
  `mkManifest` でビルドし、root は `systemRoot`（`subdir` が非 null の config は
  `systemRoot { subdir = <subdir>; }`）とする。ペアの並びと `--jobs` / `--recopy` / `--backup`
  は home-manager モジュールと同じ規則に従わなければならず、config が無いときは `layat` を
  起動してはならない。NixOS の activation は `PATH` に `config.nix.package` の `bin` を前置し、
  `NIXOS_ACTION` が `dry-activate` のときに限り `--dryrun` を付けなければならない。nix-darwin の
  activation は dry run で分岐してはならない。activation は `nix eval` / `nix build` を行っては
  ならず、engine error は activation を非 0 終了させなければならない。
---
# REQ-a220bb28-89de-42b0-a854-e7468755ecb1: NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す

## 仕様

| 層 | hook | dry-activate | PATH |
|---|---|---|---|
| NixOS | `system.activationScripts.layat`（`deps = [ "etc" "users" ]`・`supportsDryActivation = true`）| `NIXOS_ACTION=dry-activate` のときだけ `--dryrun` | `config.nix.package` の `bin` を前置 |
| nix-darwin | `system.activationScripts.postActivation.text` へ追記 | 分岐しない | そのまま |

> 規範は frontmatter が正。HM 側の起動契約は REQ-8085f194-c903-4ecb-abd8-c719fe7b3292、ペアの並び・`--jobs`・config が無いときに
> 起動しないことは REQ-c847d1af-a437-46bb-bd64-42083810d034、`--backup` / `--recopy` の配線は REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3 /
> REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c、モジュールが root を pin することは REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10、rollback を host へ
> 一本化することは REQ-844ee375-919f-4341-81e1-a5f89fd32840 の担当。
>
> `supportsDryActivation = true` が無いと、スクリプトは dry-activate 用の activate に含まれず
> `NIXOS_ACTION` の分岐が走らない。`PATH` の前置は、engine が世代コミットで呼ぶ `nix-env` が
> NixOS の activation の `PATH` に無いため。nix-darwin は固定名のスクリプトだけを実行するので、
> 独自名ではなく `postActivation` へ追記する。

## 出典

ADR-0066「NixOS / nix-darwin モジュールを system mode の配線として実装する」§1〜§4。
