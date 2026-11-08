---
id: "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
type: requirement
name: "mkManifest 自身が evalModules で入力を確認する単一の箇所になる"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `lib.mkManifest` SHALL itself run `lib.evalModules` internally to validate and normalize
  `entries` and `root`. Declaring the types only on module options would make validation
  effective on the module path (`layat.entries`) alone, so `mkManifest` SHALL own the
  validation in order to cover the CLI / entrypoint path (a direct `mkManifest` call) as
  well. The entry type definition (the submodule in `lib/types.nix`) SHALL be shared with
  `attrsOf (submodule …)` in `modules/common.nix`, and the resulting double validation on
  the module path is acceptable because it is pure and idempotent.
specification_ja: |
  `lib.mkManifest` は内部で `lib.evalModules` を回して `entries` / `root` を確認・
  正規化しなければならない。型をモジュールオプションに書くだけでは確認がモジュールの使い方
  （`layat.entries`）でしか働かないため、CLI / entrypoint の使い方（`mkManifest` 直呼び）でも
  確認を働かせる目的で `mkManifest` 自身が確認を持たなければならない。entry の型定義
  （`lib/types.nix` の submodule）は `modules/common.nix` の `attrsOf (submodule …)` と
  共有しなければならず、モジュールの使い方での二重の確認は純粋・冪等であるため許容される。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:54:35+09:00"
---
# REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34: mkManifest 自身が evalModules で入力を確認する単一の箇所になる

## 仕様

`mkManifest` は内部で `lib.evalModules` を回して `entries` / `root` を確認・正規化する。

- 型をオプションに書くだけだと確認が働くのはモジュールの使い方（`layat.entries`）のみだが、
  コアである CLI / entrypoint の使い方（`mkManifest` 直呼び）でも確認を働かせるため
  `mkManifest` 自身が `evalModules` を回す。
- `lib.types` / `mkOption` / `evalModules` は `nixpkgs.lib` のコアなので「lib は
  nixpkgs.lib のみ依存」を満たす。
- entry の型定義（`lib/types.nix` の submodule）は `modules/common.nix` の
  `attrsOf (submodule …)` と共有する。モジュールの使い方では host の `evalModules`
  （`attrsOf`）と `mkManifest` の `evalModules` で二重に確認されるが、純粋・冪等で
  害はなく、`mkManifest` を「entries が必ず通る単一の確認箇所」に保つ。

## 出典

`docs/spec.md`「lib API」→「入力検査（`evalModules` + `normalizeManifest`）」。
