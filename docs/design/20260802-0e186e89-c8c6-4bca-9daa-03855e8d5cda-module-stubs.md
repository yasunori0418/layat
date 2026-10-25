---
id: "DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda"
type: design
name: "NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする"
satisfies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
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

ADR-0066「NixOS / nix-darwin モジュールを system mode の配線として実装する」§1〜§4。
本 item は分割時点の原文の「flake.nix outputs 設計」にあったスタブ公開の項から起こし、
同 ADR で実装の形へ書き換えた。
