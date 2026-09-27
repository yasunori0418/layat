# nix-unit: manifest 構造の不変条件（schemaVersion / root 系 / store・outOfStore エントリ構造）を
# アサートする（→ ADR-0006, ADR-0010, ADR-0014）。
#
# store パスの hash 揺れを避けるため src には toString が安定する fake な flake-input 相当
# （`{ outPath = …; }`）を使う。これは srcType の store-backed 判定（`? outPath`）を通る正当な test double。
{ lib, layat }:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  norm = root: entries: layat.normalizeManifest { inherit lib root entries; };

  # passthru 検証用の fake pkgs（farm-entries.nix と同じイディオム）。runCommandLocal の attrs から
  # passthru だけを持ち帰り、derivation を組まずに `mkManifest` の passthru を純評価で取り出す。
  fakePkgs = {
    inherit lib;
    writeText = name: _text: "/nix/store/fake-${name}";
    runCommandLocal =
      _name: attrs: _script:
      attrs.passthru;
  };
  passthruOf =
    entries:
    layat.mkManifest {
      pkgs = fakePkgs;
      root = layat.projectRoot;
      inherit entries;
    };

  basic = norm layat.projectRoot {
    ".claude/skills/nix" = {
      src = fakeSrc;
      subpath = "skills/nix";
    };
  };
in
{
  testSchemaVersion = {
    expr = basic.schemaVersion;
    expected = 1;
  };

  testRootKindProject = {
    expr = basic.root.rootKind;
    expected = "project";
  };

  # project は実行時解決なので固定 root パスを持たない（→ ADR-0010）。
  testProjectHasNoFixedRoot = {
    expr = basic.root ? root;
    expected = false;
  };

  testStoreEntry = {
    expr = builtins.head basic.entries;
    expected = {
      srcKind = "store";
      src = "/nix/store/00000000000000000000000000000000-fake-src";
      subpath = "skills/nix";
      target = ".claude/skills/nix";
      method = "symlink";
    };
  };

  # out-of-store marker → clean enum 変換（→ ADR-0001, ADR-0010, ADR-0013）。
  # srcKind = "outOfStore" / src = marker の絶対パスが記録され、_layatMarker は漏れない
  # （expected は exact 一致なので余分なキーが残れば fail する）。
  testOutOfStoreEntry = {
    expr =
      builtins.head
        (norm layat.projectRoot {
          ".config/nvim" = {
            src = layat.mkOutOfStoreSymlink "/home/me/dotfiles/nvim";
            subpath = "lua";
          };
        }).entries;
    expected = {
      srcKind = "outOfStore";
      src = "/home/me/dotfiles/nvim";
      subpath = "lua";
      target = ".config/nvim";
      method = "symlink";
    };
  };

  # out-of-store entry に _layatMarker 判別タグが漏れていないことを明示アサートする。
  testOutOfStoreMarkerNotLeaked = {
    expr =
      (builtins.head
        (norm layat.projectRoot {
          ".config/nvim" = {
            src = layat.mkOutOfStoreSymlink "/home/me/dotfiles/nvim";
          };
        }).entries
      )
        ? _layatMarker;
    expected = false;
  };

  # passthru targets は正規化後 target を attrNames（キー）の辞書順で返す（→ ADR-0038）。
  testPassthruTargetsLexical = {
    expr =
      (passthruOf {
        "b/two" = {
          src = fakeSrc;
        };
        "a/one" = {
          src = fakeSrc;
        };
      }).targets;
    expected = [
      "a/one"
      "b/two"
    ];
  };

  # 明示 target 上書きはキーではなく上書き後の値で現れ、並びはキー順のまま（→ ADR-0038）。
  testPassthruTargetsOverride = {
    expr =
      (passthruOf {
        a = {
          src = fakeSrc;
          target = "zzz/override";
        };
        b = {
          src = fakeSrc;
        };
      }).targets;
    expected = [
      "zzz/override"
      "b"
    ];
  };
}
