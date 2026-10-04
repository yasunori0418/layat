package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
)

// resetResultInfo / resetEnvInfo are reset's outturn info slots, reserved for future run facts and
// held as nil pointers (see applyResultInfo).
type (
	resetResultInfo struct{}
	resetEnvInfo    struct{}
)

// resetRun is reset's concrete run instantiation, threaded from RunE into runReset.
type resetRun = outturnRun[*resetResultInfo, *resetEnvInfo]

// beginResetRun starts reset's run.
func beginResetRun(command string) *resetRun {
	return beginOutturnRun[*resetResultInfo, *resetEnvInfo](command)
}

func newResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset <name> [target...]",
		Short: "Tear placements back down to nothing (FS-only teardown; name required; no --all)",
		Long: "Teardown that returns layat.<name>'s placements to nothing. Omitting target tears down every entry; specifying targets tears down only those entries. " +
			"Symlinks are removed under the conservative invariant (only layat-managed, only as recorded) and foreign ones are kept. copy targets are deleted (confirmed due to the data-loss risk). " +
			"It does not touch the profile or generations (FS-only). A name is required (no --all). " +
			"--dryrun shows the removal targets with zero side effects and exits (no confirm / flock).",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			run := beginResetRun(cmd.Name())
			return runReset(run, args[0], args[1:], flagDryrun)
		},
	}
	cmd.Flags().BoolVar(&flagDryrun, "dryrun", false,
		"Show the removal targets (symlink / copy target) with zero side effects and exit (no confirm / flock)")
	return cmd
}

// runReset resolves rootKind (→ profileDir) via eval pre-resolution and drives engine.Reset.
// --dryrun prints the plan read-only to stdout and exits 0. Non-dryrun requires TTY confirmation / --yes.
func runReset(run *resetRun, name string, targets []string, dryrun bool) error {
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

	// --dryrun previews without side effects (no flock / confirm) and exits 0.
	if dryrun {
		res, err := engine.Reset(engine.ResetOptions{
			Name:         name,
			RootKind:     rootKind,
			FixedRoot:    fixedRoot,
			RootOverride: flagRoot,
			Targets:      targets,
			DryRun:       true,
		})
		if err != nil {
			return err
		}
		printResetPlan(res)
		return nil
	}

	// Non-dryrun is a destructive operation. Decide the confirmation policy (skip / prompt / refuse) from --yes and TTY state.
	needPrompt, err := confirmPolicy(flagYes, resetPromptAllowed(isInteractive(), flagJSON), "reset")
	if err != nil {
		return err
	}

	// Pass confirm only when a prompt is needed. The prompt shows the computed plan.
	var confirm func(*engine.ResetResult) (bool, error)
	if needPrompt {
		confirm = func(res *engine.ResetResult) (bool, error) {
			reportResetTargets(res, name)
			return promptYesNo("This will remove the above. Continue?")
		}
	}

	res, err := engine.Reset(engine.ResetOptions{
		Name:         name,
		RootKind:     rootKind,
		FixedRoot:    fixedRoot,
		RootOverride: flagRoot,
		Targets:      targets,
		Confirm:      confirm,
	})
	if res != nil {
		// A partial result on failure still carries the changes made before the failure.
		attachResetPayload(subject, res, err)
	}
	if err != nil {
		return err
	}
	if res.Aborted {
		fmt.Fprintln(os.Stderr, "layat: reset aborted")
		return nil
	}
	if flagVerbose {
		reportResetResult(res, name)
	}
	return nil
}

// printResetPlan prints reset --dryrun's removal targets to stdout, one per line. --json
// suppresses the stdout lines; the stderr nothing-to-remove notice stays.
func printResetPlan(res *engine.ResetResult) {
	if !flagJSON {
		for _, t := range res.RemovedSymlinks {
			fmt.Printf("remove-symlink\t%s\n", t)
		}
		for _, t := range res.RemovedCopies {
			fmt.Printf("remove-copy\t%s\n", t)
		}
		for _, t := range res.KeptForeign {
			fmt.Printf("keep-foreign\t%s\n", t)
		}
	}
	if len(res.RemovedSymlinks)+len(res.RemovedCopies)+len(res.KeptForeign) == 0 {
		fmt.Fprintln(os.Stderr, "layat: reset --dryrun: nothing to remove")
	}
}

// reportResetTargets prints the planned removals to stderr before the confirmation prompt.
func reportResetTargets(res *engine.ResetResult, name string) {
	fmt.Fprintf(os.Stderr, "layat: reset %s removal targets (root=%s):\n", name, res.Root)
	for _, t := range res.RemovedSymlinks {
		fmt.Fprintf(os.Stderr, "  symlink %s\n", t)
	}
	for _, t := range res.RemovedCopies {
		fmt.Fprintf(os.Stderr, "  copy    %s\n", t)
	}
	for _, t := range res.KeptForeign {
		fmt.Fprintf(os.Stderr, "  keep    %s (foreign / record mismatch; kept)\n", t)
	}
	if len(res.RemovedSymlinks)+len(res.RemovedCopies) == 0 {
		fmt.Fprintln(os.Stderr, "  (nothing to remove)")
	}
}

// reportResetResult prints the actual removal result to stderr.
func reportResetResult(res *engine.ResetResult, name string) {
	fmt.Fprintf(os.Stderr, "layat: reset %s done (root=%s)\n", name, res.Root)
	for _, t := range res.RemovedSymlinks {
		fmt.Fprintf(os.Stderr, "  removed-symlink %s\n", t)
	}
	for _, t := range res.RemovedCopies {
		fmt.Fprintf(os.Stderr, "  removed-copy    %s\n", t)
	}
	for _, t := range res.KeptForeign {
		fmt.Fprintf(os.Stderr, "  kept            %s (foreign / record mismatch)\n", t)
	}
	for _, t := range res.Pruned {
		fmt.Fprintf(os.Stderr, "  pruned          %s\n", t)
	}
	if len(res.RemovedSymlinks)+len(res.RemovedCopies) == 0 {
		fmt.Fprintln(os.Stderr, "  no-op")
	}
}

// resetPromptAllowed reports whether reset may prompt: it requires a TTY, and --json forbids it.
func resetPromptAllowed(interactive, jsonMode bool) bool {
	return interactive && !jsonMode
}

// confirmPolicy decides a destructive command's confirmation: --yes skips it, an interactive
// environment prompts, and a non-interactive one is refused. operation names the command in the
// refusal.
func confirmPolicy(yes, interactive bool, operation string) (needPrompt bool, err error) {
	if yes {
		return false, nil
	}
	if !interactive {
		return false, fmt.Errorf("layat: refusing destructive %s without --yes in a non-interactive context", operation)
	}
	return true, nil
}

// isInteractive reports whether stdin is a TTY; false under pipe / redirect / CI.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// promptYesNo reads y/N from stdin and returns true only for yes-type input (default No).
func promptYesNo(msg string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", msg)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
