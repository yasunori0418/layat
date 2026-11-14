package gitutil

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// On macOS /tmp is a symlink to /private/tmp, so compare against the resolved path.
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return dir
}

func TestToplevelFromSubdir(t *testing.T) {
	root := initRepo(t)
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Toplevel(sub)
	if err != nil {
		t.Fatalf("Toplevel: %v", err)
	}
	got, _ = filepath.EvalSymlinks(got)
	if got != root {
		t.Errorf("Toplevel = %q, want %q", got, root)
	}
}

func TestToplevelOutsideRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	// Use an isolated tmpdir, since an ancestor like $HOME being a git repo would cause false positives.
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	_, err := Toplevel(dir)
	if !errors.Is(err, ErrNotInRepo) {
		t.Fatalf("outside repo: got %v, want ErrNotInRepo", err)
	}
	if want := "cannot resolve the git toplevel (" + dir + "): outside a git repository"; err.Error() != want {
		t.Errorf("outside repo: message %q, want %q", err.Error(), want)
	}
}

// TestToplevelMissingDir: a git failure other than being outside a repository is not
// ErrNotInRepo and names the dir.
func TestToplevelMissingDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := filepath.Join(t.TempDir(), "missing")

	_, err := Toplevel(dir)
	if err == nil {
		t.Fatal("expected error for a missing dir, got nil")
	}
	if errors.Is(err, ErrNotInRepo) {
		t.Errorf("missing dir: %v must not be ErrNotInRepo", err)
	}
	if want := "cannot run git rev-parse (" + dir + "): exit status "; !strings.HasPrefix(err.Error(), want) {
		t.Errorf("missing dir: message %q, want prefix %q", err.Error(), want)
	}
	if strings.Contains(err.Error(), "\n") {
		t.Errorf("missing dir: message %q spans lines", err.Error())
	}
}

func TestToplevelGitNotInPath(t *testing.T) {
	t.Setenv("PATH", "")
	if _, err := Toplevel(t.TempDir()); !errors.Is(err, ErrGitNotFound) {
		t.Fatalf("git-not-in-PATH: got %v, want ErrGitNotFound", err)
	}
}
