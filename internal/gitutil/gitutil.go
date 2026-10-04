// Package gitutil resolves the git toplevel for project mode root resolution.
package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNotInRepo is returned when dir is outside a git repository and the toplevel cannot be resolved.
var ErrNotInRepo = errors.New("layat: outside a git repository (pass --root to set root explicitly)")

// ErrGitNotFound is returned when git is not on PATH.
var ErrGitNotFound = errors.New("layat: git is not on PATH")

// Toplevel returns the absolute git toplevel of dir via `git rev-parse --show-toplevel`.
// It fails if git is missing or dir is outside a repository.
func Toplevel(dir string) (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", ErrGitNotFound
	}

	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// git started but exited non-zero; the representative case is being outside a repository.
			return "", fmt.Errorf("%w: %s", ErrNotInRepo, strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("layat: failed to run git rev-parse: %w", err)
	}

	top := strings.TrimSpace(stdout.String())
	if top == "" {
		return "", ErrNotInRepo
	}
	return top, nil
}
