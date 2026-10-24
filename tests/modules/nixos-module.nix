# NixOS モジュールの評価アサート。nixosSystem を評価し、activation の配線と manifest の内容を検証する。
# toplevel は評価しない（fileSystems / bootloader の assertions を避ける）。実 activate は VM テストが担う。
{
  inputs,
  pkgs,
  lib,
  layatPackage,
}:
let
  # store hash 揺れを避ける fake な flake-input 相当（checks.hm-module と同じ test double）。
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  skill = {
    src = fakeSrc;
    subpath = "skills/nix";
  };
  # 共通の土台に layat の設定だけを差し替えて評価し、config を返す。
  mkNixos =
    layatCfg:
    (inputs.nixpkgs.lib.nixosSystem {
      modules = [
        ../../modules/nixos.nix
        {
          # ラッパー（flake.nixosModules.default）と同じく pin 版 layat を注入する。
          _module.args.layatPackage = layatPackage;
          nixpkgs.pkgs = pkgs;
          system.stateVersion = "24.05";
          layat = {
            enable = true;
          }
          // layatCfg;
        }
      ];
    }).config;
  # suffix 省略時に既定値 "layat-backup" が activation に渡ることも検証する。
  multi = mkNixos {
    configs.default.entries."etc/skills/nix" = skill;
    configs.work.entries."etc/work" = skill;
    backup.enable = true;
  };
  empty = mkNixos { };
  jobs3 = mkNixos {
    configs.default.entries."etc/skills/nix" = skill;
    jobs = 3;
  };
  recopy = mkNixos {
    configs.default.entries."etc/skills/nix" = skill;
    recopy = true;
  };
  # subdir は指定した config の target にだけ前置される。
  subdir = mkNixos {
    configs.nvim = {
      subdir = "etc/xdg";
      entries."nvim" = skill;
    };
    configs.default.entries."etc/skills/nix" = skill;
  };
  layatScript = c: c.system.activationScripts.layat;
  # activation の生スクリプト。
  activationScript = pkgs.writeText "layat-activation" (layatScript multi).text;
