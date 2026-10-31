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
      identity ? null,
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

      # root marker subdir: normalized, then prepended to every target. root itself never moves.
      rawSubdir = if t.isRootMarker cfg.root then cfg.root.subdir or null else null;
      subdir = if rawSubdir == null then "" else internal.normalizeRelPath lib rawSubdir;
      subdirInvalid = rawSubdir != null && (checks.isUnsafe rawSubdir || subdir == "");

      # entry marker tag → clean enum + resolved src string, in attrNames lexical order.
      rawEntries = map (key: internal.resolveEntry lib cfg.entries.${key}) (lib.attrNames cfg.entries);
      normEntries =
        if subdir == "" then rawEntries else map (e: e // { target = "${subdir}/${e.target}"; }) rawEntries;

      targets = map (e: e.target) normEntries;

      # Path shown for a root-relative src in messages; the default subpath "." is not appended.
      rootRelativeSrcPath = e: if e.subpath == "." then e.src else "${e.src}/${e.subpath}";

      # ---- Cross-field / path validation ----
      assertions = lib.concatLists [
        (lib.optional (identity == "") "layat: identity must not be empty")
        # A root subdir must stay inside root and must not normalize to empty.
        (lib.optional subdirInvalid "layat: invalid root subdir (absolute path, escapes root via `..`, or empty): ${rawSubdir}")
        # method = "copy" cannot be combined with an out-of-store / root-relative marker.
        (map (
          e: "layat: method = \"copy\" cannot be combined with an out-of-store marker (target: ${e.target})"
        ) (lib.filter (e: e.method == "copy" && e.srcKind == "outOfStore") normEntries))
        (map (
          e: "layat: method = \"copy\" cannot be combined with a root-relative marker (target: ${e.target})"
        ) (lib.filter (e: e.method == "copy" && e.srcKind == "rootRelative") normEntries))
        # An out-of-store marker path must be absolute.
        (map (e: "layat: out-of-store path must be absolute (got: ${e.src}; target: ${e.target})") (
          lib.filter (e: e.srcKind == "outOfStore" && !(lib.hasPrefix "/" e.src)) normEntries
        ))
        # A root-relative marker path must stay inside root.
        (map (
          e:
          "layat: invalid root-relative src (absolute path or escapes root via `..`): ${e.src} (target: ${e.target})"
        ) (lib.filter (e: e.srcKind == "rootRelative" && checks.isUnsafe e.src) normEntries))
        # A root-relative src/subpath must not equal target or be its ancestor / descendant (symlink loop).
        (map
          (
            e:
            "layat: root-relative src overlaps target (equal, or one contains the other): ${rootRelativeSrcPath e} (target: ${e.target})"
          )
          (
            lib.filter (
              e: e.srcKind == "rootRelative" && internal.pathsOverlap lib "${e.src}/${e.subpath}" e.target
            ) normEntries
          )
        )
        # Two keys must not resolve to the same target.
        (lib.optional (
          lib.length targets != lib.length (lib.unique targets)
        ) "layat: multiple entries resolve to the same target")
        # Reject absolute paths / `..` escapes in target (as written, before the subdir prefix) / subpath.
        (map (e: "layat: invalid target (absolute path or escapes root via `..`): ${e.target}") (
          lib.filter (e: checks.isUnsafe e.target) rawEntries
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
      }
      // lib.optionalAttrs (identity != null) { inherit identity; };
    in
    # Run every assertion through the evaluation gate with throwIf.
    lib.foldl' (acc: msg: lib.throwIf true msg acc) result assertions;

  mkManifest =
    {
      pkgs,
      root,
      entries ? { },
      identity ? null,
    }:
    let
      lib = pkgs.lib;
      norm = normalizeManifest {
        inherit
          lib
          root
          entries
          identity
          ;
      };

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
