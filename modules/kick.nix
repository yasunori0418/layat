# Engine kick for the layat modules (a plain function, not a module; each caller pins root via mkRoot):
# builds one link-farm per layat.configs.<name> with root = mkRoot <subdir or null> and the single
# `layat apply --all` command line that places them all.
{
  lib,
  pkgs,
  layatPackage,
  cfg,
  mkRoot,
}:
let
  layatLib = import ../lib;

  # link-farm derivation (manifest.json + symlink farm) per config; the engine resolves the root at runtime.
  manifests = lib.mapAttrs (
    _: c:
    layatLib.mkManifest {
      inherit pkgs;
      root = mkRoot c.subdir;
      entries = c.entries;
    }
  ) cfg.configs;

  applyCommand = layatLib.mkApplyCommand {
    inherit pkgs layatPackage manifests;
    root = "all";
  };
in
{
  inherit manifests;

  # Callers run it only when cfg.configs is non-empty (with none it would take the entrypoint path).
  # --jobs is always explicit (0 included); --recopy and a shell-quoted --backup=<suffix> only when enabled.
  command = "${applyCommand} --jobs ${toString cfg.jobs}${lib.optionalString cfg.recopy " --recopy"}${lib.optionalString cfg.backup.enable " --backup=${lib.escapeShellArg cfg.backup.suffix}"}";
}