in
pkgs.runCommandLocal "layat-nixos-module-check"
  {
    deps = lib.concatStringsSep " " (layatScript multi).deps;
    supportsDryActivation = lib.boolToString (layatScript multi).supportsDryActivation;
    nixBin = "${multi.nix.package}/bin";
    emptyLaunches = lib.boolToString (empty.system.activationScripts ? layat);
    jobs3Activation = (layatScript jobs3).text;
    recopyActivation = (layatScript recopy).text;
    subdirActivation = (layatScript subdir).text;
  }
  ''
    script=${activationScript}

    # (1) 全 config の link-farm を辞書順の <name>=<farm> で 1 回の apply --all 起動に渡し、
    #     既定の --jobs 0 と --backup=<既定 suffix> を付けること。
    test "$(grep -c ' apply ' "$script")" -eq 1 \
      || { echo "FAIL: activation が layat apply を 1 回だけ起動していません"; cat "$script"; exit 1; }
    grep -qE " apply --all --manifest 'default=/nix/store/[a-z0-9]+-layat-manifest' --manifest 'work=/nix/store/[a-z0-9]+-layat-manifest' --jobs 0 --backup=layat-backup( |$)" "$script" \
      || { echo "FAIL: activation が apply --all --manifest <name>=<farm> … --jobs 0 --backup=<suffix> の形ではありません"; cat "$script"; exit 1; }

    # (2) config ごとに別 store path の manifest が root=systemRoot を pin し、自分の entries だけを持つこと。
    farmDefault=$(grep -oE "default=/nix/store/[^ ']+" "$script" | cut -d= -f2)
    farmWork=$(grep -oE "work=/nix/store/[^ ']+" "$script" | cut -d= -f2)
    test "$farmDefault" != "$farmWork" \
      || { echo "FAIL: 2 configs の manifest が同じ store path です"; cat "$script"; exit 1; }
    for spec in "$farmDefault:etc/skills/nix:etc/work" "$farmWork:etc/work:etc/skills/nix"; do
      IFS=: read -r farm own other <<<"$spec"
      json="$farm/manifest.json"
      test -f "$json" || { echo "FAIL: $json がありません"; exit 1; }
      grep -q '"rootKind":"system"' "$json" \
        || { echo "FAIL: manifest が systemRoot を pin していません"; cat "$json"; exit 1; }
      grep -qF "\"$own\"" "$json" \
        || { echo "FAIL: configs.<name>.entries が manifest に反映されていません"; cat "$json"; exit 1; }
      ! grep -qF "\"$other\"" "$json" \
        || { echo "FAIL: 他の config の entries が manifest に混入しています"; cat "$json"; exit 1; }
    done

    # (3) etc / users の後に走り、dry-activate 用スクリプトにも含まれ、NIXOS_ACTION が dry-activate の
    #     ときだけ --dryrun を付け、PATH に nix.package の bin を前置すること。
    test "$deps" = "etc users" \
      || { echo "FAIL: deps が etc users ではありません: $deps"; exit 1; }
    test "$supportsDryActivation" = true \
      || { echo "FAIL: supportsDryActivation が true ではありません"; exit 1; }
    grep -qxF 'if [ "''${NIXOS_ACTION:-}" = dry-activate ]; then layatDryrun=--dryrun; fi' "$script" \
      && grep -qE ' apply --all .* \$layatDryrun$' "$script" \
      || { echo "FAIL: NIXOS_ACTION=dry-activate のときだけ --dryrun を起動へ付ける分岐がありません"; cat "$script"; exit 1; }
    grep -qF "PATH=$nixBin:" "$script" \
      || { echo "FAIL: PATH に nix.package の bin が前置されていません"; cat "$script"; exit 1; }

    # (4) layat.jobs の値がそのまま --jobs に乗ること。
    grep -qE " --manifest 'default=[^']+' --jobs 3( |$)" <<<"$jobs3Activation" \
      || { echo "FAIL: layat.jobs = 3 が --jobs 3 として渡っていません"; echo "$jobs3Activation"; exit 1; }

    # (5) layat.recopy = true のときだけ --recopy が同じ起動に乗ること。
    grep -qE " apply --all --manifest 'default=[^']+' --jobs 0 --recopy( |$)" <<<"$recopyActivation" \
      || { echo "FAIL: layat.recopy = true が --recopy として渡っていません"; echo "$recopyActivation"; exit 1; }
    ! grep -qF -- "--recopy" "$script" && ! grep -qF -- "--recopy" <<<"$jobs3Activation" \
      || { echo "FAIL: layat.recopy の既定で --recopy が渡っています"; cat "$script"; echo "$jobs3Activation"; exit 1; }

    # (6) config が 1 つも無いとき activation は layat を起動しないこと。
    test "$emptyLaunches" = false \
      || { echo "FAIL: configs が空でも activation が layat を起動します"; exit 1; }

    # (7) subdir 付き config だけ target に subdir が前置され、root は system のままであること。
    nvimJson="$(grep -oE "nvim=/nix/store/[^ ']+" <<<"$subdirActivation" | cut -d= -f2)/manifest.json"
    defaultJson="$(grep -oE "default=/nix/store/[^ ']+" <<<"$subdirActivation" | cut -d= -f2)/manifest.json"
    grep -qF '"etc/xdg/nvim"' "$nvimJson" && grep -q '"rootKind":"system"' "$nvimJson" \
      || { echo "FAIL: subdir が nvim の target に前置されていません"; cat "$nvimJson"; exit 1; }
    grep -qF '"etc/skills/nix"' "$defaultJson" && ! grep -qF 'etc/xdg/' "$defaultJson" \
      || { echo "FAIL: subdir が別 config に漏れています"; cat "$defaultJson"; exit 1; }

    touch "$out"
  ''
