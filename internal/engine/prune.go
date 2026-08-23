package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/yasunori0418/nput/internal/lock"
	"github.com/yasunori0418/nput/internal/paths"
)

// prune removes the profile series left behind by a root that no longer exists
// (→ ADR-0034, ADR-0036 §3, DSG-096dc893-21f4-45e3-9347-986e9275b4d1).
//
// It touches no placed artifact and thins no generations: the only verdict is
// whether the absolute root path the backref .root records still exists. A
// series whose root does exist is kept whatever else it looks like, and a
// series whose verdict cannot be reached (the backref unusable, its contents
// unlistable, the root's stat failing for a reason other than non-existence) is
// kept with a warning — no path leads from missing evidence to deletion
// (→ REQ-c44433a1-7ee7-459a-9aae-7cc42166876f).

// defaultSystemDir is the system-mode profile base (→ ADR-0036 §3). Unlike the
// user state base it is a finished base: it does not go through paths.Base.
const defaultSystemDir = "/nix/var/nix/profiles/nput"

// PruneSkipReason is why a series was left alone. Only failures appear here; a
// series kept because its root exists is not a failure and is not reported
// (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1).
type PruneSkipReason string

const (
	// PruneSkipLocked is a series holding a <name> profileDir whose try-lock
	// could not be taken — held by another process, or undecidable.
	PruneSkipLocked PruneSkipReason = "locked"
	// PruneSkipBackrefUnreadable is a series whose backref .root could not be
	// turned into a root path (unreadable, empty, or not absolute).
	PruneSkipBackrefUnreadable PruneSkipReason = "backref-unreadable"
	// PruneSkipSeriesUnreadable is a series whose <name> profileDirs could not
	// be listed: neither the set of lock keys nor the set of entities to remove
	// is known, so it is kept the same way as a missing backref.
	PruneSkipSeriesUnreadable PruneSkipReason = "series-unreadable"
	// PruneSkipRootStatFailed is a series whose root could not be statted for a
	// reason other than non-existence.
	PruneSkipRootStatFailed PruneSkipReason = "root-stat-failed"
	// PruneSkipPermissionDenied is a series whose deletion could not even begin
	// for want of privilege (→ ADR-0036 §3, the non-root run over the system base).
	PruneSkipPermissionDenied PruneSkipReason = "permission-denied"
)

// PruneOptions is the input to Prune. Prune names no config: it neither
// evaluates nor builds, and its only inputs are the two bases to scan
// (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1).
type PruneOptions struct {
	// StateDir is <state>; the base is derived through paths.Base
	// (empty = resolved with paths.StateDir).
	StateDir string
	// SystemDir is the finished system base itself — it does not go through
	// paths.Base, the system-mode layout having no nix/profiles/nput under it
	// (→ ADR-0036 §3). Empty = defaultSystemDir. Tests point it at a tmpdir so
	// no run reaches the real /nix/var/nix/profiles.
	SystemDir string
	// DryRun is a side-effect-free preview: it enumerates and judges, takes no
	// lock, and changes nothing (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
	DryRun bool
	// Confirm is the confirmation callback before deletion (nil = --yes path).
	// It is handed a preview of the verdict — the root paths the CLI presents
	// are in its Removed — which is a separate value from what Prune returns:
	// Aborted is set on the returned result, never on the preview.
	Confirm func(*PruneResult) (bool, error)
	// Warnf is the warning output sink (nil = stderr). Every skipped series is
	// reported through it.
	Warnf func(format string, args ...any)
}

// PruneSeries is one <roothash> series. RootHash / Root / Names are the fields
// --dryrun prints and the confirmation prompt lists
// (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
type PruneSeries struct {
	// RootHash is the series directory name.
	RootHash string
	// Root is the absolute root path the backref .root records.
	Root string
	// Names are the <name> profileDirs under the series.
	Names []string
	// Dir is the absolute series directory, and so the only field that says
	// which base the series came from. The <roothash> is a hash of the root
	// path (→ paths.RootHash), so the same root reached in project mode and in
	// system mode with an explicit --root puts the same directory name under
	// both bases; without Dir, neither a Removed entry nor a warning would say
	// which one it meant.
	Dir string
}

// PruneSkipped is a series left alone and why.
type PruneSkipped struct {
	Series PruneSeries
	Reason PruneSkipReason
	Detail string
}

// PruneResult is the result of Prune (in dryrun, the preview).
type PruneResult struct {
	// Removed holds the series deleted — in dryrun, the ones that would be.
	// When the deletion stage fails it holds the series completed before that
	// failure, so the caller can tell what is already gone
	// (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed). That is the only error for
	// which a result comes back at all — see Prune's own contract.
	Removed []PruneSeries
	// Skipped holds the series left alone, with the reason.
	Skipped []PruneSkipped
	DryRun  bool
	Aborted bool
}

