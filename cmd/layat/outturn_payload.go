// outturn_payload.go maps the mutation commands' engine results onto the outturn SubjectResult
// payload: items, changes, generation and warnings. The builders read the same results as the -v
// report and are pure data mapping.
package main

import (
	"fmt"
	"os"

	"github.com/yasunori0418/outturn/go"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

// outturnEntryInfo is the item.info DTO for an entry item: {target, method, subpath}. The src
// store path belongs to change.info instead.
type outturnEntryInfo struct {
	Target  string `json:"target"`
	Method  string `json:"method,omitempty"`
	Subpath string `json:"subpath,omitempty"`
}

// outturnChangeInfo is the change.info DTO: the transition's old / new values. Either side is
// omitted when unknowable.
type outturnChangeInfo struct {
	Old string `json:"old,omitempty"`
	New string `json:"new,omitempty"`
}

// outturnPayload is one subject's result payload, folded into its SubjectResult at emit time.
// TInfo is the owning command's result.info type.
type outturnPayload[TInfo any] struct {
	items      []layatItem
	changes    []layatChange
	generation *outturn.Generation
	warnings   []outturn.Warning // subject-level (not item-borne) warnings
	// info is the per-subject result.info; the mutation commands leave it nil.
	info TInfo
	// itemBorne marks that a failed item already represents the command error, so emit keeps it out
	// of errors[].
	itemBorne bool
}

// attachMutationPayload builds the apply / rollback payload from res and attaches it to s. On a
// build failure it reports to stderr and leaves the subject's items empty.
func attachMutationPayload[TInfo any](s *outturnSubject[TInfo], res *engine.Result, cmdErr error) {
	p, err := mutationPayload[TInfo](res, cmdErr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "layat: could not build the --json payload: %v\n", err)
		return
	}
	s.setPayload(p)
}

// attachResetPayload is attachMutationPayload's reset counterpart.
func attachResetPayload[TInfo any](s *outturnSubject[TInfo], res *engine.ResetResult, cmdErr error) {
	p, err := resetPayload[TInfo](res, cmdErr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "layat: could not build the --json payload: %v\n", err)
		return
	}
	s.setPayload(p)
}

// itemStatuses partitions items by reached state: the failed entry, unreached entries (skipped),
// and conflicts (failed + E_LAYAT_COLLISION). Everything else is success.
type itemStatuses struct {
	failed    string
	failedErr *outturn.Error
	unreached map[string]bool
	conflicts map[string]planner.Conflict
}

// statusFor resolves one target's item status and error under the partition.
func (s *itemStatuses) statusFor(target string) (outturn.ItemStatus, *outturn.Error) {
	if c, ok := s.conflicts[target]; ok {
		return outturn.ItemFailed, &outturn.Error{Code: "E_LAYAT_COLLISION", Message: c.Reason}
	}
	if target == s.failed {
		return outturn.ItemFailed, s.failedErr
	}
	if s.unreached[target] {
		return outturn.ItemSkipped, nil
	}
	return outturn.ItemSuccess, nil
}

// newItemStatuses assembles the partition from the engine's reached-state fields. cmdErr becomes
// the failed entry's error object.
func newItemStatuses(failedTarget string, unreached []string, conflicts []planner.Conflict, cmdErr error) *itemStatuses {
	s := &itemStatuses{failed: failedTarget, unreached: map[string]bool{}}
	for _, t := range unreached {
		s.unreached[t] = true
	}
	if len(conflicts) > 0 {
		s.conflicts = make(map[string]planner.Conflict, len(conflicts))
		for _, c := range conflicts {
			s.conflicts[c.Entry.Target] = c
		}
	}
	if failedTarget != "" && cmdErr != nil {
		e := classifyError(cmdErr)
		s.failedErr = &e
	}
	return s
}

// entryItem renders one manifest entry as an outturn item under the status partition.
func entryItem(e manifest.Entry, statuses *itemStatuses) (layatItem, error) {
	id, err := entryItemID(e.Target)
	if err != nil {
		return layatItem{}, err
	}
	status, itemErr := statuses.statusFor(e.Target)
	return layatItem{
		ID:     id,
		Kind:   "entry",
		Label:  e.Target,
		Status: status,
		Error:  itemErr,
		Info:   &outturnEntryInfo{Target: e.Target, Method: e.Method, Subpath: e.Subpath},
	}, nil
}

