---
id: "UBQ-57c32ea5-8b3d-4b5d-b1f8-daee58ecbe40"
type: ubiquitous_term
name: "module"
status: stable
name_ja: "モジュール"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:40:14+09:00"
---

# UBQ-57c32ea5-8b3d-4b5d-b1f8-daee58ecbe40: module

## Definition

An integration layer such as standalone, home-manager, the NixOS / nix-darwin modules, or a devShell `shellHook`. It is purely the wiring that kicks the engine; it never places files itself and never translates to `home.file` or `systemd.tmpfiles`.

## Avoid

"the module places files", "the module translates to native mechanisms".

## 定義

CLI 単体・home-manager・NixOS / nix-darwin モジュール・devShell の `shellHook` といった統合層。エンジンを起動する配線（wiring）に徹し、自身ではファイルを配置せず、`home.file` や `systemd.tmpfiles` へ翻訳することもない。

## 避ける表現

「モジュールがファイルを配置する」「モジュールがネイティブ機構へ翻訳する」。
