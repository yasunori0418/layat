package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// entrypointKind distinguishes a flake entrypoint (`layat.<system>.<name>`, addressed via `<flakeRef>#...`)
// from a legacy entrypoint (`layat.<name>`, addressed via `nix ... -f <path> ...`; → ADR-0007, ADR-0032).
type entrypointKind int

const (
	entrypointFlake entrypointKind = iota
	entrypointLegacy
)

// legacyEntrypointNames is the discovery order for legacy entrypoints, tried after flake.nix
// (→ docs/spec.md "entrypoint discovery", ADR-0032).
var legacyEntrypointNames = []string{"shell.nix", "default.nix"}

// entrypoint is a discovered entrypoint: a flake (flake.nix) or a legacy file (shell.nix / default.nix).
// Legacy has no per-system dimension (unlike the flake's `layat.<system>.<name>`; → ADR-0032).
type entrypoint struct {
	kind entrypointKind
	// flakeRef is the flake ref passed to `nix build`/`nix eval` (the absolute path of the directory containing flake.nix).
	flakeRef string
	// legacyPath is the absolute path to the shell.nix / default.nix file, passed to `nix ... -f`.
	legacyPath string
}

// discoverEntrypoint discovers the entrypoint in the order -f explicit → CWD autodiscovery
// (→ docs/spec.md "entrypoint discovery"). Discovery order is flake.nix → shell.nix → default.nix (→ ADR-0032).
func discoverEntrypoint(fileFlag string) (*entrypoint, error) {
	if fileFlag != "" {
		abs, err := filepath.Abs(fileFlag)
		if err != nil {
			return nil, fmt.Errorf("layat: cannot resolve the -f path (%s): %w", fileFlag, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("layat: -f path not found (%s): %w", fileFlag, err)
		}
		if info.IsDir() {
			if fileExists(filepath.Join(abs, "flake.nix")) {
				return &entrypoint{kind: entrypointFlake, flakeRef: abs}, nil
			}
			for _, legacy := range legacyEntrypointNames {
				if p := filepath.Join(abs, legacy); fileExists(p) {
					return &entrypoint{kind: entrypointLegacy, legacyPath: p}, nil
				}
			}
			return nil, fmt.Errorf("layat: no flake.nix / shell.nix / default.nix in the -f directory (%s)", abs)
		}
		switch filepath.Base(abs) {
		case "flake.nix":
			return &entrypoint{kind: entrypointFlake, flakeRef: filepath.Dir(abs)}, nil
		case "shell.nix", "default.nix":
			return &entrypoint{kind: entrypointLegacy, legacyPath: abs}, nil
		default:
			return nil, fmt.Errorf("layat: -f must point to a flake.nix, shell.nix, or default.nix (%s)", abs)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("layat: cannot get the current working directory: %w", err)
	}
	if fileExists(filepath.Join(cwd, "flake.nix")) {
		return &entrypoint{kind: entrypointFlake, flakeRef: cwd}, nil
	}
	for _, legacy := range legacyEntrypointNames {
		if p := filepath.Join(cwd, legacy); fileExists(p) {
			return &entrypoint{kind: entrypointLegacy, legacyPath: p}, nil
		}
	}
	return nil, errors.New("layat: no entrypoint found (no flake.nix / shell.nix / default.nix in the CWD; specify one with -f)")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// currentSystem returns the nix system name of the runtime environment (e.g. aarch64-darwin).
// Because the flake has a system dimension in `layat.<system>.<name>`, the CLI injects the current system (→ ADR-0007).
// Legacy entrypoints have no system dimension and ignore it (→ ADR-0032).
func currentSystem() (string, error) {
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return "", fmt.Errorf("layat: unsupported GOARCH: %s", runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "linux", "darwin":
		return arch + "-" + runtime.GOOS, nil
	default:
		return "", fmt.Errorf("layat: unsupported GOOS: %s", runtime.GOOS)
	}
}

// installableArgs returns the nix args that select `layat.<name><suffix>` for this entrypoint, to be appended
// right after the `eval`/`build` subcommand name. A flake entrypoint yields a single
// "<flakeRef>#layat.<system>.<name><suffix>" installable; a legacy entrypoint (shell.nix / default.nix) has no
// per-system dimension and yields "-f <legacyPath> layat.<name><suffix>" (→ ADR-0032, docs/spec.md addressing).
func (e *entrypoint) installableArgs(system, name, suffix string) []string {
	if e.kind == entrypointLegacy {
		return []string{"-f", e.legacyPath, "layat." + name + suffix}
	}
	return []string{fmt.Sprintf("%s#layat.%s.%s%s", e.flakeRef, system, name, suffix)}
}

// namespaceArgs returns the nix args that select the `layat.<system>` (flake) or `layat` (legacy) namespace,
// used for the batch eval of apply --all / gitignore --all (→ ADR-0024, ADR-0032).
func (e *entrypoint) namespaceArgs(system string) []string {
	if e.kind == entrypointLegacy {
		return []string{"-f", e.legacyPath, "layat"}
	}
	return []string{fmt.Sprintf("%s#layat.%s", e.flakeRef, system)}
}

// label renders a human-readable attr path for error messages (→ wrapEvalErr).
func (e *entrypoint) label(system, name string) string {
	if e.kind == entrypointLegacy {
		return "layat." + name
	}
	return fmt.Sprintf("layat.%s.%s", system, name)
}

// namespaceLabel renders a human-readable namespace path for error messages (→ wrapEvalAllErr).
func (e *entrypoint) namespaceLabel(system string) string {
	if e.kind == entrypointLegacy {
		return "layat"
	}
	return fmt.Sprintf("layat.%s", system)
}

// rootInfo is one config's root info (the value from the batch eval). It has Root only when fixed.
// Targets is the config's normalized target list, read only by apply --all's conflict preflight (→ ADR-0038).
type rootInfo struct {
	RootKind string   `json:"rootKind"`
	Root     string   `json:"root"`
	Targets  []string `json:"targets"`
}

// evalAllRoots gets the config name → rootInfo map for `apply --all` / `gitignore --all`
// in a single `nix eval` (fixing eval process launches at N→1; → docs/spec.md execution flow, ADR-0024).
// It is a cheap eval that does no build and reads only the passthru rootKind + targets (+ root for fixed; → ADR-0038).
func evalAllRoots(e *entrypoint, system string) (map[string]rootInfo, error) {
	// Extract only rootKind + targets (+ root if fixed) from each config under layat.<system>.
	apply := `cs: builtins.mapAttrs (_: c: { rootKind = c.rootKind; targets = c.targets; } // (if c ? root then { root = c.root; } else {})) cs`
	args := append([]string{"eval"}, e.namespaceArgs(system)...)
	args = append(args, "--apply", apply, "--json")
	out, err := runNixCapture(args...)
	if err != nil {
		return nil, wrapEvalAllErr(err, e.namespaceLabel(system))
	}
	var roots map[string]rootInfo
	if err := json.Unmarshal([]byte(out), &roots); err != nil {
		return nil, fmt.Errorf("layat: cannot parse the batch eval result for %s: %w", e.namespaceLabel(system), err)
	}
	return roots, nil
}

// buildManifestStorePath builds the config and returns the link-farm's store path (a read-only path).
// Because gitignore does no placement, it gets only the store path via `--no-link --print-out-paths` without laying down an out-link gcroot.
// Progress goes to stderr and the store path to stdout (→ docs/spec.md output stream discipline).
func buildManifestStorePath(e *entrypoint, system, name string) (string, error) {
	args := append([]string{"build"}, e.installableArgs(system, name, "")...)
	args = append(args, "--no-link", "--print-out-paths")
	out, err := runNixCapture(args...)
	if err != nil {
		return "", err
	}
	store := strings.TrimSpace(out)
	if store == "" {
		return "", fmt.Errorf("layat: cannot obtain the build output path for %s", e.label(system, name))
	}
	return store, nil
}

// evalRoot pre-resolves rootKind (+ the absolute path when fixed root) via a cheap nix eval before build
// (→ docs/spec.md execution flow 1, ADR-0023). This resolves profileDir and establishes the order flock → build.
func evalRoot(e *entrypoint, system, name string) (rootKind, fixedRoot string, err error) {
	args := append([]string{"eval"}, e.installableArgs(system, name, ".rootKind")...)
	args = append(args, "--raw")
	out, err := runNixCapture(args...)
	if err != nil {
		return "", "", wrapEvalErr(err, e.label(system, name))
	}
	rootKind = strings.TrimSpace(out)
	if rootKind == "fixed" {
		rootArgs := append([]string{"eval"}, e.installableArgs(system, name, ".root")...)
		rootArgs = append(rootArgs, "--raw")
		out, err := runNixCapture(rootArgs...)
		if err != nil {
			return "", "", wrapEvalErr(err, e.label(system, name))
		}
		fixedRoot = strings.TrimSpace(out)
	}
	return rootKind, fixedRoot, nil
}

// buildFunc returns the build callback injected into the engine (→ engine.BuildFunc).
// Inside the lock it runs `nix build <installable> --out-link <pending>`, reads the out-link, and returns the store path.
func buildFunc(e *entrypoint, system, name string) func(pending string) (string, error) {
	return func(pending string) (string, error) {
		args := append([]string{"build"}, e.installableArgs(system, name, "")...)
		args = append(args, "--out-link", pending)
		if err := runNixStream(args...); err != nil {
			return "", err
		}
		store, err := os.Readlink(pending)
		if err != nil {
			return "", fmt.Errorf("layat: cannot read the build output out-link (%s): %w", pending, err)
		}
		return store, nil
	}
}

// dryBuildFunc returns the build callback for --dryrun (→ engine.BuildFunc). Unlike a normal build,
// it gets the link-farm's store path via `nix build --no-link --print-out-paths` **without laying down a gcroot (out-link)**
// (dryrun is side-effect-free and creates no pending out-link; → ADR-0011, ADR-0023). The pending argument is unused.
func dryBuildFunc(e *entrypoint, system, name string) func(pending string) (string, error) {
	return func(string) (string, error) {
		return realizeNoLink(e, system, name, "")
	}
}

// realizeNoLink is dryBuildFunc's body, also run directly by apply --all's stage 1 with debugPrefix
// "[<name>] " so its --debug disclosure lines stay attributable while builds run in parallel (→ runNixCapturePrefixed).
func realizeNoLink(e *entrypoint, system, name, debugPrefix string) (string, error) {
	args := append([]string{"build"}, e.installableArgs(system, name, "")...)
	args = append(args, "--no-link", "--print-out-paths")
	out, err := runNixCapturePrefixed(debugPrefix, args...)
	if err != nil {
		return "", err
	}
	store := strings.TrimSpace(out)
	if store == "" {
		return "", fmt.Errorf("layat: nix build --print-out-paths was empty (%s)", e.label(system, name))
	}
	// --print-out-paths may return multiple lines (multi-output). The link-farm is a single output, so take the last line.
	lines := strings.Split(store, "\n")
	return strings.TrimSpace(lines[len(lines)-1]), nil
}

// nixCmdError marks a failed internal nix invocation (eval / build), so the --json error
// classification can map it to E_LAYAT_BUILD without string matching (→ issue #131, ADR-0043 §8).
// It wraps transparently: Error/Unwrap keep the existing message and chain untouched.
type nixCmdError struct{ err error }

func (e *nixCmdError) Error() string { return e.err.Error() }
func (e *nixCmdError) Unwrap() error { return e.err }

// runNixCapture captures and returns nix's stdout (for machine-readable output such as eval).
func runNixCapture(args ...string) (string, error) {
	return runNixCapturePrefixed("", args...)
}

// runNixCapturePrefixed is runNixCapture with debugPrefix put at the head of the --debug disclosure
// line (apply --all's stage 1 passes "[<name>] "; → ADR-0039). nix's own stderr is captured either way.
func runNixCapturePrefixed(debugPrefix string, args ...string) (string, error) {
	if flagDebug {
		fmt.Fprintf(os.Stderr, "%slayat: + nix %s\n", debugPrefix, strings.Join(args, " "))
	}
	cmd := exec.Command("nix", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", &nixCmdError{nixError(args, stderr.String(), err)}
	}
	return stdout.String(), nil
}

// runNixStream streams nix's output to stderr (for build progress; stdout is reserved for machine-readable output; → ADR-0023).
// eval succeeded before build = nix-command/flakes are already enabled, so experimental detection is unnecessary.
func runNixStream(args ...string) error {
	if flagDebug {
		fmt.Fprintf(os.Stderr, "layat: + nix %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command("nix", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return &nixCmdError{fmt.Errorf("layat: nix %s failed: %w", args[0], err)}
	}
	return nil
}

// nixError classifies a nix failure. For experimental-features not enabled it guides the prerequisites,
// and otherwise it returns the raw nix stderr attached without swallowing it (→ ADR-0025 §1).
func nixError(args []string, stderr string, runErr error) error {
	if isExperimentalDisabled(stderr) {
		return experimentalGuidance(stderr)
	}
	trimmed := strings.TrimSpace(stderr)
	if trimmed == "" {
		return fmt.Errorf("layat: nix %s failed: %w", args[0], runErr)
	}
	return fmt.Errorf("layat: nix %s failed:\n%s", args[0], trimmed)
}

// isExperimentalDisabled detects the nix-command / flakes not-enabled error (→ ADR-0025 §1).
func isExperimentalDisabled(stderr string) bool {
	return strings.Contains(stderr, "experimental Nix feature") ||
		strings.Contains(stderr, "experimental-features") ||
		(strings.Contains(stderr, "flakes") && strings.Contains(stderr, "disabled"))
}

// experimentalGuidance builds an error that guides the prerequisites and how to enable them (attaching the raw nix error too).
// The CLI does not add --extra-experimental-features automatically (it will not silently override environment settings; → ADR-0025 §1).
func experimentalGuidance(stderr string) error {
	return fmt.Errorf(`layat: nix's experimental-features are not enabled.
This command internally uses `+"`nix eval`"+` / `+"`nix build`"+` (the new CLI) and flakes,
so experimental-features = nix-command flakes is required.

How to enable (either one):
  - Append to ~/.config/nix/nix.conf or /etc/nix/nix.conf:
      experimental-features = nix-command flakes
  - Temporarily via an environment variable:
      export NIX_CONFIG="experimental-features = nix-command flakes"

layat does not add --extra-experimental-features automatically (it will not override your environment settings).

Original nix error:
%s`, strings.TrimSpace(stderr))
}

// wrapEvalErr makes the "layat.<name> does not exist" case of an eval failure clearer
// (experimental etc. are passed through as-is) (→ docs/spec.md error spec). The original
// error is wrapped (%w), keeping the nixCmdError marker reachable for --json classification.
func wrapEvalErr(err error, label string) error {
	msg := err.Error()
	if strings.Contains(msg, "does not provide attribute") ||
		(strings.Contains(msg, "attribute") && strings.Contains(msg, "missing")) {
		return fmt.Errorf("layat: %s not found in the entrypoint (check the config name)\n%w", label, err)
	}
	return err
}

// wrapEvalAllErr makes the "layat.<system> does not exist" case of a batch eval failure clearer.
func wrapEvalAllErr(err error, label string) error {
	msg := err.Error()
	if strings.Contains(msg, "does not provide attribute") ||
		(strings.Contains(msg, "attribute") && strings.Contains(msg, "missing")) {
		return fmt.Errorf("layat: %s not found in the entrypoint (no configs found)\n%w", label, err)
	}
	return err
}
