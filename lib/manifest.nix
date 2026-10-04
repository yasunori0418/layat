# normalizeManifest validates entries and returns pure manifest data.
# mkManifest writes that data to manifest.json and builds a symlink farm to the store src.
# Private helpers live in ./__internal.nix.
let
  internal = import ./__internal.nix;

  normalizeManifest =
    {
      lib,
      root,
      entries ? { },
    }:
    let
      t = import ./types.nix lib;
      checks = internal.pathChecks lib;

      # Validate here so both direct calls and modules go through the same checks.
      evaluated = lib.evalModules {
        modules = [
          {
            options = {
              root = lib.mkOption { type = t.rootType; };
              entries = lib.mkOption {
                type = t.entriesType;
                default = { };
              };
            };
            config = { inherit root entries; };
          }
        ];
      };
      cfg = evaluated.config;

      # root marker tag → clean enum.
      rootInfo =
        if t.isRootMarker cfg.root then
          { rootKind = cfg.root.kind; }
        else
          {
            rootKind = "fixed";
            root = cfg.root;
          };

      # entry marker tag → clean enum + resolved src string, in attrNames lexical order.
      normEntries = map (key: internal.resolveEntry lib cfg.entries.${key}) (lib.attrNames cfg.entries);

      targets = map (e: e.target) normEntries;

      # ---- Cross-field / path validation ----
      assertions = lib.concatLists [
        # systemRoot is not implemented.
        (lib.optional (
          rootInfo.rootKind == "system"
        ) "layat: root = systemRoot (system mode) is not implemented")
        # method = "copy" cannot be combined with an out-of-store marker.
        (map (
          e: "layat: method = \"copy\" cannot be combined with an out-of-store marker (target: ${e.target})"
        ) (lib.filter (e: e.method == "copy" && e.srcKind == "outOfStore") normEntries))
        # Two keys must not resolve to the same target.
        (lib.optional (
          lib.length targets != lib.length (lib.unique targets)
        ) "layat: multiple entries resolve to the same target")
        # Reject absolute paths / `..` escapes in target / subpath.
        (map (e: "layat: invalid target (absolute path or escapes root via `..`): ${e.target}") (
          lib.filter (e: checks.isUnsafe e.target) normEntries
        ))
        (map (
          e:
          "layat: invalid subpath (absolute path or escapes src via `..`): ${e.subpath} (target: ${e.target})"
        ) (lib.filter (e: checks.isUnsafe e.subpath) normEntries))
      ];

      result = {
        schemaVersion = 1;
        root = rootInfo;
        entries = normEntries;
      };
    in
    # Run every assertion through the evaluation gate with throwIf.
    lib.foldl' (acc: msg: lib.throwIf true msg acc) result assertions;

  mkManifest =
    {
      pkgs,
      root,
      entries ? { },
    }:
    let
      lib = pkgs.lib;
      norm = normalizeManifest { inherit lib root entries; };

      manifestJson = pkgs.writeText "manifest.json" (builtins.toJSON norm);

      # Farm anchors cover only store-backed entries with method = symlink.
      farmEntries = internal.farmEntries lib norm.entries;

      anchorLines = internal.anchorLines lib farmEntries;

      # Normalized targets for apply --all's cross-config conflict preflight.
      targets = map (e: e.target) norm.entries;
    in
    # Output: manifest.json + a symlink farm to the store src (GC anchors).
    pkgs.runCommandLocal "layat-manifest"
      {
        # The CLI reads rootKind / targets via `nix eval` before build.
        passthru = {
          inherit (norm.root) rootKind;
          inherit targets;
        }
        // lib.optionalAttrs (norm.root ? root) { inherit (norm.root) root; };
      }
      ''
        mkdir -p "$out"
        cp ${manifestJson} "$out/manifest.json"
        ${anchorLines}
      '';
in
{
  inherit normalizeManifest mkManifest;
}
