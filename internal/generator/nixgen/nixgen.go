// Package nixgen is the nix manifest generator: it discovers a flake / legacy entrypoint,
// pre-reads rootKind with `nix eval`, and builds the link-farm with `nix build`. The per-system
// dimension and every nix-specific error string stay inside this package.
package nixgen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/manifest"
)

// entrypointKind distinguishes a flake entrypoint (`layat.<system>.<name>`, addressed via `<flakeRef>#...`)
// from a legacy entrypoint (`layat.<name>`, addressed via `nix ... -f <path> ...`).
type entrypointKind int

const (
	entrypointFlake entrypointKind = iota
	entrypointLegacy
)

// legacyEntrypointNames is the discovery order for legacy entrypoints, tried after flake.nix.
var legacyEntrypointNames = []string{"shell.nix", "default.nix"}

// entrypoint is a discovered entrypoint: a flake (flake.nix) or a legacy file (shell.nix / default.nix).
// Legacy has no per-system dimension (unlike the flake's `layat.<system>.<name>`).
type entrypoint struct {
	kind entrypointKind
	// flakeRef is the flake ref passed to `nix build`/`nix eval` (the absolute path of the directory containing flake.nix).
	flakeRef string
	// legacyPath is the absolute path to the shell.nix / default.nix file, passed to `nix ... -f`.
	legacyPath string
}

// discoverEntrypoint discovers the entrypoint from -f, else from the CWD, in the order
// flake.nix → shell.nix → default.nix.
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

// currentSystem returns the nix system name of the runtime environment (e.g. aarch64-darwin),
// injected into a flake's `layat.<system>.<name>`. Legacy entrypoints ignore it.
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

// installableArgs returns the nix args that select `layat.<name><suffix>`, appended right after
// the `eval`/`build` subcommand: "<flakeRef>#layat.<system>.<name><suffix>" for a flake,
// "-f <legacyPath> layat.<name><suffix>" for a legacy entrypoint.
func (e *entrypoint) installableArgs(system, name, suffix string) []string {
	if e.kind == entrypointLegacy {
		return []string{"-f", e.legacyPath, "layat." + name + suffix}
	}
	return []string{fmt.Sprintf("%s#layat.%s.%s%s", e.flakeRef, system, name, suffix)}
}

// namespaceArgs returns the nix args that select the `layat.<system>` (flake) or `layat` (legacy) namespace,
// used for the batch eval of apply --all / gitignore --all.
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

// Generator is the nix implementation of generator.Generator. Diagnostics (the --debug
// disclosure and nix build's progress) go to the writer given to New, never to stdout.
type Generator struct {
	stderr io.Writer
	debug  bool
	ep     *entrypoint
	system string
}

var _ generator.Generator = (*Generator)(nil)

// New returns a nix generator writing its diagnostics to stderr; debug discloses the nix
// commands it runs.
func New(stderr io.Writer, debug bool) *Generator {
	return &Generator{stderr: stderr, debug: debug}
}

// Discover discovers the entrypoint (→ discoverEntrypoint) and the current system.
func (g *Generator) Discover(file string) error {
	ep, err := discoverEntrypoint(file)
	if err != nil {
		return discoverError(err)
	}
	system, err := currentSystem()
	if err != nil {
		return discoverError(err)
	}
	g.ep, g.system = ep, system
	return nil
}

// discoverError tags a discovery failure as Stage discover, keeping err as the cause.
func discoverError(err error) error {
	return generator.NewError(generator.NameNix, generator.StageDiscover, generator.KindFailed, err.Error(), "", "", err)
}

// Roots pre-resolves the config's rootKind (+ root when fixed) without building (→ evalRoot).
func (g *Generator) Roots(name string) (manifest.Root, error) {
	rootKind, fixedRoot, err := g.evalRoot(name)
	if err != nil {
		return manifest.Root{}, err
	}
	return manifest.Root{RootKind: rootKind, Root: fixedRoot}, nil
}

// AllRoots gets every config's root + targets in a single batch eval (→ evalAllRoots).
func (g *Generator) AllRoots() (map[string]manifest.Root, error) {
	return g.evalAllRoots()
}

