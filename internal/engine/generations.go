package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yasunori0418/layat/internal/generation"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
	"github.com/yasunori0418/layat/internal/planner"
)

// Generation operations go through the generation backend on <profileDir>/profile: rollback
// (List + Switch) lives here, the commit in Apply.

// errNoBackend is returned by every entry point that resolves a profileDir without a backend.
var errNoBackend = errors.New("layat: internal error: no generation backend configured")

// ProfileOptions is the input for fixing the profileDir (shared by Rollback / list-generations).
type ProfileOptions struct {
	Name         string
	RootKind     string
	FixedRoot    string
	RootOverride string
	WorkDir      string
	StateDir     string
	// SystemDir overrides the system-mode profile base (empty = Backend.SystemBase).
	SystemDir string
	Git       GitFunc
	// Backend owns the user and system bases the profileDir is keyed under. Required: nil is an error.
	Backend generation.Backend
}

// ProfileFor resolves root and fixes the profileDir layout (same shape as apply's preamble).
// Used in common by non-building rollback / list-generations for flock / generation reads.
func ProfileFor(opts ProfileOptions) (paths.Profile, string, error) {
	if opts.Backend == nil {
		return paths.Profile{}, "", errNoBackend
	}
	root, err := resolveRoot(opts.RootKind, opts.FixedRoot, opts.RootOverride, opts.WorkDir, opts.Git)
	if err != nil {
		return paths.Profile{}, "", err
	}
	// System mode keys under the system base; it never resolves the user <state>.
	var base string
	switch {
	case opts.RootKind == manifest.RootKindSystem && opts.SystemDir != "":
		base = opts.SystemDir
	case opts.RootKind == manifest.RootKindSystem:
		base = opts.Backend.SystemBase()
	default:
		stateDir := opts.StateDir
		if stateDir == "" {
			stateDir, err = paths.StateDir()
			if err != nil {
				return paths.Profile{}, "", err
			}
		}
		base = opts.Backend.UserBase(stateDir)
	}
	prof := paths.Resolve(base, opts.Name, opts.RootKind, root, opts.RootOverride != "")
	return prof, root, nil
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

// RollbackOptions is the input to Rollback. Which root kinds may roll back is the CLI's
// decision; the engine resolves the profileDir regardless of rootKind and converges to the previous generation.
type RollbackOptions struct {
	Name         string
	RootKind     string
	FixedRoot    string
	RootOverride string
	WorkDir      string
	StateDir     string
	// SystemDir overrides the system-mode profile base (empty = Backend.SystemBase).
	SystemDir string
	// SystemStateDir overrides the system-mode record base <state> (empty = /var/lib).
	SystemStateDir string
	// Geteuid reports the effective uid system mode checks for 0 (nil = os.Geteuid).
	Geteuid func() int

	// Backend lists and switches the generations and owns the profile base. Required: nil is an error.
	Backend generation.Backend
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

// Rollback reverts a home- or system-mode profile to one generation earlier. It converges the FS from the
// baseline N to N-1 (stale-removing N∖N-1, re-placing N-1), records N-1 as last applied, and moves
// the profile pointer last, since moving it first would shift the stale-removal baseline.
func Rollback(opts RollbackOptions) (*RollbackResult, error) {
	warnf := opts.Warnf
	if warnf == nil {
		warnf = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}

	// 1. fix profileDir (resolve root → layout · preamble shared with apply).
	prof, root, err := ProfileFor(ProfileOptions{
		Name: opts.Name, RootKind: opts.RootKind, FixedRoot: opts.FixedRoot,
		RootOverride: opts.RootOverride, WorkDir: opts.WorkDir, StateDir: opts.StateDir,
		SystemDir: opts.SystemDir, Git: opts.Git, Backend: opts.Backend,
	})
	if err != nil {
		return nil, err
	}
	if err := requireRoot(opts.RootKind, opts.Geteuid); err != nil {
		return nil, err
	}
	rec, err := recordFor(opts.StateDir, opts.SystemStateDir, opts.Name, opts.RootKind, root, opts.RootOverride != "")
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
	gens, err := opts.Backend.List(prof.Profile)
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

	// 4. baseline = what was last applied (the record, else current generation N) / target =
	//    previous generation N-1's manifest.
	b, err := loadBaseline(rec.File, prof.Profile, warnf)
	if err != nil {
		return nil, err
	}
	if b.profileUnreadable() {
		return nil, fmt.Errorf("layat: cannot read the current generation's manifest: %w", b.profileErr)
	}
	baseline := b.manifest
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
	a := &applier{opts: Options{Warnf: warnf, Origin: recordedOrigin(b.record)}, result: &Result{Root: root, ProfileDir: prof.Dir, Profile: prof.Profile}}
	a.profile = prof
	a.record = rec
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
	if err := a.runJournaled(func() error { return a.recordPlacement(paths.GenerationLink(prof.Profile, prev.Number)) }); err != nil {
		return fail(err)
	}

	// 7. finally move the profile pointer to N-1. A failure here is not unwound; re-running
	//    Rollback retries it.
	if err := opts.Backend.Switch(prof.Profile, prev.Number); err != nil {
		// Every planned FS action already succeeded, so the failure is not entry-scoped
		// (FailedTarget / Unreached stay empty) — but the partial result is still returned.
		return fail(fmt.Errorf("layat: failed to move the profile pointer (switch to generation %d): %w", prev.Number, err))
	}
	a.discardJournal()
	a.result.GenAfter = intPtr(prev.Number)

	return &RollbackResult{Result: *a.result, From: cur.Number, To: prev.Number}, nil
}
