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

// TestPreflightRootRelativeResolvesAgainstRoot checks that a rootRelative entry is looked up under
// the placement root (not the cwd), and that a missing one fails naming target and resolved path.
func TestPreflightRootRelativeResolvesAgainstRoot(t *testing.T) {
	root := realTempDir(t)
	if err := os.MkdirAll(filepath.Join(root, "dev", "skills", "x", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	present := projectManifest(rootRelativeEntry("dev/skills/x", "sub", ".claude/skills/x"))
	if err := (&applier{manifest: &present, root: root}).checkOutOfStore(); err != nil {
		t.Fatalf("existing root-relative src should pass: %v", err)
	}

	missing := projectManifest(rootRelativeEntry("dev/skills/x", "nosub", ".claude/skills/y"))
	err := (&applier{manifest: &missing, root: root}).checkOutOfStore()
	if err == nil {
		t.Fatal("expected error for missing root-relative src, got nil")
	}
	for _, want := range []string{".claude/skills/y", filepath.Join(root, "dev", "skills", "x", "nosub")} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should contain %q: %v", want, err)
		}
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should report a missing src: %v", err)
	}
}

// TestPreflightRootRelativeNonENOENTError covers the root-relative non-ENOENT branch: an Lstat
// error under a search-denied root is not reported as a missing source.
func TestPreflightRootRelativeNonENOENTError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires non-root to trigger permission error")
	}
	m := projectManifest(rootRelativeEntry("dev/skills/x", ".", ".claude/skills/x"))
	err := (&applier{manifest: &m, root: preflightErr_deniedDir(t)}).checkOutOfStore()
	if err == nil {
		t.Fatal("expected non-ENOENT error from Lstat, got nil")
	}
	if !strings.Contains(err.Error(), "cannot check root-relative link source") || strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should be the non-ENOENT branch: %v", err)
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("wrapped error should be a permission error (EACCES): %v", err)
	}
}
