---
id: "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
type: requirement
name: "全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Basic use SHALL be conceived around project mode and the standalone CLI, support for
  modules existing only to pick up the use case of wanting placement to run together with
  the switch of another module system. Every module (home-manager / NixOS / nix-darwin)
  and the devShell SHALL alike be wiring that does no more than kick the layat engine, each
  layer supplying only the root and the timing of activation. They SHALL NOT translate
  entries into a native mechanism of the platform, such as `systemd.tmpfiles` or
  `home.file`. Placement and stale removal SHALL be performed, in every layer, by the same
  engine and the same store manifest, so that the behaviour is not duplicated per layer.
  The mechanisms of the operating system SHALL remain outside the concern of layat. The
  NixOS module SHALL kick the engine from `system.activationScripts.layat`, and the
  nix-darwin module SHALL kick it from what it appends to
  `system.activationScripts.postActivation`; the hook the home-manager module uses is
  stated by [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md).
specification_ja: |
  基本的な利用は project mode（プロジェクト内配置）と CLI 単体を中心に考えなければ
  ならず、モジュール対応は、他のモジュールシステムの switch と一括で動いてほしいユースケースを
  拾うためだけに存在する。全モジュール（home-manager / NixOS / nix-darwin）と devShell は
  一律、layat エンジンをキックするだけの配線でなければならず、各層は root と activation
  タイミングだけを供給する。`systemd.tmpfiles` / `home.file` などプラットフォームの
  ネイティブ機構へ翻訳してはならない。配置・古くなった配置の除去は全層で同一の engine + store
  マニフェストが行わなければならず、振る舞いが層ごとに二重化しないようにする。OS の機構は
  layat の関心外でなければならない。NixOS モジュールは `system.activationScripts.layat` から、
  nix-darwin モジュールは `system.activationScripts.postActivation` への追記から engine を
  起動しなければならない（home-manager モジュールが使うフックは
  [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md) の担当）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9: 全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない

## 仕様

基本的な利用は **project mode（プロジェクト内配置）と standalone CLI** を中心に考える。
モジュール対応は、他のモジュールシステムの switch と**一括で動いてほしいユースケース**を拾う
ためだけに存在する。全モジュール（HM / NixOS / nix-darwin）と devShell は
**一律「layat エンジンをキックするだけ」の配線**であり、各層は root と activation タイミングを
供給するだけ。`systemd.tmpfiles` / `home.file` へは翻訳しない。配置・stale 除去は全層で同一の
engine + store マニフェスト。

- NixOS モジュールは `system.activationScripts.layat` から engine を起動する
  **配線**に徹する。`systemd.tmpfiles` へは翻訳しない。OS の機構（tmpfiles 等）は layat の
  関心外
- nix-darwin モジュールは `system.activationScripts.postActivation` への追記から engine を起動する

> **上は原文の写しを [NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md) に合わせて改めたもので、規範は frontmatter が正**。各層が供給する
> root の具体は [home mode の root は層ごとに定まった供給元から解決する](home-mode-root-resolution.md)（HM の `$HOME`）と [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)（NixOS / nix-darwin の `systemRoot`）、HM の
> activation フックと起動契約は [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md)、NixOS / nix-darwin の起動契約は [NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す](system-module-activation-contract.md)、devShell の起動タイミングは [devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する](devshell-wiring.md)、
> module の使い方での rollback を host 世代へ一本化することは [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md)、engine 自身が
> ネイティブ FS 操作で配置することは [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md) の担当。
>
> CLI が PATH 常駐の一次 UX であることと、その導入形（CLI 単体のグローバル導入・
> project mode の devShell 同梱）は [layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](cli-primary-ux-installation.md) の担当。本 item はモジュール対応の側から見た
> 位置づけ（一括 switch 用途を拾うためだけに存在する配線であること）を規定する。

## 出典

`docs/spec.md`「モジュール別動作仕様」の前文と、「NixOS モジュール（将来拡張）」
「nix-darwin モジュール（将来拡張）」節。将来拡張 2 節のうち root の供給元は [home mode の root は層ごとに定まった供給元から解決する](home-mode-root-resolution.md)、
rollback の一本化は [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md) が担当し、残る「activation フックから engine を起動する
配線に徹する」を本 item に畳んだ。NixOS / nix-darwin の root は、その後 [NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md) で
[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md) の `systemRoot` pin へ移った。

決定の実体は [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md)・[配置は全層で engine が行い、ネイティブ機構へ翻訳しない](../adr/engine-owns-placement-no-native-translation.md) で、
モジュール対応を一括 switch 用途に限る positioning は [project mode を第一の使い方とする](../adr/project-first.md)が定めている。NixOS / nix-darwin の hook を
具体化したのは [NixOS モジュールは system.activationScripts.layat から起動する](../adr/nixos-activation-script-hook.md)・[nix-darwin モジュールは postActivation に起動を追記する](../adr/darwin-post-activation-hook.md)。
