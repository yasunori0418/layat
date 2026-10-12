# nix-unit: eval 時に throw する検査ゲート群（systemRoot 未実装・copy 併用・絶対/`..` escape・
# root 相対 src の target 重なり・重複 target・未知キー・素文字列 src・共有 entriesType）をアサートする。
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

  # out-of-store marker のパスは評価時に確定する絶対パスでなければならない。
  testOutOfStoreRelativePathRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/x" = {
          src = layat.mkOutOfStoreSymlink "rel/path";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "must be absolute";
  };

  # root 相対 marker のパスは絶対パスであってはならない。
  testRootRelativeAbsolutePathRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/x" = {
          src = layat.mkRootRelativeSymlink "/abs";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "invalid root-relative src";
  };

  # root 相対 marker のパスは `..` で root の外へ出てはならない。
  testRootRelativeEscapingPathRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/x" = {
          src = layat.mkRootRelativeSymlink "../x";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "invalid root-relative src";
  };

  # method = "copy" と root 相対 marker の併用は out-of-store と同じく拒否。
  testCopyRootRelativeRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/x" = {
          src = layat.mkRootRelativeSymlink "dev/x";
          method = "copy";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "cannot be combined with a root-relative marker";
  };

  # src/subpath が target の祖先（symlink ループになる）。
  testRootRelativeSrcAncestorOfTargetRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/skills/x" = {
          src = layat.mkRootRelativeSymlink ".claude";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "overlaps target";
  };

  # target が src/subpath の祖先。
  testRootRelativeTargetAncestorOfSrcRejected = {
    expr =
      (norm layat.projectRoot {
        "dev" = {
          src = layat.mkRootRelativeSymlink "dev";
          subpath = "skills/x";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "overlaps target";
  };

  # src/subpath と target が同一（`./` や subpath で表記が揺れても同一とみなす）。
  testRootRelativeSrcEqualsTargetRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/skills/x" = {
          src = layat.mkRootRelativeSymlink "./.claude/skills";
          subpath = "x/";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "overlaps target";
  };

  # root 内に留まる `..` を畳むと target の祖先になる src。
  testRootRelativeDotDotFoldedOverlapRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/skills/x" = {
          src = layat.mkRootRelativeSymlink "dev/../.claude";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "overlaps target";
  };

  # src = "." は root 自身なので、どの target とも重なる。
  testRootRelativeRootItselfRejected = {
    expr =
      (norm layat.projectRoot {
        ".claude/x" = {
          src = layat.mkRootRelativeSymlink ".";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "overlaps target";
  };

  # subpath を省いた entry のメッセージは src をそのまま出し、既定値 "." を "/." として付けない。
  testRootRelativeOverlapMessageOmitsDefaultSubpath = {
    expr =
      (norm layat.projectRoot {
        ".claude/x" = {
          src = layat.mkRootRelativeSymlink ".claude";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "other\\): \\.claude \\(target: \\.claude/x\\)";
  };

  # subpath 付きの entry のメッセージは src と subpath を "/" で繋いだ形で出す。
  testRootRelativeOverlapMessageJoinsSubpath = {
    expr =
      (norm layat.projectRoot {
        ".claude" = {
          src = layat.mkRootRelativeSymlink ".claude";
          subpath = "skills/x";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "other\\): \\.claude/skills/x \\(target: \\.claude\\)";
  };

  # 先頭の文字列が一致するだけの兄弟パスは重なりではない（コンポーネント単位で比較する）。
  testRootRelativeSiblingPrefixAccepted = {
    expr =
      map (e: e.target)
        (norm layat.projectRoot {
          ".claude-x/skills" = {
            src = layat.mkRootRelativeSymlink ".claude";
          };
        }).entries;
    expected = [ ".claude-x/skills" ];
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
