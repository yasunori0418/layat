// Package engine is the placement core: it resolves root, fixes the profileDir layout,
// takes a flock, places entries via native FS ops, removes stale links conservatively,
// and commits a generation with `nix-env --set` (injectable so tests do not call nix).
package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/gitutil"
	"github.com/yasunori0418/layat/internal/lock"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
	"github.com/yasunori0418/layat/internal/planner"
)

// CommitFunc is the commit point that records a generation after a successful placement.
// The default is nix-env --profile <profileLink> --set <linkFarm>; tmpdir tests substitute it.
type CommitFunc func(profileLink, linkFarm string) error

// BuildFunc builds the link-farm in-lock with pending (<profileDir>/.pending) as the out-link
// and returns the built link-farm's store path. When nil, opts.LinkFarm is used as pre-built.
type BuildFunc func(pending string) (linkFarm string, err error)

// GitFunc resolves the git toplevel in project mode. The default is gitutil.Toplevel.
type GitFunc func(dir string) (string, error)

// Options is the input to Apply.
type Options struct {
	// LinkFarm is the link-farm directory containing manifest.json and the GC anchor symlink farm.
	// Pre-built link-farm used only on the path that does not pass Build (tmpdir tests).
	LinkFarm string
	// Name is the config name (uniquely identifies a profile; derived from the entrypoint's layat.<name>).
	Name string
	// RootKind is the root kind obtained via eval pre-resolution.
	// Required on the Build path since the manifest is not yet built. When empty, obtained from LinkFarm's manifest.
	RootKind string
	// FixedRoot is the absolute path when rootKind=fixed (from eval pre-resolution's passthru.root).
	// When empty and Build=nil, LinkFarm's manifest.Root.Root is used.
	FixedRoot string
	// RootOverride is the --root override (empty = none). When set, uses the roothash key in all modes.
	RootOverride string
	// WorkDir is the starting point for project-mode git toplevel resolution (empty = os.Getwd).
	WorkDir string
	// StateDir overrides the profile base <state> (empty = resolved via paths.StateDir · mainly for tests).
	StateDir string
	// NoWait makes the flock a try-lock (shellHook path; ErrSkipped if held).
	NoWait bool
	// Recopy is the apply --recopy modifier (unconditionally overwrite/re-copy all copy targets in the config from src).
	// An opt-in path that breaks place-once. The normal apply of the symlink part (stale removal + generation commit) is unchanged.
	Recopy bool
	// Backup is the apply --backup modifier: a foreign occupant that would otherwise be a conflict
	// (or a copy foreign skip) is renamed aside to "<target>.<BackupSuffix>" and the entry placed
	// fresh, instead of stopping.
	Backup bool
	// BackupSuffix is the apply --backup rename suffix. Empty defaults to "layat-backup".
	BackupSuffix string
	// DryRun is a side-effect-free read-only preview (apply --dryrun).
	// When true it runs the planner read-only, packs the plan into Result and returns,
	// taking none of FS writes / --set / flock / pending gcroot. It builds (src resolution) but does not place.
	DryRun bool

	// Build substitutes the in-lock build (nil = use opts.LinkFarm as pre-built).
	Build BuildFunc
	// Git substitutes git toplevel resolution (nil = gitutil.Toplevel).
	Git GitFunc
	// Commit substitutes the generation commit (nil = nix-env --set).
	Commit CommitFunc
	// Warnf is the warning output sink (nil = stderr). Used to surface foreign symlinks etc.
	Warnf func(format string, args ...any)
}

