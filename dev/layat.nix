# layat の project mode config（dogfood）を flake-parts module として切り出す（→ ADR-0029）。
# root flake が公開する flakeModules.default（perSystem.layat.<name> を flake.layat.<system>.<name>
# へ転置する機構）を前提に、perSystem.layat.skills へ mattpocock/skills の manifest を宣言する。
# dev/flake.nix の imports に並べて読み込む。
#
# `layat apply skills -f <dev flake>` でビルドし、各 skill を .claude/skills/<name> へ
# store-symlink 配置する。root = projectRoot（git toplevel）なので配置先は repo root 配下。
# 配置物は .gitignore 済み（.claude/skills/*）の ephemeral。
{ inputs, ... }:
let
  layatLib = inputs.root.lib;

  # 展開する skill を明示列挙する（mattpocock/skills の skills/ 配下の相対パス）。
  # 本来は skills/<category> を builtins.readDir で動的列挙したい（既 realise の store パス /
  # flake input の readDir は IFD を起こさない）が、現行運用（skills-lock.json）の skill 集合を
  # 明示列挙して忠実に再現する。skills-lock.json は vercel skills 用に残置（両者は別経路）。
  skillSubpaths = [
    "productivity/grilling"
    "productivity/handoff"
  ];

  # skill ごとに { ".claude/skills/<name>" = entry; } を組む。
  # target = .claude/skills/<skill 名>、配置元は skills/<category>/<name> の subpath。
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
      # perSystem.layat.skills → flake.layat.<system>.skills へ自動転置される（root flakeModule）。
      # pkgs は perSystem 由来（= nixpkgs.legacyPackages.<system>）で packages.layat と一貫する。
      layat.skills = layatLib.mkManifest {
        inherit pkgs;
        root = layatLib.projectRoot;
        entries = skillEntries;
      };
    };
}