// entryChange renders one change for an entry item, deriving the itemId from the target.
func entryChange(target string, kind outturn.ChangeKind, reversible bool, info *outturnChangeInfo) (layatChange, error) {
	id, err := entryItemID(target)
	if err != nil {
		return layatChange{}, err
	}
	return layatChange{Kind: kind, ItemID: id, Reversible: reversible, Info: info}, nil
}

// changeInfoOrNil packs old/new into a change info, or nil when both are unknowable.
func changeInfoOrNil(old, new string) *outturnChangeInfo {
	if old == "" && new == "" {
		return nil
	}
	return &outturnChangeInfo{Old: old, New: new}
}

// mutationPayload maps an apply / rollback engine.Result onto the outturn payload: the full item
// inventory, the changes that happened in op order, the generation observation, and warnings
// (plus W_LAYAT_UNWOUND when the run was rolled back).
func mutationPayload[TInfo any](res *engine.Result, cmdErr error) (*outturnPayload[TInfo], error) {
	statuses := newItemStatuses(res.FailedTarget, res.Unreached, res.Conflicts, cmdErr)
	p := &outturnPayload[TInfo]{itemBorne: res.FailedTarget != "" || len(res.Conflicts) > 0}

	// Items: the new manifest's entries, then stale-removed old entries it does not shadow.
	inventory := map[string]bool{}
	for _, e := range res.Entries {
		item, err := entryItem(e, statuses)
		if err != nil {
			return nil, err
		}
		p.items = append(p.items, item)
		inventory[e.Target] = true
	}
	oldEntries := map[string]manifest.Entry{}
	for _, e := range res.RemovalEntries {
		if inventory[e.Target] {
			continue
		}
		if _, seen := oldEntries[e.Target]; seen {
			continue
		}
		oldEntries[e.Target] = e
		item, err := entryItem(e, statuses)
		if err != nil {
			return nil, err
		}
		p.items = append(p.items, item)
		inventory[e.Target] = true
	}

	// Changes. oldDest is a target's dest before this run: the pre-re-link readlink for replaces,
	// the recorded dest for removals.
	removalByTarget := map[string]manifest.Entry{}
	for _, e := range res.RemovalEntries {
		removalByTarget[e.Target] = e
	}
	oldDest := func(target string) string {
		if d, ok := res.ReplacedDests[target]; ok {
			return d
		}
		if e, ok := removalByTarget[target]; ok {
			return planner.LinkDest(e)
		}
		return ""
	}
	newDest := map[string]string{}
	for _, e := range res.Entries {
		newDest[e.Target] = planner.LinkDest(e)
	}
	removed := map[string]bool{}
	for _, t := range res.Removed {
		removed[t] = true
	}
	rePlaced := map[string]bool{}
	addChange := func(target string, kind outturn.ChangeKind, reversible bool, info *outturnChangeInfo) error {
		c, err := entryChange(target, kind, reversible, info)
		if err != nil {
			return err
		}
		p.changes = append(p.changes, c)
		return nil
	}
	place := func(targets []string, modifyAlways, reversible, skipNoop bool) error {
		for _, t := range targets {
			rePlaced[t] = true
			kind := outturn.ChangeAdd
			old := ""
			if modifyAlways || removed[t] {
				// A re-link or unlink-then-re-place is one modify from old to new.
				kind = outturn.ChangeModify
				old = oldDest(t)
			}
			if skipNoop && old != "" && old == newDest[t] {
				// A re-link to the same dest is a noop and is not a change.
				continue
			}
			if err := addChange(t, kind, reversible, changeInfoOrNil(old, newDest[t])); err != nil {
				return err
			}
		}
		return nil
	}
	if err := place(res.Placed, false, true, false); err != nil {
		return nil, err
	}
	if err := place(res.Replaced, true, true, true); err != nil {
		return nil, err
	}
	if err := place(res.Copied, false, true, false); err != nil {
		return nil, err
	}
	// A recopy overwrite is irreversible and has no old value.
	if err := place(res.Recopied, true, false, false); err != nil {
		return nil, err
	}
	for _, t := range res.Removed {
		if rePlaced[t] {
			continue // coalesced into the modify above
		}
		if err := addChange(t, outturn.ChangeRemove, true, changeInfoOrNil(oldDest(t), "")); err != nil {
			return nil, err
		}
	}

	// Generation: the run's before / after observation.
	p.generation = &outturn.Generation{Profile: res.Profile, Before: res.GenBefore, After: res.GenAfter}

	p.warnings = attachWarnings(p.items, res.Warnings)
	if res.Unwound {
		p.warnings = append(p.warnings, outturn.Warning{
			Code:    "W_LAYAT_UNWOUND",
			Message: "the undo journal rolled this run's filesystem changes back after the failure; the listed changes did not survive on disk",
		})
	}
	return p, nil
}

