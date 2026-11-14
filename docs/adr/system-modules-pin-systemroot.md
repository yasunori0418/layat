---
id: "ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c"
type: adr
name: "NixOS / nix-darwin モジュールは root を systemRoot に固定する"
section: modules
status: stable
sources:
  - id: "issue-519"
    resource: "https://github.com/yasunori0418/layat/issues/519"
    title: "#519"
  - id: "issue-493"
    resource: "https://github.com/yasunori0418/layat/issues/493"
    title: "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §1 を書き直した"
justifies:
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
references:
  - "ADR-3d450af1-1c79-48ce-8ee9-b989dcb41ad7"
  - "ADR-ad5bd91d-e061-4a1e-9eef-651f4c9ce880"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-70debe94-c7ce-492b-b938-4c62f7a781e7"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98"
  - "ADR-2230e78f-770a-447e-9c5d-6eba1e9a7359"
  - "ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e"
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-a0d0cf15-8234-42be-b83e-9b8271850475"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# NixOS / nix-darwin モジュールは root を systemRoot に固定する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [配置は全層で engine が行い、ネイティブ機構へ翻訳しない](engine-owns-placement-no-native-translation.md), [モジュールは root と起動の時機だけを供給して engine を起動する](modules-wire-engine-kick-only.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md), [out-of-store のリンクは lib の関数が包むだけにして engine が張る](out-of-store-linked-by-engine.md), [古くなった配置の除去は記録どおりの symlink だけに限る](conservative-removal-invariant.md), [engine を単一の固定 Go バイナリにする](engine-fixed-go-binary.md), [`configs.<name>.subdir` で config ごとに target の基点を寄せる](module-config-subdir-option.md), [モジュールは subdir の確認を lib に委ねる](module-delegates-subdir-validation.md), [NixOS モジュールは system.activationScripts.layat から起動する](nixos-activation-script-hook.md), [モジュールの activation は apply --all --manifest を 1 回だけ起動する](module-activation-single-apply-all.md), [activation の起動文字列を modules/kick.nix で組み立てる](modules-kick-builder.md), [モジュールは configs 間の target 衝突を評価時に確認しない](modules-no-eval-conflict-assertion.md), [NixOS の dry-activate では --dryrun を付けて起動する](nixos-dry-activate-dryrun.md), [NixOS の activation の PATH に nix を前置する](nixos-activation-path-nix.md), [system mode の last-applied 記録を /var/lib/layat に置く](system-record-base-var-lib.md), [root の明示を mkManifest 層で必須にする](root-explicit-at-mkmanifest.md), GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §1 を書き直した

## 背景

モジュールは engine を起動する配線であり、root と activation の時機だけを供給する。NixOS / nix-darwin
モジュールの root を何にするかについて、当初は `layat.user` で配置先のユーザーを受け取り、その home を
root にする想定だった。一方、ユーザーの home への配置は home-manager モジュールが既に担っている。
home-manager モジュールは NixOS / nix-darwin の中でも使える。

## 決定

NixOS / nix-darwin モジュールは、どの config の root も `systemRoot` に固定する。config の `subdir` が
null でなければ `systemRoot { subdir = <subdir>; }` とする。配置先のユーザーを受け取る `layat.user`
オプションは設けない。

root の供給元とオプションの規範は [home mode の root は層ごとに定まった供給元から解決する](../requirements/home-mode-root-resolution.md) と
[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) に置く。

## 根拠

host の activation は root 権限で実行される。ここからユーザーの home へ配置すると、ファイルの所有者の扱い・
ユーザーの解決・profile 基底の選び方をモジュールが抱えることになる。home への配置は home-manager
モジュールが担っており、同じ home へ配置する手段を 2 つ持たない。

`subdir` の扱いを home-manager モジュールの `homeRoot` と同じ形にそろえるので、利用者は統合層を
変えても config の書き方を変えずに済む。

## 棄却した代替案

### `layat.user` を受け取り、そのユーザーの home へ home mode で配置する

当初の想定。home-manager モジュールと重なる配置の手段を root 権限で持つことになる。
