# Private helpers of manifest.nix, exposed as `layat.__internal.<name>` for unit tests.
# Not a public API. Each helper takes nixpkgs.lib explicitly.
let
  # Whether following `..` makes the depth go negative (escapes outside base).
  escapesBase =
    lib: p:
    let
      comps = lib.filter (c: c != "" && c != ".") (lib.splitString "/" p);
      step =
        acc: c:
        if acc.bad then
          acc
        else if c == ".." then
          {
            bad = acc.depth == 0;
            depth = if acc.depth == 0 then 0 else acc.depth - 1;
          }
        else
          {
            bad = false;
            depth = acc.depth + 1;
          };
    in
    (lib.foldl' step {
      bad = false;
      depth = 0;
    } comps).bad;

  # Absolute paths and paths that escape outward via `..` are unsafe.
  pathChecks = lib: {
    isUnsafe = p: lib.hasPrefix "/" p || escapesBase lib p;
  };

  # Path components with `.` dropped and `..` folded into its parent (an excess `..` is dropped).
  relComponents =
    lib: p:
    lib.foldl' (
      acc: c: if c == ".." then lib.take (lib.max 0 (lib.length acc - 1)) acc else acc ++ [ c ]
    ) [ ] (lib.filter (c: c != "" && c != ".") (lib.splitString "/" p));

  # Relative path rejoined from relComponents ("./a/" -> "a"); "" when nothing remains.
  normalizeRelPath = lib: p: lib.concatStringsSep "/" (relComponents lib p);

  # Whether two relative paths are equal or one is an ancestor of the other, per component.
  pathsOverlap =
    lib: a: b:
    let
      ca = relComponents lib a;
      cb = relComponents lib b;
      isPrefix = x: y: lib.take (lib.length x) y == x;
    in
    isPrefix ca cb || isPrefix cb ca;

  # GC anchor name for the symlink farm: sha256 short hex of target.
  anchorName = lib: target: lib.substring 0 32 (builtins.hashString "sha256" target);

  # entry marker tag → clean enum + resolved src string.
  resolveEntry =
    lib: e:
    let
      t = import ./types.nix lib;
      srcInfo =
        if t.isOutOfStoreMarker e.src then
          {
            srcKind = "outOfStore";
            src = e.src.path;
          }
        else if t.isRootRelativeMarker e.src then
          {
            srcKind = "rootRelative";
            src = e.src.path;
          }
        else
          {
            srcKind = "store";
            src = toString e.src;
          };
    in
    {
      inherit (srcInfo) srcKind src;
      inherit (e) subpath target method;
    };

  # Farm anchors cover only store-backed entries with method = symlink.
  farmEntries = lib: entries: lib.filter (e: e.srcKind == "store" && e.method == "symlink") entries;

  # Shell lines that place one GC anchor per farm entry.
  anchorLines =
    lib: entries:
    lib.concatMapStringsSep "\n" (
      e: "ln -s ${lib.escapeShellArg e.src} \"$out/${anchorName lib e.target}\""
    ) entries;
in
{
  inherit
    escapesBase
    pathChecks
    normalizeRelPath
    pathsOverlap
    anchorName
    resolveEntry
    farmEntries
    anchorLines
    ;
}
