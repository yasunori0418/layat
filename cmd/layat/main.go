// Command layat is the CLI that drives the placement engine (internal/engine). The engine owns
// flock, the in-lock build, placement and commit; the CLI handles entrypoint discovery and nix eval.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/manifest"
)

// version is the layat version shown by `layat --version` and used as the envelope's tool.version.
// The nix build injects it via ldflags (-X main.version=...); a plain `go build` leaves it "dev".
var version = "dev"

// Note: cobra's Version field adds only a `--version` flag, not a `version` subcommand.

// Global flags.
var (
	flagFile        string // -f/--file: specify the entrypoint explicitly
	flagRoot        string // --root: explicitly override the resolved root
	flagNoWait      bool   // --no-wait: skip without waiting on flock contention (for shellHook)
	flagVerbose     bool   // -v/--verbose: print the placement report (summary + per-target lines); silent on success by default
	flagJSON        bool   // --json: write an outturn envelope (single JSON document) to stdout at command completion
	flagDebug       bool   // --debug: disclose the internally run nix commands on stderr
	flagRecopy      bool   // --recopy: apply modifier; unconditionally re-copy every copy target from src, overwriting
	flagYes         bool   // -y/--yes: skip the confirmation prompt of a destructive command (reset / prune; for scripts / CI)
	flagDryrun      bool   // --dryrun: apply / reset / prune modifier; show the plan with zero side effects
	flagProjectRoot bool   // --project-root: apply --all modifier; apply only projectRoot configs
	flagHomeRoot    bool   // --home-root: apply --all modifier; apply only homeRoot configs
	flagSystemRoot  bool   // --system-root: apply --all modifier; apply only systemRoot configs (future seam)
	flagManifest    string // --manifest: apply a pre-built manifest (link-farm) directly (for module activation)
	flagGenerator   string // --generator: the manifest generator to use; "" = not specified
	// flagBackup / flagBackupEnabled are --backup[=suffix]: bare --backup uses "layat-backup", a custom
	// suffix needs the "=" form. flagBackupEnabled records whether the flag was given at all.
	flagBackup        string
	flagBackupEnabled bool
)

// exitError is an error carrying a specific exit code, such as apply --dryrun's conflict (2).
// With an empty msg, main exits with the code alone.
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

