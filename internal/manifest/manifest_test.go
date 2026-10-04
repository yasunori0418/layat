package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, FileName)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadValid(t *testing.T) {
	dir := writeManifest(t, `{
	  "schemaVersion": 1,
	  "root": { "rootKind": "project" },
	  "entries": [
	    { "srcKind": "store", "src": "/nix/store/aaa-source", "subpath": "skills/nix", "target": ".claude/skills/nix", "method": "symlink" }
	  ]
	}`)

	m, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.SchemaVersion != 1 {
		t.Errorf("schemaVersion = %d, want 1", m.SchemaVersion)
	}
	if m.Root.RootKind != RootKindProject {
		t.Errorf("rootKind = %q, want project", m.Root.RootKind)
	}
	// project omits the path in the manifest, so it decodes to the empty string.
	if m.Root.Root != "" {
		t.Errorf("project root = %q, want empty", m.Root.Root)
	}
	if len(m.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(m.Entries))
	}
	e := m.Entries[0]
	if e.SrcKind != SrcKindStore || e.Src != "/nix/store/aaa-source" || e.Subpath != "skills/nix" || e.Target != ".claude/skills/nix" || e.Method != MethodSymlink {
		t.Errorf("entry mismatch: %+v", e)
	}
}

func TestLoadRejectsNewerSchema(t *testing.T) {
	dir := writeManifest(t, `{ "schemaVersion": 2, "root": { "rootKind": "project" }, "entries": [] }`)
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for schemaVersion 2, got nil")
	}
	// The sentinel must be detectable via errors.Is so the CLI can add skew guidance (→ docs/spec.md).
	if !errors.Is(err, ErrSchemaVersionUnsupported) {
		t.Errorf("error should wrap ErrSchemaVersionUnsupported, got %v", err)
	}
}

// Both edges of the schemaVersion lower bound; the upper bound is TestLoadRejectsNewerSchema.
func TestLoadSchemaVersionBoundary(t *testing.T) {
	for _, tt := range []struct {
		name    string
		version string
		wantErr bool
	}{
		{"negative", "-1", true},
		{"zero", "0", true},
		{"lowest accepted", "1", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeManifest(t, `{ "schemaVersion": `+tt.version+`, "root": { "rootKind": "project" }, "entries": [] }`)
			_, err := Load(dir)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("schemaVersion %s should be accepted, got %v", tt.version, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error for schemaVersion %s, got nil", tt.version)
			}
			if !strings.Contains(err.Error(), "is invalid") {
				t.Errorf("error should report an invalid schemaVersion, got %v", err)
			}
			// Below the minimum is a broken document, not version skew, so it must not wrap
			// the skew sentinel.
			if errors.Is(err, ErrSchemaVersionUnsupported) {
				t.Errorf("error should not wrap ErrSchemaVersionUnsupported, got %v", err)
			}
		})
	}
}

// A missing root object and an explicitly empty rootKind both reach the same emptiness check.
func TestLoadRejectsRootKindEmptiedByEitherSpelling(t *testing.T) {
	for _, tt := range []struct {
		name string
		doc  string
	}{
		{"explicitly empty", `{ "schemaVersion": 1, "root": { "rootKind": "" }, "entries": [] }`},
		{"root object omitted", `{ "schemaVersion": 1, "entries": [] }`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeManifest(t, tt.doc))
			if err == nil {
				t.Fatal("expected error for empty rootKind, got nil")
			}
			// Pin the rejection to the rootKind check, not to decoding or any later guard.
			if !strings.Contains(err.Error(), "root.rootKind") {
				t.Errorf("error should report an empty root.rootKind, got %v", err)
			}
		})
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	dir := writeManifest(t, `{ "schemaVersion": 1, "root": { "rootKind": "project" }, "entries": [], "bogus": true }`)
	if _, err := Load(dir); err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("expected error for missing manifest.json, got nil")
	}
}

// TestLoadRootPathAllowedOnlyForFixed covers the kind × path-presence table: only fixed may carry
// a path. Accepting rows mean only that this layer does not reject them (e.g. fixed without a
// path is stopped later by root resolution).
func TestLoadRootPathAllowedOnlyForFixed(t *testing.T) {
	// An empty path means the key is left out ("root": "" decodes the same).
	for _, tt := range []struct {
		name    string
		kind    string
		path    string
		wantErr bool
	}{
		{"project with path", RootKindProject, "/should/be/ignored", true},
		{"home with path", RootKindHome, "/should/be/ignored", true},
		{"system with path", RootKindSystem, "/should/be/ignored", true},
		{"project without path", RootKindProject, "", false},
		{"home without path", RootKindHome, "", false},
		{"system without path", RootKindSystem, "", false},
		{"fixed with path", RootKindFixed, "/opt/x", false},
		{"fixed without path", RootKindFixed, "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := `{ "rootKind": "` + tt.kind + `" }`
			if tt.path != "" {
				root = `{ "rootKind": "` + tt.kind + `", "root": "` + tt.path + `" }`
			}
			_, err := Load(writeManifest(t, `{ "schemaVersion": 1, "root": `+root+`, "entries": [] }`))
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("rootKind %q with root %q should be accepted, got %v", tt.kind, tt.path, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error for rootKind %q carrying a path, got nil", tt.kind)
			}
			// Pin the rejection to this check, not to decoding or any other guard.
			if !strings.Contains(err.Error(), "root.root must be omitted") {
				t.Errorf("error should report that root.root must be omitted, got %v", err)
			}
		})
	}
}

func TestLoadFixedRootHasPath(t *testing.T) {
	dir := writeManifest(t, `{ "schemaVersion": 1, "root": { "rootKind": "fixed", "root": "/opt/x" }, "entries": [] }`)
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Root.Root != "/opt/x" {
		t.Errorf("fixed root = %q, want /opt/x", m.Root.Root)
	}
}
