package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
	"github.com/yasunori0418/layat/internal/planner"
)

// Generation operations use `nix-env --profile <profileDir>/profile`: rollback
// (--switch-generation) and listing (--list-generations) live here, commit (--set) in commit.go.
// All are injectable so tmpdir tests do not call nix.

// Generation is one generation of a profile (one line of nix-env --list-generations).
type Generation struct {
	Number  int    // generation number
	Date    string // creation timestamp (nix-env's display verbatim; carried unparsed)
	Current bool   // whether it is the current generation (the one the profile link points at)
}

// ListGenerationsFunc is the generation-list retrieval point (default nix-env --list-generations). Substituted in tmpdir tests.
type ListGenerationsFunc func(profileLink string) ([]Generation, error)

// SwitchGenerationFunc is the profile pointer move point (default nix-env --switch-generation). Substituted in tmpdir tests.
type SwitchGenerationFunc func(profileLink string, gen int) error

// ProfileOptions is the input for fixing the profileDir (shared by Rollback / list-generations).
type ProfileOptions struct {
	Name         string
	RootKind     string
	FixedRoot    string
	RootOverride string
	WorkDir      string
	StateDir     string
	Git          GitFunc
}

// ProfileFor resolves root and fixes the profileDir layout (same shape as apply's preamble).
// Used in common by non-building rollback / list-generations for flock / generation reads.
func ProfileFor(opts ProfileOptions) (paths.Profile, string, error) {
	root, err := resolveRoot(opts.RootKind, opts.FixedRoot, opts.RootOverride, opts.WorkDir, opts.Git)
	if err != nil {
		return paths.Profile{}, "", err
	}
	stateDir := opts.StateDir
	if stateDir == "" {
		stateDir, err = paths.StateDir()
		if err != nil {
			return paths.Profile{}, "", err
		}
	}
	prof := paths.Resolve(stateDir, opts.Name, opts.RootKind, root, opts.RootOverride != "")
	return prof, root, nil
}

// ListGenerations returns the generation list for profileLink (default nix-env implementation). Used by the CLI's list-generations.
func ListGenerations(profileLink string) ([]Generation, error) {
	return nixEnvListGenerations(profileLink)
}

// observeGeneration returns the generation number the profile link points at, read from its
// "<base>-<N>-link" destination without a subprocess. It returns nil when there is no profile
// or the destination does not parse.
func observeGeneration(profileLink string) *int {
	dest, err := os.Readlink(profileLink)
	if err != nil {
		return nil
	}
	base := filepath.Base(dest)
	prefix := filepath.Base(profileLink) + "-"
	if !strings.HasPrefix(base, prefix) || !strings.HasSuffix(base, "-link") {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(base, prefix), "-link"))
	if err != nil || n < 0 {
		return nil
	}
	return &n
}

// intPtr copies n onto the heap for the nil-able generation observation fields.
func intPtr(n int) *int { return &n }

// RollbackOptions is the input to Rollback. Rollback is home mode only, but that decision is
// the CLI's; the engine resolves the profileDir regardless of rootKind and converges to the previous generation.
type RollbackOptions struct {
	Name         string
	RootKind     string
	FixedRoot    string
	RootOverride string
	WorkDir      string
	StateDir     string

	// ListGenerations / SwitchGeneration substitute the nix-env calls (nil = default nix-env implementation).
	ListGenerations  ListGenerationsFunc
	SwitchGeneration SwitchGenerationFunc
	// Git substitutes git toplevel resolution (nil = gitutil.Toplevel). Unused in home mode.
	Git GitFunc
	// Warnf is the warning output sink (nil = stderr).
	Warnf func(format string, args ...any)
}

// RollbackResult is the result report of Rollback: Result plus the generation transition From→To.
// On failure the partial result carries From == To == the current generation.
type RollbackResult struct {
	Result
	From int // current generation N before rolling back
	To   int // previous generation N-1 rolled back to
}

