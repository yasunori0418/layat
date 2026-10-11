package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yasunori0418/layat/internal/generation"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
	"github.com/yasunori0418/layat/internal/planner"
)

// reset is an FS-only teardown that reverts the baseline's entries (the recorded truth, which a
// rebuilt config can diverge from) to a not-placed state: symlinks under the stale-removal
// invariant, copy targets deleted outright. The profile and generations are untouched.

// ResetOptions is the input to Reset. Reset does not build (it reads the previous generation's manifest).
type ResetOptions struct {
	Name         string
	RootKind     string
	FixedRoot    string
	RootOverride string
	WorkDir      string
	StateDir     string
	Git          GitFunc
	// Backend owns the profile base the profileDir is keyed under. Required: nil is an error.
	Backend generation.Backend

	// Targets narrows the teardown-target targets (root-relative · empty = all entries).
	// Specifying a target not present in the previous generation's manifest is an error.
	Targets []string
	// DryRun is a side-effect-free preview (just computes and returns the removal targets · no flock / confirm / FS deletion).
	DryRun bool
	// Confirm is the confirmation callback before performing deletion (nil = run without confirmation · --yes path / dryrun).
	// It is passed the computed plan; returning false aborts (Result.Aborted = true). The CLI handles the TTY prompt.
	Confirm func(*ResetResult) (bool, error)
	// Warnf is the warning output sink (nil = stderr). Used to surface kept foreign symlinks etc.
	Warnf func(format string, args ...any)
}

// ResetResult is the result of Reset (dryrun is a preview · at run time the actual deletion result).
type ResetResult struct {
	Root            string   // resolved absolute root
	ProfileDir      string   // the fixed profileDir
	RemovedSymlinks []string // removed (in dryrun, to-be-removed) symlink targets
	RemovedCopies   []string // removed (in dryrun, to-be-removed) copy targets
	KeptForeign     []string // symlink targets kept for not satisfying the conservative invariant (foreign / record mismatch)
	Pruned          []string // empty ancestor directories rmdir-ed after a removal (not computed in dryrun)
	DryRun          bool     // was a read-only preview
	Aborted         bool     // aborted at the confirmation prompt

	// Warnings are the planner's entry-scoped warnings in structured form (kind + target), for
	// the CLI to map onto outturn warnings; KeptForeign above stays the preview-oriented view of
	// the same data.
	Warnings []planner.Warning
	// Entries are the selected teardown entries (the previous generation's manifest narrowed by
	// Targets) — reset's full inventory, so the CLI can list every entry as an outturn item with
	// its method/subpath even when it produced no removal.
	Entries []manifest.Entry
	// FailedTarget is the root-relative target whose removal failed ("" when not entry-scoped) and
	// Unreached the planned removals never attempted. Both empty on success; when set, the partial
	// ResetResult is returned alongside the error.
	FailedTarget string
	Unreached    []string
	// GenBefore / GenAfter are the profile generation numbers observed for the run. Reset is an
	// FS-only teardown that never moves the profile pointer, so before == after; nil when the
	// profile link is not a parsable generation link.
	GenBefore *int
	GenAfter  *int
}

