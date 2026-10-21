# nix-unit: root マーカーの subdir 適用形（marker の形・全 target への前置・root 側への不漏洩・
# subdir の拒否ゲート）をアサートする。
# src は toString が安定する fake な flake-input 相当（`{ outPath = …; }`）を使う。
{ lib, layat }:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  norm = root: entries: layat.normalizeManifest { inherit lib root entries; };

  twoEntries = {
    "nvim" = {
      src = fakeSrc;
    };
    "git/config" = {
      src = fakeSrc;
      subpath = "git";
    };
  };

  targetsOf = m: map (e: e.target) m.entries;
in
{
  # ---- marker の形 ----------------------------------------------------------------
  # 単体 marker は __functor を除けば従来の 2 フィールドだけ（exact 一致で余分なキーを拒む）。
  testRootSubdirBareProjectRootShape = {
    expr = removeAttrs layat.projectRoot [ "__functor" ];
    expected = {
      _layatMarker = "root";
      kind = "project";
    };
  };

  testRootSubdirBareHomeRootShape = {
    expr = removeAttrs layat.homeRoot [ "__functor" ];
    expected = {
      _layatMarker = "root";
      kind = "home";
    };
  };

  testRootSubdirBareSystemRootShape = {
    expr = removeAttrs layat.systemRoot [ "__functor" ];
    expected = {
      _layatMarker = "root";
      kind = "system";
    };
  };

  # 適用形は subdir を足した marker。exact 一致なので __functor が残れば落ちる。
  testRootSubdirAppliedMarkerShape = {
    expr = layat.homeRoot { subdir = "x"; };
    expected = {
      _layatMarker = "root";
      kind = "home";
      subdir = "x";
    };
  };

  testRootSubdirAppliedMarkerHasNoFunctor = {
    expr = (layat.homeRoot { subdir = "x"; }) ? __functor;
    expected = false;
  };

  # ---- 適用形の引数 ----------------------------------------------------------------
  testRootSubdirMissingArgumentRejected = {
    expr = layat.homeRoot { };
    expectedError.type = "TypeError";
    expectedError.msg = "without required argument";
  };

  testRootSubdirUnknownArgumentRejected = {
    expr = layat.homeRoot {
      subdir = "x";
      prefix = "y";
    };
    expectedError.type = "TypeError";
    expectedError.msg = "unexpected argument";
  };

  testRootSubdirReapplyRejected = {
    expr = layat.homeRoot { subdir = "a"; } { subdir = "b"; };
    expectedError.type = "TypeError";
    expectedError.msg = "not a function";
  };

  # ---- 全 target への前置 ----------------------------------------------------------
  # entries は attrNames の辞書順で並ぶ。
  testRootSubdirPrefixesEveryTarget = {
    expr = targetsOf (norm (layat.homeRoot { subdir = ".config"; }) twoEntries);
    expected = [
      ".config/git/config"
      ".config/nvim"
    ];
  };

  # 明示 target にも前置する。
  testRootSubdirPrefixesExplicitTarget = {
    expr = targetsOf (
      norm (layat.projectRoot { subdir = "dev"; }) {
        "k" = {
          src = fakeSrc;
          target = "tools/k";
        };
      }
    );
    expected = [ "dev/tools/k" ];
  };

  testRootSubdirTrailingSlashNormalized = {
    expr = targetsOf (
      norm (layat.homeRoot { subdir = "a/"; }) {
        "x" = {
          src = fakeSrc;
        };
      }
    );
    expected = [ "a/x" ];
  };

  testRootSubdirLeadingDotNormalized = {
    expr = targetsOf (
      norm (layat.homeRoot { subdir = "./a"; }) {
        "x" = {
          src = fakeSrc;
        };
      }
    );
    expected = [ "a/x" ];
  };

  # 前置以外のフィールドは単体 marker のときと変わらない。
  testRootSubdirOtherFieldsUnchanged = {
    expr =
      map (e: removeAttrs e [ "target" ])
        (norm (layat.homeRoot { subdir = ".config"; }) twoEntries).entries;
    expected = map (e: removeAttrs e [ "target" ]) (norm layat.homeRoot twoEntries).entries;
  };

  # ---- root 側へ漏らさない ----------------------------------------------------------
  testRootSubdirRootKindUnchanged = {
    expr = map (r: (norm r { }).root.rootKind) [
      (layat.projectRoot { subdir = "a"; })
      (layat.homeRoot { subdir = "a"; })
      (layat.systemRoot { subdir = "a"; })
    ];
    expected = map (r: (norm r { }).root.rootKind) [
      layat.projectRoot
      layat.homeRoot
      layat.systemRoot
    ];
  };

  # root オブジェクトは rootKind だけ（exact 一致で subdir の不在まで見る）。
  testRootSubdirRootObjectHasNoSubdir = {
    expr = (norm (layat.homeRoot { subdir = ".config"; }) twoEntries).root;
    expected = {
      rootKind = "home";
    };
  };

  testRootSubdirSchemaVersionUnchanged = {
    expr = (norm (layat.homeRoot { subdir = ".config"; }) twoEntries).schemaVersion;
    expected = 1;
  };

  # ---- subdir の拒否ゲート -----------------------------------------------------------
  testRootSubdirAbsoluteRejected = {
    expr = (norm (layat.homeRoot { subdir = "/abs"; }) twoEntries).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "root subdir";
  };

  testRootSubdirEscapingRejected = {
    expr = (norm (layat.homeRoot { subdir = "../x"; }) twoEntries).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "root subdir";
  };

  testRootSubdirEmptyRejected = {
    expr = (norm (layat.homeRoot { subdir = ""; }) twoEntries).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "root subdir";
  };

  testRootSubdirDotRejected = {
    expr = (norm (layat.homeRoot { subdir = "."; }) twoEntries).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "root subdir";
  };

  testRootSubdirFoldsToEmptyRejected = {
    expr = (norm (layat.homeRoot { subdir = "a/.."; }) twoEntries).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "root subdir";
  };

  # entries が空でも subdir の検査は効く。
  testRootSubdirRejectedWithoutEntries = {
    expr = (norm (layat.homeRoot { subdir = "/abs"; }) { }).root;
    expectedError.type = "ThrownError";
    expectedError.msg = "root subdir";
  };

  # target の安全性検査は前置前の値に掛かる（前置で `..` 脱出を隠さない）。
  testRootSubdirEscapingTargetStillRejected = {
    expr =
      (norm (layat.homeRoot { subdir = "a"; }) {
        "k" = {
          src = fakeSrc;
          target = "../x";
        };
      }).entries;
    expectedError.type = "ThrownError";
    expectedError.msg = "invalid target";
  };
}
