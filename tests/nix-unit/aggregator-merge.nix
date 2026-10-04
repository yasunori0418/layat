# nix-unit: アグリゲータの衝突検査（`tests/nix-unit-lib.nix` の `mergeTests`）をダミー入力でアサートする。
# ダミーの値は名前と値の対応の取り違えを見分けられるよう全て違う値にする。
# `layat` は使わないが、シグネチャは他の leaf に揃える。
{ lib, layat }:
let
  inherit (import ../nix-unit-lib.nix { inherit lib; }) mergeTests;

  # 衝突の無い 2 ファイル。
  disjoint = [
    {
      file = "alpha.nix";
      tests = {
        testAlphaOne = 1;
        testAlphaTwo = 2;
      };
    }
    {
      file = "beta.nix";
      tests = {
        testBetaOne = 3;
      };
    }
  ];

  # 同名 `testDup` を 2 ファイルが定義する。
  twoWayCollision = [
    {
      file = "alpha.nix";
      tests = {
        testAlphaOne = 1;
        testDup = 2;
      };
    }
    {
      file = "beta.nix";
      tests = {
        testDup = 3;
      };
    }
  ];

  # 同名 `testDup` を 3 ファイルが定義する。
  threeWayCollision = [
    {
      file = "alpha.nix";
      tests.testDup = 1;
    }
    {
      file = "beta.nix";
      tests.testDup = 2;
    }
    {
      file = "gamma.nix";
      tests.testDup = 3;
    }
  ];

  # 衝突するテスト名が 2 組ある（報告が 2 行になる）。
  twoCollisionGroups = [
    {
      file = "alpha.nix";
      tests = {
        testDupA = 1;
        testDupB = 2;
      };
    }
    {
      file = "beta.nix";
      tests = {
        testDupA = 3;
        testDupB = 4;
      };
    }
  ];
in
{
  # 衝突が無ければ全ファイルのテスト名がマージ結果に揃う。
  testAggregatorMergeUnionsNames = {
    expr = lib.attrNames (mergeTests disjoint);
    expected = [
      "testAlphaOne"
      "testAlphaTwo"
      "testBetaOne"
    ];
  };

  # 値もそのまま写る（名前だけ集めて中身を落とす実装にならない担保）。
  testAggregatorMergeKeepsValues = {
    expr = mergeTests disjoint;
    expected = {
      testAlphaOne = 1;
      testAlphaTwo = 2;
      testBetaOne = 3;
    };
  };

  # 空入力は空 attrset（衝突なしの境界の下端）。
  testAggregatorMergeEmptyInput = {
    expr = mergeTests [ ];
    expected = { };
  };

  # 1 ファイルだけなら、テストを何件持っていても衝突しない（`length owners > 1` の境界の下側）。
  testAggregatorMergeSingleModuleManyTests = {
    expr = mergeTests [
      {
        file = "solo.nix";
        tests = {
          testSoloOne = 1;
          testSoloTwo = 2;
          testSoloThree = 3;
          testSoloFour = 4;
        };
      }
    ];
    expected = {
      testSoloOne = 1;
      testSoloTwo = 2;
      testSoloThree = 3;
      testSoloFour = 4;
    };
  };

  # 2 ファイルが同名を定義したら throw する。メッセージに衝突したテスト名と
  # 両方のファイル名が出る（後勝ちで消えた側を追えるようにするため）。
  testAggregatorMergeTwoWayCollisionThrows = {
    expr = mergeTests twoWayCollision;
    expectedError.type = "ThrownError";
    expectedError.msg = "  - testDup: alpha\\.nix, beta\\.nix";
  };

  # 衝突していないテスト名は報告に出ない。`expectedError.msg` は ECMAScript の正規表現で、
  # `$` は文字列末尾を指し、複数行は `(.|\n)` で舐める。
  testAggregatorMergeCollisionReportExcludesInnocent = {
    expr = mergeTests twoWayCollision;
    expectedError.type = "ThrownError";
    expectedError.msg = "^(.|\n)*テスト名がファイル横断で衝突((?!testAlphaOne)(.|\n))*$";
  };

  # 3 ファイル以上なら owners が 3 件とも列挙される（2 件で打ち切らない）。
  testAggregatorMergeThreeWayCollisionListsAllOwners = {
    expr = mergeTests threeWayCollision;
    expectedError.type = "ThrownError";
    expectedError.msg = "  - testDup: alpha\\.nix, beta\\.nix, gamma\\.nix";
  };

  # 衝突が複数組あれば報告が複数行になる（最初の 1 組で止めない）。
  testAggregatorMergeMultipleCollisionGroupsReportEachLine = {
    expr = mergeTests twoCollisionGroups;
    expectedError.type = "ThrownError";
    expectedError.msg = "  - testDupA: alpha\\.nix, beta\\.nix\n  - testDupB: alpha\\.nix, beta\\.nix";
  };
}