// Prune deletes the <roothash> series whose recorded root no longer exists,
// under the user state base and the system base. The field name Removed rather
// than Pruned is deliberate: Result.Pruned already means the empty ancestor
// directories rmdir-ed after a removal (→ REQ-8409db86-a1ba-4053-86dc-588985cc1ca7).
//
// On error the result is nil except when the deletion stage itself failed: only
// then is anything already gone, and the partial result comes back alongside
// the error so the caller can report it. Everything before that stage —
// resolving the state dir, the confirmation callback — fails with nothing
// deleted, so there is no partial state to describe.
func Prune(opts PruneOptions) (*PruneResult, error) {
	warnf := opts.Warnf
	if warnf == nil {
		warnf = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}

	stateDir := opts.StateDir
	if stateDir == "" {
		s, err := paths.StateDir()
		if err != nil {
			return nil, err
		}
		stateDir = s
	}
	systemDir := opts.SystemDir
	if systemDir == "" {
		systemDir = defaultSystemDir
	}

	res := &PruneResult{DryRun: opts.DryRun}

	// 1. judge both bases. A base that does not exist holds no series and is
	// not reported; a base that exists but cannot be listed is reported by
	// name, and the other base is still processed (→ RISK-f522a51b-db4b-4164-bd20-63fa4a0cb27d).
	var candidates []PruneSeries
	for _, base := range []string{paths.Base(stateDir), systemDir} {
		found, err := pruneJudgeBase(base, res, warnf)
		if err != nil {
			// paths.ListRootHashSeries already names the base in its own wrap,
			// so this only tags the warning as prune's.
			warnf("nput: prune: %v", err)
			continue
		}
		candidates = append(candidates, found...)
	}

	// 2. dryrun stops at the verdict: no lock is taken and nothing is removed.
	// Here the candidates are the answer, so they go straight into Removed.
	if opts.DryRun {
		res.Removed = candidates
		return res, nil
	}

	// 3. confirm before deleting. The preview handed to Confirm is a separate
	// value: res.Removed goes on to mean "actually deleted", and mutating the
	// object the callback still holds would empty a list the CLI captured to
	// display (→ ADR-0034 §2 の root パス一覧).
	if opts.Confirm != nil {
		// Skipped is shared rather than copied: from here on it only ever
		// grows by append, so the callback's view stays the snapshot it was
		// handed. An in-place update to an existing entry would break that.
		preview := &PruneResult{Removed: candidates, Skipped: res.Skipped}
		ok, err := opts.Confirm(preview)
		if err != nil {
			// Nothing has been deleted, so there is no partial result to hand
			// back; a result carrying the judged candidates in Removed would
			// read as "these were deleted" (→ Reset does the same).
			return nil, err
		}
		if !ok {
			res.Aborted = true
			return res, nil
		}
	}

	// 4. delete. Removed only ever holds series that completed, so a failure
	// partway leaves it holding what finished before it.
	for _, s := range candidates {
		reason, err := pruneRemoveSeries(s)
		if err != nil {
			if reason != "" {
				// The series was left untouched — its lock could not be taken,
				// or its deletion could not begin for want of privilege
				// (→ ADR-0036 §3). It is kept and the rest carries on.
				pruneSkip(res, warnf, s, reason, err)
				continue
			}
			// Either something was already removed, or the reason was not one
			// prune can give meaningful guidance about. The series is neither
			// gone nor untouched, so it is reported as an error and the run
			// stops (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
			return res, err
		}
		res.Removed = append(res.Removed, s)
	}
	return res, nil
}

// pruneJudgeBase enumerates one base and returns the series to delete, adding
// the ones kept for a failed verdict to res.Skipped. A base that does not exist
// yields no series and no error (→ REQ-c44433a1-7ee7-459a-9aae-7cc42166876f).
func pruneJudgeBase(base string, res *PruneResult, warnf func(string, ...any)) ([]PruneSeries, error) {
	listed, err := paths.ListRootHashSeries(base)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	// ListRootHashSeries makes no promise about the order of the series it
	// returns, so the deletion order — and with it what lands in Removed
	// before a failure — is fixed here rather than left to the enumeration.
	sort.Slice(listed, func(i, j int) bool { return listed[i].RootHash < listed[j].RootHash })

	var candidates []PruneSeries
	for _, l := range listed {
		s := PruneSeries{
			RootHash: l.RootHash,
			Root:     l.Root,
			Names:    l.Names,
			Dir:      filepath.Join(base, l.RootHash),
		}
		// Load-bearing: the lock loop below takes the profileDirs in this
		// order, so the prefix a skipped series has to release is
		// deterministic rather than whatever the enumeration happened to give.
		sort.Strings(s.Names)

		// The backref could not be turned into a root path, so there is
		// nothing to judge.
		if l.BackrefErr != nil {
			pruneSkip(res, warnf, s, PruneSkipBackrefUnreadable, l.BackrefErr)
			continue
		}
		// The <name> profileDirs are unknown, so neither the lock keys nor the
		// entities to remove are. Names is nil here just as it is for a series
		// holding nothing but the backref, so this must be decided on NamesErr
		// and never on Names alone (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1).
		if l.NamesErr != nil {
			pruneSkip(res, warnf, s, PruneSkipSeriesUnreadable, l.NamesErr)
			continue
		}

		// The single condition: only a root that does not exist is an orphan.
		// os.Stat follows symlinks, so a dangling symlink root falls out here
		// with no branch of its own; a root replaced by a regular file returns
		// nil and is kept (the kind is not looked at).
		switch _, err := os.Stat(s.Root); {
		case err == nil:
			continue
		case errors.Is(err, fs.ErrNotExist):
			candidates = append(candidates, s)
		default:
			pruneSkip(res, warnf, s, PruneSkipRootStatFailed,
				fmt.Errorf("nput: cannot stat root (%s): %w", s.Root, err))
		}
	}
	return candidates, nil
}