// Reset reverts the placed entities of the target entries to a not-placed state. It owns
// profileDir resolution · blocking flock · reading the previous generation's manifest ·
// conservative symlink removal + copy deletion.
func Reset(opts ResetOptions) (*ResetResult, error) {
	warnf := opts.Warnf
	if warnf == nil {
		warnf = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}

	// 1. fix profileDir (resolve root → layout · preamble shared with apply / rollback).
	prof, root, err := ProfileFor(ProfileOptions{
		Name: opts.Name, RootKind: opts.RootKind, FixedRoot: opts.FixedRoot,
		RootOverride: opts.RootOverride, WorkDir: opts.WorkDir, StateDir: opts.StateDir, Git: opts.Git,
		Backend: opts.Backend,
	})
	if err != nil {
		return nil, err
	}
	rec, err := recordFor(opts.StateDir, opts.Name, opts.RootKind, root, opts.RootOverride != "")
	if err != nil {
		return nil, err
	}
	res := &ResetResult{Root: root, ProfileDir: prof.Dir, DryRun: opts.DryRun}

	// With neither a profile (the previous generation link) nor a record, apply has never run = a
	// no-op with zero teardown targets.
	if _, err := os.Stat(prof.Profile); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("layat: cannot check profile (%s): %w", prof.Profile, err)
		}
		if _, err := os.Lstat(rec.File); os.IsNotExist(err) {
			return res, nil
		}
	}

	// Observe the generation once: the FS-only teardown never moves the profile pointer, so the
	// same observation serves as both before and after.
	res.GenBefore = observeGeneration(prof.Profile)
	if res.GenBefore != nil {
		res.GenAfter = intPtr(*res.GenBefore)
	}

	// 2. at run time, serialize with concurrent apply / reset via a blocking flock.
	//    dryrun is read-only, so it does not take a flock.
	if !opts.DryRun {
		l, err := acquireProfileLock(prof.Dir, true)
		if err != nil {
			return nil, err
		}
		defer func() { _ = l.Release() }()
	}

	// 3. read the baseline (the record, else the previous generation) and narrow the target entries.
	b, err := loadBaseline(rec.File, prof.Profile, warnf)
	if err != nil {
		return nil, err
	}
	if b.record == nil && b.profileErr != nil {
		return nil, fmt.Errorf("layat: cannot read the previous generation's manifest (%s): %w", prof.Profile, b.profileErr)
	}
	if b.manifest == nil {
		return res, nil
	}
	prev := b.manifest
	entries, err := selectResetEntries(prev.Entries, opts.Targets)
	if err != nil {
		return nil, err
	}
	res.Entries = entries

	// 4. for symlinks, compute a conservative-invariant removal plan with the planner (next=nil makes all targets remove candidates).
	//    copy is a region the planner never removes, so handle it separately.
	var symEntries, copyEntries []manifest.Entry
	for _, e := range entries {
		if e.Method == manifest.MethodCopy {
			copyEntries = append(copyEntries, e)
		} else {
			symEntries = append(symEntries, e)
		}
	}
	symManifest := &manifest.Manifest{SchemaVersion: prev.SchemaVersion, Root: prev.Root, Entries: symEntries}
	plan, err := planner.Compute(symManifest, nil, root, planner.OSFS, planner.Options{})
	if err != nil {
		return nil, err
	}

	// 5. make the existing copy targets removal candidates (absent ones are a no-op · → docs/spec.md error spec).
	copyTargets := make([]string, 0, len(copyEntries))
	for _, e := range copyEntries {
		targetAbs := filepath.Join(root, filepath.Clean(e.Target))
		if _, err := os.Lstat(targetAbs); err == nil {
			copyTargets = append(copyTargets, targetAbs)
			res.RemovedCopies = append(res.RemovedCopies, e.Target)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("layat: cannot lstat copy target (%s): %w", targetAbs, err)
		}
	}

	// For the preview (dryrun / confirm display), pack the to-be-removed symlinks and the kept foreign.
	res.Warnings = plan.Warnings
	for _, a := range plan.Remove {
		res.RemovedSymlinks = append(res.RemovedSymlinks, a.Entry.Target)
	}
	for _, w := range plan.Warnings {
		if w.Kind == planner.WarnStaleMismatch || w.Kind == planner.WarnStaleNonSymlink {
			res.KeptForeign = append(res.KeptForeign, w.Target)
		}
	}

	// 6. dryrun returns the computed plan and finishes (no FS deletion · no confirm).
	if opts.DryRun {
		return res, nil
	}

	// 7. confirmation (data-loss risk). The CLI handles the TTY prompt / --yes.
	if opts.Confirm != nil {
		proceed, err := opts.Confirm(res)
		if err != nil {
			return nil, err
		}
		if !proceed {
			res.Aborted = true
			return res, nil
		}
	}

	// 8. reflect onto the real FS: staleremove for symlinks (with drift re-verification), then
	//    delete copy targets. A mid-teardown failure returns the partial ResetResult alongside
	//    the error, mirroring Apply's stage-failure contract.
	a := &applier{opts: Options{Warnf: warnf}, result: &Result{Root: root, ProfileDir: prof.Dir}}
	a.profile = prof
	a.root = root
	a.emitWarnings(plan.Warnings, false)
	plannedCopies := res.RemovedCopies // preview view: copy removals planned but not yet attempted
	if err := a.removeStale(plan.Remove); err != nil {
		res.RemovedSymlinks = a.result.Removed
		res.RemovedCopies = nil // the copy stage was never reached
		res.Pruned = a.result.Pruned
		res.FailedTarget = a.result.FailedTarget
		res.Unreached = resetUnreached(plan.Remove, a.result.Removed, a.result.FailedTarget, plannedCopies)
		return res, err
	}
	res.RemovedSymlinks = a.result.Removed // actually removed (excludes those kept due to drift)

	removedCopies := make([]string, 0, len(copyTargets))
	for i, targetAbs := range copyTargets {
		if err := os.RemoveAll(targetAbs); err != nil {
			res.RemovedCopies = removedCopies
			res.Pruned = a.result.Pruned
			res.FailedTarget = plannedCopies[i]
			res.Unreached = plannedCopies[i+1:]
			return res, fmt.Errorf("layat: cannot remove copy target (%s): %w", targetAbs, err)
		}
		removedCopies = append(removedCopies, plannedCopies[i])
		if err := a.pruneEmptyAncestors(targetAbs); err != nil {
			a.opts.Warnf("layat: could not prune an empty ancestor directory: %v", err)
		}
	}
	res.RemovedCopies = removedCopies
	res.Pruned = a.result.Pruned // covers both the symlink half (removeStale) and the copy half above

	if err := recordAfterReset(rec, root, b, entries, plan.Remove, res); err != nil {
		return res, err
	}
	return res, nil
}

