package generator

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yasunori0418/layat/internal/manifest"
)

// writeLinkFarm creates a link-farm directory holding manifest.json with content.
func writeLinkFarm(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, manifest.FileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Every implementation of the contract satisfies Generator (→ ADR-0055 §1, §5).
var (
	_ Generator = (*Prebuilt)(nil)
	_ Generator = (*Fake)(nil)
)

// TestErrorKeepsCauseAndMessage pins generator.Error's error face: Error() is Message as-is, and
// the cause stays reachable through Unwrap, so errors.Is / errors.As see past it (→ ADR-0055 §6).
func TestErrorKeepsCauseAndMessage(t *testing.T) {
	cause := &fs.PathError{Op: "stat", Path: "/nope", Err: fs.ErrNotExist}
	e := NewError("nix", StageDiscover, KindFailed, "layat: -f path not found (/nope)", "", "", cause)

	if e.Error() != "layat: -f path not found (/nope)" {
		t.Errorf("Error() = %q, want the Message", e.Error())
	}
	if !errors.Is(e, fs.ErrNotExist) {
		t.Error("errors.Is(e, fs.ErrNotExist) = false, want the cause reachable through Unwrap")
	}
	var pe *fs.PathError
	if !errors.As(e, &pe) {
		t.Error("errors.As(e, *fs.PathError) = false, want the cause reachable through Unwrap")
	}
	if e.Generator != "nix" || e.Stage != StageDiscover || e.Kind != KindFailed {
		t.Errorf("fields = %q/%q/%q, want nix/discover/Failed", e.Generator, e.Stage, e.Kind)
	}
}

// TestErrorSurvivesWrapping pins that the marker survives a %w chain (the CLI classifies by errors.As).
func TestErrorSurvivesWrapping(t *testing.T) {
	e := NewError("nix", StageBuild, KindFailed, "layat: nix build failed", "", "", nil)
	wrapped := errors.Join(errors.New("context"), e)
	var got *Error
	if !errors.As(wrapped, &got) || got != e {
		t.Fatalf("errors.As through a wrap = %v, want the original *Error", got)
	}
	if e.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil without a cause", e.Unwrap())
	}
}

// TestPrebuiltDiscoverResolvesAbsolutePath covers prebuilt Discover = the given link-farm path,
// made absolute (→ ADR-0055 §5, ADR-0026).
func TestPrebuiltDiscoverResolvesAbsolutePath(t *testing.T) {
	dir := writeLinkFarm(t, `{"schemaVersion":1,"root":{"rootKind":"home"},"entries":[]}`)
	t.Chdir(filepath.Dir(dir))

	p := &Prebuilt{}
	if err := p.Discover(filepath.Base(dir)); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	for _, op := range []struct {
		name string
		run  func() (string, error)
	}{
		{"Build", func() (string, error) { return p.Build("web", "/ignored/.pending") }},
		{"DryBuild", func() (string, error) { return p.DryBuild("web") }},
	} {
		got, err := op.run()
		if err != nil {
			t.Fatalf("%s: %v", op.name, err)
		}
		if got != dir {
			t.Errorf("%s = %q, want the absolute link-farm path %q", op.name, got, dir)
		}
	}
}

// TestPrebuiltRootsReadsManifest covers prebuilt Roots = the link-farm's manifest.json root, with
// Targets derived from the entries (the manifest schema itself carries no targets · → ADR-0055 §2).
func TestPrebuiltRootsReadsManifest(t *testing.T) {
	dir := writeLinkFarm(t, `{
	  "schemaVersion": 1,
	  "root": { "rootKind": "fixed", "root": "/srv/app" },
	  "entries": [
	    { "srcKind": "store", "src": "/nix/store/aaa-source", "subpath": "", "target": "a", "method": "symlink" },
	    { "srcKind": "store", "src": "/nix/store/aaa-source", "subpath": "", "target": "b/c", "method": "copy" }
	  ]
	}`)
	p := &Prebuilt{}
	if err := p.Discover(dir); err != nil {
		t.Fatalf("Discover: %v", err)
	}

	want := manifest.Root{RootKind: manifest.RootKindFixed, Root: "/srv/app", Targets: []string{"a", "b/c"}}
	got, err := p.Roots("web")
	if err != nil {
		t.Fatalf("Roots: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Roots = %+v, want %+v", got, want)
	}
}

