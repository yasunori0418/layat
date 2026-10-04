# nix-unit: eval 時に throw する検査ゲート群（systemRoot 未実装・copy×outOfStore・絶対/`..` escape・
# 重複 target・未知キー・素文字列 src・共有 entriesType）をアサートする。
# src は toString が安定する fake な flake-input 相当（`{ outPath = …; }`）を使う。
{ lib, layat }:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  norm = root: entries: layat.normalizeManifest { inherit lib root entries; };
in
{
  # systemRoot は未実装。
  testSystemRootUnimplemented = {
    expr =
      (norm layat.systemRoot {
        "x" = {
          src = fakeSrc;
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "system mode";
  };

  # method = "copy" かつ out-of-store marker は意図矛盾。
  testCopyOutOfStoreRejected = {
    expr =
      (norm layat.projectRoot {
        ".config/x" = {
          src = layat.mkOutOfStoreSymlink "/home/me/dots";
          method = "copy";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "out-of-store";
  };

  # target が絶対パス。
  testAbsoluteTargetRejected = {
    expr =
      (norm layat.projectRoot {
        "/etc/x" = {
          src = fakeSrc;
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "target";
  };

  # target が `..` で root の外。
  testEscapingTargetRejected = {
    expr =
      (norm layat.projectRoot {
        "../../etc/x" = {
          src = fakeSrc;
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "target";
  };

  # subpath が `..` で src の外。
  testEscapingSubpathRejected = {
    expr =
      (norm layat.projectRoot {
        ".config/x" = {
          src = fakeSrc;
          subpath = "../escape";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "subpath";
  };

  # 別キーで target を同値に明示上書きした衝突。
  testDuplicateTargetRejected = {
    expr =
      (norm layat.projectRoot {
        "a" = {
          src = fakeSrc;
          target = ".config/same";
        };
        "b" = {
          src = fakeSrc;
          target = ".config/same";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "same target";
  };

  # 未知キー（タイポ / 旧名）は submodule strict で弾く。
  testUnknownKeyRejected = {
    expr =
      (norm layat.projectRoot {
        ".config/x" = {
          src = fakeSrc;
          source = "skills/nix"; # 旧名（正しくは subpath）
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "source";
  };

  # 素の文字列 src は拒否（out-of-store は marker で opt-in）。
  testStringSrcRejected = {
    expr =
      (norm layat.projectRoot {
        ".config/x" = {
          src = "/home/me/dots";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "src";
  };

  # modules/common.nix が共有する entriesType を evalModules で直接検査する。
  # 未知キーはモジュール経路でも eval エラーになる。
  testSharedEntriesTypeUnknownKey = {
    expr =
      let
        t = import ../../lib/types.nix lib;
        evaluated = lib.evalModules {
          modules = [
            { options.entries = lib.mkOption { type = t.entriesType; }; }
            {
              entries.".config/x" = {
                src = fakeSrc;
                bogus = true; # 未知キー
              };
            }
          ];
        };
      in
      evaluated.config.entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "bogus";
  };
}
