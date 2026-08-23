package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/nput/internal/engine"
)

// pruneInfo is prune's envelope-wide info: the series it deleted (in --dryrun, would delete) and
// the ones it left alone with the reason (→ issue #134, niface ADR-0018; the init shape). prune
// names no config, so it registers no subject and its result.info slot stays an anonymous
// *struct{} left nil — the inventory has nowhere else to ride.
//
// The deleted series are removed, never pruned: Result.Pruned already means the empty ancestor
// directories rmdir-ed after a removal (→ REQ-8409db86-a1ba-4053-86dc-588985cc1ca7), and the two
// meanings must not collide in one document.
//
// Carried as a pointer: the --json-without---yes refusal and the state-dir resolution both fail
// before any scan exists, and only a nil pointer keeps the envelope's info omitted there. A value
// struct would emit "info":{"removed":null,"skipped":null} on those paths, which reads as a run
// that looked and found nothing (→ issue #196 §4).
type pruneInfo struct {
	Removed []pruneSeriesRow  `json:"removed"`
	Skipped []pruneSkippedRow `json:"skipped"`
}

// pruneSeriesRow is one <roothash> series of the inventory. The fields are the ones --dryrun
// prints and the confirmation prompt lists (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed), plus
// dir: the <roothash> is a hash of the root path, so the same name can stand under both scan
// bases and dir is the only field that says which one this series came from.
type pruneSeriesRow struct {
	RootHash string   `json:"roothash"`
	Root     string   `json:"root"`
	Names    []string `json:"names"`
	Dir      string   `json:"dir"`
}

// pruneSkippedRow is one series left alone. reason is the engine's own vocabulary verbatim
// (locked / backref-unreadable / series-unreadable / root-stat-failed / permission-denied ·
// → DSG-096dc893-21f4-45e3-9347-986e9275b4d1): a CLI-side re-spelling would drift from the
// engine's. detail carries the underlying failure, which the reason alone does not name — it is
// always present (the engine records a skip only with the cause that produced it), so it is not
// omitempty: a consumer never has to handle its absence.
type pruneSkippedRow struct {
	Series pruneSeriesRow `json:"series"`
	Reason string         `json:"reason"`
	Detail string         `json:"detail"`
}

// pruneRun is prune's concrete run instantiation, threaded from RunE into runPrune.
type pruneRun = nifaceRun[*struct{}, *pruneInfo]

// beginPruneRun starts prune's run (→ beginNifaceRun, beginApplyRun).
func beginPruneRun(command string) *pruneRun {
	return beginNifaceRun[*struct{}, *pruneInfo](command)
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
			"removable disk or a network mount — being taken for a deleted one while it is unmounted (see ADR-0034).\n" +
			"--dryrun shows the same series with zero side effects and exits (no confirm / flock).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			run := beginPruneRun(cmd.Name())
			return runPrune(run, flagDryrun, isInteractive())
		},
	}
	cmd.Flags().BoolVar(&flagDryrun, "dryrun", false,
		"Show the series that would be deleted with zero side effects and exit (no confirm / flock; see ADR-0034 §2)")
	return cmd
}

// runPrune scans the profile bases and drives engine.Prune. --dryrun prints the plan read-only to
// stdout and exits 0. Non-dryrun requires TTY confirmation / --yes, and the prompt lists the root
// path of every series first (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
//
// prune registers no subject: it names no config, so its inventory rides in the envelope-wide
// info and results stays [] (the init shape · niface ADR-0018, → issue #164).
//
// dryrun must be flagDryrun's value (RunE passes exactly that, as reset's does): the envelope's
// own dryRun field is captured from the flag by nifaceRun.begin, so passing anything else here
// would emit a document whose dryRun disagrees with what the run did.
//
// interactive carries no such constraint — it is the seam the TTY check comes in through, so the
// policy stays exercisable without a terminal, and a caller passing something other than
// isInteractive() simply gets that policy (→ confirmPolicy, which reset feeds the same way).
func runPrune(run *pruneRun, dryrun, interactive bool) error {
	// --dryrun: a side-effect-free preview (no flock / confirm). It stays available under --json
	// without --yes — the refusal below guards the deletion, and a preview deletes nothing.
	if dryrun {
		res, err := pruneFn(pruneOptions(true, nil))
		if err != nil {
			return err
		}
		run.setEnvelopeInfo(pruneInfoFrom(res))
		printPrunePlan(res)
		return nil
	}

	// Non-dryrun is a destructive operation. Decide the confirmation policy (skip / prompt /
	// refuse) from --yes and TTY state, before anything is scanned: under --json the refusal is
	// the fail-fast path, and its envelope must not carry an inventory that reads as a completed
	// scan (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed, ADR-0043 §8).
	needPrompt, err := confirmPolicy(flagYes, prunePromptAllowed(interactive, flagJSON), "prune")
	if err != nil {
		return err
	}

	// Pass confirm only when a prompt is needed (--yes skips the listing along with the prompt ·
	// → ADR-0034 §2).
	var confirm func(*engine.PruneResult) (bool, error)
	if needPrompt {
		confirm = prunePrompt
	}

	res, err := pruneFn(pruneOptions(false, confirm))
	if res != nil && !res.Aborted {
		// Also on a mid-deletion failure: the partial result keeps the series completed before it,
		// so the envelope says what is already gone rather than dropping it (→ engine.Prune の契約,
		// REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
		//
		// An aborted run is the exception: nothing was deleted, so nothing may be reported as
		// deleted. Removed is empty in the engine as it stands, but the CLI does not lean on that
		// — an inventory built from a declined run's result is a report of a deletion that did
		// not happen (→ resetPayload, which drops its changes on Aborted for the same reason).
		//
		// The skipped series go with it rather than being kept the way resetPayload keeps its
		// items: a declined run judged them but acted on nothing, and the engine has already put
		// every skip on stderr as a warning, so nothing is lost by leaving the document empty.
		run.setEnvelopeInfo(pruneInfoFrom(res))
	}
	if err != nil {
		return err
	}
	if res.Aborted {
		// Unreachable under --json today: that path requires --yes (above), which leaves Confirm
		// nil so nothing can decline. Were the --yes requirement ever relaxed, the two outcomes
		// would already be distinguishable in the document — a declined run carries no info at
		// all (above), while a run that found no orphan carries an empty removed.
		fmt.Fprintln(os.Stderr, "nput: prune aborted")
		return nil
	}
	if flagVerbose {
		reportPruneResult(res)
	}
	return nil
}

