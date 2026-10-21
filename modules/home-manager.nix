# home-manager module: builds one link-farm per layat.configs.<name> with root = homeRoot (applied to
# { subdir } when the config sets one) and passes them all to a single `layat apply --all` from home.activation.
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

  # link-farm derivation (manifest.json + symlink farm) per config; the engine resolves $HOME at runtime.
  manifests = lib.mapAttrs (
    _: c:
    layatLib.mkManifest {
      inherit pkgs;
      root = if c.subdir == null then layatLib.homeRoot else layatLib.homeRoot { subdir = c.subdir; };
      entries = c.entries;
    }
  ) cfg.configs;

  # `--manifest <name>=<farm>` per config in lexicographic name order.
  manifestArgs = lib.concatMapStringsSep " " (
    name: "--manifest ${lib.escapeShellArg "${name}=${manifests.${name}}"}"
  ) (lib.attrNames manifests);
in
{
  imports = [ ./common.nix ];

  # With no config, layat is not started: `apply --all` without --manifest takes the entrypoint path.
  config = lib.mkIf (cfg.enable && cfg.configs != { }) {
    # Apply the pre-built link-farms after writeBoundary so it does not collide with home.file's
    # placement; `run` honors --dry-run.
    # --jobs is always explicit (0 included); --recopy and a shell-quoted --backup=<suffix> only when enabled.
    home.activation.layat = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      run ${lib.getExe layatPackage} apply --all ${manifestArgs} --jobs ${toString cfg.jobs}${lib.optionalString cfg.recopy " --recopy"}${lib.optionalString cfg.backup.enable " --backup=${lib.escapeShellArg cfg.backup.suffix}"}
    '';
  };
}
