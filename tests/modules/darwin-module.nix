# nix-darwin モジュールの評価アサート。darwinSystem を評価し、postActivation への配線を検証する。
# 起動の形は評価時に検査するので、darwin 以外の system から drvPath を評価しても退行が露見する。
{
  inputs,
  pkgs,
  lib,
  layatPackage,
}:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  skill = {
    src = fakeSrc;
    subpath = "skills/nix";
  };
  mkDarwin =
    layatCfg:
    (inputs.nix-darwin.lib.darwinSystem {
      modules = [
        ../../modules/nix-darwin.nix
        {
          _module.args.layatPackage = layatPackage;
          nixpkgs.pkgs = pkgs;
          system.stateVersion = 6;
          # 利用者の追記と共存すること（text は types.lines）。
          system.activationScripts.postActivation.text = "echo user-post-activation";
          layat = {
            enable = true;
          }
          // layatCfg;
        }
      ];
    }).config;
  multi = mkDarwin {
    configs.default.entries."etc/skills/nix" = skill;
    configs.work.entries."etc/work" = skill;
    backup.enable = true;
  };
  empty = mkDarwin { };
  postActivation = c: c.system.activationScripts.postActivation.text;
  multiText = postActivation multi;
  applyLines = lib.filter (l: lib.hasInfix " apply " l) (lib.splitString "\n" multiText);
  launch = builtins.head applyLines;
  # 評価時の検査。失敗すれば drvPath の評価が throw する。
  evalChecks = [
    {
      ok = lib.length applyLines == 1;
      msg = "postActivation が layat apply を 1 回だけ起動していません";
    }
    {
      ok =
        builtins.match ".* apply --all --manifest 'default=/nix/store/[a-z0-9]+-layat-manifest' --manifest 'work=/nix/store/[a-z0-9]+-layat-manifest' --jobs 0 --backup=layat-backup" launch
        != null;
      msg = "postActivation が apply --all --manifest <name>=<farm> … --jobs 0 --backup=<suffix> の形ではありません";
    }
    {
      ok = !lib.hasInfix "--dryrun" multiText;
      msg = "postActivation に --dryrun が含まれています";
    }
    {
      ok = lib.hasInfix "echo user-post-activation" multiText;
      msg = "利用者の postActivation 追記が失われています";
    }
    {
      ok = !lib.hasInfix " apply " (postActivation empty);
      msg = "configs が空でも postActivation が layat を起動します";
    }
  ];
  evalAsserts = lib.all (c: lib.assertMsg c.ok "${c.msg}:\n${multiText}") evalChecks;
  activationScript = pkgs.writeText "layat-postactivation" multiText;
in
builtins.seq evalAsserts (
  pkgs.runCommandLocal "layat-darwin-module-check" { } ''
    script=${activationScript}

    # config ごとの manifest が root=systemRoot を pin していること。
    for name in default work; do
      json="$(grep -oE "$name=/nix/store/[^ ']+" "$script" | cut -d= -f2)/manifest.json"
      grep -q '"rootKind":"system"' "$json" \
        || { echo "FAIL: $name の manifest が systemRoot を pin していません"; cat "$json"; exit 1; }
    done

    touch "$out"
  ''
)
