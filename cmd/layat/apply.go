package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
)

var (
	flagApplyAll  bool // --all: apply all of layat.* in parallel, reported in lexical order (narrowable by root filter)
	flagApplyJobs int  // --jobs: apply --all's build and placement concurrency (0 = the logical CPU count; → ADR-0039)
)

// applyResultInfo / applyEnvInfo are apply's niface info slots (→ issue #196). apply's record
// lives entirely in items / changes, so both are empty seat types held as nil pointers: the
// omitempty on result.info / the envelope's info keeps them out of the document (a non-pointer
// struct{} would emit "info":{} and change the output). They exist so a later issue can put
// mutation run facts (profile / trunk root / retention ...) on them by adding fields alone,
// without touching the run's type arguments or any RunE instantiation.
//
// The asymmetry against the read commands is deliberate, not an oversight (→ issue #196 §5):
// unused slots get a named seat only where information is expected to land later, which is the
// mutation commands' two slots (ncompose reconstructs "what this run did" from them · niface
// ADR-0018). Read commands record no run-scoped state and init registers no subject, so their
// unused slots stay anonymous *struct{} — a named seat there would promise a future that is
// not planned. Read "named seat" as "reserved", "*struct{}" as "nothing goes here".
type (
	applyResultInfo struct{}
	applyEnvInfo    struct{}
)

// applyRun is apply's concrete run instantiation, threaded from RunE into the run functions.
// applySubject is one config's handle within it — what the payload builders attach to.
type (
	applyRun     = nifaceRun[*applyResultInfo, *applyEnvInfo]
	applySubject = nifaceSubject[*applyResultInfo]
)

// beginApplyRun starts apply's run (→ beginNifaceRun). It is the only production site that
// spells apply's info type pair outside the alias (the tests have one more, newApplyTestRun),
// and because it returns the alias type, changing applyRun's parameters without updating it is
// a compile error rather than a stale instantiation.
func beginApplyRun(command string) *applyRun {
	return beginNifaceRun[*applyResultInfo, *applyEnvInfo](command)
}

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [name]",
		Short: "Build layat.<name>, create a new generation, and apply it (defaults to layat.default)",
		Long: "Build and place the entrypoint's layat.<name>. " +
			"Omitting name applies layat.default (the flake default convention; an error if undefined). " +
			"--all applies all of layat.* in parallel and reports them in lexical order; --project-root / --home-root / --system-root narrow by root mode.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// beginApplyRun also publishes the run to nifaceReport, so main emits the envelope
			// after Execute returns whichever path below runs.
			run := beginApplyRun(cmd.Name())
			flagBackupEnabled = cmd.Flags().Changed("backup")
			if flagApplyAll {
				if len(args) > 0 {
					return fmt.Errorf("layat: apply cannot combine <name> with --all")
				}
				return runApplyAll(run)
			}
			if err := ensureNoRootFilter("apply --all"); err != nil {
				return err
			}
			// Changed, not the value: the default 0 is itself a valid --all setting, so an explicit
			// --jobs 0 on a named apply must be rejected too.
			if cmd.Flags().Changed("jobs") {
				return fmt.Errorf("layat: --jobs is a modifier for apply --all")
			}
			name := "default"
			if len(args) == 1 {
				name = args[0]
			}
			return runApply(run, name)
		},
	}
	cmd.Flags().BoolVar(&flagApplyAll, "all", false, "Apply all of layat.* in parallel, reporting in lexical order (continues on partial failure; exits non-zero if any fails)")
	cmd.Flags().IntVar(&flagApplyJobs, "jobs", 0,
		"With --all, build and then place up to N configs in parallel (0 = the logical CPU count; see ADR-0039)")
	cmd.Flags().BoolVar(&flagRecopy, "recopy", false,
		"Unconditionally re-copy every copy target from src, overwriting (discards local edits; see ADR-0020)")
	cmd.Flags().BoolVar(&flagDryrun, "dryrun", false,
		"Show place/replace/remove/conflict/no-op with zero side effects (exit 2 on conflict; see ADR-0006)")
	cmd.Flags().StringVar(&flagManifest, "manifest", "",
		"Apply a pre-built manifest (link-farm path) directly (host/module activation seam; no entrypoint discovery or nix eval/build; see ADR-0026)")
	cmd.Flags().StringVar(&flagBackup, "backup", "",
		"Back up an occupying foreign entity to \"<target>.<suffix>\" before placing, instead of stopping on conflict (bare --backup uses suffix \"layat-backup\"; \"=\" form required for a custom suffix, e.g. --backup=bak; see ADR-0045)")
	cmd.Flags().Lookup("backup").NoOptDefVal = "layat-backup"
	return cmd
}