// prunePrompt is the confirmation callback handed to the engine. What it lists is the engine's
// preview — a value distinct from the result Prune returns, so the listing the user just read
// cannot be emptied by the deletion stage (→ engine.PruneOptions.Confirm).
func prunePrompt(preview *engine.PruneResult) (bool, error) {
	reportPruneTargets(preview)
	return promptYesNo("This will delete the above profile series. Continue?")
}

// pruneFn is the engine entry point runPrune drives, indirected so the CLI's own decisions (which
// preview reaches the prompt, what an aborted run reports, what the envelope ends up carrying)
// are observable without a state dir on disk. Production never reassigns it; the TTY side needs
// no seam because runPrune takes interactivity as an argument.
var pruneFn = engine.Prune

// pruneOptions builds the engine options for one prune run. The scan bases are left at their
// defaults (the user state base and /nix/var/nix/profiles/nput · → ADR-0036 §3): only the engine's
// own tests point them elsewhere. Warnf is left nil so the engine's own default (stderr, one line
// per warning) is used — the same thing runReset does, and duplicating the formatting here would
// give the same knowledge two places to drift apart.
func pruneOptions(dryrun bool, confirm func(*engine.PruneResult) (bool, error)) engine.PruneOptions {
	return engine.PruneOptions{
		DryRun:  dryrun,
		Confirm: confirm,
	}
}

// prunePromptAllowed reports whether prune may prompt interactively: it requires a TTY, and --json
// unconditionally forbids it — machine consumption never prompts, so without --yes the
// confirmPolicy refuse path fails fast exactly like the non-interactive case (→ ADR-0043 §8, the
// same shape as resetPromptAllowed).
func prunePromptAllowed(interactive, jsonMode bool) bool {
	return interactive && !jsonMode
}

// printPrunePlan prints prune --dryrun's series to stdout (it owns the machine-readable output;
// one line per series, tab-separated as elsewhere · → docs/spec.md stream discipline, ADR-0023).
// It is not suppressed under silent-on-success (the stdout-ownership principle; → ADR-0031).
// Under --json the stdout lines are suppressed at this single chokepoint (the envelope owns
// stdout); the stderr nothing-to-delete notice stays — human diagnostics coexist (→ ADR-0043 §2).
//
// The <name> profiles of a series ride in the last field, comma-separated: a series holding
// nothing but the backref has none, and the field is then empty rather than absent, so the line
// keeps its arity.
func printPrunePlan(res *engine.PruneResult) {
	if !flagJSON {
		for _, s := range res.Removed {
			fmt.Printf("remove-series\t%s\t%s\t%s\n", s.RootHash, s.Root, strings.Join(s.Names, ","))
		}
	}
	if len(res.Removed) == 0 {
		fmt.Fprintln(os.Stderr, "nput: prune --dryrun: nothing to delete")
	}
}

// reportPruneTargets prints the series about to be deleted to stderr before the confirmation
// prompt (treated as progress; stdout is reserved for machine-readable output). Every root path
// appears: it is the only guard against an out-of-store root being taken for a deleted one while
// it is unmounted (→ ADR-0034 §2, REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
func reportPruneTargets(res *engine.PruneResult) {
	fmt.Fprintf(os.Stderr, "nput: prune will delete %d orphan profile series:\n", len(res.Removed))
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

// reportPruneResult prints what was actually deleted to stderr (stdout is reserved for
// machine-readable output; → ADR-0023). The skipped series are already on stderr as warnings from
// the engine, so they are not repeated here.
func reportPruneResult(res *engine.PruneResult) {
	fmt.Fprintf(os.Stderr, "nput: prune done (%d series deleted, %d skipped)\n", len(res.Removed), len(res.Skipped))
	for _, s := range res.Removed {
		fmt.Fprintf(os.Stderr, "  removed-series %s (root=%s)\n", s.Dir, s.Root)
	}
	if len(res.Removed) == 0 {
		fmt.Fprintln(os.Stderr, "  no-op")
	}
}

// pruneInfoFrom maps the engine result onto the --json inventory. It is pure data mapping — the
// same result the -v report reads, so the document and the human report cannot disagree
// (→ niface_payload.go の単一結果源). Both arrays stay non-nil so a run that deleted nothing still
// emits "removed": [] rather than null.
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

// pruneSeriesRowFrom converts one series. names stays a non-nil array even for a series holding
// nothing but the backref (engine leaves it nil there): a consumer distinguishing null from []
// would be distinguishing nothing — both mean the series has no <name> profile.
func pruneSeriesRowFrom(s engine.PruneSeries) pruneSeriesRow {
	names := s.Names
	if names == nil {
		names = []string{}
	}
	return pruneSeriesRow{RootHash: s.RootHash, Root: s.Root, Names: names, Dir: s.Dir}
}
