# mkApplyCommand returns one `layat apply --all --manifest <name>=<farm> …` command line that
# places prebuilt link-farms without eval / build / entrypoint lookup. Runtime flags
# (--jobs / --recopy / --backup / --no-wait) are left for the caller to append.
let
  rootFlags = {
    all = "";
    project = " --project-root";
    home = " --home-root";
    system = " --system-root";
  };

  mkApplyCommand =
    {
      pkgs,
      layatPackage,
      manifests,
      root ? null,
    }:
    let
      lib = pkgs.lib;
      validRoots = lib.concatMapStringsSep ", " (r: ''"${r}"'') (lib.attrNames rootFlags);
      rootFlag =
        if root == null then
          throw "layat.lib.mkApplyCommand: root is required (one of ${validRoots})"
        else if !(builtins.isString root && rootFlags ? ${root}) then
          throw "layat.lib.mkApplyCommand: invalid root ${
            lib.generators.toPretty { } root
          } (expected one of ${validRoots})"
        else
          rootFlags.${root};

      # `--manifest <name>=<farm>` per config in lexicographic name order; selection is the CLI filter's job.
      manifestArgs = lib.concatMapStringsSep " " (
        name: "--manifest ${lib.escapeShellArg "${name}=${manifests.${name}}"}"
      ) (lib.attrNames manifests);
    in
    "${lib.getExe layatPackage} apply --all${rootFlag} ${manifestArgs}";
in
{
  inherit mkApplyCommand;
}
