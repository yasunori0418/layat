# nix-darwin module: kicks the engine (modules/kick.nix) from postActivation with root = systemRoot
# (applied to { subdir } when the config sets one); nix-darwin only runs its fixed-name scripts.
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
    system.activationScripts.postActivation.text = ''
      ${kick.command}
    '';
  };
}
