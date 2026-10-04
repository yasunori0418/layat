# nix-unit: normalizeManifest のデフォルト適用・明示上書き・target 辞書順をアサートする。
# src は toString が安定する fake な flake-input 相当（`{ outPath = …; }`）を使う。
{ lib, layat }:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  norm = root: entries: layat.normalizeManifest { inherit lib root entries; };
in
{
  # ---- デフォルト適用（subpath="." / target=属性キー / method="symlink"）-----
  testDefaultsApplied = {
    expr =
      builtins.head
        (norm layat.projectRoot {
          ".config/foo" = {
            src = fakeSrc;
          };
        }).entries;
    expected = {
      srcKind = "store";
      src = "/nix/store/00000000000000000000000000000000-fake-src";
      subpath = ".";
      target = ".config/foo";
      method = "symlink";
    };
  };

  # 明示上書きが反映される
  testExplicitOverrides = {
    expr =
      builtins.head
        (norm layat.projectRoot {
          "label" = {
            src = fakeSrc;
            target = ".config/bar";
            subpath = "sub/dir";
            method = "copy";
          };
        }).entries;
    expected = {
      srcKind = "store";
      src = "/nix/store/00000000000000000000000000000000-fake-src";
      subpath = "sub/dir";
      target = ".config/bar";
      method = "copy";
    };
  };

  # entries は target（属性キー）の辞書順で決定的に配列化される。
  testEntriesSortedByTarget = {
    expr =
      map (e: e.target)
        (norm layat.projectRoot {
          "b" = {
            src = fakeSrc;
          };
          "a" = {
            src = fakeSrc;
          };
          "c" = {
            src = fakeSrc;
          };
        }).entries;
    expected = [
      "a"
      "b"
      "c"
    ];
  };
}
