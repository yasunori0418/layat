package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
)

// pruneInfo is prune's envelope-wide info: the series it deleted (in --dryrun, would delete) and
// the ones it skipped with the reason. It is a pointer so failures before the scan omit info.
// "removed" avoids "pruned", which already means the rmdir-ed empty ancestor directories.
type pruneInfo struct {
	Removed []pruneSeriesRow  `json:"removed"`
	Skipped []pruneSkippedRow `json:"skipped"`
}

// pruneSeriesRow is one <roothash> series of the inventory. dir tells which scan base the series
// came from, since the same <roothash> can stand under both.
type pruneSeriesRow struct {
	RootHash string   `json:"roothash"`
	Root     string   `json:"root"`
	Names    []string `json:"names"`
	Dir      string   `json:"dir"`
}

// pruneSkippedRow is one series left alone. reason is the engine's vocabulary verbatim, and detail
// always carries the underlying failure.
type pruneSkippedRow struct {
	Series pruneSeriesRow `json:"series"`
	Reason string         `json:"reason"`
	Detail string         `json:"detail"`
}

// pruneRun is prune's concrete run instantiation, threaded from RunE into runPrune.
type pruneRun = outturnRun[*struct{}, *pruneInfo]

// beginPruneRun starts prune's run.
func beginPruneRun(command string) *pruneRun {
	return beginOutturnRun[*struct{}, *pruneInfo](command)
}

func newPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete the orphan profile series whose recorded root no longer exists (no name; scan only)",
		Long: "Scan the user state base and the system profile base for <roothash> series whose backref .root " +
			"records a root path that no longer exists, and delete those series whole. It takes no name, " +
			"discovers no entrypoint, and runs no nix eval / build — it only looks at the profile state on disk.\n\n" +
			"A series whose root does exist is left alone whatever else it looks like, and so is one whose verdict " +
			"cannot be reached (an unusable backref, an unlistable series, a stat failing for any reason but " +
			"non-existence) — each with a warning. Placed artifacts are never touched and the generations of a " +
			"kept series are never thinned; free the store with nix-collect-garbage as before.\n\n" +
			"Before deleting, the root paths of every series are listed and confirmed (--yes skips the prompt; a " +
			"non-TTY without --yes aborts). That listing is the only guard against an out-of-store root — one on a " +
			"removable disk or a network mount — being taken for a deleted one while it is unmounted.\n" +
			"--dryrun shows the same series with zero side effects and exits (no confirm / flock).\n\n" +
			"Alongside the user state base a system base is scanned. " + systemBaseEnv +
			" points that one elsewhere (for tests / isolated harnesses, not a way to target a base).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			run := beginPruneRun(cmd.Name())
			return runPrune(run, flagDryrun, isInteractive())
		},
	}
	cmd.Flags().BoolVar(&flagDryrun, "dryrun", false,
		"Show the series that would be deleted with zero side effects and exit (no confirm / flock)")
	return cmd
}

// runPrune scans the profile bases and drives engine.Prune. --dryrun prints the plan to stdout;
// otherwise deletion requires TTY confirmation or --yes. dryrun must be flagDryrun's value, and
// interactive is the TTY check passed in by the caller.
func runPrune(run *pruneRun, dryrun, interactive bool) error {
	// --dryrun previews without side effects (no flock / confirm), also under --json without --yes.
	if dryrun {
		res, err := pruneFn(pruneOptions(true, nil))
		if err != nil {
			return err
		}
		run.setEnvelopeInfo(pruneInfoFrom(res))
		printPrunePlan(res)
		return nil
	}

	// Decide the confirmation policy (skip / prompt / refuse) from --yes and TTY state before any scan.
	needPrompt, err := confirmPolicy(flagYes, prunePromptAllowed(interactive, flagJSON), "prune")
	if err != nil {
		return err
	}

	// Pass confirm only when a prompt is needed.
	var confirm func(*engine.PruneResult) (bool, error)
	if needPrompt {
		confirm = prunePrompt
	}

	res, err := pruneFn(pruneOptions(false, confirm))
	if res != nil && !res.Aborted {
		// A partial result on failure still records the series already deleted. An aborted run deleted
		// nothing, so it reports no inventory.
		run.setEnvelopeInfo(pruneInfoFrom(res))
	}
	if err != nil {
		return err
	}
	if res.Aborted {
		// Unreachable under --json, which requires --yes and so cannot decline.
		fmt.Fprintln(os.Stderr, "layat: prune aborted")
		return nil
	}
	if flagVerbose {
		reportPruneResult(res)
	}
	return nil
}