// Result is the result report of Apply (for dryrun / report display · test verification).
type Result struct {
	Root       string   // resolved absolute root path
	ProfileDir string   // the fixed profileDir
	Profile    string   // the profile link (<profileDir>/profile) — the outturn generation.profile value
	Placed     []string // newly placed symlink targets
	Replaced   []string // targets whose existing symlink was re-linked
	Copied     []string // copy targets newly copied via place-once
	Recopied   []string // existing copy targets overwritten/re-copied by --recopy
	Removed    []string // stale-removed targets
	Pruned     []string // empty ancestor directories rmdir-ed after a removal
	BackedUp   []string // targets renamed aside to "<target>.<suffix>" under apply --backup
	Skipped    bool     // skipped on try-lock contention (NoWait path)
	DryRun     bool     // read-only preview (Placed etc. are "to be placed" plans)
	// Conflicts are the planner-detected conflicts, in structured form. Populated on the dryrun
	// path and on the non-dryrun conflict stop, where the partial Result is returned alongside
	// the aggregate error.
	Conflicts []planner.Conflict
	// GenerationSkipped indicates that the project-mode generation skip committed no new
	// generation (omitted --set) and only drifted entries were repaired.
	GenerationSkipped bool

	// Entries is the new manifest's full entry inventory, exposed regardless of whether an
	// entry produced any FS action.
	Entries []manifest.Entry
	// RemovalEntries are the previous-generation entries behind this run's planned symlink
	// removals (pre-removal + stale), recorded at plan time regardless of completion. A
	// method-change target also present in Entries is shadowed there.
	RemovalEntries []manifest.Entry
	// ReplacedDests records, for each re-linked target in Replaced, the symlink destination
	// actually on disk immediately before the re-link (the foreign dest for a foreign replace).
	ReplacedDests map[string]string
	// Warnings are the planner's entry-scoped warnings in structured form (kind + target).
	// The human-readable text is emitted through Warnf alongside.
	Warnings []planner.Warning
	// FailedTarget is the root-relative target of the entry whose FS action failed, "" when
	// the failure was not entry-scoped. Its presence in an op list above means the action was
	// attempted, not completed.
	FailedTarget string
	// Unreached lists the root-relative targets of planned actions never attempted because
	// an earlier failure stopped the run. Empty on success.
	Unreached []string
	// Unwound reports that the undo journal rolled this run's FS writes back after a failure:
	// the op lists above then describe performed-then-reverted actions, not surviving state.
	Unwound bool
	// GenBefore / GenAfter are the profile generation numbers observed at run start / end,
	// nil when unobservable (no profile yet, or a link that is not a generation link).
	GenBefore *int
	GenAfter  *int
}

// ErrSkipped indicates a skip on the NoWait path because another apply is in progress.
var ErrSkipped = lock.ErrLocked

// acquireProfileLock takes the profileDir flock and wraps the error. Shared by Apply /
// Rollback / Reset; NoWait handling and Release stay with each caller.
func acquireProfileLock(dir string, wait bool) (*lock.Lock, error) {
	l, err := lock.Acquire(dir, wait)
	if err != nil {
		return nil, fmt.Errorf("layat: failed to acquire flock (%s): %w", dir, err)
	}
	return l, nil
}

