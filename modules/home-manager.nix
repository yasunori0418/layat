# home-manager module: builds a link-farm from layat.entries with root = homeRoot and
# passes it to `layat apply --manifest` from home.activation.
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

  # link-farm derivation (manifest.json + symlink farm); the engine resolves $HOME at runtime.
  manifest = layatLib.mkManifest {
    inherit pkgs;
    root = layatLib.homeRoot;
    entries = cfg.entries;
  };
in
{
  imports = [ ./common.nix ];

  config = lib.mkIf cfg.enable {
    # Apply the pre-built link-farm after writeBoundary; `run` honors --dry-run.
    # layat.backup.enable adds --backup=<suffix> with the user-supplied suffix shell-quoted.
    home.activation.layat = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      run ${lib.getExe layatPackage} apply --manifest ${manifest}${lib.optionalString cfg.backup.enable " --backup=${lib.escapeShellArg cfg.backup.suffix}"}
    '';
  };
}
