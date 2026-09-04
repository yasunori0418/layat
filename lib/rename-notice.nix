# Rename notice shared by the Nix-side warning layers (→ ADR-0054 §6).
#
# nput is being renamed to layat. Until the rename lands, three layers announce it:
# `mkManifest` via lib.warn and `modules/common.nix` via config.warnings (both read this
# file), and the CLI via a stderr line (cmd/nput/main.go holds its own copy of the text —
# a Go const cannot read a Nix expression). Keeping the Nix side in one file means the
# date lives in two literals in the tree rather than three.
#
# The date is a LOWER BOUND ("on or after"): the rename lands when both the notice period
# (14 days) and the prune epic have completed, whichever is later, so promising an upper
# bound would make the message a lie if the latter slips (→ ADR-0054 §6).
#
# This whole file is removed by the rename PR (→ Issue #388), which drops the three
# notice layers before performing the mechanical replacement.
let
  # The earliest date the rename may land. Confirm this is at least 14 days after the
  # merge date before merging the notice PR (→ Issue #387).
  renameDate = "2026-09-22";
in
{
  inherit renameDate;

  # A single English line (→ ADR-0054 §6: English, and never inside the --json envelope).
  # `lib.warn` and `config.warnings` both render a plain string, so no trailing newline.
  message =
    "nput will be renamed to layat on or after ${renameDate}. "
    + "The flake input URL, the `nput.*` module options, `home.activation.nput` and "
    + "`#nput.<system>.<name>` will all change, and `--json` consumers will see "
    + "`E_LAYAT_*` / `W_LAYAT_*` codes and `tool.name = \"layat\"`. "
    + "See the \"Migrating from nput\" section of "
    + "https://github.com/yasunori0418/nput#migrating-from-nput . "
    + "To stay on the old name, pin `github:yasunori0418/nput/legacy-nput`.";
}
