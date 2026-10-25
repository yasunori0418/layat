# The layat lib public API. Depends on nixpkgs.lib only.
let
  markers = import ./out-of-store.nix;
  manifest = import ./manifest.nix;
  applyCommand = import ./apply-command.nix;
in
{
  # Private helpers exposed only for unit tests. Not a stable public API.
  __internal = import ./__internal.nix;

  # lib.mkManifest { pkgs, root, entries } -> derivation (manifest.json + symlink farm)
  inherit (manifest) mkManifest;

  # normalizeManifest { lib, root, entries } -> attrset (validated pure data)
  inherit (manifest) normalizeManifest;

  # lib.mkApplyCommand { pkgs, layatPackage, manifests, root } -> string (`layat apply --all --manifest …`)
  # root is required: "all" (no filter) / "project" / "home" / "system" (adds --<root>-root)
  inherit (applyCommand) mkApplyCommand;

  # lib.mkOutOfStoreSymlink "/abs/path" -> marker (passed to src)
  # lib.mkRootRelativeSymlink "rel/path" -> marker (passed to src; relative to the placement root)
  # lib.projectRoot / homeRoot / systemRoot -> marker (passed to root)
  # lib.homeRoot { subdir = "rel/path"; } -> marker (prefixes every target with subdir; root stays put)
  inherit (markers)
    mkOutOfStoreSymlink
    mkRootRelativeSymlink
    projectRoot
    homeRoot
    systemRoot
    ;
}