// runApplyManifest applies the pre-built link-farm passed via --manifest directly to the engine
// (the module activation path). It does no entrypoint discovery, no rootKind pre-resolution eval,
// and no nix build; the engine reads rootKind from manifest.json (an HM module pins homeRoot, so home).
// It drives engine.Apply's Build=nil path (a pre-built LinkFarm) from the CLI (→ engine.Options).
func runApplyManifest(subject *applySubject, name string) error {
	linkFarm, err := filepath.Abs(flagManifest)
	if err != nil {
		return fmt.Errorf("layat: cannot resolve the --manifest path (%s): %w", flagManifest, err)
	}

	res, err := engine.Apply(engine.Options{
		Name:         name,
		LinkFarm:     linkFarm,
		RootOverride: flagRoot,
		NoWait:       flagNoWait,
		Recopy:       flagRecopy,
		Backup:       flagBackupEnabled,
		BackupSuffix: flagBackup,
	})
	if res != nil {
		// Also on failure: a partial result carries the reached/unreached item partition and
		// the changes that actually happened before the stop (→ issue #131, niface ADR-0020).
		attachMutationPayload(subject, res, err)
	}
	if err != nil {
		if errors.Is(err, engine.ErrSkipped) {
			if flagVerbose {
				fmt.Fprintln(os.Stderr, "layat: skipped because another apply is in progress (run layat apply manually)")
			}
			return nil
		}
		return err
	}

	if flagVerbose {
		reportResult(res, name)
	}
	return nil
}

// runApply drives the "execution flow" in docs/spec.md:
// entrypoint discovery → rootKind pre-resolution eval → engine.Apply (flock → in-lock build → place → --set → remove .pending).
// When --manifest is given it does no entrypoint discovery and no nix eval/build, passing the pre-built link-farm
// directly to the engine (the module activation path; → docs/spec.md "per-module behavior spec", ADR-0003, ADR-0007).
func runApply(run *applyRun, name string) error {
	// The config name is the niface subject; errors from here on are subject-borne (→ issue #130).
	// A named apply registers exactly one, so the run's results[] holds N=1 (→ issue #164).
	subject := run.beginSubject(name)
	if flagManifest != "" {
		// --manifest fixes the source to a link-farm, so it conflicts in meaning with the
		// entrypoint discovery flags (the positional name is orthogonal as a profile selector and coexists; → ADR-0026).
		if flagFile != "" {
			return errors.New("layat: --manifest cannot be combined with -f (--manifest fixes the source to a pre-built link-farm)")
		}
		return runApplyManifest(subject, name)
	}

	ep, err := discoverEntrypoint(flagFile)
	if err != nil {
		return err
	}
	system, err := currentSystem()
	if err != nil {
		return err
	}

	// 1. Pre-resolve rootKind before build (to establish the order profileDir resolution → flock → build; → ADR-0023).
	rootKind, fixedRoot, err := evalRoot(ep, system, name)
	if err != nil {
		return err
	}

	// 1.5 --dryrun is a side-effect-free preview (takes no flock / pending gcroot; runs only build read-only; → ADR-0023).
	if flagDryrun {
		res, err := engine.Apply(engine.Options{
			Name:         name,
			RootKind:     rootKind,
			FixedRoot:    fixedRoot,
			RootOverride: flagRoot,
			Recopy:       flagRecopy,
			Backup:       flagBackupEnabled,
			BackupSuffix: flagBackup,
			DryRun:       true,
			Build:        dryBuildFunc(ep, system, name),
		})
		if err != nil {
			return err
		}
		// The dryrun rides the same payload builder as the real apply, so parity is
		// structural — same schema by construction, only the observed values differ
		// (→ issue #132). cmdErr is nil here: a conflict is item-borne (failed item +
		// E_LAYAT_COLLISION inside the payload) and the exit-2 decision comes below,
		// after the plan is printed — the envelope still carries the payload alongside.
		attachMutationPayload(subject, res, nil)
		printApplyPlan(res)
		// exit 2 if there are conflicts (a pre-gate for CI; → docs/spec.md exit code table).
		if len(res.Conflicts) > 0 {
			return &exitError{code: 2}
		}
		return nil
	}

	// 2. Drive the engine (flock acquisition, in-lock build, placement, commit, and .pending removal are owned by the engine).
	res, err := applyOne(ep, system, name, rootKind, fixedRoot)
	if res != nil {
		// Also on failure: a partial result carries the reached/unreached item partition and
		// the changes that actually happened before the stop (→ issue #131, niface ADR-0020).
		attachMutationPayload(subject, res, err)
	}
	if err != nil {
		if errors.Is(err, engine.ErrSkipped) {
			// A try-lock skip is a normal skip (exit 0; → docs/spec.md exit code table).
			if flagVerbose {
				fmt.Fprintln(os.Stderr, "layat: skipped because another apply is in progress (run layat apply manually)")
			}
			return nil
		}
		return err
	}

	if flagVerbose {
		reportResult(res, name)
	}
	return nil
}

