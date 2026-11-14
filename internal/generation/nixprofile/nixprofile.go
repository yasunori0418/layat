// Package nixprofile is the generation backend on nix profiles: Commit and Switch run nix-env,
// List scans the generation links read-only, and the profileDirs are keyed under
// <state>/nix/profiles/layat (user) and /nix/var/nix/profiles/layat (system).
package nixprofile

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yasunori0418/layat/internal/generation"
)

// systemBase is the system-mode profile base, a finished base that takes no <state>.
const systemBase = "/nix/var/nix/profiles/layat"

// backend is the nix-env implementation of generation.Backend.
type backend struct{}

// New returns the nix profile backend.
func New() generation.Backend {
	return backend{}
}

// CommandError is a nix-env run that failed. Op is the nix-env operation with its argument
// (`--set`, `--switch-generation N`), Profile the profile link, and Stderr what nix-env wrote to
// stderr, which Error() leaves out because it already went to the terminal.
type CommandError struct {
	Op      string
	Profile string
	Stderr  string
	Err     error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("cannot run nix-env %s (%s): %v", e.Op, e.Profile, e.Err)
}

func (e *CommandError) Unwrap() error { return e.Err }

// runNixEnv runs `nix-env --profile <profileLink> <args...>`; op names the operation in errors.
// nix output goes to stderr because stdout is reserved for machine-readable output; stderr is
// also kept for the CommandError.
func runNixEnv(profileLink, op string, args ...string) error {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("cannot run nix-env %s (%s): %w", op, profileLink, err)
	}
	var stderr bytes.Buffer
	cmd := exec.Command("nix-env", append([]string{"--profile", profileLink}, args...)...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	if err := cmd.Run(); err != nil {
		return &CommandError{Op: op, Profile: profileLink, Stderr: strings.TrimSpace(stderr.String()), Err: err}
	}
	return nil
}

// Commit atomically swaps the profile to linkFarm via `nix-env --profile <profileLink> --set
// <linkFarm>`.
func (backend) Commit(profileLink, linkFarm string) error {
	return runNixEnv(profileLink, "--set", "--set", linkFarm)
}

// List scans the profile's generation links without running nix-env, which takes a lock in
// the profileDir even to list.
func (backend) List(profileLink string) ([]generation.Generation, error) {
	return generation.ListLinks(profileLink)
}

// Switch moves the profile pointer via `nix-env --profile <profileLink> --switch-generation
// <gen>`.
func (backend) Switch(profileLink string, gen int) error {
	n := strconv.Itoa(gen)
	return runNixEnv(profileLink, "--switch-generation "+n, "--switch-generation", n)
}

// UserBase returns <state>/nix/profiles/layat. The home (no --root) profileDir is <name>
// directly under it; the roothash series is <roothash>/<name>.
func (backend) UserBase(stateDir string) string {
	return filepath.Join(stateDir, "nix", "profiles", "layat")
}

// SystemBase returns /nix/var/nix/profiles/layat.
func (backend) SystemBase() string {
	return systemBase
}
