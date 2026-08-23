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
	// It is handed the judged result — the root paths the CLI presents are in
	// it — and returning false aborts with PruneResult.Aborted set.
	Confirm func(*PruneResult) (bool, error)
	// Warnf is the warning output sink (nil = stderr). Every skipped series is
	// reported through it.
	Warnf func(format string, args ...any)
}

// PruneSeries is one <roothash> series. The three fields are the ones --dryrun
// prints and the confirmation prompt lists (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
type PruneSeries struct {
	// RootHash is the series directory name.
	RootHash string
	// Root is the absolute root path the backref .root records.
	Root string
	// Names are the <name> profileDirs under the series.
	Names []string

	// dir is the absolute series directory. Unexported: it is Prune's own
	// working state, not part of what the CLI reports.
	dir string
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
	// On an error it holds the series completed before the failure, so the
	// caller can tell what is already gone (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
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
			warnf("nput: prune: cannot list profile base (%s): %v", base, err)
			continue
		}
		candidates = append(candidates, found...)
	}
	res.Removed = candidates

	// 2. dryrun stops at the verdict: no lock is taken and nothing is removed.
	if opts.DryRun {
		return res, nil
	}

	// 3. confirm before deleting. The judged result carries the root paths the
	// CLI presents (→ ADR-0034 §2).
	if opts.Confirm != nil {
		ok, err := opts.Confirm(res)
		if err != nil {
			return res, err
		}
		if !ok {
			res.Aborted = true
			res.Removed = nil
			return res, nil
		}
	}

	// 4. delete. Removed is rebuilt as series actually complete, so a failure
	// partway leaves it holding only what finished before it.
	res.Removed = nil
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

	// Enumeration order follows os.ReadDir, which is sorted; sorting the
	// series keeps the deletion order (and so what lands in Removed before a
	// failure) the same across bases.
	sort.Slice(listed, func(i, j int) bool { return listed[i].RootHash < listed[j].RootHash })

	var candidates []PruneSeries
	for _, l := range listed {
		s := PruneSeries{
			RootHash: l.RootHash,
			Root:     l.Root,
			Names:    l.Names,
			dir:      filepath.Join(base, l.RootHash),
		}
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
func pruneSkip(res *PruneResult, warnf func(string, ...any), s PruneSeries, reason PruneSkipReason, cause error) {
	res.Skipped = append(res.Skipped, PruneSkipped{Series: s, Reason: reason, Detail: cause.Error()})
	warnPruneSkip(warnf, s, reason, cause)
}

func warnPruneSkip(warnf func(string, ...any), s PruneSeries, reason PruneSkipReason, cause error) {
	warnf("nput: prune: skipped series %s (%s): %v", s.RootHash, reason, cause)
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
		dir := filepath.Join(s.dir, n)
		l, lerr := lock.Acquire(dir, false)
		if lerr != nil {
			// ErrLocked means another process holds the profileDir; any other
			// acquisition failure leaves the lock state undecided. Neither is
			// allowed to fall through to deletion, so both skip the series
			// whole and release what was already taken.
			releaseAll()
			return PruneSkipLocked, pruneLockSkip(s, dir, lerr)
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
		target := filepath.Join(s.dir, n)
		if rerr := os.RemoveAll(target); rerr != nil {
			return removalReason(rerr),
				fmt.Errorf("nput: cannot remove profile directory of series %s (%s): %w", s.RootHash, target, rerr)
		}
		removedAny = true
	}
	backref := filepath.Join(s.dir, ".root")
	if rerr := os.Remove(backref); rerr != nil && !errors.Is(rerr, fs.ErrNotExist) {
		return removalReason(rerr),
			fmt.Errorf("nput: cannot remove backref of series %s (%s): %w", s.RootHash, backref, rerr)
	}
	removedAny = true
	if rerr := os.Remove(s.dir); rerr != nil {
		return "", fmt.Errorf("nput: cannot remove series %s (%s): %w", s.RootHash, s.dir, rerr)
	}
	return "", nil
}

// pruneLockSkip turns a failed acquisition into the error carried by a locked
// skip. The caller maps it to PruneSkipLocked; it is separate only so the
// message names the profileDir the lock keys on.
func pruneLockSkip(s PruneSeries, dir string, cause error) error {
	if errors.Is(cause, lock.ErrLocked) {
		return fmt.Errorf("nput: profile directory of series %s is locked by another process (%s)", s.RootHash, dir)
	}
	return fmt.Errorf("nput: cannot lock profile directory of series %s (%s): %w", s.RootHash, dir, cause)
}
