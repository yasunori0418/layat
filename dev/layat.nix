# layat の dogfood config。perSystem.layat.skills へ mattpocock/skills の manifest を宣言する。
# `layat apply skills -f <dev flake>` で各 skill を .claude/skills/<name> へ store-symlink 配置する。
{ inputs, ... }:
let
  layatLib = inputs.root.lib;

  # 展開する skill（mattpocock/skills の skills/ 配下の相対パス）。
  skillSubpaths = [
    "productivity/grilling"
    "productivity/handoff"
  ];

  # skill ごとに { ".claude/skills/<name>" = entry; } を組む。
  skillEntries = builtins.listToAttrs (
    map (p: {
      name = ".claude/skills/${baseNameOf p}";
      value = {
        src = inputs.matt-skills;
        subpath = "skills/${p}";
      };
    }) skillSubpaths
  );
in
{
  perSystem =
    { pkgs, ... }:
    {
      # flake.layat.<system>.skills へ転置される。
      layat.skills = layatLib.mkManifest {
        inherit pkgs;
        root = layatLib.projectRoot;
        entries = skillEntries;
      };
    };
}