// TestPrebuiltRootsRejectsTargetsKey covers that Targets stays out of the manifest.json schema v1:
// a root.targets key is still an unknown field the load rejects (→ ADR-0055 §2).
func TestPrebuiltRootsRejectsTargetsKey(t *testing.T) {
	dir := writeLinkFarm(t, `{"schemaVersion":1,"root":{"rootKind":"project","targets":["x"]},"entries":[]}`)
	p := &Prebuilt{}
	if err := p.Discover(dir); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if _, err := p.Roots("web"); err == nil {
		t.Fatal("Roots accepted a manifest.json carrying root.targets, want the unknown field rejected")
	}
}

// TestPrebuiltAllRootsFails covers that a single pre-built link-farm has no config list to enumerate,
// even when its manifest.json is readable.
func TestPrebuiltAllRootsFails(t *testing.T) {
	p := &Prebuilt{}
	if err := p.Discover(writeLinkFarm(t, `{"schemaVersion":1,"root":{"rootKind":"home"},"entries":[]}`)); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	all, err := p.AllRoots()
	var ge *Error
	if all != nil || !errors.As(err, &ge) || ge.Generator != NamePrebuilt {
		t.Errorf("AllRoots = %v, %v; want nil and a prebuilt *generator.Error", all, err)
	}
}

// TestPrebuiltRootsFailureKeepsCause covers an unreadable manifest.json: a prebuilt-tagged
// generator.Error whose cause chain still reaches fs.ErrNotExist (the CLI classifies prebuilt
// failures by the cause, not as E_LAYAT_BUILD · → ADR-0055 §7).
func TestPrebuiltRootsFailureKeepsCause(t *testing.T) {
	p := &Prebuilt{}
	if err := p.Discover(t.TempDir()); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	_, err := p.Roots("web")
	var ge *Error
	if !errors.As(err, &ge) {
		t.Fatalf("Roots error = %v, want a *generator.Error", err)
	}
	if ge.Generator != NamePrebuilt || ge.Stage != StageRoots {
		t.Errorf("Generator/Stage = %q/%q, want %q/%q", ge.Generator, ge.Stage, NamePrebuilt, StageRoots)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("errors.Is(err, fs.ErrNotExist) = false, want the cause kept: %v", err)
	}
}

// TestFakeRecordsCallsAndReturnsStubs covers the test double: every operation is recorded in call
// order, and each one returns what its stub returns (zero values when unset).
func TestFakeRecordsCallsAndReturnsStubs(t *testing.T) {
	buildErr := errors.New("stub build failed")
	f := &Fake{
		RootsFunc: func(name string) (manifest.Root, error) {
			return manifest.Root{RootKind: manifest.RootKindHome, Targets: []string{name}}, nil
		},
		BuildFunc: func(string, string) (string, error) { return "", buildErr },
		DryBuildFunc: func(name string) (string, error) {
			return "/nix/store/xxx-" + name, nil
		},
	}

	if err := f.Discover("/proj"); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	r, err := f.Roots("web")
	if err != nil || r.RootKind != manifest.RootKindHome || !reflect.DeepEqual(r.Targets, []string{"web"}) {
		t.Errorf("Roots = %+v, %v; want the stubbed root", r, err)
	}
	all, err := f.AllRoots()
	if err != nil || all != nil {
		t.Errorf("AllRoots = %v, %v; want zero values when unset", all, err)
	}
	if _, err := f.Build("web", "/p/.pending"); !errors.Is(err, buildErr) {
		t.Errorf("Build err = %v, want the stubbed error", err)
	}
	if got, _ := f.DryBuild("web"); got != "/nix/store/xxx-web" {
		t.Errorf("DryBuild = %q, want the stubbed path", got)
	}

	want := []Call{
		{Op: "Discover", Args: []string{"/proj"}},
		{Op: "Roots", Args: []string{"web"}},
		{Op: "AllRoots"},
		{Op: "Build", Args: []string{"web", "/p/.pending"}},
		{Op: "DryBuild", Args: []string{"web"}},
	}
	if got := f.Calls(); !reflect.DeepEqual(got, want) {
		t.Errorf("Calls() = %+v, want %+v", got, want)
	}
}