// evalAllRoots gets the config name → root map for `apply --all` / `gitignore --all` in a single
// `nix eval` that builds nothing and reads only rootKind + targets (+ root for fixed).
func (g *Generator) evalAllRoots() (map[string]manifest.Root, error) {
	// Extract only rootKind + targets (+ root if fixed) from each config under layat.<system>.
	apply := `cs: builtins.mapAttrs (_: c: { rootKind = c.rootKind; targets = c.targets; } // (if c ? root then { root = c.root; } else {})) cs`
	args := append([]string{"eval"}, g.ep.namespaceArgs(g.system)...)
	args = append(args, "--apply", apply, "--json")
	out, err := g.runNixCapture(args...)
	if err != nil {
		return nil, wrapEvalAllErr(err, g.ep.namespaceLabel(g.system))
	}
	// The batch eval's own shape (manifest.Root keeps Targets out of the manifest.json schema).
	var raw map[string]struct {
		RootKind string   `json:"rootKind"`
		Root     string   `json:"root"`
		Targets  []string `json:"targets"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("layat: cannot parse the batch eval result for %s: %w", g.ep.namespaceLabel(g.system), err)
	}
	roots := make(map[string]manifest.Root, len(raw))
	for name, r := range raw {
		roots[name] = manifest.Root{RootKind: r.RootKind, Root: r.Root, Targets: r.Targets}
	}
	return roots, nil
}

// evalRoot pre-resolves rootKind (+ the absolute path when fixed root) via a cheap nix eval, so
// profileDir is known before the flock and the build.
func (g *Generator) evalRoot(name string) (rootKind, fixedRoot string, err error) {
	args := append([]string{"eval"}, g.ep.installableArgs(g.system, name, ".rootKind")...)
	args = append(args, "--raw")
	out, err := g.runNixCapture(args...)
	if err != nil {
		return "", "", wrapEvalErr(err, g.ep.label(g.system, name))
	}
	rootKind = strings.TrimSpace(out)
	if rootKind == "fixed" {
		rootArgs := append([]string{"eval"}, g.ep.installableArgs(g.system, name, ".root")...)
		rootArgs = append(rootArgs, "--raw")
		out, err := g.runNixCapture(rootArgs...)
		if err != nil {
			return "", "", wrapEvalErr(err, g.ep.label(g.system, name))
		}
		fixedRoot = strings.TrimSpace(out)
	}
	return rootKind, fixedRoot, nil
}

// Build runs `nix build <installable> --out-link <pending>` (the engine calls it inside the lock;
// → engine.BuildFunc), reads the out-link, and returns the store path.
func (g *Generator) Build(name, pending string) (string, error) {
	args := append([]string{"build"}, g.ep.installableArgs(g.system, name, "")...)
	args = append(args, "--out-link", pending)
	if err := g.runNixStream(args...); err != nil {
		return "", err
	}
	store, err := os.Readlink(pending)
	if err != nil {
		return "", fmt.Errorf("layat: cannot read the build output out-link (%s): %w", pending, err)
	}
	return store, nil
}

// DryBuild gets the link-farm's store path via `nix build --no-link --print-out-paths` without
// laying down a gcroot (out-link).
func (g *Generator) DryBuild(name string) (string, error) {
	args := append([]string{"build"}, g.ep.installableArgs(g.system, name, "")...)
	args = append(args, "--no-link", "--print-out-paths")
	out, err := g.runNixCapture(args...)
	if err != nil {
		return "", err
	}
	store := strings.TrimSpace(out)
	if store == "" {
		return "", fmt.Errorf("layat: nix build --print-out-paths was empty (%s)", g.ep.label(g.system, name))
	}
	// --print-out-paths may return multiple lines (multi-output). The link-farm is a single output, so take the last line.
	lines := strings.Split(store, "\n")
	return strings.TrimSpace(lines[len(lines)-1]), nil
}

// stageOf maps a nix subcommand to the contract stage its failure belongs to: eval pre-reads the
// roots, build builds.
func stageOf(args []string) generator.Stage {
	if args[0] == "eval" {
		return generator.StageRoots
	}
	return generator.StageBuild
}

// runNixCapture captures and returns nix's stdout (for machine-readable output such as eval).
func (g *Generator) runNixCapture(args ...string) (string, error) {
	var stdout bytes.Buffer
	if err := g.runNix(args, &stdout); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

// runNixStream streams nix's output to the diagnostics writer (for build progress).
func (g *Generator) runNixStream(args ...string) error {
	return g.runNix(args, nil)
}

// runNix runs nix with its stderr teed to the diagnostics writer (in whole lines) and to a capture
// for the failure's Stderr. A nil stdout shares the same tee, so the writer is never written from
// two goroutines at once. A failure is a generator.Error classified by nixFailure.
func (g *Generator) runNix(args []string, stdout io.Writer) error {
	if g.debug {
		_, _ = fmt.Fprintf(g.stderr, "layat: + nix %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command("nix", args...)
	var stderr bytes.Buffer
	lines := &lineWriter{w: g.stderr}
	tee := io.MultiWriter(&stderr, lines)
	if stdout == nil {
		stdout = tee
	}
	cmd.Stdout = stdout
	cmd.Stderr = tee
	err := cmd.Run()
	lines.flush()
	if err != nil {
		return nixFailure(args, stderr.String(), err)
	}
	return nil
}

// lineWriter passes writes on to w in whole lines, keeping the unterminated rest until flush,
// so a writer shared by parallel generators never sees a line in pieces. Writing to w is
// best-effort: a failing writer neither fails nix nor cuts the stderr capture short.
type lineWriter struct {
	w   io.Writer
	buf []byte
}

func (l *lineWriter) Write(b []byte) (int, error) {
	l.buf = append(l.buf, b...)
	i := bytes.LastIndexByte(l.buf, '\n')
	if i < 0 {
		return len(b), nil
	}
	_, _ = l.w.Write(l.buf[:i+1])
	l.buf = append(l.buf[:0], l.buf[i+1:]...)
	return len(b), nil
}

// flush writes the unterminated rest once the command has exited.
func (l *lineWriter) flush() {
	if len(l.buf) > 0 {
		_, _ = l.w.Write(l.buf)
		l.buf = l.buf[:0]
	}
}

// nixFailure classifies a failed nix command by its stderr: experimental-features disabled is
// PrerequisiteMissing, a missing attribute while reading the roots is NotFound, anything else
// Failed. The raw stderr is kept in Stderr.
func nixFailure(args []string, stderr string, runErr error) error {
	cause := fmt.Errorf("layat: nix %s failed: %w", args[0], runErr)
	kind, message, guidance := generator.KindFailed, cause.Error(), ""
	switch {
	case isExperimentalDisabled(stderr):
		kind = generator.KindPrerequisiteMissing
		message = fmt.Sprintf("layat: nix's experimental-features are not enabled (nix %s failed)", args[0])
		guidance = experimentalGuidance
	case stageOf(args) == generator.StageRoots && isMissingAttribute(stderr):
		kind = generator.KindNotFound
	}
	return generator.NewError(generator.NameNix, stageOf(args), kind, message, guidance, strings.TrimSpace(stderr), cause)
}

// isExperimentalDisabled detects the nix-command / flakes not-enabled error.
func isExperimentalDisabled(stderr string) bool {
	return strings.Contains(stderr, "experimental Nix feature") ||
		strings.Contains(stderr, "experimental-features") ||
		(strings.Contains(stderr, "flakes") && strings.Contains(stderr, "disabled"))
}

// isMissingAttribute detects nix reporting that the addressed attribute does not exist.
func isMissingAttribute(stderr string) bool {
	return strings.Contains(stderr, "does not provide attribute") ||
		(strings.Contains(stderr, "attribute") && strings.Contains(stderr, "missing"))
}

// experimentalGuidance guides the prerequisites and how to enable them.
const experimentalGuidance = `This command internally uses ` + "`nix eval`" + ` / ` + "`nix build`" + ` (the new CLI) and flakes,
so experimental-features = nix-command flakes is required.

How to enable (either one):
  - Append to ~/.config/nix/nix.conf or /etc/nix/nix.conf:
      experimental-features = nix-command flakes
  - Temporarily via an environment variable:
      export NIX_CONFIG="experimental-features = nix-command flakes"

layat does not add --extra-experimental-features automatically (it will not override your environment settings).`

// wrapEvalErr makes the "layat.<name> does not exist" case of an eval failure clearer; any other
// failure is passed through as-is.
func wrapEvalErr(err error, label string) error {
	return notFound(err, label, "check the config name")
}

// wrapEvalAllErr makes the "layat.<system> does not exist" case of a batch eval failure clearer.
func wrapEvalAllErr(err error, label string) error {
	return notFound(err, label, "no configs found; define configs under "+label)
}

// notFound rewrites a NotFound generator.Error's summary to name the missing label and gives it
// guidance; the stage, the captured stderr, and the cause are kept.
func notFound(err error, label, guidance string) error {
	var ge *generator.Error
	if !errors.As(err, &ge) || ge.Kind != generator.KindNotFound {
		return err
	}
	message := fmt.Sprintf("layat: %s not found in the entrypoint (nix eval failed)", label)
	return generator.NewError(ge.Generator, ge.Stage, ge.Kind, message, guidance, ge.Stderr, ge.Unwrap())
}