// Apply places the manifest's entries and commits a generation on success, in the order
// flock → in-lock build (opts.Build, or pre-built opts.LinkFarm) → placement → --set →
// .pending removal.
func Apply(opts Options) (*Result, error) {
	a := &applier{opts: opts, result: &Result{}}
	if a.opts.Warnf == nil {
		a.opts.Warnf = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}

	// root kind: on the Build path the manifest is not yet built, so use eval-pre-resolved opts.RootKind.
	// On the pre-built LinkFarm path (tests), read the manifest first to obtain rootKind.
	rootKind := opts.RootKind
	fixedRoot := opts.FixedRoot
	if opts.Build == nil {
		m, err := manifest.Load(opts.LinkFarm)
		if err != nil {
			return nil, fmt.Errorf("layat: cannot read the link-farm's manifest (%s): %w", opts.LinkFarm, err)
		}
		a.manifest = m
		if rootKind == "" {
			rootKind = m.Root.RootKind
		}
		if fixedRoot == "" {
			fixedRoot = m.Root.Root
		}
	}

	// 1. resolve root → fix profileDir.
	prof, root, err := ProfileFor(ProfileOptions{
		Name: opts.Name, RootKind: rootKind, FixedRoot: fixedRoot,
		RootOverride: opts.RootOverride, WorkDir: opts.WorkDir, StateDir: opts.StateDir, Git: opts.Git,
	})
	if err != nil {
		return nil, err
	}
	a.root = root
	a.result.Root = root
	a.profile = prof
	a.result.ProfileDir = a.profile.Dir
	a.result.Profile = a.profile.Profile

	// 1.2 observe the profile generation at run start, and again on every return path (deferred).
	a.result.GenBefore = observeGeneration(a.profile.Profile)
	defer func() { a.result.GenAfter = observeGeneration(a.profile.Profile) }()

	// 1.5 dryrun short-circuits here: no mkdir / flock / placement / --set / pending gcroot;
	//     the planner runs read-only and the plan is packed into Result.
	if opts.DryRun {
		return a.dryRun()
	}

	// 2. prepare profileDir / backref (the flock opens profileDir, so create it first).
	if err := a.ensureProfileDir(); err != nil {
		return nil, err
	}

	// 3. acquire a flock per resolved profileDir and serialize.
	l, err := acquireProfileLock(a.profile.Dir, !opts.NoWait)
	if err != nil {
		if opts.NoWait && errors.Is(err, lock.ErrLocked) {
			a.result.Skipped = true
			return a.result, ErrSkipped
		}
		return nil, err
	}
	defer func() { _ = l.Release() }()

	// 4. build the link-farm in-lock.
	//    Closing the build inside the lock structurally removes .pending contention among concurrent applies.
	if opts.Build != nil {
		linkFarm, err := opts.Build(a.profile.Pending)
		if err != nil {
			return nil, err
		}
		m, err := manifest.Load(linkFarm)
		if err != nil {
			return nil, fmt.Errorf("layat: cannot read the built link-farm's manifest (%s): %w", linkFarm, err)
		}
		a.opts.LinkFarm = linkFarm
		a.manifest = m
	}

	// 5. read the previous generation's manifest (absent = first run = zero stale removals).
	prev := a.loadPrevManifest()

	// 5.5 expose the new manifest's full entry inventory (not just the diff).
	a.result.Entries = a.manifest.Entries

	// 6. compute the place/replace/remove plan with the planner (pure logic · → internal/planner).
	plan, err := planner.Compute(prev, a.manifest, a.root, planner.OSFS, a.plannerOptions())
	if err != nil {
		return nil, err
	}
	a.recordRemovalPlan(plan)
	if len(plan.Conflicts) > 0 {
		// Return the partial Result (full inventory + structured conflicts + everything
		// unreached via fail), not nil.
		a.result.Conflicts = plan.Conflicts
		return a.fail(plan, reportConflicts(a.opts.Warnf, plan.Conflicts))
	}
	a.emitWarnings(plan.Warnings, opts.Recopy)

	// 6.5 check out-of-store link target existence just before placement (no dangling).
	//     Closed before any FS change, so on absence it places nothing and stops with an error.
	if err := a.checkOutOfStore(); err != nil {
		return nil, err
	}

	// 7. project mode only: if the new link-farm equals the previous generation, commit nothing
	//    (omit --set) and only repair drifted entries. Other root kinds commit every time.
	if rootKind == manifest.RootKindProject && prev != nil {
		same, err := generationUnchanged(a.profile.Profile, a.opts.LinkFarm)
		if err != nil {
			// When the previous generation's link-farm cannot be resolved, fall back to the safe side: normal apply (commit a new generation).
			a.opts.Warnf("layat: could not resolve the previous generation's link-farm; recommitting without a generation skip: %v", err)
		} else if same {
			// The repair is journaled like normal placement; with no commit on this path, the
			// journal is discarded as soon as the repair succeeds.
			if err := a.runJournaled(func() error { return a.repairDrift(plan, opts.Recopy) }); err != nil {
				return nil, err
			}
			a.discardJournal()
			a.result.GenerationSkipped = true
			a.cleanupPending()
			return a.result, nil
		}
	}

	// 8. reflect the plan onto the FS: PreRemove, Backup, symlinks, copies, then stale removal.
	//    A failure in any stage unwinds every journaled write of this run and returns the partial
	//    Result with FailedTarget / Unreached / Unwound.
	if err := a.runJournaled(func() error { return a.preRemove(plan.PreRemove) }); err != nil {
		return a.fail(plan, err)
	}
	if err := a.runJournaled(func() error { return a.backup(plan.Backup) }); err != nil {
		return a.fail(plan, err)
	}
	if err := a.runJournaled(func() error { return a.place(plan.Place) }); err != nil {
		return a.fail(plan, err)
	}
	if err := a.runJournaled(func() error { return a.materializeCopies(plan, opts.Recopy) }); err != nil {
		return a.fail(plan, err)
	}
	if err := a.runJournaled(func() error { return a.removeStale(plan.Remove) }); err != nil {
		return a.fail(plan, err)
	}

	// 9. generation commit. A commit failure is not unwound; re-apply converges.
	commit := opts.Commit
	if commit == nil {
		commit = nixEnvCommit
	}
	if err := commit(a.profile.Profile, a.opts.LinkFarm); err != nil {
		// Not entry-scoped, so no FailedTarget / Unreached, but the partial Result is still returned.
		return a.result, fmt.Errorf("layat: generation commit (nix-env --set) failed: %w", err)
	}
	a.discardJournal()

	// 10. remove .pending after --set succeeds (the generation link inherits the gcroot).
	a.cleanupPending()

	return a.result, nil
}

