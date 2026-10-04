package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/yasunori0418/layat/internal/planner"
)

// removeStale applies the planner's Remove actions, re-verifying the conservative invariant
// against the real FS before each unlink; drifted targets are kept with a warning. After each
// unlink it prunes ancestors left empty, warning (not failing) when a prune fails.
func (a *applier) removeStale(actions []planner.RemoveAction) error {
	for _, act := range actions {
		if !reverifyStale(act) {
			a.opts.Warnf("layat: keeping stale symlink because it drifted after planning: %s", act.Entry.Target)
			continue
		}
		if err := os.Remove(act.TargetAbs); err != nil {
			return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot remove stale symlink (%s): %w", act.TargetAbs, err))
		}
		a.result.Removed = append(a.result.Removed, act.Entry.Target)
		a.journalRelinkedSymlink(act.TargetAbs, planner.LinkDest(act.Entry))
		if err := a.pruneEmptyAncestors(act.TargetAbs); err != nil {
			a.opts.Warnf("layat: could not prune an empty ancestor directory: %v", err)
		}
	}
	return nil
}

// preRemove removes, before place, the self-recorded stale objects the planner scheduled to clear
// a placement target, folding them into result.Removed / result.Pruned. On drift it aborts instead
// of skipping, and it never walks pruneEmptyAncestors.
func (a *applier) preRemove(actions []planner.RemoveAction) error {
	for _, act := range actions {
		switch act.Kind {
		case planner.RemoveRmdir:
			// os.Remove failing with ENOTEMPTY is the emptiness re-verification. An already-absent
			// dir counts as success but is not folded into result.Pruned. The mode is captured
			// only so undo can recreate an identical directory.
			mode := os.FileMode(0o755)
			if info, lerr := os.Lstat(act.TargetAbs); lerr == nil {
				mode = info.Mode().Perm()
			}
			err := os.Remove(act.TargetAbs)
			switch {
			case err == nil:
				a.result.Pruned = append(a.result.Pruned, act.TargetAbs)
				a.journalRemovedEmptyDir(act.TargetAbs, mode)
			case os.IsNotExist(err):
				// already absent; the Rmdir's goal is met, nothing to report.
			case errors.Is(err, syscall.ENOTEMPTY), errors.Is(err, syscall.EEXIST):
				return fmt.Errorf("layat: directory gained content after planning; cannot migrate this placement target safely (%s); re-run apply to converge", act.TargetAbs)
			default:
				return fmt.Errorf("layat: cannot remove empty directory for migration (%s): %w", act.TargetAbs, err)
			}
		default: // planner.RemoveUnlink
			if !reverifyStale(act) {
				return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: recorded symlink changed after planning; cannot migrate this placement target safely (%s); re-run apply to converge", act.Entry.Target))
			}
			if err := os.Remove(act.TargetAbs); err != nil {
				return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot remove recorded symlink for migration (%s): %w", act.TargetAbs, err))
			}
			a.result.Removed = append(a.result.Removed, act.Entry.Target)
			a.journalRelinkedSymlink(act.TargetAbs, planner.LinkDest(act.Entry))
		}
	}
	return nil
}

// pruneEmptyAncestors walks removedAbs's parent chain toward a.root, rmdir-ing each ancestor
// left empty by a removal into result.Pruned. The walk stops at root, at a non-empty ancestor,
// and at a symlink anywhere in the path from root, re-checked on every iteration.
func (a *applier) pruneEmptyAncestors(removedAbs string) error {
	root := filepath.Clean(a.root)
	dir := filepath.Dir(removedAbs)
	for {
		dir = filepath.Clean(dir)
		if dir == root || !isWithinRoot(root, dir) {
			return nil
		}
		hasSymlink, err := pathHasSymlinkComponent(root, dir)
		if err != nil {
			return err
		}
		if hasSymlink {
			return nil
		}
		// The mode is captured before the removal purely so undo can recreate an identical
		// directory; it plays no role in the TOCTOU-safe emptiness re-check below.
		mode := os.FileMode(0o755)
		if info, lerr := os.Lstat(dir); lerr == nil {
			mode = info.Mode().Perm()
		}
		if err := os.Remove(dir); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			// non-empty dir → ENOTEMPTY on Linux, EEXIST on some BSD/Darwin rmdir(2) implementations;
			// both mean "left something behind, stop here" rather than a real failure.
			if errors.Is(err, syscall.ENOTEMPTY) || errors.Is(err, syscall.EEXIST) {
				return nil
			}
			return fmt.Errorf("layat: cannot remove empty ancestor directory (%s): %w", dir, err)
		}
		a.result.Pruned = append(a.result.Pruned, dir)
		a.journalRemovedEmptyDir(dir, mode)
		dir = filepath.Dir(dir)
	}
}

// pathHasSymlinkComponent reports whether any path component strictly between root and
// dir (dir itself included) is a symlink, lstat-ing each component from root downward
// so an ancestor symlink is caught before Remove(dir) would resolve through it.
func pathHasSymlinkComponent(root, dir string) (bool, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false, fmt.Errorf("layat: cannot resolve %q relative to root (%s): %w", dir, root, err)
	}
	cur := root
	for _, comp := range strings.Split(rel, string(filepath.Separator)) {
		if comp == "" || comp == "." {
			continue
		}
		cur = filepath.Join(cur, comp)
		info, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, fmt.Errorf("layat: cannot lstat ancestor directory (%s): %w", cur, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}
	return false, nil
}

// isWithinRoot reports whether dir is root or a descendant of root, guarding the
// upward walk from ever stepping outside root's subtree (defense in depth alongside
// the dir == root stop condition above).
func isWithinRoot(root, dir string) bool {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// reverifyStale re-checks the conservative invariant on the real FS right before
// unlink: the target must still be a symlink pointing to the recorded dest.
func reverifyStale(act planner.RemoveAction) bool {
	info, err := os.Lstat(act.TargetAbs)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	onDisk, err := os.Readlink(act.TargetAbs)
	if err != nil {
		return false
	}
	return onDisk == planner.LinkDest(act.Entry)
}