// pruneSkip records a skipped series and warns. Every reason warns: leaving one
// of them silent would break the "report the reason as a warning" rule
// (→ REQ-c44433a1-7ee7-459a-9aae-7cc42166876f).
// The series is named by its directory rather than by the <roothash> alone: the
// same <roothash> can stand under both bases (→ PruneSeries.Dir).
func pruneSkip(res *PruneResult, warnf func(string, ...any), s PruneSeries, reason PruneSkipReason, cause error) {
	res.Skipped = append(res.Skipped, PruneSkipped{Series: s, Reason: reason, Detail: cause.Error()})
	warnf("nput: prune: skipped series %s (%s): %v", s.Dir, reason, cause)
}

// pruneRemoveSeries deletes one series, locks and all. On failure it returns
// the reason the caller may fold into Skipped, or "" when the failure must be
// reported as an error instead. A lock that cannot be taken is always
// skippable — nothing has been touched yet. A removal failure is skippable only
// when nothing had been removed and the reason was privilege: the progress
// state is checked first and the reason only after, so a permission failure
// that came after something was already gone stays an error
// (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1 の 2×2).
func pruneRemoveSeries(s PruneSeries) (PruneSkipReason, error) {
	// A series with no <name> has no lock key at all; the enumeration having
	// succeeded (NamesErr was checked before this point) is what makes an
	// empty Names mean "no key" rather than "key unknown".
	locks := make([]*lock.Lock, 0, len(s.Names))
	releaseAll := func() {
		for _, l := range locks {
			_ = l.Release()
		}
	}
	for _, n := range s.Names {
		dir := filepath.Join(s.Dir, n)
		// The same acquire-and-wrap point Apply / Rollback / Reset share, so
		// the flock failure wording stays single-sourced in the engine.
		l, lerr := acquireProfileLock(dir, false)
		if lerr != nil {
			// Either another process holds the profileDir (lock.ErrLocked) or
			// the lock state could not be decided at all. Neither is allowed
			// to fall through to deletion, so both skip the series whole and
			// release what was already taken. The two need no telling apart
			// here — acquireProfileLock has named the profileDir and carried
			// ErrLocked's own wording, and pruneSkip prefixes the series and
			// the reason, so the error goes back as it came.
			releaseAll()
			return PruneSkipLocked, lerr
		}
		locks = append(locks, l)
	}
	// The locks are held across the removal: a holder that released first
	// would let another process take the profileDir mid-deletion.
	defer releaseAll()

	// <name> first and .root last. A .root removed first would take the series
	// out of the enumeration for good, leaving a half-deleted series that
	// neither prune nor a person can find again. os.RemoveAll on the series
	// directory cannot express this order, so it is not used.
	removedAny := false
	// removalReason decides, for a failure of the removal itself, whether the
	// series may still be called untouched.
	removalReason := func(rerr error) PruneSkipReason {
		if !removedAny && errors.Is(rerr, fs.ErrPermission) {
			return PruneSkipPermissionDenied
		}
		return ""
	}
	for _, n := range s.Names {
		target := filepath.Join(s.Dir, n)
		if rerr := os.RemoveAll(target); rerr != nil {
			return removalReason(rerr),
				fmt.Errorf("nput: cannot remove profile directory of series %s (%s): %w", s.RootHash, target, rerr)
		}
		removedAny = true
	}
	backref := filepath.Join(s.Dir, ".root")
	if rerr := os.Remove(backref); rerr != nil && !errors.Is(rerr, fs.ErrNotExist) {
		return removalReason(rerr),
			fmt.Errorf("nput: cannot remove backref of series %s (%s): %w", s.RootHash, backref, rerr)
	}
	// Past this point the backref is gone, so the series is definitely no
	// longer untouched and the rmdir below can only fail as an error — hence
	// the "" rather than another removalReason call.
	if rerr := os.Remove(s.Dir); rerr != nil {
		return "", fmt.Errorf("nput: cannot remove series %s (%s): %w", s.RootHash, s.Dir, rerr)
	}
	return "", nil
}
