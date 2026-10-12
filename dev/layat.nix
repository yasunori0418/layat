# layat の dogfood config。perSystem.layat.skills へ mattpocock/skills と dev/skills/ の manifest を宣言する。
# `layat apply skills -f <dev flake>` で各 skill を .claude/skills/<name> へ配置する。
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

  # dev/skills/（開発中のスキル正本）は編集を即時反映させるため、root 相対の可変 symlink で張る。
  devSkillEntries = builtins.listToAttrs (
    map
      (name: {
        name = ".claude/skills/${name}";
        value.src = layatLib.mkRootRelativeSymlink "dev/skills/${name}";
      })
      (
        builtins.attrNames (
          inputs.nixpkgs.lib.filterAttrs (_: type: type == "directory") (builtins.readDir ./skills)
        )
      )
  );

  # mattpocock/skills と dev/skills/ で名前が重なると後者が黙って勝つため、重なりを評価時に拒否する。
  duplicateSkillNames = map baseNameOf (
    builtins.attrNames (builtins.intersectAttrs skillEntries devSkillEntries)
  );
  entries =
    inputs.nixpkgs.lib.throwIf (duplicateSkillNames != [ ])
      "layat dev: skill name collides between mattpocock/skills and dev/skills: ${builtins.concatStringsSep ", " duplicateSkillNames}"
      (skillEntries // devSkillEntries);
in
{
  perSystem =
    { pkgs, ... }:
    {
      # flake.layat.<system>.skills へ転置される。
      layat.skills = layatLib.mkManifest {
        inherit pkgs entries;
        root = layatLib.projectRoot;
      };
    };
}
