package lock

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestTryLockConflict(t *testing.T) {
	dir := t.TempDir()

	l1, err := Acquire(dir, true)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	defer func() { _ = l1.Release() }()

	// A try-lock while held returns ErrLocked.
	_, err = Acquire(dir, false)
	if !errors.Is(err, ErrLocked) {
		t.Fatalf("try-lock while held: got %v, want ErrLocked", err)
	}
	var fe *FlockError
	if !errors.As(err, &fe) || fe.Dir != dir {
		t.Fatalf("try-lock while held: got %v, want a FlockError on %s", err, dir)
	}
}

// TestFlockFailureIsFlockError: a failed blocking flock comes back as a FlockError that names the
// profileDir and unwraps to the syscall error.
func TestFlockFailureIsFlockError(t *testing.T) {
	dir := t.TempDir()
	orig := flock
	t.Cleanup(func() { flock = orig })
	flock = func(int, int) error { return syscall.ENOLCK }

	_, err := Acquire(dir, true)
	var fe *FlockError
	if !errors.As(err, &fe) || fe.Dir != dir {
		t.Fatalf("Acquire: got %v, want a FlockError on %s", err, dir)
	}
	if !errors.Is(err, syscall.ENOLCK) {
		t.Errorf("Acquire: %v does not unwrap to ENOLCK", err)
	}
	if errors.Is(err, ErrLocked) {
		t.Errorf("Acquire: %v must not be ErrLocked", err)
	}
	if want := "cannot acquire flock (" + dir + "): "; !strings.HasPrefix(err.Error(), want) {
		t.Errorf("Acquire: message %q, want prefix %q", err.Error(), want)
	}
}

// TestOpenFailureIsNotFlockError: a profileDir that cannot be opened keeps its fs cause and is
// not a FlockError.
func TestOpenFailureIsNotFlockError(t *testing.T) {
	_, err := Acquire(filepath.Join(t.TempDir(), "missing"), true)
	if err == nil {
		t.Fatal("Acquire on a missing dir: got nil error")
	}
	var fe *FlockError
	if errors.As(err, &fe) {
		t.Errorf("Acquire on a missing dir: %v is a FlockError", err)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Acquire on a missing dir: %v does not unwrap to fs.ErrNotExist", err)
	}
}

func TestReleaseAllowsReacquire(t *testing.T) {
	dir := t.TempDir()

	l1, err := Acquire(dir, false)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if err := l1.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	l2, err := Acquire(dir, false)
	if err != nil {
		t.Fatalf("re-Acquire after release: %v", err)
	}
	_ = l2.Release()
}

func TestBlockingWaitsForRelease(t *testing.T) {
	dir := t.TempDir()

	l1, err := Acquire(dir, true)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	acquired := make(chan struct{})
	go func() {
		l2, err := Acquire(dir, true) // blocking: wait until acquired.
		if err != nil {
			t.Errorf("blocking Acquire: %v", err)
			close(acquired)
			return
		}
		_ = l2.Release()
		close(acquired)
	}()

	select {
	case <-acquired:
		t.Fatal("blocking Acquire returned before lock was released")
	case <-time.After(100 * time.Millisecond):
		// Still not acquired = blocking is in effect.
	}

	if err := l1.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	select {
	case <-acquired:
		// Acquired after release.
	case <-time.After(2 * time.Second):
		t.Fatal("blocking Acquire did not proceed after release")
	}
}
