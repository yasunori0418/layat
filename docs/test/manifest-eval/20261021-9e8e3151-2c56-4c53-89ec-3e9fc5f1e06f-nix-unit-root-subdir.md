---
id: "CASE-9e8e3151-2c56-4c53-89ec-3e9fc5f1e06f"
type: test_case
name: "nix-unit: root-subdir.nix — root マーカーの subdir 適用形"
target: "tests/nix-unit/root-subdir.nix"
covers:
  - "TC-7d713505-5673-4e15-99bf-e98965e18892"
---
# CASE-9e8e3151-2c56-4c53-89ec-3e9fc5f1e06f: nix-unit root-subdir.nix

## 対象

`tests/nix-unit/root-subdir.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。TP-36e90d5d-4524-4294-bc72-ee263bb02782
が定めるとおり、テスト名はファイル横断で一意でなければならない。本ファイルは
`testRootSubdir*` 接頭辞を使う）

## 検証内容

root マーカーの subdir 適用形を、marker 単体と `normalizeManifest` の出力の両面からアサートする。

- 単体の `projectRoot` / `homeRoot` / `systemRoot` を `removeAttrs m [ "__functor" ]` で見ると、
  従来の 2 フィールドに exact 一致する
- `homeRoot { subdir = "x"; }` が `subdir` 付きの marker に exact 一致し、`__functor` を持たない
- `homeRoot { }`・未知キー付きの引数・再適用がそれぞれ `TypeError` になる
- 既定 target と明示 target の全 entry に前置が掛かり、`"a/"` と `"./a"` がどちらも `a/x` になる。
  target 以外のフィールドは単体 marker で正規化した結果と一致する
- 3 マーカーとも適用形の `rootKind` が単体と同じで、`root` は `{ rootKind = "home"; }` に
  exact 一致し、`schemaVersion` は 1 のまま
- `"/abs"`・`"../x"`・`""`・`"."`・`"a/.."` の `subdir` が `"root subdir"` を含む `ThrownError` になる。
  entries が空でも `root` の参照で同じエラーになる
- `subdir = "a"` の下でも target `"../x"` が `"invalid target"` で拒否される
- root 相対 src との重なりは前置後の target で判定する。前置後と一致する src は `"overlaps target"` で
  拒否され、前置前の target とだけ一致する src は受理される
- `subdir = null` と非文字列の `subdir` が型検査（`"is not of type"`）で拒否される
- 正規化ヘルパ `__internal.normalizeRelPath` が `.`・空成分・末尾スラッシュを除き、`..` を親へ畳み、
  何も残らない値を `""` にする

## 出典

`tests/nix-unit/root-subdir.nix` の実装。規範は REQ-08c8f639-790c-4970-a954-c9567538cf9b。
