# Marker constructors. A marker carries what the engine resolves at runtime.
# The `_layatMarker` tag lets `lib/types.nix` tell markers from store-backed attrsets;
# it never reaches `manifest.json`.
let
  mkRootMarker = kind: {
    _layatMarker = "root";
    inherit kind;
    __functor = self: { subdir }: removeAttrs self [ "__functor" ] // { inherit subdir; };
  };
in
{
  # Out-of-store symlink to an absolute local path; the engine creates the link.
  mkOutOfStoreSymlink = path: {
    _layatMarker = "outOfStore";
    inherit path;
  };

  # Out-of-store symlink to a path relative to the placement root.
  mkRootRelativeSymlink = path: {
    _layatMarker = "rootRelative";
    inherit path;
  };

  # root markers. The engine resolves the concrete path at runtime.
  # Applying one (`homeRoot { subdir = "rel"; }`) returns a non-applicable marker carrying subdir.
  projectRoot = mkRootMarker "project";
  homeRoot = mkRootMarker "home";
  systemRoot = mkRootMarker "system";
}
