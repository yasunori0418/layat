---
id: "REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3"
type: requirement
name: "layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The `layat.backup` submodule, common to every module, SHALL wire `--backup=<suffix>` onto
  the `layat apply --all --manifest <name>=<link-farm> …` invocation performed at
  activation only where
  `layat.backup.enable` is true, setting it aside being an explicit opt-in by the user; the
  suffix SHALL be the value of `layat.backup.suffix`, whose default REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 states.
  `layat.backup` SHALL be an option of the invocation-wiring layer, independent of `entries`
  (the manifest v1 contract), and SHALL NOT affect the manifest itself.
specification_ja: |
  全モジュール共通の `layat.backup` submodule は、`layat.backup.enable` が true のときに限り、
  activation で行う `layat apply --all --manifest <name>=<link-farm> …` 起動へ
  `--backup=<suffix>` を配線しなければ
  ならない（退避はユーザーの明示 opt-in であるため）。suffix は `layat.backup.suffix` の値で
  なければならず、そのデフォルトは REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 が定める。`layat.backup` は `entries`
  （manifest v1 契約）とは独立な起動配線レイヤーのオプションであり、manifest 自体に影響して
  はならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3: layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない

## 仕様

`layat.backup`（submodule・全モジュール共通）は activation の `layat apply --manifest` 起動へ
`--backup=<suffix>` を配線する。`enable = true` かつ `suffix` 省略で既定 `layat-backup` が
使われる。`entries`（manifest v1 契約・`lib/types.nix`）とは独立な起動配線レイヤーの
オプションで、manifest 自体には影響しない。

> **上は原文の写しで、規範は frontmatter が正**。オプションの型とデフォルト値そのものは
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、`--backup[=<suffix>]` が何を退避しどう振る舞うかは [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md)、
> module activation が `apply --all --manifest <name>=<link-farm> …` で engine を kick すること自体は [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](apply-manifest-direct.md) /
> [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md) の担当。本 item は `layat.backup` が manifest ではなく起動側に働くという
> レイヤーの所属と、その**配線の発動条件**（`enable` が true のときに限る）を規定する。
> 発動条件は [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md) が定めるオプションの型・デフォルト値とは別の規範で、退避が
> ユーザーの明示 opt-in であるという [モジュールに layat.backup を起動配線のオプションとして持つ](../adr/module-backup-option.md) の設計意図（既定 false・記録外実体を勝手に
> 動かさない）を担保する。
>
> **2026-10-15 追記（[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)）**: 上の写しの `layat apply --manifest` 起動は、全 config を渡す
> `layat apply --all --manifest <name>=<link-farm> …` の 1 回起動と読み替える。`--backup` は
> その起動に載る全 config へ一律に付く。

## 出典

`docs/spec.md`「モジュールオプション仕様」→「共通オプション（全モジュール）」節の
`layat.backup` の段落。

決定の実体は [モジュールに layat.backup を起動配線のオプションとして持つ](../adr/module-backup-option.md)「モジュールに layat.backup を起動配線のオプションとして持つ」で、
module の使い方への配線は同 ADR が Issue #169 とともに定めている。配線先を全 config の
`apply --all --manifest <name>=<link-farm> …` 起動にしたのは [モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)。
