# nix-unit アグリゲータ: `tests/nix-unit/` 配下の全 `*.nix` を `{ lib, layat }` で import する。
# マージとテスト名の重複検査は `tests/nix-unit-lib.nix` の `mergeTests` が行う。
{ lib, layat }:
let
  inherit (import ./nix-unit-lib.nix { inherit lib; }) mergeTests;

  dir = ./nix-unit;
  testFiles = lib.filterAttrs (name: type: type == "regular" && lib.hasSuffix ".nix" name) (
    builtins.readDir dir
  );
  # { file, tests } の組で持ち回り、衝突時にどのファイル同士かを示せるようにする。
  modules = lib.mapAttrsToList (name: _type: {
    file = name;
    tests = import (dir + "/${name}") { inherit lib layat; };
  }) testFiles;
in
mergeTests modules