// cleanupPending removes the .pending out-link after --set succeeds (or after a generation skip).
// pending is only created on the build path, so it is removed only on that path.
func (a *applier) cleanupPending() {
	if a.opts.Build == nil {
		return
	}
	if err := os.Remove(a.profile.Pending); err != nil && !os.IsNotExist(err) {
		a.opts.Warnf("layat: could not remove the .pending out-link (%s): %v", a.profile.Pending, err)
	}
}

type applier struct {
	opts     Options
	manifest *manifest.Manifest
	profile  paths.Profile
	root     string
	result   *Result
	journal  []undoOp
}

// plannerOptions translates the apply --backup modifier (opts.Backup / opts.BackupSuffix) into
// planner.Options for planner.Compute.
func (a *applier) plannerOptions() planner.Options {
	return planner.Options{Backup: a.opts.Backup, Suffix: a.opts.BackupSuffix}
}

// runJournaled runs an FS-mutating stage and, if it fails, unwinds the whole journal recorded
// so far in this Apply/Rollback call. On success the journal is kept; the caller discards it.
func (a *applier) runJournaled(stage func() error) error {
	if err := stage(); err != nil {
		a.unwind(err)
		a.result.Unwound = true
		return err
	}
	return nil
}

// recordRemovalPlan records the previous-generation entries behind the plan's symlink
// removals in Result.RemovalEntries at plan time. Rmdir actions carry no entry and are skipped.
func (a *applier) recordRemovalPlan(plan planner.Plan) {
	for _, acts := range [][]planner.RemoveAction{plan.PreRemove, plan.Remove} {
		for _, act := range acts {
			if act.Kind != planner.RemoveUnlink {
				continue
			}
			a.result.RemovalEntries = append(a.result.RemovalEntries, act.Entry)
		}
	}
}

// recordReplacedDest records the on-disk symlink destination a re-linked target pointed at
// immediately before this run replaced it (→ Result.ReplacedDests).
func (a *applier) recordReplacedDest(target, prevDest string) {
	if a.result.ReplacedDests == nil {
		a.result.ReplacedDests = map[string]string{}
	}
	a.result.ReplacedDests[target] = prevDest
}

// entryFailed records target as result.FailedTarget when err is non-nil and returns err
// unchanged. Only the first failure is recorded; target == "" records nothing.
func (a *applier) entryFailed(target string, err error) error {
	if err != nil && target != "" && a.result.FailedTarget == "" {
		a.result.FailedTarget = target
	}
	return err
}

// fail finalizes a mid-run stage failure: it fills result.Unreached with planned targets that are
// neither completed nor the FailedTarget (including the manifest's copy entries under --recopy)
// and returns the partial Result alongside err.
func (a *applier) fail(plan planner.Plan, err error) (*Result, error) {
	done := map[string]bool{}
	if a.result.FailedTarget != "" {
		done[a.result.FailedTarget] = true
	}
	for _, list := range [][]string{
		a.result.Placed, a.result.Replaced, a.result.Copied, a.result.Recopied,
		a.result.Removed, a.result.BackedUp,
	} {
		for _, t := range list {
			done[t] = true
		}
	}
	unreached := func(target string) {
		if target == "" || done[target] {
			return
		}
		done[target] = true
		a.result.Unreached = append(a.result.Unreached, target)
	}
	for _, r := range plan.PreRemove {
		unreached(r.Entry.Target)
	}
	for _, b := range plan.Backup {
		unreached(b.Entry.Target)
	}
	for _, p := range plan.Place {
		unreached(p.Entry.Target)
	}
	for _, c := range plan.Copies {
		unreached(c.Entry.Target)
	}
	if a.opts.Recopy && a.manifest != nil {
		for _, e := range a.manifest.Entries {
			if e.Method == manifest.MethodCopy {
				unreached(e.Target)
			}
		}
	}
	for _, r := range plan.Remove {
		unreached(r.Entry.Target)
	}
	return a.result, err
}