// printApplyPlan prints the apply --dryrun plan to stdout (it owns the machine-readable output; one action per line;
// → docs/spec.md stream discipline, ADR-0023, ADR-0024). It is not suppressed even under silent-on-success (the stdout-ownership principle; → ADR-0031).
// conflict lines are also put on stdout as part of the plan, with the exit code (exit 2) complementing machine discrimination.
// Under --json it prints nothing: stdout belongs to the niface envelope alone, and gating here —
// the single chokepoint for every call site — keeps that contract testable (→ ADR-0043 §2, issue #130).
func printApplyPlan(res *engine.Result) {
	if flagJSON {
		return
	}
	for _, t := range res.Placed {
		fmt.Printf("place\t%s\n", t)
	}
	for _, t := range res.Replaced {
		fmt.Printf("replace\t%s\n", t)
	}
	for _, t := range res.Copied {
		fmt.Printf("copy\t%s\n", t)
	}
	for _, t := range res.Removed {
		fmt.Printf("remove\t%s\n", t)
	}
	for _, t := range res.BackedUp {
		fmt.Printf("backup\t%s\n", t)
	}
	for _, c := range res.Conflicts {
		fmt.Printf("conflict\t%s: %s\n", c.Entry.Target, c.Reason)
	}
}

// applyOne runs engine.Apply for one config (shared by runApply / runApplyAll). rootKind / fixedRoot
// come from evalRoot pre-resolution for the single case and from the batch eval (evalAllRoots) for --all. Only build is done per config.
func applyOne(ep *entrypoint, system, name, rootKind, fixedRoot string) (*engine.Result, error) {
	return engine.Apply(engine.Options{
		Name:         name,
		RootKind:     rootKind,
		FixedRoot:    fixedRoot,
		RootOverride: flagRoot,
		NoWait:       flagNoWait,
		Recopy:       flagRecopy,
		Backup:       flagBackupEnabled,
		BackupSuffix: flagBackup,
		Build:        buildFunc(ep, system, name),
	})
}

