---
id: "DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda"
type: design
name: "NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする"
status: stable
satisfies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:01:21+09:00"
---
# DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda: NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする

## 設計

`nixosModules.default`（`modules/nixos.nix`）と `darwinModules.default`（`modules/nix-darwin.nix`）は、
`common.nix` の共通オプションを import し、`modules/kick.nix` が返す起動文字列を activation へ
置くだけの配線である。

```nix
# modules/kick.nix（module ではなく純粋関数）
{ lib, pkgs, layatPackage, cfg, mkRoot }:   # cfg = config.layat / mkRoot = subdir → root マーカー
{ manifests = { <name> = <link-farm>; … };
  command   = "layat apply --all --manifest <name>=<farm> … --jobs <n>[ --recopy][ --backup=<suffix>]"; }
```

| 層 | `mkRoot` | activation |
|---|---|---|
| home-manager | `homeRoot`（subdir 適用形を含む）| `home.activation.layat` |
| NixOS | `systemRoot`（subdir 適用形を含む）| `system.activationScripts.layat`（`deps = [ "etc" "users" ]`・`supportsDryActivation = true`）|
| nix-darwin | `systemRoot`（subdir 適用形を含む）| `system.activationScripts.postActivation.text` へ追記 |

- **起動の組み立てを `kick.nix` 1 か所に置く**。REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 が求める共通オプション集合は
  `common.nix` の import で、その集合から起動を組む規則は `kick.nix` で全モジュールにそろう。
  モジュールごとに残るのは root マーカーと hook だけで、REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9 の「配線に徹する」が
  ファイルの中身として見える
- **NixOS 固有の起動環境はモジュールが持つ**。`PATH` への `config.nix.package` の前置と、
  `NIXOS_ACTION=dry-activate` のときの `--dryrun` は `nixos.nix` が `command` の外側で付ける
  （REQ-a220bb28-89de-42b0-a854-e7468755ecb1）。nix-darwin は dry run で分岐しない
- **`layatPackage` は flake の wrapper が注入する**。HM と同じく `_module.args.layatPackage` で渡し、
  `layat.package` オプションは公開しない

モジュールの評価結果は eval アサートの check、NixOS の実 activate は VM テストで検証する
（DSG-901351ea-f01b-4bb0-a470-b890d92c980b）。

## 出典

ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c「NixOS / nix-darwin モジュールは root を systemRoot に固定する」、
ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e「NixOS モジュールは system.activationScripts.layat から起動する」、
ADR-eee039fc-40af-42ac-bc65-979d6eaa0d06「nix-darwin モジュールは postActivation に起動を追記する」、
ADR-f58e9981-bacc-4937-98b8-bddce0a057db「モジュールの activation は apply --all --manifest を 1 回だけ起動する」、
ADR-a0d0cf15-8234-42be-b83e-9b8271850475「activation の起動文字列を modules/kick.nix で組み立てる」、
ADR-1338a293-3f69-48f5-9164-efb4ec6994d6「モジュールは configs 間の target 衝突を評価時に確認しない」、
ADR-1743c957-a732-4a44-a89a-4c03ab277f0c「NixOS の dry-activate では --dryrun を付けて起動する」、
ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70「NixOS の activation の PATH に nix を前置する」。
本 item は分割時点の原文の「flake.nix outputs 設計」にあったスタブ公開の項から起こし、
同 ADR で実装の形へ書き換えた。
