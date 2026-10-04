# nix-unit: fixed root（root に絶対パス文字列を渡すと `rootKind = "fixed"` + `root` になる）をアサートする。
# marker 側（固定 root を持たない）は structure.nix が見る。
# src は toString が安定する fake な flake-input 相当（`{ outPath = …; }`）を使う。
{ lib, layat }:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  norm = root: entries: layat.normalizeManifest { inherit lib root entries; };

  fixed = norm "/srv/deploy" {
    ".config/foo" = {
      src = fakeSrc;
    };
  };
in
{
  # marker でない文字列を root に渡すと fixed になる。
  testFixedRootKind = {
    expr = fixed.root.rootKind;
    expected = "fixed";
  };

  # fixed のときだけ絶対パスを併記する。project（structure.nix の否定側）との対。
  testFixedRootPath = {
    expr = fixed.root.root;
    expected = "/srv/deploy";
  };

  # root オブジェクトは rootKind と root のちょうど 2 フィールド。exact 一致で見るので
  # 余分なキーが混ざれば落ちる（entry 側の shape アサートと同じ方針）。
  testFixedRootObjectShape = {
    expr = fixed.root;
    expected = {
      rootKind = "fixed";
      root = "/srv/deploy";
    };
  };

  # 渡した文字列がそのまま写ることを別のパスでも確かめる（特定の値に依存した通り方を避ける）。
  # entries は上と同条件に揃え、差分がパス値の 1 軸だけになるようにする。
  testFixedRootPathVerbatim = {
    expr =
      (norm "/opt/layat" {
        ".config/foo" = {
          src = fakeSrc;
        };
      }).root;
    expected = {
      rootKind = "fixed";
      root = "/opt/layat";
    };
  };

  # entry の正規化は root 種別に依らない。project root との同値比較で見る
  # （真偽値へ畳まず、ずれたフィールドが出力に出るようにする）。
  testFixedRootEntryUnaffected = {
    expr = fixed.entries;
    expected =
      (norm layat.projectRoot {
        ".config/foo" = {
          src = fakeSrc;
        };
      }).entries;
  };

  # homeRoot marker は fixed にならず絶対パスも持たない（exact 一致で `root` の不在まで見る）。
  testFixedRootHomeMarkerShape = {
    expr = (norm layat.homeRoot { }).root;
    expected = {
      rootKind = "home";
    };
  };
}
