package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/manifest"
)

var (
	flagApplyAll  bool // --all: apply all of layat.* in parallel, reported in lexical order (narrowable by root filter)
	flagApplyJobs int  // --jobs: apply --all's build and placement concurrency (0 = the logical CPU count)
)

// applyResultInfo / applyEnvInfo are apply's outturn info slots, reserved for future run facts.
// They are held as nil pointers, so omitempty keeps them out of the document.
type (
	applyResultInfo struct{}
	applyEnvInfo    struct{}
)

// applyRun is apply's concrete run instantiation, threaded from RunE into the run functions.
// applySubject is one config's handle within it — what the payload builders attach to.
type (
	applyRun     = outturnRun[*applyResultInfo, *applyEnvInfo]
	applySubject = outturnSubject[*applyResultInfo]
)

// beginApplyRun starts apply's run with apply's info type pair.
func beginApplyRun(command string) *applyRun {
	return beginOutturnRun[*applyResultInfo, *applyEnvInfo](command)
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
			// beginApplyRun publishes the run to outturnReport, so main emits the envelope after Execute.
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
			// An explicit --jobs 0 is rejected too, since 0 is a valid --all value.
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
		"With --all, build and then place up to N configs in parallel (0 = the logical CPU count)")
	cmd.Flags().BoolVar(&flagRecopy, "recopy", false,
		"Unconditionally re-copy every copy target from src, overwriting (discards local edits)")
	cmd.Flags().BoolVar(&flagDryrun, "dryrun", false,
		"Show place/replace/remove/conflict/no-op with zero side effects (exit 2 on conflict)")
	cmd.Flags().StringVar(&flagManifest, "manifest", "",
		"Apply a pre-built manifest (link-farm path) directly (host/module activation seam; no entrypoint discovery or nix eval/build)")
	cmd.Flags().StringVar(&flagBackup, "backup", "",
		"Back up an occupying foreign entity to \"<target>.<suffix>\" before placing, instead of stopping on conflict (bare --backup uses suffix \"layat-backup\"; \"=\" form required for a custom suffix, e.g. --backup=bak)")
	cmd.Flags().Lookup("backup").NoOptDefVal = "layat-backup"
	return cmd
}

