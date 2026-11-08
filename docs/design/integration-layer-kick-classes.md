---
id: "DSG-98d7fa5d-c590-441a-b682-0ab7afa14233"
type: design
name: "engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる"
status: stable
satisfies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-dec58330-6dad-47f7-8f56-2402764a89c7"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "REQ-a0bdf6db-6c0c-476c-916a-61ee4e4510d9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# DSG-98d7fa5d-c590-441a-b682-0ab7afa14233: engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる

## 設計

統合層は 5 つ（CLI 単体 / home-manager / devShell / NixOS / nix-darwin）あるが、
**engine の起動方法は 2 クラスに束ねられる**。配置から世代コミットまでは両クラスで
同一の engine の処理を通り、異なるのは link-farm をどう手に入れるかだけになる。

| クラス | 該当層 | link-farm の取得 |
|---|---|---|
| **entrypoint 駆動** | CLI 単体 / devShell | ユーザーが `layat.<name>` を entrypoint に公開し、CLI が発見 → `nix eval`（rootKind 先取り）→ `nix build` |
| **ビルド済み manifest** | home-manager / NixOS / nix-darwin | モジュール評価時に `mkManifest` でビルドした link-farm を、activation から `layat apply --manifest <link-farm>` へ渡す |

層ごとのクラス割り当てと起動方法は次の通り。

| 層 | クラス | 起動方法 |
|---|---|---|
| CLI 単体 | entrypoint 駆動 | `layat apply <name>` を明示実行 |
| devShell | entrypoint 駆動 | `shellHook` から `layat apply <name>` |
| home-manager | ビルド済み manifest | `home.activation` から `apply --manifest` |
| NixOS | ビルド済み manifest | `system.activationScripts.layat` から `apply --manifest` |
| nix-darwin | ビルド済み manifest | `system.activationScripts.postActivation` への追記から `apply --manifest` |

層ごとの root の供給元は [home mode の root は層ごとに定まった供給元から解決する](../requirements/home-mode-root-resolution.md) / [project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない](../requirements/project-mode-root-resolution.md) / [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md)、ユーザー向け
rollback を公開する層は [rollback と list-generations は home / system mode 限定にする](../requirements/rollback-list-generations-home-only.md) / [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](../requirements/module-mode-profile-internal.md) が定める。どちらも起動クラスとは
独立に決まるため、本 item の担当ではない。

**2 クラスに束ねる**ことが実現手段になっているのは次の点。

- **[全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) が求める「各層は engine をキックするだけの配線」**を、層の数（5）ではなく
  クラスの数（2）で担保できる。層を増やしても新しい engine の処理が増えず、
  既存 2 クラスのどちらかへ割り当てるだけで済む
- **クラスの分かれ目が「entrypoint を持つか」に一致する**。モジュールの `layat.configs` は
  モジュール config 内にあり flake output に現れないため、entrypoint 発見では到達できない。
  そこで [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](../requirements/apply-manifest-direct.md) の `apply --manifest` と [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](../requirements/hm-activation-contract.md) の activation kick が
  必要になる。逆に entrypoint 駆動側（[devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する](../requirements/devshell-wiring.md) の devShell 含む）は
  activation 内で `nix build` / `eval` を行わない制約が無いので entrypoint 発見の処理を使える
- **層ごとに違う要素がクラスの分岐に持ち込まれない**。root の供給元は engine から見れば
  `manifest.json` の root の種類に還元され、rollback の公開有無は profile の使い方の差に
  留まる。どちらも link-farm の取得方法とは無関係なので、2 クラスの外側で決まる

## 出典

分割時点の原文（縮退前）の `docs/design.md`「モジュール統合設計」→「各統合層の動作」の
統合層テーブルと engine kick 2 クラスの説明、および「実行タイミング」節。

なお同原文の統合層テーブルは home-manager 行に「MVP は profile `<name>` =
`default` 固定の 1 profile・役割分離は不可」と書いているが、本 item はこれを採らない。
[HM モジュールは layat.configs.<name> ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md) が `layat.configs.<name>` を導入し、[HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](../requirements/hm-named-configs-profiles.md) が HM 経由でも役割分離した
独立 profile を取れることを規範化しているため。