// recordAfterReset rewrites the record without the selected entries reset took down, keeping the
// ones it left in place (foreign or drifted) and the record's link-farm. With nothing left the
// record is deleted.
func recordAfterReset(rec paths.Record, root string, b baseline, selected []manifest.Entry, planned []planner.RemoveAction, res *ResetResult) error {
	down := make(map[string]bool, len(selected))
	for _, e := range selected {
		down[e.Target] = true
	}
	for _, t := range res.KeptForeign {
		delete(down, t)
	}
	removed := make(map[string]bool, len(res.RemovedSymlinks))
	for _, t := range res.RemovedSymlinks {
		removed[t] = true
	}
	for _, act := range planned {
		if act.Kind == planner.RemoveUnlink && !removed[act.Entry.Target] {
			delete(down, act.Entry.Target)
		}
	}

	m := *b.manifest
	m.Entries = nil
	for _, e := range b.manifest.Entries {
		if !down[e.Target] {
			m.Entries = append(m.Entries, e)
		}
	}
	if len(m.Entries) == 0 {
		return removeRecord(rec.File)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("layat: cannot encode the manifest for the last-applied record: %w", err)
	}
	return writeRecord(rec, root, b.linkFarm(), raw, time.Now())
}

// resetUnreached lists the planned removals never attempted once a symlink-stage failure
// stopped the run: the unlink targets neither removed nor failed, then every planned copy
// removal. Drift-kept targets are folded in, as Apply's fail() does.
func resetUnreached(planned []planner.RemoveAction, removed []string, failed string, plannedCopies []string) []string {
	done := map[string]bool{failed: true}
	for _, t := range removed {
		done[t] = true
	}
	var out []string
	for _, act := range planned {
		if act.Kind != planner.RemoveUnlink || done[act.Entry.Target] {
			continue
		}
		done[act.Entry.Target] = true
		out = append(out, act.Entry.Target)
	}
	return append(out, plannedCopies...)
}

// selectResetEntries narrows the previous generation's manifest entries by Targets (empty = all entries).
// If a specified target does not exist in the previous generation, it is an error (a target layat did not place is not a teardown target).
func selectResetEntries(entries []manifest.Entry, targets []string) ([]manifest.Entry, error) {
	if len(targets) == 0 {
		return entries, nil
	}
	byTarget := make(map[string]manifest.Entry, len(entries))
	for _, e := range entries {
		byTarget[e.Target] = e
	}
	out := make([]manifest.Entry, 0, len(targets))
	var unknown []string
	for _, t := range targets {
		key := filepath.Clean(t)
		e, ok := byTarget[key]
		if !ok {
			unknown = append(unknown, t)
			continue
		}
		out = append(out, e)
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("layat: reset target not found in the previous generation's manifest (not a target layat placed): %v", unknown)
	}
	return out, nil
}
