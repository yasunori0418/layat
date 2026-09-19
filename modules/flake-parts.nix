# flake-parts module: transpose perSystem.layat.<name> to flake.layat.<system>.<name> (→ ADR-0029).
# Reuse flake-parts' official mkTransposedPerSystemModule (the same transposition mechanism as packages / legacyPackages),
# rather than hand-writing our own mkPerSystemOption + transposition (ADR-0029 remaining task 1).
#
# The option is lazyAttrsOf package. The consumer writes, in perSystem,
#   layat.<name> = inputs.layat.lib.mkManifest { inherit pkgs; root = ...; entries = { ... }; };
# The transposition target flake.layat.<system>.<name> must be a buildable derivation that the CLI invokes
# via `nix build .#layat.<system>.<name>` (it stores the result of mkManifest = a derivation・ADR decision 2).
#
# The module does pure transposition only. It does not inject mkManifest or the markers into perSystem arguments
# (ADR decision 5). The consumer references layat.lib of the same input directly.
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
      Attrset that exposes layat's named manifests (the result of `layat.lib.mkManifest` = a derivation) (→ ADR-0007, ADR-0029).

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
