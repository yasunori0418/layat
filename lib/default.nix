# The layat lib public API. Depends on nixpkgs.lib only.
let
  markers = import ./out-of-store.nix;
  manifest = import ./manifest.nix;
in
{
  # Private helpers exposed only for unit tests. Not a stable public API.
  __internal = import ./__internal.nix;

  # lib.mkManifest { pkgs, root, entries } -> derivation (manifest.json + symlink farm)
  inherit (manifest) mkManifest;

  # normalizeManifest { lib, root, entries } -> attrset (validated pure data)
  inherit (manifest) normalizeManifest;

  # lib.mkOutOfStoreSymlink "/abs/path" -> marker (passed to src)
  # lib.projectRoot / homeRoot / systemRoot -> marker (passed to root)
  inherit (markers)
    mkOutOfStoreSymlink
    projectRoot
    homeRoot
    systemRoot
    ;
}