// runApplyAll applies all of the entrypoint's layat.* (→ docs/spec.md execution flow, ADR-0024, ADR-0039).
// rootKind and targets are taken in a single batch eval (collapsing process launches N→1); build is per config for atomicity.
// It runs in two stages (→ ADR-0039): stage 1 realizes every selected config's build ahead of time on a
// worker pool of --jobs, and stage 2 applies them on a worker pool of the same size, where the engine's
// in-lock build is then a cache hit (the build stays inside the lock; stage 1 only warms the store).
// Stage 2's execution order is not deterministic; results[] keeps the lexical order of the subjects,
// which are registered before any config runs.
// It continues with the rest on a partial failure, shows an aggregate at the end, and exits non-zero if any one fails.
// Each selected config becomes one SubjectResult in results[], the same shape a named apply
// emits with N=1 (→ issue #164); the failures below stay on their own subject, so a partial
// failure still carries every succeeded config's result.
func runApplyAll(run *applyRun) error {
	filter, err := selectedRootFilter()
	if err != nil {
		return err
	}
	jobs, err := resolveApplyJobs(flagApplyJobs)
	if err != nil {
		return err
	}
	ep, err := discoverEntrypoint(flagFile)
	if err != nil {
		return err
	}
	system, err := currentSystem()
	if err != nil {
		return err
	}

	// 1. Get rootKind + targets in a single batch eval (config name → rootInfo map; → ADR-0024, ADR-0038).
	roots, err := evalAllRoots(ep, system)
	if err != nil {
		return err
	}

	// 2. Sort lexically and, if a root filter (--project-root etc.) is given, narrow to that mode only.
	names := make([]string, 0, len(roots))
	for name := range roots {
		names = append(names, name)
	}
	sort.Strings(names)
	var selected []string
	for _, name := range names {
		if filter == "" || roots[name].RootKind == filter {
			selected = append(selected, name)
		}
	}
	if len(selected) == 0 {
		if flagVerbose {
			fmt.Fprintln(os.Stderr, "layat: apply --all: no matching configs")
		}
		return nil
	}

	// 2.3 Stop before any build when two selected configs claim the same normalized target in the
	//     same root (--dryrun included; a named apply is not checked; → ADR-0038).
	if err := detectCrossConfigConflicts(roots, selected, flagRoot); err != nil {
		return err
	}

	// 2.4 Stage 1: realize the selected configs' builds in parallel (read-only: --no-link, no gcroot).
	//     A config that fails here never reaches stage 2 (→ skipFailedPrebuilds).
	built := prebuildAll(selected, jobs, func(name string) (string, error) {
		return realizeNoLink(ep, system, name, "["+name+"] ")
	})

	// 2.5 --dryrun is a side-effect-free preview (takes no flock / --set / pending gcroot; runs only build
	//     read-only). It aggregates each selected config's plan to stdout and decides the exit code by
	//     priority error(1) > conflict(2) > 0 (→ docs/spec.md, ADR-0024).
	if flagDryrun {
		return runApplyAllDryRun(run, ep, system, selected, jobs, roots, built)
	}

	// 3. Stage 2: apply the configs in parallel, each independently. Continue on partial failure and aggregate failures (each config is independently atomic).
	applied, skipped, failures := aggregateApply(run, selected, jobs, skipFailedPrebuilds(built, func(name string) (*engine.Result, error) {
		ri := roots[name]
		return applyOne(ep, system, name, ri.RootKind, ri.Root)
	}))

	// 4. Aggregate report and exit code (priority error(1) > conflict(2) > 0; → docs/spec.md, ADR-0024).
	if flagVerbose {
		fmt.Fprintf(os.Stderr, "layat: apply --all done (applied %d / skipped %d / failed %d / selected %d)\n",
			applied, skipped, failures, len(selected))
	}
	// conflict(2) arises only on the --dryrun (#13) read-only path. Non-dryrun --all yields only error/0.
	code := applyAllExitCode(failures > 0, false)
	if code == 0 {
		return nil
	}
	return &exitCodeError{code: code, msg: fmt.Sprintf("layat: apply --all: %d config(s) failed", failures)}
}

// detectCrossConfigConflicts is apply --all's cross-config target conflict preflight (→ ADR-0038).
// It groups the selected configs into buckets that resolve to the same root — one per rootKind for
// project / home / system, one per root string value for fixed, and a single bucket for everything
// when rootOverride (--root) is set — and errors when a normalized target appears in two configs of
// one bucket. Configs outside selected are ignored. The error is a plain one (exit 1, and
// E_LAYAT_FAILED on the --json top-level errors[] since no subject is registered yet).
func detectCrossConfigConflicts(roots map[string]rootInfo, selected []string, rootOverride string) error {
	bucketOf := func(ri rootInfo) string {
		switch {
		case rootOverride != "":
			return "--root " + rootOverride
		case ri.RootKind == manifest.RootKindFixed:
			return "fixed root " + ri.Root
		default:
			return ri.RootKind + " root"
		}
	}
	// bucket → target → the first selected config that claimed it.
	owners := map[string]map[string]string{}
	var conflicts []string
	for _, name := range selected {
		ri := roots[name]
		bucket := bucketOf(ri)
		if owners[bucket] == nil {
			owners[bucket] = map[string]string{}
		}
		for _, target := range ri.Targets {
			if first, dup := owners[bucket][target]; dup {
				conflicts = append(conflicts, fmt.Sprintf("  %s: claimed by both %s and %s (%s)", target, first, name, bucket))
				continue
			}
			owners[bucket][target] = name
		}
	}
	if len(conflicts) == 0 {
		return nil
	}
	return fmt.Errorf("layat: apply --all: selected configs place the same target in the same root (nothing was built or placed; → ADR-0038):\n%s",
		strings.Join(conflicts, "\n"))
}

