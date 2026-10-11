// Package nixprofile is the generation backend on nix profiles: generations are the
// `nix-env --profile <profileLink>` generations, and the profileDirs are keyed under
// <state>/nix/profiles/layat (user) and /nix/var/nix/profiles/layat (system).
package nixprofile

import (
	"bytes"
	"fmt"
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

// List runs `nix-env --profile <profileLink> --list-generations` and parses its stdout.
func (backend) List(profileLink string) ([]generation.Generation, error) {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return nil, fmt.Errorf("nix-env is not on PATH: %w", err)
	}
	cmd := exec.Command("nix-env", "--profile", profileLink, "--list-generations")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		trimmed := strings.TrimSpace(stderr.String())
		if trimmed != "" {
			return nil, fmt.Errorf("layat: nix-env --list-generations failed: %w\n%s", err, trimmed)
		}
		return nil, fmt.Errorf("layat: nix-env --list-generations failed: %w", err)
	}
	return parseGenerations(stdout.String())
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

// parseGenerations parses the output of `nix-env --list-generations`. Each line is of the
// form "<number>   <timestamp>   [(current)]" (leading and separating whitespace is variable).
func parseGenerations(out string) ([]generation.Generation, error) {
	var gens []generation.Generation
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("layat: cannot parse a line of the generation list: %q", line)
		}
		g := generation.Generation{Number: n}
		end := len(fields)
		if fields[end-1] == "(current)" {
			g.Current = true
			end--
		}
		g.Date = strings.Join(fields[1:end], " ")
		gens = append(gens, g)
	}
	return gens, nil
}
