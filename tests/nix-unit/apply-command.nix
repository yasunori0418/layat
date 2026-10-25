# nix-unit: mkApplyCommand が組む `apply --all --manifest …` の文字列をアサートする。
# pkgs は lib だけを持つ fake、layatPackage は `lib.getExe` が純評価できる fake（derivation 扱いの attrset）を使う。
# farm も toString が安定する fake な outPath で、store hash の揺れを避ける。
{ lib, layat }:
let
  fakePkgs = { inherit lib; };
  fakeLayat = {
    type = "derivation";
    outPath = "/nix/store/00000000000000000000000000000000-layat";
    meta.mainProgram = "layat";
  };
  farm = name: {
    outPath = "/nix/store/00000000000000000000000000000000-${name}";
  };
  exe = "/nix/store/00000000000000000000000000000000-layat/bin/layat";

  # 名前の辞書順と attrset の宣言順が食い違うように並べる。
  manifests = {
    zeta = farm "zeta";
    alpha = farm "alpha";
  };
  manifestArgs = "--manifest 'alpha=/nix/store/00000000000000000000000000000000-alpha' --manifest 'zeta=/nix/store/00000000000000000000000000000000-zeta'";

  mk =
    root:
    layat.mkApplyCommand {
      pkgs = fakePkgs;
      layatPackage = fakeLayat;
      inherit manifests root;
    };
in
{
  # "all" だけがフィルタなしの bare `--all` を出す。
  testApplyCommandRootAll = {
    expr = mk "all";
    expected = "${exe} apply --all ${manifestArgs}";
  };

  testApplyCommandRootProject = {
    expr = mk "project";
    expected = "${exe} apply --all --project-root ${manifestArgs}";
  };

  testApplyCommandRootHome = {
    expr = mk "home";
    expected = "${exe} apply --all --home-root ${manifestArgs}";
  };

  testApplyCommandRootSystem = {
    expr = mk "system";
    expected = "${exe} apply --all --system-root ${manifestArgs}";
  };

  # `<name>=<farm>` 全体を 1 語として escapeShellArg で包む（空白を含む名前でも語が割れない）。
  testApplyCommandEscapesManifestArg = {
    expr = layat.mkApplyCommand {
      pkgs = fakePkgs;
      layatPackage = fakeLayat;
      manifests = {
        "my cfg" = farm "cfg";
      };
      root = "all";
    };
    expected = "${exe} apply --all --manifest 'my cfg=/nix/store/00000000000000000000000000000000-cfg'";
  };

  # 実行ファイルは layatPackage の meta.mainProgram から引く（PATH 上の layat に頼らない）。
  testApplyCommandUsesPackageExe = {
    expr = lib.head (
      lib.splitString " " (
        layat.mkApplyCommand {
          pkgs = fakePkgs;
          layatPackage = {
            type = "derivation";
            outPath = "/nix/store/11111111111111111111111111111111-layat-pinned";
            meta.mainProgram = "layat";
          };
          inherit manifests;
          root = "all";
        }
      )
    );
    expected = "/nix/store/11111111111111111111111111111111-layat-pinned/bin/layat";
  };

  # root は既定値を持たない。省略すると throw する。
  testApplyCommandRootMissingRejected = {
    expr = layat.mkApplyCommand {
      pkgs = fakePkgs;
      layatPackage = fakeLayat;
      inherit manifests;
    };
    expectedError.type = "ThrownError";
    expectedError.msg = "root is required";
  };

  testApplyCommandRootInvalidRejected = {
    expr = mk "fixed";
    expectedError.type = "ThrownError";
    expectedError.msg = "invalid root";
  };
}
