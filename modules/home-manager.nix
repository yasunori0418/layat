# home-manager module: kicks the engine (modules/kick.nix) from home.activation with root = homeRoot
# (applied to { subdir } when the config sets one).
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
      subdir: if subdir == null then layatLib.homeRoot else layatLib.homeRoot { inherit subdir; };
  };
in
{
  imports = [ ./common.nix ];

  # With no config, layat is not started: `apply --all` without --manifest takes the entrypoint path.
  config = lib.mkIf (cfg.enable && cfg.configs != { }) {
    # Apply the pre-built link-farms after writeBoundary so it does not collide with home.file's
    # placement; `run` honors --dry-run.
    home.activation.layat = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      run ${kick.command}
    '';
  };
}
