package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/planner"
)

// ensureParentDir creates targetAbs's parent directory, wrapping any failure with the path.
// The directories it creates are not journaled for undo; empty leftovers are swept by
// pruneEmptyAncestors.
func ensureParentDir(targetAbs string) error {
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		return fmt.Errorf("layat: cannot create parent directory (%s): %w", filepath.Dir(targetAbs), err)
	}
	return nil
}

// place executes the planner's Place actions as native symlinks (new or re-link).
// Result op lists are appended only after an action's final FS write succeeds; a failing
// action is recorded as result.FailedTarget instead.
func (a *applier) place(actions []planner.PlaceAction) error {
	for _, act := range actions {
		if err := ensureParentDir(act.TargetAbs); err != nil {
			return a.entryFailed(act.Entry.Target, err)
		}

		if act.Kind == planner.PlaceReplace || act.Kind == planner.PlaceForeign {
			// Re-link is unlink + symlink (no rename-based atomic swap).
			// The foreign-overwrite warning is already emitted via planner.Warnings by emitWarnings.
			prevDest, err := os.Readlink(act.TargetAbs)
			if err != nil {
				return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot read existing symlink before re-link (%s): %w", act.TargetAbs, err))
			}
			if err := os.Remove(act.TargetAbs); err != nil {
				return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot remove existing symlink (%s): %w", act.TargetAbs, err))
			}
			// Journaled before the re-symlink so undo can restore prevDest even if it fails.
			a.journalRelinkedSymlink(act.TargetAbs, prevDest)
			if err := os.Symlink(act.Dest, act.TargetAbs); err != nil {
				return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot create symlink (%s -> %s): %w", act.TargetAbs, act.Dest, err))
			}
			a.result.Replaced = append(a.result.Replaced, act.Entry.Target)
			a.recordReplacedDest(act.Entry.Target, prevDest)
			continue
		}

		// Only PlaceNew reaches here; assert it so a future PlaceKind cannot silently fall
		// through to a fresh-symlink creation it was never classified for.
		if act.Kind != planner.PlaceNew {
			return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: internal: unhandled place kind %d (target: %s)", act.Kind, act.Entry.Target))
		}
		if err := os.Symlink(act.Dest, act.TargetAbs); err != nil {
			return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot create symlink (%s -> %s): %w", act.TargetAbs, act.Dest, err))
		}
		a.journalPlacedSymlink(act.TargetAbs)
		a.result.Placed = append(a.result.Placed, act.Entry.Target)
	}
	return nil
}