// runApplyAllDryRun drives apply --all --dryrun. It builds each selected config read-only and
// aggregates the plan to stdout (taking none of FS writes / flock / --set / pending gcroot; → ADR-0023).
// It decides the exit code by priority error(1) > conflict(2) > 0 (→ docs/spec.md, ADR-0024) and carries it in an
// empty-msg exitError (symmetric with the single apply --dryrun conflict=2; main exits with the code alone).
// Stage 1 (built) is shared with the real apply; a config whose build failed there is not planned.
// The read-only applies run on the same --jobs pool as the real apply's stage 2.
func runApplyAllDryRun(run *applyRun, ep *entrypoint, system string, selected []string, jobs int, roots map[string]rootInfo, built map[string]prebuildResult) error {
	code := aggregateDryRun(run, selected, jobs, skipFailedPrebuilds(built, func(name string) (*engine.Result, error) {
		ri := roots[name]
		return engine.Apply(engine.Options{
			Name:         name,
			RootKind:     ri.RootKind,
			FixedRoot:    ri.Root,
			RootOverride: flagRoot,
			Recopy:       flagRecopy,
			Backup:       flagBackupEnabled,
			BackupSuffix: flagBackup,
			DryRun:       true,
			Build:        dryBuildFunc(ep, system, name),
		})
	}))
	if code == 0 {
		return nil
	}
	return &exitError{code: code}
}

// prebuildResult is one config's stage-1 outcome: the realized link-farm store path, or the build error.
type prebuildResult struct {
	storePath string
	err       error
}

// prebuildAll is apply --all's stage 1 (→ ADR-0039): it runs build for every selected config on a
// worker pool of min(jobs, len(selected)) goroutines and collects config name → outcome. It does not
// stop on a failure (each config's outcome is independent), and returns only after every build ends.
// Workers write to their own slot of an index-addressed slice, so the collection needs no lock.
func prebuildAll(selected []string, jobs int, build func(name string) (string, error)) map[string]prebuildResult {
	results := make([]prebuildResult, len(selected))
	queue := make(chan int)
	var wg sync.WaitGroup
	for range min(jobs, len(selected)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range queue {
				store, err := build(selected[i])
				results[i] = prebuildResult{storePath: store, err: err}
			}
		}()
	}
	for i := range selected {
		queue <- i
	}
	close(queue)
	wg.Wait()

	built := make(map[string]prebuildResult, len(selected))
	for i, name := range selected {
		built[name] = results[i]
	}
	return built
}

// skipFailedPrebuilds wraps stage 2's per-config function so a config whose stage-1 build failed
// returns that error without being applied. The aggregators then settle it like any other failure —
// its own subject in selection order, counted, and reported on stderr — so results[] keeps the
// lexical order.
func skipFailedPrebuilds(built map[string]prebuildResult, fn func(name string) (*engine.Result, error)) func(name string) (*engine.Result, error) {
	return func(name string) (*engine.Result, error) {
		if err := built[name].err; err != nil {
			return nil, err
		}
		return fn(name)
	}
}

// forEachConfig runs work(i) for every index of selected on a worker pool of min(jobs, len(selected))
// goroutines and returns once all of them are done (apply --all's stage 2; → ADR-0039). Each work
// call must touch only its own index's state, so the callers need no lock.
func forEachConfig(selected []string, jobs int, work func(i int)) {
	queue := make(chan int)
	var wg sync.WaitGroup
	for range min(jobs, len(selected)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range queue {
				work(i)
			}
		}()
	}
	for i := range selected {
		queue <- i
	}
	close(queue)
	wg.Wait()
}

// beginSubjects registers one niface subject per selected config, in selection (lexical) order,
// before any config runs: beginSubject appends to the run's subject list, which the workers must
// not share, and the registration order is results[]'s order whatever order the configs finish in.
func beginSubjects(run *applyRun, selected []string) []*applySubject {
	subjects := make([]*applySubject, len(selected))
	for i, name := range selected {
		subjects[i] = run.beginSubject(name)
	}
	return subjects
}

// applyOutcome is one config's settled stage-2 outcome under apply --all.
type applyOutcome struct {
	skipped bool  // a try-lock skip (a normal skip, not a failure)
	err     error // the config's failure; nil on success and on a skip
}