// dryRun is the read-only short-circuit of apply --dryrun: it builds the manifest without a
// gcroot, plans against the previous generation, and packs the plan into Result without flock /
// FS writes / --set. Conflicts are recorded in Result.Conflicts instead of returned as an error.
func (a *applier) dryRun() (*Result, error) {
	// On the build path (CLI) the manifest is not yet obtained, so resolve src via a read-only build.
	// In dryrun the CLI injects `nix build --no-link --print-out-paths` (no gcroot).
	if a.opts.Build != nil {
		linkFarm, err := a.opts.Build(a.profile.Pending)
		if err != nil {
			return nil, err
		}
		m, err := manifest.Load(linkFarm)
		if err != nil {
			return nil, fmt.Errorf("layat: cannot read the built link-farm's manifest (%s): %w", linkFarm, err)
		}
		a.opts.LinkFarm = linkFarm
		a.manifest = m
	}

	prev := a.loadPrevManifest()
	a.result.Entries = a.manifest.Entries
	plan, err := planner.Compute(prev, a.manifest, a.root, planner.OSFS, a.plannerOptions())
	if err != nil {
		return nil, err
	}
	a.recordRemovalPlan(plan)
	a.emitWarnings(plan.Warnings, a.opts.Recopy)

	a.result.DryRun = true
	for _, p := range plan.Place {
		if p.Kind == planner.PlaceNew {
			a.result.Placed = append(a.result.Placed, p.Entry.Target)
		} else {
			a.result.Replaced = append(a.result.Replaced, p.Entry.Target)
		}
	}
	for _, c := range plan.Copies {
		a.result.Copied = append(a.result.Copied, c.Entry.Target)
	}
	for _, r := range plan.PreRemove {
		if r.Kind == planner.RemoveRmdir {
			// Rmdir actions carry no manifest Entry; report the absolute directory path, as
			// result.Pruned does on a real run.
			a.result.Pruned = append(a.result.Pruned, r.TargetAbs)
			continue
		}
		a.result.Removed = append(a.result.Removed, r.Entry.Target)
	}
	for _, r := range plan.Remove {
		a.result.Removed = append(a.result.Removed, r.Entry.Target)
	}
	for _, b := range plan.Backup {
		a.result.BackedUp = append(a.result.BackedUp, b.Entry.Target)
	}
	a.result.Conflicts = plan.Conflicts
	return a.result, nil
}

// resolveRoot resolves the absolute placement root from rootKind (+ the absolute path when
// fixed root). Pure resolution logic shared by Apply /
// Rollback / ProfileFor; on `--root` override it uses the override path regardless of kind.
func resolveRoot(rootKind, fixedRoot, rootOverride, workDir string, git GitFunc) (string, error) {
	if rootOverride != "" {
		return filepath.Abs(rootOverride)
	}
	switch rootKind {
	case manifest.RootKindProject:
		dir := workDir
		if dir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("layat: cannot get cwd: %w", err)
			}
			dir = cwd
		}
		if git == nil {
			git = gitutil.Toplevel
		}
		return git(dir)
	case manifest.RootKindHome:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("layat: cannot resolve $HOME: %w", err)
		}
		return home, nil
	case manifest.RootKindFixed:
		if fixedRoot == "" {
			return "", fmt.Errorf("layat: rootKind=fixed but no root path provided")
		}
		return filepath.Abs(fixedRoot)
	case manifest.RootKindSystem:
		return "", fmt.Errorf("layat: root = systemRoot (system mode) is not implemented")
	case "":
		return "", fmt.Errorf("layat: rootKind is undetermined (eval prefetch or a manifest is required)")
	default:
		return "", fmt.Errorf("layat: unknown rootKind: %q", rootKind)
	}
}

