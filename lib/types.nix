# entry submodule + srcType / rootType / marker custom types.
# Shared by `mkManifest` and `modules/common.nix`; depends on nixpkgs.lib only.
lib:
let
  inherit (lib) types mkOption mkDefault;
  inherit (lib.options) mergeEqualOption;
  inherit (builtins) isAttrs isString;

  # Marker discrimination (distinguished by the `_layatMarker` tag attached by `out-of-store.nix`).
  isOutOfStoreMarker = x: isAttrs x && (x._layatMarker or null) == "outOfStore";
  isRootMarker = x: isAttrs x && (x._layatMarker or null) == "root";

  # store-backed src: path / derivation / flake input (`{ outPath = …; }`). Bare strings are rejected.
  isStoreBacked =
    x:
    lib.isPath x
    || lib.isDerivation x
    || (isAttrs x && x ? outPath && (x._layatMarker or null) == null);

  # srcType = either storeBacked outOfStoreMarker.
  srcType = types.mkOptionType {
    name = "layatSrc";
    description = "store-backed source (path / derivation / flake input) or out-of-store marker";
    check = x: isStoreBacked x || isOutOfStoreMarker x;
    merge = mergeEqualOption;
  };

  # rootType = either str rootMarker. Used only by `mkManifest`; modules pin root.
  rootType = types.mkOptionType {
    name = "layatRoot";
    description = "absolute path string or root marker (projectRoot / homeRoot / systemRoot)";
    check = x: isString x || isRootMarker x;
    merge = mergeEqualOption;
  };

  # entry submodule. The attribute key is the default target; unknown keys are rejected.
  entryModule =
    { name, ... }:
    {
      options = {
        src = mkOption {
          type = srcType;
          description = "Placement source. A store-backed value (path / set) or an out-of-store marker.";
        };
        subpath = mkOption {
          type = types.str;
          default = ".";
          description = "Relative path inside src. Omitted = the whole repository (→ ADR-0008).";
        };
        target = mkOption {
          type = types.str;
          # Default = attribute key.
          default = name;
          defaultText = "attribute key";
          description = "Placement target relative to root. Defaults to the attribute key when omitted.";
        };
        method = mkOption {
          type = types.enum [
            "symlink"
            "copy"
          ];
          default = "symlink";
          description = "Placement method (formerly named mode; → ADR-0015).";
        };
      };
    };

  entriesType = types.attrsOf (types.submodule entryModule);
in
{
  inherit
    isOutOfStoreMarker
    isRootMarker
    isStoreBacked
    srcType
    rootType
    entryModule
    entriesType
    ;
}
