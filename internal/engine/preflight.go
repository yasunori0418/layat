package engine

import (
	"errors"
	"fmt"
	"io/fs"
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
		var what, src string
		switch e.SrcKind {
		case manifest.SrcKindOutOfStore:
			what, src = "out-of-store link target", planner.LinkDest(e)
		case manifest.SrcKindRootRelative:
			what, src = "root-relative link source", filepath.Join(a.root, e.Src, e.Subpath)
		default:
			continue
		}
		if _, err := os.Lstat(src); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("%s does not exist (target: %s -> %s); will not create a dangling symlink", what, e.Target, src)
			}
			return fmt.Errorf("cannot check %s (target: %s -> %s): %w", what, e.Target, src, err)
		}
	}
	return nil
}
