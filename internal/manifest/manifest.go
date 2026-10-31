// Package manifest reads and validates manifest.json, the contract between the Nix side
// (lib.mkManifest) and the engine. Any schemaVersion newer than SchemaVersion is rejected.
package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrSchemaVersionUnsupported indicates that a manifest newer than SchemaVersion was read.
// The CLI detects it via errors.Is to add schemaVersion skew guidance.
var ErrSchemaVersionUnsupported = errors.New("layat: schemaVersion is newer than the engine supports")

// SchemaVersion is the latest manifest.json version the engine can interpret.
const SchemaVersion = 1

// FileName is the fixed manifest name embedded in the link-farm derivation.
const FileName = "manifest.json"

// Enums for src kind, placement method, and root kind.
const (
	SrcKindStore        = "store"
	SrcKindOutOfStore   = "outOfStore"
	SrcKindRootRelative = "rootRelative"

	MethodSymlink = "symlink"
	MethodCopy    = "copy"

	RootKindProject = "project"
	RootKindHome    = "home"
	RootKindSystem  = "system"
	RootKindFixed   = "fixed"
)

// Root is the placement target base. Only fixed holds an absolute path; the other kinds are
// resolved at runtime. Targets is the config's target list a generator's Roots returns; it is
// not part of manifest.json, so Load rejects a root.targets key.
type Root struct {
	RootKind string   `json:"rootKind"`
	Root     string   `json:"root,omitempty"`
	Targets  []string `json:"-"`
}

// Entry is a single placement definition. Its identity is Target (the diff key for stale removal).
type Entry struct {
	SrcKind string `json:"srcKind"`
	Src     string `json:"src"`
	Subpath string `json:"subpath"`
	Target  string `json:"target"`
	Method  string `json:"method"`
}

// Manifest is the top level of manifest.json. Identity is the optional user-declared identifier
// of the config; empty when left out.
type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	Root          Root    `json:"root"`
	Entries       []Entry `json:"entries"`
	Identity      string  `json:"identity,omitempty"`
}

// Load reads manifest.json inside the link-farm directory and validates schemaVersion.
func Load(linkFarm string) (*Manifest, error) {
	return LoadFile(filepath.Join(linkFarm, FileName))
}

// LoadFile reads a manifest.json file directly and validates schemaVersion.
func LoadFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("layat: cannot read manifest.json: %w", err)
	}

	var m Manifest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("layat: cannot parse manifest.json (%s): %w", path, err)
	}

	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("layat: invalid manifest.json (%s): %w", path, err)
	}
	return &m, nil
}

func (m *Manifest) validate() error {
	// The engine rejects any schemaVersion newer than the version it supports.
	if m.SchemaVersion > SchemaVersion {
		return fmt.Errorf("schemaVersion %d is unsupported (this engine supports up to v%d): %w", m.SchemaVersion, SchemaVersion, ErrSchemaVersionUnsupported)
	}
	if m.SchemaVersion < 1 {
		return fmt.Errorf("schemaVersion %d is invalid", m.SchemaVersion)
	}
	if m.Root.RootKind == "" {
		return fmt.Errorf("root.rootKind is empty")
	}
	switch m.Root.RootKind {
	case RootKindProject, RootKindHome, RootKindSystem, RootKindFixed:
	default:
		return fmt.Errorf("unknown root.rootKind %q", m.Root.RootKind)
	}
	// Only fixed carries a path; beside the other kinds it would be silently ignored.
	if m.Root.RootKind != RootKindFixed && m.Root.Root != "" {
		return fmt.Errorf("root.root must be omitted: it is only allowed when rootKind is %q, got %q", RootKindFixed, m.Root.RootKind)
	}
	for _, e := range m.Entries {
		switch e.SrcKind {
		case SrcKindStore, SrcKindOutOfStore, SrcKindRootRelative:
		default:
			return fmt.Errorf("unknown srcKind %q (target: %s)", e.SrcKind, e.Target)
		}
		switch e.Method {
		case MethodSymlink, MethodCopy:
		default:
			return fmt.Errorf("unknown method %q (target: %s)", e.Method, e.Target)
		}
	}
	return nil
}
