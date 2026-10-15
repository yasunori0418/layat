# layat options shared by the HM / NixOS / nix-darwin modules; each module pins root itself.
# The entry type is shared with lib/types.nix, so unknown keys are eval errors.
# Each layat.configs.<name> is one profile; layat.entries is a deprecated alias of configs.default.entries.
{ config, lib, ... }:
let
  layatTypes = import ../lib/types.nix lib;
in
{
  imports = [
    (lib.mkRenamedOptionModule [ "layat" "entries" ] [ "layat" "configs" "default" "entries" ])
  ];

  options.layat = {
    enable = lib.mkEnableOption "layat (symlink / copy placement of fetched git repositories)";

    configs = lib.mkOption {
      type = lib.types.attrsOf (
        lib.types.submodule {
          options.entries = lib.mkOption {
            type = layatTypes.entriesType;
            default = { };
            example = lib.literalExpression ''
              {
                # attribute key = root-relative target (identifier)
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
        }
      );
      default = { };
      description = ''
        Named configs. Each `<name>` gets its own link-farm and its own profile `<name>`,
        so configs can be rolled back independently. `layat.entries` is a deprecated alias
        of `layat.configs.default.entries`.
      '';
    };

    backup = lib.mkOption {
      type = lib.types.submodule {
        options = {
          enable = lib.mkEnableOption ''
            back up an occupying foreign entity to "<target>.<suffix>" before placing it,
            instead of stopping on conflict (wires activation's apply --backup)
          '';
          suffix = lib.mkOption {
            type = lib.types.str;
            default = "layat-backup";
            description = ''
              The backup rename suffix (activation wires apply --backup=<suffix>). The
              backup destination becomes "<target>.<suffix>".
            '';
          };
        };
      };
      default = { };
      description = ''
        apply --backup wiring: renames an occupying foreign entity aside instead of
        conflicting. This is a placement modifier orthogonal to `entries`
        and does not touch the manifest v1 contract (lib/types.nix) — activation only
        adds `--backup=<suffix>` to the `layat apply --all` invocation when enabled. It
        applies to every config.
      '';
    };
  };
}
