---
id: "REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51"
type: requirement
name: "lib は nixpkgs.lib のみに依存する純データ生成器である"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  `lib` (`mkManifest`, `mkApplyCommand` and the marker functions) SHALL be a pure data generator that
  depends only on nixpkgs.lib. It SHALL NOT carry placement logic (filesystem operations,
  profile swap, stale removal), and SHALL NOT introduce a dependency on home-manager,
  NixOS or nix-darwin.
specification_ja: |
  `lib`（`mkManifest` / `mkApplyCommand` / マーカー群）は nixpkgs.lib のみに依存する
  純データ生成器でなければならない。配置ロジック（FS 操作・profile swap・古くなった配置の除去）を持ってはならず、
  home-manager / NixOS / nix-darwin への依存を持ち込んではならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51: lib は nixpkgs.lib のみに依存する純データ生成器である

## 仕様

`lib`（`mkManifest` / `mkApplyCommand` / マーカー群）は nixpkgs.lib のみに依存する純データ生成器。
配置ロジックは持たない。

- 依存は nixpkgs.lib に限る。`lib.types` / `mkOption` / `evalModules` は nixpkgs.lib の
  コアなのでこの制約を満たす。
- 実際の配置（place / replace / remove・profile swap・古くなった配置の除去）は engine の責務で、
  lib は一切行わない。
- 統合層（home-manager モジュール等）へ依存しないため、lib 単体を任意の Nix 環境から
  取り込める。

## 出典

`docs/spec.md`「lib API」。
