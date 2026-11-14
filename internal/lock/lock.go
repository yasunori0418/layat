// Package lock serializes concurrent apply / reset / rollback on a profileDir via an advisory
// flock, which the OS releases on process exit, so no stale lock remains after a crash.
package lock

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// ErrLocked is wrapped in a FlockError by a non-blocking acquisition (try-lock) when another
// holder is active.
var ErrLocked = errors.New("locked by another process")

// FlockError is a failure to take the flock on the profileDir Dir. Err is the cause: ErrLocked
// for a held try-lock, the syscall error otherwise.
type FlockError struct {
	Dir string
	Err error
}

func (e *FlockError) Error() string {
	return fmt.Sprintf("cannot acquire flock (%s): %v", e.Dir, e.Err)
}

func (e *FlockError) Unwrap() error { return e.Err }

// flock is the flock system call; tests replace it to make the blocking flock fail.
var flock = syscall.Flock

// Lock is an exclusive flock acquired on a profileDir.
type Lock struct {
	f *os.File
}

// Acquire takes an exclusive flock on dir (the profileDir), which must already exist.
// blocking=true waits until acquired; blocking=false returns a FlockError wrapping ErrLocked
// if held. A failure to open dir is not a FlockError.
func Acquire(dir string, blocking bool) (*Lock, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot open the profileDir for flock: %w", err)
	}

	how := syscall.LOCK_EX
	if !blocking {
		how |= syscall.LOCK_NB
	}
	if err := flock(int(f.Fd()), how); err != nil {
		_ = f.Close()
		if !blocking && (errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)) {
			err = ErrLocked
		}
		return nil, &FlockError{Dir: dir, Err: err}
	}
	return &Lock{f: f}, nil
}

// Release releases the flock and closes the file.
func (l *Lock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	err := syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	if cerr := l.f.Close(); err == nil {
		err = cerr
	}
	l.f = nil
	return err
}
