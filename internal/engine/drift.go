package engine

import (
	"fmt"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/planner"
)

// Generation skip + lstat drift repair (project mode only): when the new link-farm equals the
// previous generation no generation is committed, and only drifted entries are re-converged.

// generationUnchanged reports whether the new link-farm resolves to the same store path
// as the profile link's current generation. It returns an error if either cannot be
// resolved; the caller then falls back to a normal apply.
func generationUnchanged(profileLink, newLinkFarm string) (bool, error) {
	prev, err := filepath.EvalSymlinks(profileLink)
	if err != nil {
		return false, err
	}
	next, err := filepath.EvalSymlinks(newLinkFarm)
	if err != nil {
		return false, err
	}
	return prev == next, nil
}

// repairDrift re-converges only the drifted entries during a generation skip, without stale
// removal or a generation commit: it applies Backup, re-links PlaceNew / PlaceForeign, and places
// absent copies (or recopies all on recopy). plan.PreRemove is always empty on this path.
func (a *applier) repairDrift(plan planner.Plan, recopy bool) error {
	// Guards the invariant that any PreRemove changes the derivation and so never reaches here.
	if len(plan.PreRemove) > 0 {
		return fmt.Errorf("layat: internal invariant violated: generation-skip drift repair received %d pre-removal(s) (→ ADR-0046, ADR-0047)", len(plan.PreRemove))
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
