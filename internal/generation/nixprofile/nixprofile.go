// Package nixprofile is the generation backend on nix profiles: Commit and Switch run nix-env,
// List scans the generation links read-only, and the profileDirs are keyed under
// <state>/nix/profiles/layat (user) and /nix/var/nix/profiles/layat (system).
package nixprofile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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

// Commit atomically swaps the profile to linkFarm via `nix-env --profile <profileLink> --set
// <linkFarm>`. nix output goes to stderr because stdout is reserved for machine-readable output.
func (backend) Commit(profileLink, linkFarm string) error {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("nix-env is not on PATH: %w", err)
	}
	cmd := exec.Command("nix-env", "--profile", profileLink, "--set", linkFarm)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// List scans the profile's generation links without running nix-env, which takes a lock in
// the profileDir even to list.
func (backend) List(profileLink string) ([]generation.Generation, error) {
	return generation.ListLinks(profileLink)
}

// Switch moves the profile pointer via `nix-env --profile <profileLink> --switch-generation
// <gen>`. nix output is routed to stderr (stdout is reserved for machine-readable output).
func (backend) Switch(profileLink string, gen int) error {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("nix-env is not on PATH: %w", err)
	}
	cmd := exec.Command("nix-env", "--profile", profileLink, "--switch-generation", strconv.Itoa(gen))
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