// prunePrompt is the confirmation callback handed to the engine; it lists the engine's preview.
func prunePrompt(preview *engine.PruneResult) (bool, error) {
	reportPruneTargets(preview)
	return promptYesNo("This will delete the above profile series. Continue?")
}

// pruneFn is the engine entry point runPrune drives, replaceable in tests.
var pruneFn = engine.Prune

// systemBaseEnv overrides the system scan base (/nix/var/nix/profiles/layat) for one run, so tests
// can point it at a throwaway directory. It is not a supported way to prune another base.
const systemBaseEnv = "LAYAT_SYSTEM_PROFILE_BASE"

// pruneOptions builds the engine options for one prune run. StateDir and Warnf keep the engine's
// defaults; SystemDir comes verbatim from systemBaseEnv when set.
func pruneOptions(dryrun bool, confirm func(*engine.PruneResult) (bool, error)) engine.PruneOptions {
	return engine.PruneOptions{
		DryRun:    dryrun,
		Confirm:   confirm,
		SystemDir: os.Getenv(systemBaseEnv),
	}
}

// prunePromptAllowed reports whether prune may prompt: it requires a TTY, and --json forbids it.
func prunePromptAllowed(interactive, jsonMode bool) bool {
	return interactive && !jsonMode
}

// printPrunePlan prints prune --dryrun's series to stdout, one tab-separated line per series with
// comma-separated profile names last (empty when none). --json suppresses the stdout lines; the
// stderr nothing-to-delete notice stays.
func printPrunePlan(res *engine.PruneResult) {
	if !flagJSON {
		for _, s := range res.Removed {
			fmt.Printf("remove-series\t%s\t%s\t%s\n", s.RootHash, s.Root, strings.Join(s.Names, ","))
		}
	}
	if len(res.Removed) == 0 {
		fmt.Fprintln(os.Stderr, "layat: prune --dryrun: nothing to delete")
	}
}

// reportPruneTargets lists the series about to be deleted, with every root path, on stderr before
// the confirmation prompt. The root paths keep an unmounted root from passing for a deleted one.
func reportPruneTargets(res *engine.PruneResult) {
	fmt.Fprintf(os.Stderr, "layat: prune will delete %d orphan profile series:\n", len(res.Removed))
	for _, s := range res.Removed {
		fmt.Fprintf(os.Stderr, "  root %s\n", s.Root)
		fmt.Fprintf(os.Stderr, "    series %s\n", s.Dir)
		if len(s.Names) > 0 {
			fmt.Fprintf(os.Stderr, "    profiles %s\n", strings.Join(s.Names, " "))
		}
	}
	if len(res.Removed) == 0 {
		fmt.Fprintln(os.Stderr, "  (nothing to delete)")
	}
}

// reportPruneResult prints what was deleted to stderr. Skipped series are already engine warnings.
func reportPruneResult(res *engine.PruneResult) {
	fmt.Fprintf(os.Stderr, "layat: prune done (%d series deleted, %d skipped)\n", len(res.Removed), len(res.Skipped))
	for _, s := range res.Removed {
		fmt.Fprintf(os.Stderr, "  removed-series %s (root=%s)\n", s.Dir, s.Root)
	}
	if len(res.Removed) == 0 {
		fmt.Fprintln(os.Stderr, "  no-op")
	}
}

// pruneInfoFrom maps the engine result onto the --json inventory. Both arrays stay non-nil, so an
// empty run emits "removed": [].
func pruneInfoFrom(res *engine.PruneResult) *pruneInfo {
	info := &pruneInfo{
		Removed: make([]pruneSeriesRow, 0, len(res.Removed)),
		Skipped: make([]pruneSkippedRow, 0, len(res.Skipped)),
	}
	for _, s := range res.Removed {
		info.Removed = append(info.Removed, pruneSeriesRowFrom(s))
	}
	for _, sk := range res.Skipped {
		info.Skipped = append(info.Skipped, pruneSkippedRow{
			Series: pruneSeriesRowFrom(sk.Series),
			Reason: string(sk.Reason),
			Detail: sk.Detail,
		})
	}
	return info
}

// pruneSeriesRowFrom converts one series; names stays a non-nil array.
func pruneSeriesRowFrom(s engine.PruneSeries) pruneSeriesRow {
	names := s.Names
	if names == nil {
		names = []string{}
	}
	return pruneSeriesRow{RootHash: s.RootHash, Root: s.Root, Names: names, Dir: s.Dir}
}
