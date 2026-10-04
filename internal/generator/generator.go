// Package generator defines the contract the CLI obtains a link-farm (manifest.json + its GC
// anchor) through. The nix implementation lives in nixgen; the prebuilt implementation
// (--manifest) and the test double live here.
package generator

import "github.com/yasunori0418/layat/internal/manifest"

// Generator is the contract every manifest generator implements. Discover runs first and the
// other operations act on what it found; a config is addressed by its name alone.
type Generator interface {
	// Discover finds the entrypoint from file (the -f value; "" = autodiscovery from the CWD).
	Discover(file string) error
	// Roots returns the config's root without building it, so profileDir is settled before the flock.
	Roots(name string) (manifest.Root, error)
	// AllRoots returns every config's root in one go, keyed by config name (apply --all /
	// gitignore --all).
	AllRoots() (map[string]manifest.Root, error)
	// Build lays the config's link-farm down at pending and returns its path, which must be
	// committable as a generation, i.e. a store path.
	Build(name, pending string) (string, error)
	// DryBuild returns the config's link-farm path without laying down a gcroot (--dryrun and
	// gitignore).
	DryBuild(name string) (string, error)
}

// Generator names carried by Error.Generator.
const (
	NameNix      = "nix"
	NamePrebuilt = "prebuilt"
)

// Stage is the contract step a generator failure happened in.
type Stage string

const (
	StageDiscover Stage = "discover"
	StageRoots    Stage = "roots"
	// StageBuild covers DryBuild failures too.
	StageBuild Stage = "build"
)

// Kind classifies a generator failure. It never affects the --json code.
type Kind string

const (
	// KindPrerequisiteMissing is an unmet prerequisite (for nix: experimental-features).
	KindPrerequisiteMissing Kind = "PrerequisiteMissing"
	// KindNotFound is a config name the entrypoint does not have.
	KindNotFound Kind = "NotFound"
	// KindFailed is any other failure.
	KindFailed Kind = "Failed"
)

// Error is a generator failure. Error() returns Message, the one-line summary, and Unwrap
// returns the cause so errors.Is / errors.As see past it.
type Error struct {
	Generator string // generator name (NameNix / NamePrebuilt)
	Stage     Stage
	Kind      Kind
	Message   string // the summary; for a failed internal command it names that command
	Guidance  string // how to resolve it ("" if none)
	Stderr    string // the captured raw diagnostics
	cause     error
}

// NewError builds an Error that wraps cause (nil for none).
func NewError(generator string, stage Stage, kind Kind, message, guidance, stderr string, cause error) *Error {
	return &Error{
		Generator: generator,
		Stage:     stage,
		Kind:      kind,
		Message:   message,
		Guidance:  guidance,
		Stderr:    stderr,
		cause:     cause,
	}
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.cause }
