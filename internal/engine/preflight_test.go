package engine

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// preflightErr_deniedDir creates a directory with mode 0000 and returns its path, so
// os.Lstat on any path inside it fails with EACCES. Cleanup restores 0755.
func preflightErr_deniedDir(t *testing.T) string {
	t.Helper()
	denied := filepath.Join(realTempDir(t), "denied")
	if err := os.Mkdir(denied, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(denied, 0o755) })
	if err := os.Chmod(denied, 0o000); err != nil {
		t.Fatal(err)
	}
	return denied
}

// TestPreflightOutOfStoreNonENOENTError covers checkOutOfStore's non-ENOENT branch: an os.Lstat
// error on an out-of-store link target is reported as "cannot check out-of-store link target",
// not as a missing target. EACCES is induced from a search-denied parent directory.
func TestPreflightOutOfStoreNonENOENTError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires non-root to trigger permission error")
	}
	denied := preflightErr_deniedDir(t)
	// LinkDest = <denied>/child; Lstat fails with EACCES because <denied> lacks search permission.
	m := projectManifest(outOfStoreEntry(denied, "child", ".config/nvim"))
	a := &applier{manifest: &m}

	err := a.checkOutOfStore()
	if err == nil {
		t.Fatal("expected non-ENOENT error from Lstat, got nil")
	}
	if !strings.Contains(err.Error(), "cannot check out-of-store link target") {
		t.Errorf("error should be the non-ENOENT branch: %v", err)
	}
	if strings.Contains(err.Error(), "does not exist") {
		t.Errorf("non-ENOENT Lstat error must not be reported as a missing target: %v", err)
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("wrapped error should be a permission error (EACCES): %v", err)
	}
}
