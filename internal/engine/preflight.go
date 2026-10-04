package engine

import (
	"fmt"
	"os"

	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

// checkOutOfStore verifies, just before placement, that every out-of-store entry's link target
// (planner.LinkDest) exists, and fails otherwise so no dangling symlink is created. Store links
// and copy entries are not checked here.
func (a *applier) checkOutOfStore() error {
	for _, e := range a.manifest.Entries {
		if e.SrcKind != manifest.SrcKindOutOfStore {
			continue
		}
		dest := planner.LinkDest(e)
		if _, err := os.Lstat(dest); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("layat: out-of-store link target does not exist (target: %s -> %s); will not create a dangling symlink (→ ADR-0001)", e.Target, dest)
			}
			return fmt.Errorf("layat: cannot check out-of-store link target (target: %s -> %s): %w", e.Target, dest, err)
		}
	}
	return nil
}