// aggregateApply runs each selected config via applyFn and aggregates applied / skipped / failure counts,
// continuing on partial failure (each config is independently atomic; → docs/spec.md "continue on partial
// failure"). It does not swallow ErrSkipped (a try-lock skip is normal) or other errors; a skip is counted
// and reported to stderr under flagVerbose, a failure is counted and always reported to stderr, and either
// way the rest continue (a seam that injects the apply implementation for testability, mirroring aggregateDryRun).
//
// The configs run on a worker pool of jobs (→ ADR-0039). Each worker settles only its own config —
// its subject's payload and outcome — and the counts are gathered after every config is done, so
// they need no lock.
//
// Each config also gets its own niface subject, settled with that config's own outcome (→ issue
// #164): the counts drive the aggregate exit code as before, while the subjects carry the per-config
// results — including every succeeded one alongside a partial failure.
func aggregateApply(run *applyRun, selected []string, jobs int, applyFn func(name string) (*engine.Result, error)) (applied, skipped, failures int) {
	subjects := beginSubjects(run, selected)
	outcomes := make([]applyOutcome, len(selected))
	forEachConfig(selected, jobs, func(i int) {
		name, subject, o := selected[i], subjects[i], &outcomes[i]
		res, err := applyFn(name)
		if res != nil {
			// Also on failure: a partial result carries the reached/unreached item partition and
			// the changes that actually happened before the stop (→ issue #131, niface ADR-0020).
			attachMutationPayload(subject, res, err)
		}
		// subjectErr is what this config's subject settles on, which is not always err: a try-lock
		// skip is a normal skip (exit 0), so its subject succeeds — symmetric with the named apply,
		// which returns nil on ErrSkipped (→ docs/spec.md exit codes). Deciding it before the
		// reporting below keeps finish to a single unconditional call, so no branch can forget it.
		subjectErr := err
		switch {
		case err == nil:
			if flagVerbose {
				reportResult(res, name)
			}
		case errors.Is(err, engine.ErrSkipped):
			subjectErr = nil
			o.skipped = true
			if flagVerbose {
				fmt.Fprintf(os.Stderr, "layat: skipped apply %s (another apply is in progress)\n", name)
			}
		default:
			o.err = err
			// Do not swallow partial failures; print to stderr and continue (→ docs/spec.md "continue on partial failure").
			fmt.Fprintf(os.Stderr, "layat: apply %s failed: %v\n", name, err)
		}
		subject.finish(subjectErr)
	})
	for i := range outcomes {
		o := &outcomes[i]
		switch {
		case o.skipped:
			skipped++
		case o.err != nil:
			failures++
		default:
			applied++
		}
	}
	return applied, skipped, failures
}

// dryRunOutcome is one config's settled read-only outcome under apply --all --dryrun.
type dryRunOutcome struct {
	err      error // the config's build / eval failure
	conflict bool  // the plan has at least one conflict
}

// aggregateDryRun runs each selected config read-only via applyDry, prints the plan to stdout,
// aggregates error / conflict, and returns the exit code (a seam that injects the apply implementation for testability).
// It does not swallow a config's build / eval failure (error); it prints to stderr, continues, and reflects it in the final code.
// Like aggregateApply it runs the configs on a worker pool of jobs (→ ADR-0039).
//
// Like aggregateApply it settles one niface subject per config, riding the same payload builder as
// the real apply so the dryrun's SubjectResult is the same shape by construction (→ issue #164). A
// conflict is item-borne — the conflicting entry is a failed item carrying E_LAYAT_COLLISION — and
// still puts that subject in error, symmetric with the named apply --dryrun (→ layat ADR-0043 §6,
// niface ADR-0002).
func aggregateDryRun(run *applyRun, selected []string, jobs int, applyDry func(name string) (*engine.Result, error)) int {
	subjects := beginSubjects(run, selected)
	outcomes := make([]dryRunOutcome, len(selected))
	forEachConfig(selected, jobs, func(i int) {
		name, subject, o := selected[i], subjects[i], &outcomes[i]
		res, err := applyDry(name)
		if err != nil {
			o.err = err
			// Do not swallow partial failures; print to stderr and continue (→ docs/spec.md "continue on partial failure").
			fmt.Fprintf(os.Stderr, "layat: apply %s --dryrun failed: %v\n", name, err)
			subject.finish(err)
			return
		}
		attachMutationPayload(subject, res, nil)
		printApplyPlan(res)
		o.conflict = len(res.Conflicts) > 0
		// No subject-level error either way: a conflict is already failed items carrying
		// E_LAYAT_COLLISION, and the payload's item-borne mark is what puts this subject in error
		// (→ nifaceSubject.itemBorne) — the same mechanism aggregateApply relies on for an
		// entry-scoped failure, so both settle a config the one way.
		subject.finish(nil)
	})
	var anyError, anyConflict bool
	for i := range outcomes {
		o := &outcomes[i]
		anyError = anyError || o.err != nil
		anyConflict = anyConflict || o.conflict
	}
	return applyAllExitCode(anyError, anyConflict)
}

