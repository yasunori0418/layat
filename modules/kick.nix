# Engine kick shared by the HM / NixOS / nix-darwin modules (a plain function, not a module):
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

  # `--manifest <name>=<farm>` per config in lexicographic name order.
  manifestArgs = lib.concatMapStringsSep " " (
    name: "--manifest ${lib.escapeShellArg "${name}=${manifests.${name}}"}"
  ) (lib.attrNames manifests);
in
{
  inherit manifests;

  # --jobs is always explicit (0 included); --recopy and a shell-quoted --backup=<suffix> only when enabled.
  command = "${lib.getExe layatPackage} apply --all ${manifestArgs} --jobs ${toString cfg.jobs}${lib.optionalString cfg.recopy " --recopy"}${lib.optionalString cfg.backup.enable " --backup=${lib.escapeShellArg cfg.backup.suffix}"}";
}
