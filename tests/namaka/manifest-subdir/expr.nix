# namaka: root マーカーの subdir 適用形で生成した normalizeManifest 出力のスナップショット回帰。
# target は `.config/` 前置済みで、root は rootKind だけになる。
{ lib, layat }:
layat.normalizeManifest {
  inherit lib;
  root = layat.homeRoot { subdir = ".config"; };
  entries = {
    "nvim" = {
      src = {
        outPath = "/nix/store/00000000000000000000000000000000-fake-src";
      };
      subpath = "nvim";
    };
    # subpath / target / method 省略 → デフォルト適用
    "git" = {
      src = {
        outPath = "/nix/store/11111111111111111111111111111111-other";
      };
    };
  };
}