// runApplyManifest applies the pre-built link-farm given by --manifest through the prebuilt
// generator, without entrypoint discovery, rootKind eval or nix build.
func runApplyManifest(subject *applySubject, name string) error {
	gen, err := newGenerator()
	if err != nil {
		return err
	}
	if err := gen.Discover(flagManifest); err != nil {
		return err
	}
	linkFarm, err := gen.DryBuild(name)
	if err != nil {
		return err
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
		// A partial result on failure still carries the reached items and the changes made.
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

// runApply builds and applies one named config, or applies a pre-built link-farm under --manifest.
func runApply(run *applyRun, name string) error {
	// The config name is the run's single outturn subject.
	subject := run.beginSubject(name)
	if flagManifest != "" {
		// --manifest fixes the source to a link-farm, so it conflicts with -f.
		if flagFile != "" {
			return &inputError{err: errors.New("layat: --manifest cannot be combined with -f (--manifest fixes the source to a pre-built link-farm)")}
		}
		return runApplyManifest(subject, name)
	}

	gen, err := newGenerator()
	if err != nil {
		return err
	}
	if err := gen.Discover(flagFile); err != nil {
		return err
	}

	// 1. Pre-resolve rootKind before build.
	root, err := gen.Roots(name)
	if err != nil {
		return err
	}
	rootKind, fixedRoot := root.RootKind, root.Root

	// 1.5 --dryrun previews without side effects (no flock / pending gcroot; build is read-only).
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
			Build:        dryBuildFunc(gen, name),
		})
		if err != nil {
			return err
		}
		// The dryrun uses the real apply's payload builder. A conflict is item-borne, so cmdErr is nil
		// and exit 2 is decided after the plan is printed.
		attachMutationPayload(subject, res, nil)
		printApplyPlan(res)
		// Exit 2 if there are conflicts.
		if len(res.Conflicts) > 0 {
			return &exitError{code: 2}
		}
		return nil
	}

	// 2. Drive the engine (flock, in-lock build, placement, commit and .pending removal).
	res, err := applyOne(gen, name, rootKind, fixedRoot)
	if res != nil {
		// A partial result on failure still carries the reached items and the changes made.
		attachMutationPayload(subject, res, err)
	}
	if err != nil {
		if errors.Is(err, engine.ErrSkipped) {
			// A try-lock skip is a normal skip (exit 0).
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

// printApplyPlan prints the apply --dryrun plan to stdout, one action per line.
// It prints even when silent on success, and prints nothing under --json.
func printApplyPlan(res *engine.Result) {
	fprintApplyPlan(os.Stdout, res)
}

// fprintApplyPlan is printApplyPlan writing to w.
func fprintApplyPlan(w io.Writer, res *engine.Result) {
	if flagJSON {
		return
	}
	for _, t := range res.Placed {
		_, _ = fmt.Fprintf(w, "place\t%s\n", t)
	}
	for _, t := range res.Replaced {
		_, _ = fmt.Fprintf(w, "replace\t%s\n", t)
	}
	for _, t := range res.Copied {
		_, _ = fmt.Fprintf(w, "copy\t%s\n", t)
	}
	for _, t := range res.Removed {
		_, _ = fmt.Fprintf(w, "remove\t%s\n", t)
	}
	for _, t := range res.BackedUp {
		_, _ = fmt.Fprintf(w, "backup\t%s\n", t)
	}
	for _, c := range res.Conflicts {
		_, _ = fmt.Fprintf(w, "conflict\t%s: %s\n", c.Entry.Target, c.Reason)
	}
}

// applyOne runs engine.Apply for one config; only build is done per config.
func applyOne(gen generator.Generator, name, rootKind, fixedRoot string) (*engine.Result, error) {
	return engine.Apply(engine.Options{
		Name:         name,
		RootKind:     rootKind,
		FixedRoot:    fixedRoot,
		RootOverride: flagRoot,
		NoWait:       flagNoWait,
		Recopy:       flagRecopy,
		Backup:       flagBackupEnabled,
		BackupSuffix: flagBackup,
		Build:        func(pending string) (string, error) { return gen.Build(name, pending) },
	})
}

// dryBuildFunc returns the --dryrun build callback: the generator's DryBuild, which lays down no
// gcroot. The pending argument is unused.
func dryBuildFunc(gen generator.Generator, name string) engine.BuildFunc {
	return func(string) (string, error) { return gen.DryBuild(name) }
}

// runApplyAll applies all selected configs of layat.* in two stages: stage 1 prebuilds them in
// parallel, stage 2 applies them in parallel. It continues on partial failure, reports in lexical
// order, and exits non-zero if any config fails.
func runApplyAll(run *applyRun) error {
	// --manifest fixes the source to one link-farm, which --all cannot apply per config.
	if flagManifest != "" {
		return &inputError{err: errors.New("layat: --manifest cannot be combined with --all (--manifest fixes the source to a pre-built link-farm)")}
	}
	filter, err := selectedRootFilter()
	if err != nil {
		return err
	}
	jobs, err := resolveApplyJobs(flagApplyJobs)
	if err != nil {
		return err
	}
	// The generator is selected once; stage 1's per-config generators reuse the name.
	genName, err := selectGenerator()
	if err != nil {
		return err
	}
	gen := newGeneratorTo(genName, os.Stderr)
	if err := gen.Discover(flagFile); err != nil {
		return err
	}

	// 1. Get rootKind + targets of every config in a single batch eval.
	roots, err := gen.AllRoots()
	if err != nil {
		return err
	}

	// 2. Sort lexically and narrow to the root filter's mode if given.
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

	// 2.3 Stop before any build when two selected configs claim the same target in the same root.
	if err := detectCrossConfigConflicts(roots, selected, flagRoot); err != nil {
		return err
	}

	// 2.4 Stage 1: prebuild the selected configs in parallel (read-only). Diagnostics carry a
	// "[<name>] " line prefix.
	built := prebuildAll(selected, jobs, func(name string) (string, error) {
		g := newGeneratorTo(genName, &linePrefixWriter{w: os.Stderr, prefix: "[" + name + "] "})
		if err := g.Discover(flagFile); err != nil {
			return "", err
		}
		return g.DryBuild(name)
	})

	// 2.5 --dryrun previews without side effects and decides the exit code error(1) > conflict(2) > 0.
	if flagDryrun {
		return runApplyAllDryRun(run, gen, selected, jobs, roots, built)
	}

	// 3. Stage 2: apply the configs in parallel, each independently atomic.
	applied, skipped, failures := aggregateApply(run, selected, jobs, skipFailedPrebuilds(built, func(name string) (*engine.Result, error) {
		ri := roots[name]
		return applyOne(gen, name, ri.RootKind, ri.Root)
	}))

	// 4. Aggregate report and exit code.
	if flagVerbose {
		fmt.Fprintf(os.Stderr, "layat: apply --all done (applied %d / skipped %d / failed %d / selected %d)\n",
			applied, skipped, failures, len(selected))
	}
	// conflict(2) arises only on the --dryrun path.
	code := applyAllExitCode(failures > 0, false)
	if code == 0 {
		return nil
	}
	return &exitCodeError{code: code, msg: fmt.Sprintf("layat: apply --all: %d config(s) failed", failures)}
}

// detectCrossConfigConflicts errors when a normalized target appears in two selected configs that
// resolve to the same root. Buckets are per rootKind, per fixed root value, or one under --root.
func detectCrossConfigConflicts(roots map[string]manifest.Root, selected []string, rootOverride string) error {
	bucketOf := func(ri manifest.Root) string {
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
	return fmt.Errorf("layat: apply --all: selected configs place the same target in the same root (nothing was built or placed):\n%s",
		strings.Join(conflicts, "\n"))
}

// runApplyAllDryRun drives apply --all --dryrun: it plans each config read-only and returns the
// exit code error(1) > conflict(2) > 0 in an empty-msg exitError. Configs whose prebuild failed
// are not planned.
func runApplyAllDryRun(run *applyRun, gen generator.Generator, selected []string, jobs int, roots map[string]manifest.Root, built map[string]prebuildResult) error {
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
			Build:        dryBuildFunc(gen, name),
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

// prebuildAll is apply --all's stage 1: it builds every selected config on a pool of jobs workers
// and returns config name → outcome once every build ends.
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
// returns that error without being applied.
func skipFailedPrebuilds(built map[string]prebuildResult, fn func(name string) (*engine.Result, error)) func(name string) (*engine.Result, error) {
	return func(name string) (*engine.Result, error) {
		if err := built[name].err; err != nil {
			return nil, err
		}
		return fn(name)
	}
}

// configOutput buffers one config's CLI output under apply --all, flushed in lexical order.
// Output written by the engine and nix themselves is not buffered.
type configOutput struct {
	stdout, stderr bytes.Buffer
}

// flush writes the buffered output to os.Stdout / os.Stderr.
func (o *configOutput) flush() {
	_, _ = o.stdout.WriteTo(os.Stdout)
	_, _ = o.stderr.WriteTo(os.Stderr)
}

// forEachConfig runs work(i) for every index of selected on a pool of jobs workers and returns
// once all are done. Each work call touches only its own index's state.
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

// beginSubjects registers one outturn subject per selected config in lexical order, before any
// config runs.
func beginSubjects(run *applyRun, selected []string) []*applySubject {
	subjects := make([]*applySubject, len(selected))
	for i, name := range selected {
		subjects[i] = run.beginSubject(name)
	}
	return subjects
}

// applyOutcome is one config's settled stage-2 outcome under apply --all.
type applyOutcome struct {
	out     configOutput
	skipped bool  // a try-lock skip (a normal skip, not a failure)
	err     error // the config's failure; nil on success and on a skip
}

// aggregateApply runs each selected config via applyFn on a pool of jobs workers and returns the
// applied / skipped / failure counts. Each config settles its own subject; output is flushed in
// lexical order, and failures are always reported to stderr.
func aggregateApply(run *applyRun, selected []string, jobs int, applyFn func(name string) (*engine.Result, error)) (applied, skipped, failures int) {
	subjects := beginSubjects(run, selected)
	outcomes := make([]applyOutcome, len(selected))
	forEachConfig(selected, jobs, func(i int) {
		name, subject, o := selected[i], subjects[i], &outcomes[i]
		res, err := applyFn(name)
		if res != nil {
			// A partial result on failure still carries the reached items and the changes made.
			attachMutationPayload(subject, res, err)
		}
		// A try-lock skip settles the subject as a success (exit 0).
		subjectErr := err
		switch {
		case err == nil:
			if flagVerbose {
				fprintResult(&o.out.stderr, res, name)
			}
		case errors.Is(err, engine.ErrSkipped):
			subjectErr = nil
			o.skipped = true
			if flagVerbose {
				fmt.Fprintf(&o.out.stderr, "layat: skipped apply %s (another apply is in progress)\n", name)
			}
		default:
			o.err = err
			// Report the failure to stderr and continue.
			fmt.Fprintf(&o.out.stderr, "layat: apply %s failed: %v\n", name, err)
		}
		subject.finish(subjectErr)
	})
	for i := range outcomes {
		o := &outcomes[i]
		o.out.flush()
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
	out      configOutput
	err      error // the config's build / eval failure
	conflict bool  // the plan has at least one conflict
}

// aggregateDryRun plans each selected config read-only via applyDry on a pool of jobs workers and
// returns the exit code. Each config settles its own subject; a conflict is a failed item carrying
// E_LAYAT_COLLISION.
func aggregateDryRun(run *applyRun, selected []string, jobs int, applyDry func(name string) (*engine.Result, error)) int {
	subjects := beginSubjects(run, selected)
	outcomes := make([]dryRunOutcome, len(selected))
	forEachConfig(selected, jobs, func(i int) {
		name, subject, o := selected[i], subjects[i], &outcomes[i]
		res, err := applyDry(name)
		if err != nil {
			o.err = err
			// Report the failure to stderr and continue.
			fmt.Fprintf(&o.out.stderr, "layat: apply %s --dryrun failed: %v\n", name, err)
			subject.finish(err)
			return
		}
		attachMutationPayload(subject, res, nil)
		fprintApplyPlan(&o.out.stdout, res)
		o.conflict = len(res.Conflicts) > 0
		// A conflict is already item-borne, so the subject settles with no subject-level error.
		subject.finish(nil)
	})
	var anyError, anyConflict bool
	for i := range outcomes {
		o := &outcomes[i]
		o.out.flush()
		anyError = anyError || o.err != nil
		anyConflict = anyConflict || o.conflict
	}
	return applyAllExitCode(anyError, anyConflict)
}

// applyAllExitCode decides apply --all's exit code by priority error(1) > conflict(2) > 0, so a
// conflict never hides an error.
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

// selectedRootFilter returns the manifest.RootKind* chosen by --project-root / --home-root /
// --system-root, "" for none, and an error for more than one.
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

// resolveApplyJobs resolves --jobs: 0 is the logical CPU count, a negative value is an error.
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

// ensureNoRootFilter errors when a root filter is used outside --all.
func ensureNoRootFilter(modifier string) error {
	if flagProjectRoot || flagHomeRoot || flagSystemRoot {
		return fmt.Errorf("layat: --project-root / --home-root / --system-root are modifiers for %s", modifier)
	}
	return nil
}

// reportResult prints the placement report to stderr.
func reportResult(res *engine.Result, name string) {
	fprintResult(os.Stderr, res, name)
}

// fprintResult is reportResult writing to w.
func fprintResult(w io.Writer, res *engine.Result, name string) {
	_, _ = fmt.Fprintf(w, "layat: apply %s done (root=%s)\n", name, res.Root)
	for _, t := range res.Placed {
		_, _ = fmt.Fprintf(w, "  placed   %s\n", t)
	}
	for _, t := range res.Replaced {
		_, _ = fmt.Fprintf(w, "  replaced %s\n", t)
	}
	for _, t := range res.Copied {
		_, _ = fmt.Fprintf(w, "  copied   %s\n", t)
	}
	for _, t := range res.Recopied {
		_, _ = fmt.Fprintf(w, "  recopied %s\n", t)
	}
	for _, t := range res.Removed {
		_, _ = fmt.Fprintf(w, "  removed  %s\n", t)
	}
	for _, t := range res.Pruned {
		_, _ = fmt.Fprintf(w, "  pruned   %s\n", t)
	}
	for _, t := range res.BackedUp {
		_, _ = fmt.Fprintf(w, "  backedUp %s\n", t)
	}
	if len(res.Placed)+len(res.Replaced)+len(res.Copied)+len(res.Recopied)+len(res.Removed)+len(res.Pruned)+len(res.BackedUp) == 0 {
		_, _ = fmt.Fprintln(w, "  no-op")
	}
}