// Rollback reverts a home-mode profile to one generation earlier. It converges the FS from current
// generation N to N-1 (stale-removing N∖N-1, re-placing N-1) and moves the profile pointer last,
// since moving it first would shift the stale-removal baseline.
func Rollback(opts RollbackOptions) (*RollbackResult, error) {
	warnf := opts.Warnf
	if warnf == nil {
		warnf = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}
	listFn := opts.ListGenerations
	if listFn == nil {
		listFn = nixEnvListGenerations
	}
	switchFn := opts.SwitchGeneration
	if switchFn == nil {
		switchFn = nixEnvSwitchGeneration
	}

	// 1. fix profileDir (resolve root → layout · preamble shared with apply).
	prof, root, err := ProfileFor(ProfileOptions{
		Name: opts.Name, RootKind: opts.RootKind, FixedRoot: opts.FixedRoot,
		RootOverride: opts.RootOverride, WorkDir: opts.WorkDir, StateDir: opts.StateDir, Git: opts.Git,
	})
	if err != nil {
		return nil, err
	}

	// If profileDir is absent, apply has never run → no generation to roll back to.
	if _, err := os.Stat(prof.Dir); err != nil {
		return nil, fmt.Errorf("layat: no profile (apply has never run): %s", prof.Dir)
	}

	// 2. serialize with concurrent apply / rollback via a blocking flock.
	l, err := acquireProfileLock(prof.Dir, true)
	if err != nil {
		return nil, err
	}
	defer func() { _ = l.Release() }()

	// 3. identify current generation N and previous generation N-1 from the generation list.
	gens, err := listFn(prof.Profile)
	if err != nil {
		return nil, err
	}
	curIdx := -1
	for i, g := range gens {
		if g.Current {
			curIdx = i
		}
	}
	if curIdx < 0 {
		return nil, fmt.Errorf("layat: cannot identify the current generation (profile: %s)", prof.Profile)
	}
	if curIdx == 0 {
		return nil, fmt.Errorf("layat: no previous generation (this is the oldest generation, cannot rollback)")
	}
	cur := gens[curIdx]
	prev := gens[curIdx-1]

	// 4. baseline = current generation N's manifest (current FS state) / target = previous generation N-1's manifest.
	baseline, err := manifest.Load(prof.Profile)
	if err != nil {
		return nil, fmt.Errorf("layat: cannot read the current generation's manifest: %w", err)
	}
	target, err := manifest.Load(paths.GenerationLink(prof.Profile, prev.Number))
	if err != nil {
		return nil, fmt.Errorf("layat: cannot read the previous generation's manifest (generation %d): %w", prev.Number, err)
	}

	// 5. compute the plan for N∖N-1 stale removal · N-1 entry re-placement with the planner (reusing the apply engine with (baseline, target) substituted).
	plan, err := planner.Compute(baseline, target, root, planner.OSFS, planner.Options{})
	if err != nil {
		return nil, err
	}

	// Generation numbers come from the listing above; GenAfter is set after the pointer move.

	// 6. reflect the plan onto the FS in Apply's stages: PreRemove, symlinks, place-once copies
	//    (recopy always off), then stale removal. There is no Backup stage.
	a := &applier{opts: Options{Warnf: warnf}, result: &Result{Root: root, ProfileDir: prof.Dir, Profile: prof.Profile}}
	a.profile = prof
	a.root = root
	a.result.GenBefore = intPtr(cur.Number)
	// Full inventory = the generation being rolled back to (its entries are the FS end state).
	a.result.Entries = target.Entries
	a.recordRemovalPlan(plan)
	// A failure in any stage unwinds this call's journaled FS writes and returns the partial
	// result, with GenAfter pinned at the unmoved current generation.
	fail := func(err error) (*RollbackResult, error) {
		res, _ := a.fail(plan, err)
		res.GenAfter = intPtr(cur.Number)
		return &RollbackResult{Result: *res, From: cur.Number, To: cur.Number}, err
	}
	if len(plan.Conflicts) > 0 {
		// Same partial-result contract as Apply's conflict stop: structured conflicts and
		// everything unreached.
		a.result.Conflicts = plan.Conflicts
		return fail(reportConflicts(warnf, plan.Conflicts))
	}
	a.emitWarnings(plan.Warnings, false)
	if err := a.runJournaled(func() error { return a.preRemove(plan.PreRemove) }); err != nil {
		return fail(err)
	}
	if err := a.runJournaled(func() error { return a.place(plan.Place) }); err != nil {
		return fail(err)
	}
	if err := a.runJournaled(func() error { return a.materializeCopies(plan, false) }); err != nil {
		return fail(err)
	}
	if err := a.runJournaled(func() error { return a.removeStale(plan.Remove) }); err != nil {
		return fail(err)
	}

	// 7. finally move the profile pointer to N-1. A failure here is not unwound; re-running
	//    Rollback retries it.
	if err := switchFn(prof.Profile, prev.Number); err != nil {
		// Every planned FS action already succeeded, so the failure is not entry-scoped
		// (FailedTarget / Unreached stay empty) — but the partial result is still returned.
		return fail(fmt.Errorf("layat: failed to move the profile pointer (--switch-generation %d): %w", prev.Number, err))
	}
	a.discardJournal()
	a.result.GenAfter = intPtr(prev.Number)

	return &RollbackResult{Result: *a.result, From: cur.Number, To: prev.Number}, nil
}

// nixEnvListGenerations is the default generation-list retrieval (nix-env --profile <p> --list-generations).
// It captures and parses stdout (ingesting machine-readable output).
func nixEnvListGenerations(profileLink string) ([]Generation, error) {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return nil, fmt.Errorf("nix-env is not on PATH: %w", err)
	}
	cmd := exec.Command("nix-env", "--profile", profileLink, "--list-generations")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		trimmed := strings.TrimSpace(stderr.String())
		if trimmed != "" {
			return nil, fmt.Errorf("layat: nix-env --list-generations failed: %w\n%s", err, trimmed)
		}
		return nil, fmt.Errorf("layat: nix-env --list-generations failed: %w", err)
	}
	return parseGenerations(stdout.String())
}

// nixEnvSwitchGeneration is the default profile pointer move (nix-env --profile <p> --switch-generation <gen>).
// nix output is routed to stderr (stdout is reserved for machine-readable output).
func nixEnvSwitchGeneration(profileLink string, gen int) error {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("nix-env is not on PATH: %w", err)
	}
	cmd := exec.Command("nix-env", "--profile", profileLink, "--switch-generation", strconv.Itoa(gen))
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// parseGenerations parses the output of `nix-env --list-generations`. Each line is of the
// form "<number>   <timestamp>   [(current)]" (leading and separating whitespace is variable).
func parseGenerations(out string) ([]Generation, error) {
	var gens []Generation
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("layat: cannot parse a line of the generation list: %q", line)
		}
		g := Generation{Number: n}
		end := len(fields)
		if fields[end-1] == "(current)" {
			g.Current = true
			end--
		}
		g.Date = strings.Join(fields[1:end], " ")
		gens = append(gens, g)
	}
	return gens, nil
}
