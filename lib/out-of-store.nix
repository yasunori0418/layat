# Marker constructors. A marker carries what the engine resolves at runtime.
# The `_layatMarker` tag lets `lib/types.nix` tell markers from store-backed attrsets;
# it never reaches `manifest.json`.
{
  # Out-of-store symlink to an absolute local path; the engine creates the link.
  mkOutOfStoreSymlink = path: {
    _layatMarker = "outOfStore";
    inherit path;
  };

  # root markers. The engine resolves the concrete path at runtime.
  projectRoot = {
    _layatMarker = "root";
    kind = "project";
  };
  homeRoot = {
    _layatMarker = "root";
    kind = "home";
  };
  systemRoot = {
    _layatMarker = "root";
    kind = "system";
  };
}
