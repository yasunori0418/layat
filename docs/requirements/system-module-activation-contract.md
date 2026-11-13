---
id: "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
type: requirement
name: "NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す"
section: modules
status: stable
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
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:04:55+09:00"
---
# NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す

## 仕様

| 層 | hook | dry-activate | PATH |
|---|---|---|---|
| NixOS | `system.activationScripts.layat`（`deps = [ "etc" "users" ]`・`supportsDryActivation = true`）| `NIXOS_ACTION=dry-activate` のときだけ `--dryrun` | `config.nix.package` の `bin` を前置 |
| nix-darwin | `system.activationScripts.postActivation.text` へ追記 | 分岐しない | そのまま |

> 規範は frontmatter が正。HM 側の起動契約は [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md)、ペアの並び・`--jobs`・config が無いときに
> 起動しないことは [HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う](hm-activation-per-config-kick.md)、`--backup` / `--recopy` の配線は [layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない](backup-wiring-layer.md) /
> [layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける](recopy-wiring-layer.md)、モジュールが root を pin することは [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、rollback を host へ
> 一本化することは [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md) の担当。
>
> `supportsDryActivation = true` が無いと、スクリプトは dry-activate 用の activate に含まれず
> `NIXOS_ACTION` の分岐が実行されない。`PATH` の前置は、engine が世代コミットで呼ぶ `nix-env` が
> NixOS の activation の `PATH` に無いため。nix-darwin は固定名のスクリプトだけを実行するので、
> 独自名ではなく `postActivation` へ追記する。

## 出典

[NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md)・[NixOS モジュールは system.activationScripts.layat から起動する](../adr/nixos-activation-script-hook.md)・[nix-darwin モジュールは postActivation に起動を追記する](../adr/darwin-post-activation-hook.md)・[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)・[activation の起動文字列を modules/kick.nix で組み立てる](../adr/modules-kick-builder.md)・[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)・[NixOS の dry-activate では --dryrun を付けて起動する](../adr/nixos-dry-activate-dryrun.md)・[NixOS の activation の PATH に nix を前置する](../adr/nixos-activation-path-nix.md)。