func (a *applier) ensureProfileDir() error {
	if err := os.MkdirAll(a.profile.Dir, 0o755); err != nil {
		return fmt.Errorf("layat: cannot create profileDir (%s): %w", a.profile.Dir, err)
	}
	// Place backref .root at the <roothash> level (reverse-lookup seam for orphan profiles).
	if a.profile.Backref != "" {
		if err := os.MkdirAll(a.profile.BackrefDir, 0o755); err != nil {
			return fmt.Errorf("layat: cannot create backref directory (%s): %w", a.profile.BackrefDir, err)
		}
		if err := os.WriteFile(a.profile.Backref, []byte(a.root+"\n"), 0o644); err != nil {
			return fmt.Errorf("layat: cannot write backref (%s): %w", a.profile.Backref, err)
		}
	}
	return nil
}

// loadPrevManifest reads the manifest.json pointed at by profileDir/profile (the symlink to
// the previous generation's link-farm). On the first run (profile absent) it returns nil
// (zero removal targets).
func (a *applier) loadPrevManifest() *manifest.Manifest {
	if _, err := os.Stat(a.profile.Profile); err != nil {
		return nil
	}
	m, err := manifest.Load(a.profile.Profile)
	if err != nil {
		// Even if the previous generation cannot be read, do not block new placement (just give up stale removal).
		a.opts.Warnf("layat: could not read the previous generation's manifest; skipping stale removal: %v", err)
		return nil
	}
	return m
}

// reportConflicts lists every planner-detected conflict to stderr (warnf), each followed by a
// one-line guidance for its kind, then returns a single count-bearing aggregate error.
func reportConflicts(warnf func(format string, args ...any), conflicts []planner.Conflict) error {
	for _, c := range conflicts {
		warnf("layat: conflict: %s (target: %s)", c.Reason, c.Entry.Target)
		warnf("layat:   → %s", conflictGuidance(c.Kind))
	}
	return fmt.Errorf("layat: %d conflict(s) detected; stopped without placing (see above)", len(conflicts))
}

// conflictGuidance returns the one-line remediation hint for a conflict kind.
func conflictGuidance(kind planner.ConflictKind) string {
	switch kind {
	case planner.ConflictForeignEntity:
		return "move or remove the existing file/directory manually, then re-apply"
	case planner.ConflictForeignAncestor:
		return "check what created this symlink (another config / tool / manual placement)"
	case planner.ConflictSelfContradictoryAncestor:
		return "the manifest keeps both this ancestor and an entry nested beneath it; fix the entry definitions"
	case planner.ConflictCopyStructureMismatch:
		return "the copy entry's structure no longer matches the existing target; fix the entry definition"
	case planner.ConflictDirMigrationFailed:
		return "move or remove the directory's non-migratable contents manually (or use --backup to back up the whole directory), then re-apply"
	case planner.ConflictBackupTargetExists:
		return "a previous --backup was left at \"<target>.<suffix>\"; move or remove it manually, then re-run --backup"
	default:
		return "review the entry definition and the existing target"
	}
}

// emitWarnings emits the planner's non-fatal warnings to opts.Warnf and records them on
// Result.Warnings. With recopy=true the copy foreign skip warning is suppressed on both, since
// recopy overwrites foreign targets too.
func (a *applier) emitWarnings(ws []planner.Warning, recopy bool) {
	for _, w := range ws {
		switch w.Kind {
		case planner.WarnForeignReplace:
			a.opts.Warnf("layat: overwriting an unrecorded symlink (foreign; last-wins): %s", w.Target)
		case planner.WarnStaleMismatch:
			a.opts.Warnf("layat: keeping stale symlink because it mismatches the record: %s", w.Target)
		case planner.WarnStaleNonSymlink:
			a.opts.Warnf("layat: keeping stale target because it is not a symlink: %s", w.Target)
		case planner.WarnCopyOrphan:
			a.opts.Warnf("layat: copy entry vanished but the target is not removed (orphan; clear it with reset): %s", w.Target)
		case planner.WarnCopyForeign:
			if recopy {
				continue
			}
			a.opts.Warnf("layat: skipped copy because a real file already exists at the copy target (foreign; place-once): %s", w.Target)
		}
		a.result.Warnings = append(a.result.Warnings, w)
	}
}