// rootCmdLong is the root --help text, which discloses the internally run nix commands.
const rootCmdLong = `layat lays contents at root-relative targets, as the manifest says.
It does not generate configuration (configuration is written in Nix and evaluated by nix build).

Internal nix commands (disclosed for transparency; you can run them by hand selectively):
  init <template>   nix flake init -t <ref>#<template>
  apply <name>      nix eval <ep>#layat.<system>.<name>.rootKind --raw
                    nix build <ep>#layat.<system>.<name> --out-link <profileDir>/.pending
  apply --all       nix eval <ep>#layat.<system> --apply '<rootKind map>' --json
                    nix build <ep>#layat.<system>.<name> (per config)
  gitignore <name>  nix eval <ep>#layat.<system>.<name>.rootKind --raw
                    nix build <ep>#layat.<system>.<name> --no-link --print-out-paths
  rollback /        nix eval <ep>#layat.<system>.<name>.rootKind --raw
  list-generations

For a legacy entrypoint (shell.nix / default.nix; no per-system dimension), the
above take the -f form instead: nix eval -f <ep> layat.<name>.rootKind / nix build -f <ep> layat.<name> ...

Pass --debug to print the actual nix commands to stderr as they run.

Generator selection (the manifest generator; only nix today), first match wins:
  1. --generator <name>
  2. LAYAT_GENERATOR
  3. layat.toml in the -f directory (the file's directory for a file), else the CWD (not searched upward)
  4. $XDG_CONFIG_HOME/layat/config.toml (~/.config/layat/config.toml when unset)
  5. nix
Settings files take only the key generator = "<name>"; an unknown key or name is an input error.
apply --manifest reads none of these; prune and init ignore them.`

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "layat",
		Short: "Lays contents at root-relative targets, as the manifest says.",
		Long:  rootCmdLong,
		// `layat --version` prints the version with cobra's default template. There is no -v shorthand,
		// since -v is --verbose.
		Version: version,
		// main prints errors exactly once, so cobra's usage / error display is suppressed.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// Each subcommand's RunE begins its own outturn run, so help / completion never emit an envelope.
	pf := root.PersistentFlags()
	pf.StringVarP(&flagFile, "file", "f", "", "Specify the entrypoint explicitly (overrides autodiscovery)")
	pf.StringVar(&flagRoot, "root", "", "Override the resolved root explicitly (all modes)")
	pf.BoolVar(&flagNoWait, "no-wait", false, "Skip without waiting on flock contention (for shellHook)")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "Print the placement report (summary + per-target lines); silent on success by default")
	pf.BoolVar(&flagJSON, "json", false, "Write an outturn-conformant JSON envelope to stdout (machine-readable; orthogonal to -v)")
	pf.BoolVar(&flagDebug, "debug", false, "Disclose the internal nix commands on stderr")
	pf.StringVar(&flagGenerator, "generator", "",
		"Manifest generator (nix). Precedence: --generator > LAYAT_GENERATOR > layat.toml (-f dir, else CWD) > "+
			"$XDG_CONFIG_HOME/layat/config.toml > nix; the settings files take only generator = \"<name>\"; "+
			"ignored by prune / init, rejected with apply --manifest")
	pf.BoolVarP(&flagYes, "yes", "y", false, "Skip the confirmation prompt of a destructive command (reset / prune; for scripts / CI)")
	pf.BoolVar(&flagProjectRoot, "project-root", false, "Modifier for apply --all: apply only projectRoot configs")
	pf.BoolVar(&flagHomeRoot, "home-root", false, "Modifier for apply --all: apply only homeRoot configs")
	pf.BoolVar(&flagSystemRoot, "system-root", false, "Modifier for apply --all: apply only systemRoot configs (system mode not yet implemented)")

	root.AddCommand(newInitCmd())
	root.AddCommand(newApplyCmd())
	root.AddCommand(newResetCmd())
	root.AddCommand(newRollbackCmd())
	root.AddCommand(newListGenerationsCmd())
	root.AddCommand(newGitignoreCmd())
	root.AddCommand(newPruneCmd())
	return root
}

// exitCodeX is the interface of errors that carry an exit code (such as apply --all's aggregate).
type exitCodeX interface{ ExitCode() int }

// exitCodeError is an error that explicitly carries an exit code.
type exitCodeError struct {
	code int
	msg  string
}

func (e *exitCodeError) Error() string { return e.msg }
func (e *exitCodeError) ExitCode() int { return e.code }

func main() {
	err := newRootCmd().Execute()

	// Emit the outturn envelope once on stdout before the exit-code handling; its status mirrors
	// the exit code.
	if flagJSON && outturnReport.began() {
		if emitErr := outturnReport.emit(err); emitErr != nil {
			fmt.Fprintf(os.Stderr, "layat: cannot write the --json envelope: %v\n", emitErr)
			if err == nil {
				// The command succeeded but the envelope was not written, so do not exit 0.
				os.Exit(1)
			}
		}
	}

	if err != nil {
		// An exitError exits with its code alone; the plan is already on stdout.
		var ee *exitError
		if errors.As(err, &ee) {
			if ee.msg != "" {
				fmt.Fprintln(os.Stderr, ee.msg)
			}
			os.Exit(ee.code)
		}

		// A generator failure is formatted by printGeneratorError; anything else is printed as-is.
		var ge *generator.Error
		if errors.As(err, &ge) {
			printGeneratorError(os.Stderr, ge)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		// On a manifest schemaVersion skew between the CLI and the flake pin, add the cause and the fix.
		if errors.Is(err, manifest.ErrSchemaVersionUnsupported) {
			fmt.Fprintln(os.Stderr, "\nlayat: the layat version pinned by the CLI (engine) and by the flake may be out of sync.\n"+
				"  The flake's layat input is generating a manifest newer than the CLI.\n"+
				"  Update the CLI, or lower the flake's layat input to match the CLI so both versions align.")
		}
		// An error carrying an exit code (such as apply --all's aggregate) exits with that code.
		var ec exitCodeX
		if errors.As(err, &ec) {
			os.Exit(ec.ExitCode())
		}
		os.Exit(1)
	}
}
