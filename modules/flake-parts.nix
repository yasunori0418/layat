# flake-parts module: transpose perSystem.layat.<name> to flake.layat.<system>.<name>.
# The consumer writes, in perSystem,
#   layat.<name> = inputs.layat.lib.mkManifest { inherit pkgs; root = ...; entries = { ... }; };
# and the CLI builds it via `nix build .#layat.<system>.<name>`.
{ lib, flake-parts-lib, ... }:
let
  inherit (lib)
    mkOption
    types
    ;
  inherit (flake-parts-lib)
    mkTransposedPerSystemModule
    ;
in
mkTransposedPerSystemModule {
  name = "layat";
  option = mkOption {
    type = types.lazyAttrsOf types.package;
    default = { };
    description = ''
      Attrset that exposes layat's named manifests (the result of `layat.lib.mkManifest` = a derivation).

      Declaring `perSystem.layat.<name>` automatically transposes it to the top-level `flake.layat.<system>.<name>`,
      yielding a buildable derivation the CLI invokes via `nix build .#layat.<system>.<name>`.

      ```nix
      perSystem = { pkgs, ... }: {
        layat.default = inputs.layat.lib.mkManifest {
          inherit pkgs;
          root = inputs.layat.lib.projectRoot;
          entries.".config/foo" = { src = inputs.foo; subpath = "."; };
        };
      };
      ```
    '';
  };
  file = ./flake-parts.nix;
}
