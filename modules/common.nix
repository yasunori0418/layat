# layat options shared by the HM / NixOS / nix-darwin modules; each module pins root itself.
# The entry type is shared with lib/types.nix, so unknown keys are eval errors.
# The HM module has a single profile (`default`).
{ config, lib, ... }:
let
  layatTypes = import ../lib/types.nix lib;
in
{
  options.layat = {
    enable = lib.mkEnableOption "layat (symlink / copy placement of fetched git repositories)";

    entries = lib.mkOption {
      type = layatTypes.entriesType;
      default = { };
      example = lib.literalExpression ''
        {
          # attribute key = root-relative target (identifier; → ADR-0014)
          ".claude/skills/nix" = {
            src = inputs.claude-skills;
            subpath = "skills/nix";
          };
        }
      '';
      description = ''
        Attrset of placement definitions. The attribute key = the placement target
        (identifier), and each value is an entry submodule (src / subpath / target /
        method). The type is shared with lib/types.nix, and unknown keys are an eval
        error (→ docs/spec.md "entries schema").
      '';
    };

    backup = lib.mkOption {
      type = lib.types.submodule {
        options = {
          enable = lib.mkEnableOption ''
            back up an occupying foreign entity to "<target>.<suffix>" before placing it,
            instead of stopping on conflict (wires activation's apply --backup; → ADR-0045)
          '';
          suffix = lib.mkOption {
            type = lib.types.str;
            default = "layat-backup";
            description = ''
              The backup rename suffix (activation wires apply --backup=<suffix>). The
              backup destination becomes "<target>.<suffix>" (→ ADR-0045).
            '';
          };
        };
      };
      default = { };
      description = ''
        apply --backup wiring: renames an occupying foreign entity aside instead of
        conflicting (→ ADR-0045). This is a placement modifier orthogonal to `entries`
        and does not touch the manifest v1 contract (lib/types.nix) — activation only
        adds `--backup=<suffix>` to the `layat apply --manifest` invocation when enabled.
      '';
    };
  };
}