// applyAllExitCode decides apply --all's exit code by priority error(1) > conflict(2) > 0
// (→ docs/spec.md "output streams and exit codes", ADR-0024). It does not take the plain maximum (2 > 1)
// (because a conflict would hide serious eval / engine errors in CI). conflict arises only on the --dryrun path.
func applyAllExitCode(anyError, anyConflict bool) int {
	switch {
	case anyError:
		return 1
	case anyConflict:
		return 2
	default:
		return 0
	}
}

// selectedRootFilter returns the root mode filter from --project-root / --home-root / --system-root
// (none gives ""; specifying more than one is an error; → ADR-0017). The return value is a manifest.RootKind* string.
func selectedRootFilter() (string, error) {
	var modes []string
	if flagProjectRoot {
		modes = append(modes, manifest.RootKindProject)
	}
	if flagHomeRoot {
		modes = append(modes, manifest.RootKindHome)
	}
	if flagSystemRoot {
		modes = append(modes, manifest.RootKindSystem)
	}
	if len(modes) > 1 {
		return "", fmt.Errorf("layat: --project-root / --home-root / --system-root may be specified only one at a time")
	}
	if len(modes) == 0 {
		return "", nil
	}
	return modes[0], nil
}

// resolveApplyJobs resolves --jobs to apply --all's stage-1 build concurrency: 0 (the default) is
// the logical CPU count, a positive value is taken as-is, and a negative one is an error (→ ADR-0039).
func resolveApplyJobs(jobs int) (int, error) {
	switch {
	case jobs < 0:
		return 0, fmt.Errorf("layat: --jobs must be 0 (the logical CPU count) or a positive integer, got %d", jobs)
	case jobs == 0:
		return runtime.NumCPU(), nil
	default:
		return jobs, nil
	}
}

// ensureNoRootFilter errors when a root filter is used outside --all
// (the filter is a modifier for --all; in a named apply <name> pins a single config, so it is meaningless; → ADR-0017).
func ensureNoRootFilter(modifier string) error {
	if flagProjectRoot || flagHomeRoot || flagSystemRoot {
		return fmt.Errorf("layat: --project-root / --home-root / --system-root are modifiers for %s", modifier)
	}
	return nil
}

// reportResult prints the placement report to stderr (stdout is reserved for machine-readable output; → ADR-0023).
func reportResult(res *engine.Result, name string) {
	fmt.Fprintf(os.Stderr, "layat: apply %s done (root=%s)\n", name, res.Root)
	for _, t := range res.Placed {
		fmt.Fprintf(os.Stderr, "  placed   %s\n", t)
	}
	for _, t := range res.Replaced {
		fmt.Fprintf(os.Stderr, "  replaced %s\n", t)
	}
	for _, t := range res.Copied {
		fmt.Fprintf(os.Stderr, "  copied   %s\n", t)
	}
	for _, t := range res.Recopied {
		fmt.Fprintf(os.Stderr, "  recopied %s\n", t)
	}
	for _, t := range res.Removed {
		fmt.Fprintf(os.Stderr, "  removed  %s\n", t)
	}
	for _, t := range res.Pruned {
		fmt.Fprintf(os.Stderr, "  pruned   %s\n", t)
	}
	for _, t := range res.BackedUp {
		fmt.Fprintf(os.Stderr, "  backedUp %s\n", t)
	}
	if len(res.Placed)+len(res.Replaced)+len(res.Copied)+len(res.Recopied)+len(res.Removed)+len(res.Pruned)+len(res.BackedUp) == 0 {
		fmt.Fprintln(os.Stderr, "  no-op")
	}
}
