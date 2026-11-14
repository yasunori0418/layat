// Package gitutil resolves the git toplevel for project mode root resolution.
package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ErrNotInRepo is wrapped when dir is outside a git repository and the toplevel cannot be resolved.
var ErrNotInRepo = errors.New("outside a git repository")

// ErrGitNotFound is returned when git is not on PATH.
var ErrGitNotFound = errors.New("git is not on PATH")

// Toplevel returns the absolute git toplevel of dir via `git rev-parse --show-toplevel`.
// It fails if git is missing, dir is outside a repository (ErrNotInRepo), or git fails otherwise.
func Toplevel(dir string) (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", ErrGitNotFound
	}

	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	// git's messages are matched below, so they must not be translated.
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.Contains(stderr.String(), "not a git repository") {
			return "", fmt.Errorf("cannot resolve the git toplevel (%s): %w", dir, ErrNotInRepo)
		}
		if line, _, _ := strings.Cut(strings.TrimSpace(stderr.String()), "\n"); line != "" {
			return "", fmt.Errorf("cannot run git rev-parse (%s): %w: %s", dir, err, line)
		}
		return "", fmt.Errorf("cannot run git rev-parse (%s): %w", dir, err)
	}

	top := strings.TrimSpace(stdout.String())
	if top == "" {
		return "", fmt.Errorf("cannot resolve the git toplevel (%s): %w", dir, ErrNotInRepo)
	}
	return top, nil
}