// resetPayload maps an engine.ResetResult onto the outturn payload: the selected entries as items,
// the removals as changes (copy deletions irreversible), and no generation. An aborted run has no
// changes.
func resetPayload[TInfo any](res *engine.ResetResult, cmdErr error) (*outturnPayload[TInfo], error) {
	statuses := newItemStatuses(res.FailedTarget, res.Unreached, nil, cmdErr)
	p := &outturnPayload[TInfo]{itemBorne: res.FailedTarget != ""}

	byTarget := map[string]manifest.Entry{}
	for _, e := range res.Entries {
		item, err := entryItem(e, statuses)
		if err != nil {
			return nil, err
		}
		p.items = append(p.items, item)
		byTarget[e.Target] = e
	}
	if !res.Aborted {
		for _, t := range res.RemovedSymlinks {
			info := changeInfoOrNil(planner.LinkDest(byTarget[t]), "")
			c, err := entryChange(t, outturn.ChangeRemove, true, info)
			if err != nil {
				return nil, err
			}
			p.changes = append(p.changes, c)
		}
		for _, t := range res.RemovedCopies {
			// No info: a copy deletion destroys untracked on-disk content.
			c, err := entryChange(t, outturn.ChangeRemove, false, nil)
			if err != nil {
				return nil, err
			}
			p.changes = append(p.changes, c)
		}
	}

	p.warnings = attachWarnings(p.items, res.Warnings)
	return p, nil
}

// attachWarnings attaches each planner warning to its target's item, or returns it as a
// subject-level warning when the target is outside the inventory. items is mutated in place.
func attachWarnings(items []layatItem, warnings []planner.Warning) []outturn.Warning {
	itemIdx := map[string]int{}
	for i, it := range items {
		itemIdx[it.Info.Target] = i
	}
	var subject []outturn.Warning
	for _, w := range warnings {
		nw := outturnWarning(w)
		if i, ok := itemIdx[w.Target]; ok {
			items[i].Warnings = append(items[i].Warnings, nw)
			continue
		}
		subject = append(subject, nw)
	}
	return subject
}

// outturnWarning translates one planner warning into a W_LAYAT_* outturn warning. The message
// mirrors the stderr text; the target rides in detail.
func outturnWarning(w planner.Warning) outturn.Warning {
	var code, msg string
	switch w.Kind {
	case planner.WarnForeignReplace:
		code, msg = "W_LAYAT_FOREIGN_SYMLINK", "overwriting an unrecorded symlink (foreign; last-wins)"
	case planner.WarnStaleMismatch:
		code, msg = "W_LAYAT_STALE_MISMATCH", "keeping stale symlink because it mismatches the record"
	case planner.WarnStaleNonSymlink:
		code, msg = "W_LAYAT_STALE_NON_SYMLINK", "keeping stale target because it is not a symlink"
	case planner.WarnCopyOrphan:
		code, msg = "W_LAYAT_COPY_ORPHAN", "copy entry vanished but the target is not removed (orphan; clear it with reset)"
	case planner.WarnCopyForeign:
		code, msg = "W_LAYAT_COPY_FOREIGN", "skipped copy because a real file already exists at the copy target (foreign; place-once)"
	default:
		// Fallback for an unknown planner.WarnKind.
		code, msg = "W_LAYAT_WARNING", "unclassified planner warning"
	}
	return outturn.Warning{Code: code, Message: msg, Detail: map[string]any{"target": w.Target}}
}
