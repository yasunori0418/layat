package engine

import (
	"fmt"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/planner"
)

// Generation skip + lstat drift repair (project mode only): when the new link-farm equals the
// previous generation no generation is committed, and only drifted entries are re-converged.

// generationUnchanged reports whether the new link-farm resolves to the same store path as the
// profile's current generation and, when there is a record, the record's link-farm. Without a
// profile it is false; an unresolvable link-farm is an error and the caller falls back to a normal apply.
func generationUnchanged(b baseline, newLinkFarm string) (bool, error) {
	if b.profileErr != nil {
		return false, b.profileErr
	}
	if b.profileLinkFarm == "" {
		return false, nil
	}
	next, err := filepath.EvalSymlinks(newLinkFarm)
	if err != nil {
		return false, err
	}
	return next == b.profileLinkFarm && (b.record == nil || b.record.LinkFarm == next), nil
}

// repairDrift re-converges only the drifted entries during a generation skip, without stale
// removal or a generation commit: it applies Backup, re-links PlaceNew / PlaceForeign, and places
// absent copies (or recopies all on recopy). plan.PreRemove is always empty on this path.
func (a *applier) repairDrift(plan planner.Plan, recopy bool) error {
	// Guards the invariant that any PreRemove changes the derivation and so never reaches here.
	if len(plan.PreRemove) > 0 {
		return fmt.Errorf("layat: internal invariant violated: generation-skip drift repair received %d pre-removal(s)", len(plan.PreRemove))
	}
	if err := a.backup(plan.Backup); err != nil {
		return err
	}
	drifted := make([]planner.PlaceAction, 0, len(plan.Place))
	for _, act := range plan.Place {
		if act.Kind == planner.PlaceReplace {
			continue // symlink pointing at the recorded dest = in-sync. No drift, no re-link needed.
		}
		drifted = append(drifted, act)
	}
	if err := a.place(drifted); err != nil {
		return err
	}
	return a.materializeCopies(plan, recopy)
}
