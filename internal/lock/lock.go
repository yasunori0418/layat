// Package lock serializes concurrent apply / reset / rollback on a profileDir via an advisory
// flock, which the OS releases on process exit, so no stale lock remains after a crash.
package lock

import (
	"errors"
	"os"
	"syscall"
)

// ErrLocked is returned by a non-blocking acquisition (try-lock) when another holder is active.
var ErrLocked = errors.New("layat: profileDir is locked by another process")

// Lock is an exclusive flock acquired on a profileDir.
type Lock struct {
	f *os.File
}

// Acquire takes an exclusive flock on dir (the profileDir), which must already exist.
// blocking=true waits until acquired; blocking=false returns ErrLocked if held.
func Acquire(dir string, blocking bool) (*Lock, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	how := syscall.LOCK_EX
	if !blocking {
		how |= syscall.LOCK_NB
	}
	if err := syscall.Flock(int(f.Fd()), how); err != nil {
		_ = f.Close()
		if !blocking && (errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)) {
			return nil, ErrLocked
		}
		return nil, err
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
