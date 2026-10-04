package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
)

// rollbackResultInfo / rollbackEnvInfo are rollback's outturn info slots, reserved for future run
// facts and held as nil pointers (see applyResultInfo).
type (
	rollbackResultInfo struct{}
	rollbackEnvInfo    struct{}
)

// rollbackRun is rollback's concrete run instantiation, threaded from RunE into runRollback.
type rollbackRun = outturnRun[*rollbackResultInfo, *rollbackEnvInfo]

// beginRollbackRun starts rollback's run.
func beginRollbackRun(command string) *rollbackRun {
	return beginOutturnRun[*rollbackResultInfo, *rollbackEnvInfo](command)
}

func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <name>",
		Short: "Roll layat.<name> back to the previous generation (home mode only; name required)",
		Long: "Roll the home mode profile back one generation. Because moving the profile pointer alone does not change the FS at an arbitrary root, " +
			"it re-converges the FS (treating current generation N as baseline and previous generation N-1 as target: conservatively stale-removes N∖N-1 and re-places N-1) " +
			"before moving the profile pointer. A name is required (no --all); errors out if there is no previous generation.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			run := beginRollbackRun(cmd.Name())
			return runRollback(run, args[0])
		},
	}
}

// runRollback confirms rootKind via eval pre-resolution (home mode only) and drives engine.Rollback.
func runRollback(run *rollbackRun, name string) error {
	// The config name is the run's single outturn subject.
	subject := run.beginSubject(name)
	gen, err := newGenerator()
	if err != nil {
		return err
	}
	if err := gen.Discover(flagFile); err != nil {
		return err
	}

	root, err := gen.Roots(name)
	if err != nil {
		return err
	}
	rootKind, fixedRoot := root.RootKind, root.Root
	if rootKind != manifest.RootKindHome {
		return fmt.Errorf("layat: rollback is home mode only (layat.%s has rootKind=%q; project / fixed do not expose generations)", name, rootKind)
	}

	res, err := engine.Rollback(engine.RollbackOptions{
		Name:         name,
		RootKind:     rootKind,
		FixedRoot:    fixedRoot,
		RootOverride: flagRoot,
	})
	if res != nil {
		// The From→To transition rides generation.before/after, also for a partial result.
		attachMutationPayload(subject, &res.Result, err)
	}
	if err != nil {
		return err
	}

	if flagVerbose {
		reportRollback(res, name)
	}
	return nil
}

// reportRollback prints the generation transition and placement diff to stderr.
func reportRollback(res *engine.RollbackResult, name string) {
	fmt.Fprintf(os.Stderr, "layat: rollback %s done (generation %d → %d, root=%s)\n", name, res.From, res.To, res.Root)
	for _, t := range res.Placed {
		fmt.Fprintf(os.Stderr, "  placed   %s\n", t)
	}
	for _, t := range res.Replaced {
		fmt.Fprintf(os.Stderr, "  replaced %s\n", t)
	}
	for _, t := range res.Copied {
		fmt.Fprintf(os.Stderr, "  copied   %s\n", t)
	}
	for _, t := range res.Removed {
		fmt.Fprintf(os.Stderr, "  removed  %s\n", t)
	}
	if len(res.Placed)+len(res.Replaced)+len(res.Copied)+len(res.Removed) == 0 {
		fmt.Fprintln(os.Stderr, "  no-op")
	}
}
