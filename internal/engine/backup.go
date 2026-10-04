package engine

import (
	"fmt"
	"os"

	"github.com/yasunori0418/layat/internal/planner"
)

// backup applies the planner's Backup actions, renaming each foreign object at TargetAbs aside to
// BackupAbs so placement lands on an absent target. BackupAbs is kept after a successful commit,
// and the run aborts if it already exists rather than clobber a prior backup.
func (a *applier) backup(actions []planner.BackupAction) error {
	for _, act := range actions {
		if _, err := os.Lstat(act.BackupAbs); err == nil {
			return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: backup destination already exists (%s); cannot safely back up %s; re-run apply to converge", act.BackupAbs, act.TargetAbs))
		} else if !os.IsNotExist(err) {
			return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot lstat backup destination (%s): %w", act.BackupAbs, err))
		}
		if err := os.Rename(act.TargetAbs, act.BackupAbs); err != nil {
			return a.entryFailed(act.Entry.Target, fmt.Errorf("layat: cannot rename aside for backup (%s -> %s): %w", act.TargetAbs, act.BackupAbs, err))
		}
		// Journaled before placement so undo can restore BackupAbs even if a later stage fails.
		a.journalBackedUp(act.TargetAbs, act.BackupAbs)
		a.result.BackedUp = append(a.result.BackedUp, act.Entry.Target)
		a.opts.Warnf("layat: backed up existing target before placement: %s -> %s", act.TargetAbs, act.BackupAbs)
	}
	return nil
}
