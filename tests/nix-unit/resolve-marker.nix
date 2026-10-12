# nix-unit: `layat.__internal.resolveEntry` の src 文字列化・marker→enum 変換と、
# `lib/types.nix` の `isRootMarker` / `isOutOfStoreMarker` / `isRootRelativeMarker` の判別をアサートする。
# src は toString が安定する fake な flake-input 相当（`{ outPath = …; }`）を使う。
{ lib, layat }:
let
  inherit (layat.__internal) resolveEntry;
  types = import ../../lib/types.nix lib;
  inherit (types) isRootMarker isOutOfStoreMarker isRootRelativeMarker;

  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };

  resolve = resolveEntry lib;

  # store-backed src（path / derivation / flake input 相当）→ srcKind="store" / src=toString。
  storeEntry = resolve {
    src = fakeSrc;
    subpath = "skills/nix";
    target = ".claude/skills/nix";
    method = "symlink";
  };

  # out-of-store marker → srcKind="outOfStore" / src=marker の path。
  outOfStoreEntry = resolve {
    src = layat.mkOutOfStoreSymlink "/home/me/dotfiles/nvim";
    subpath = "lua";
    target = ".config/nvim";
    method = "copy";
  };

  # root 相対 marker → srcKind="rootRelative" / src=marker の root 相対 path。
  rootRelativeEntry = resolve {
    src = layat.mkRootRelativeSymlink "dev/skills/x";
    subpath = ".";
    target = ".claude/skills/x";
    method = "symlink";
  };
in
{
  # ---- store src の文字列化（toString）------------------------------------------------
  testResolveMarkerStoreSrcKind = {
    expr = storeEntry.srcKind;
    expected = "store";
  };

  testResolveMarkerStoreSrcString = {
    expr = storeEntry.src;
    expected = "/nix/store/00000000000000000000000000000000-fake-src";
  };

  # store entry の exact 形状。余分なキー（_layatMarker 等）が残れば exact 一致で fail する。
  testResolveMarkerStoreEntryShape = {
    expr = storeEntry;
    expected = {
      srcKind = "store";
      src = "/nix/store/00000000000000000000000000000000-fake-src";
      subpath = "skills/nix";
      target = ".claude/skills/nix";
      method = "symlink";
    };
  };

  # ---- out-of-store marker → clean enum 変換（src=marker の path）------------------------
  testResolveMarkerOutOfStoreSrcKind = {
    expr = outOfStoreEntry.srcKind;
    expected = "outOfStore";
  };

  testResolveMarkerOutOfStoreSrcPath = {
    expr = outOfStoreEntry.src;
    expected = "/home/me/dotfiles/nvim";
  };

  testResolveMarkerOutOfStoreEntryShape = {
    expr = outOfStoreEntry;
    expected = {
      srcKind = "outOfStore";
      src = "/home/me/dotfiles/nvim";
      subpath = "lua";
      target = ".config/nvim";
      method = "copy";
    };
  };

  # ---- root 相対 marker → clean enum 変換（src=marker の root 相対 path）----------------
  testResolveMarkerRootRelativeMarkerShape = {
    expr = layat.mkRootRelativeSymlink "dev/skills/x";
    expected = {
      _layatMarker = "rootRelative";
      path = "dev/skills/x";
    };
  };

  testResolveMarkerRootRelativeSrcKind = {
    expr = rootRelativeEntry.srcKind;
    expected = "rootRelative";
  };

  testResolveMarkerRootRelativeSrcPath = {
    expr = rootRelativeEntry.src;
    expected = "dev/skills/x";
  };

  testResolveMarkerRootRelativeEntryShape = {
    expr = rootRelativeEntry;
    expected = {
      srcKind = "rootRelative";
      src = "dev/skills/x";
      subpath = ".";
      target = ".claude/skills/x";
      method = "symlink";
    };
  };

  # ---- _layatMarker 判別タグが出力に漏れないこと（Go contract は clean enum）----
  testResolveMarkerNoTagLeakStore = {
    expr = storeEntry ? _layatMarker;
    expected = false;
  };

  testResolveMarkerNoTagLeakOutOfStore = {
    expr = outOfStoreEntry ? _layatMarker;
    expected = false;
  };

  testResolveMarkerNoTagLeakRootRelative = {
    expr = rootRelativeEntry ? _layatMarker;
    expected = false;
  };

  # normalizeManifest 経由でも rootRelative へ正規化され、タグが漏れない。
  testResolveMarkerRootRelativeNormalized = {
    expr =
      (layat.normalizeManifest {
        inherit lib;
        root = layat.projectRoot;
        entries.".claude/skills/x".src = layat.mkRootRelativeSymlink "dev/skills/x";
      }).entries;
    expected = [
      {
        srcKind = "rootRelative";
        src = "dev/skills/x";
        subpath = ".";
        target = ".claude/skills/x";
        method = "symlink";
      }
    ];
  };

  # ---- isOutOfStoreMarker の判別 -------------------------------------------------------
  testResolveMarkerIsOutOfStoreTrue = {
    expr = isOutOfStoreMarker (layat.mkOutOfStoreSymlink "/home/me/x");
    expected = true;
  };

  # root marker は outOfStore ではない（タグ値で判別）。
  testResolveMarkerIsOutOfStoreFalseOnRoot = {
    expr = isOutOfStoreMarker layat.projectRoot;
    expected = false;
  };

  # store-backed src（マーカーでない attrset）も outOfStore ではない。
  testResolveMarkerIsOutOfStoreFalseOnStore = {
    expr = isOutOfStoreMarker fakeSrc;
    expected = false;
  };

  # root 相対 marker は outOfStore ではない（タグ値で判別）。
  testResolveMarkerIsOutOfStoreFalseOnRootRelative = {
    expr = isOutOfStoreMarker (layat.mkRootRelativeSymlink "dev/x");
    expected = false;
  };

  # ---- isRootRelativeMarker の判別 -----------------------------------------------------
  testResolveMarkerIsRootRelativeTrue = {
    expr = isRootRelativeMarker (layat.mkRootRelativeSymlink "dev/x");
    expected = true;
  };

  testResolveMarkerIsRootRelativeFalseOnOutOfStore = {
    expr = isRootRelativeMarker (layat.mkOutOfStoreSymlink "/home/me/x");
    expected = false;
  };

  testResolveMarkerIsRootRelativeFalseOnRoot = {
    expr = isRootRelativeMarker layat.projectRoot;
    expected = false;
  };

  testResolveMarkerIsRootRelativeFalseOnStore = {
    expr = isRootRelativeMarker fakeSrc;
    expected = false;
  };

  # ---- isRootMarker の判別 -------------------------------------------------------------
  testResolveMarkerIsRootTrueProject = {
    expr = isRootMarker layat.projectRoot;
    expected = true;
  };

  testResolveMarkerIsRootTrueHome = {
    expr = isRootMarker layat.homeRoot;
    expected = true;
  };

  testResolveMarkerIsRootTrueSystem = {
    expr = isRootMarker layat.systemRoot;
    expected = true;
  };

  # out-of-store marker は root ではない。
  testResolveMarkerIsRootFalseOnOutOfStore = {
    expr = isRootMarker (layat.mkOutOfStoreSymlink "/home/me/x");
    expected = false;
  };

  # 文字列 root（絶対パス）は marker ではない。
  testResolveMarkerIsRootFalseOnString = {
    expr = isRootMarker "/abs/root";
    expected = false;
  };
}
