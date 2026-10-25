# NixOS module: kicks the engine (modules/kick.nix) from system.activationScripts.layat with
# root = systemRoot (applied to { subdir } when the config sets one).
# layatPackage (the pinned layat CLI) is injected as _module.args by flake.nix.
{
  config,
  lib,
  pkgs,
  layatPackage,
  ...
}:
let
  cfg = config.layat;
  layatLib = import ../lib;

  kick = import ./kick.nix {
    inherit
      lib
      pkgs
      layatPackage
      cfg
      ;
    mkRoot =
      subdir: if subdir == null then layatLib.systemRoot else layatLib.systemRoot { inherit subdir; };
  };
in
{
  imports = [ ./common.nix ];

  # With no config, layat is not started: `apply --all` without --manifest takes the entrypoint path.
  config = lib.mkIf (cfg.enable && cfg.configs != { }) {
    # Runs after /etc and users exist. The engine commits generations with nix-env, which the
    # activation PATH lacks; dry-activate previews with --dryrun.
    system.activationScripts.layat = {
      deps = [
        "etc"
        "users"
      ];
      supportsDryActivation = true;
      text = ''
        layatDryrun=
        if [ "''${NIXOS_ACTION:-}" = dry-activate ]; then layatDryrun=--dryrun; fi
        PATH=${config.nix.package}/bin:$PATH ${kick.command} $layatDryrun
      '';
    };
  };
}
