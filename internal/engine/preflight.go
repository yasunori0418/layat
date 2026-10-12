package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

// checkOutOfStore verifies, just before placement, that every out-of-store and root-relative
// entry's source exists, and fails otherwise so no dangling symlink is created. A root-relative
// source is resolved under the placement root. Store links and copy entries are not checked here.
func (a *applier) checkOutOfStore() error {
	for _, e := range a.manifest.Entries {
		switch e.SrcKind {
		case manifest.SrcKindOutOfStore:
			dest := planner.LinkDest(e)
			if _, err := os.Lstat(dest); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("layat: out-of-store link target does not exist (target: %s -> %s); will not create a dangling symlink", e.Target, dest)
				}
				return fmt.Errorf("layat: cannot check out-of-store link target (target: %s -> %s): %w", e.Target, dest, err)
			}
		case manifest.SrcKindRootRelative:
			src := filepath.Join(a.root, e.Src, e.Subpath)
			if _, err := os.Lstat(src); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("layat: root-relative link source does not exist (target: %s -> %s); will not create a dangling symlink", e.Target, src)
				}
				return fmt.Errorf("layat: cannot check root-relative link source (target: %s -> %s): %w", e.Target, src, err)
			}
		}
	}
	return nil
}
