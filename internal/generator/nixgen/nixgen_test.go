package nixgen

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/generator"
)

// writeFile creates an empty file at dir/name (content is irrelevant; discoverEntrypoint only checks existence).
func writeFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
		t.Fatalf("writeFile(%s): %v", name, err)
	}
}

// TestDiscoverEntrypoint_FileFlag covers -f pointing directly at a file (→ ADR-0032 discovery order).
func TestDiscoverEntrypoint_FileFlag(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		wantKind entrypointKind
	}{
		{"flake.nix", "flake.nix", entrypointFlake},
		{"shell.nix", "shell.nix", entrypointLegacy},
		{"default.nix", "default.nix", entrypointLegacy},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, c.file)
			path := filepath.Join(dir, c.file)

			ep, err := discoverEntrypoint(path)
			if err != nil {
				t.Fatalf("discoverEntrypoint(%s): %v", path, err)
			}
			if ep.kind != c.wantKind {
				t.Errorf("kind = %v, want %v", ep.kind, c.wantKind)
			}
			switch c.wantKind {
			case entrypointFlake:
				if ep.flakeRef != dir {
					t.Errorf("flakeRef = %q, want %q", ep.flakeRef, dir)
				}
			case entrypointLegacy:
				if ep.legacyPath != path {
					t.Errorf("legacyPath = %q, want %q", ep.legacyPath, path)
				}
			}
		})
	}
}

// TestDiscoverEntrypoint_FileFlagRejectsUnknownName covers -f pointing at a file that is none of the three.
func TestDiscoverEntrypoint_FileFlagRejectsUnknownName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.nix")
	path := filepath.Join(dir, "config.nix")

	if _, err := discoverEntrypoint(path); err == nil {
		t.Fatal("expected an error for an unrecognized -f file name, got nil")
	}
}

// TestDiscoverEntrypoint_FileFlagMissing covers -f pointing at a path that does not exist.
func TestDiscoverEntrypoint_FileFlagMissing(t *testing.T) {
	if _, err := discoverEntrypoint(filepath.Join(t.TempDir(), "nope.nix")); err == nil {
		t.Fatal("expected an error for a missing -f path, got nil")
	}
}

// TestDiscoverEntrypoint_FileFlagDir covers -f pointing at a directory, applying the same
// flake.nix -> shell.nix -> default.nix priority as CWD autodiscovery (→ ADR-0032).
func TestDiscoverEntrypoint_FileFlagDir(t *testing.T) {
	cases := []struct {
		name       string
		files      []string
		wantKind   entrypointKind
		wantLegacy string // expected legacy file name, if wantKind == entrypointLegacy
	}{
		{"flake.nix only", []string{"flake.nix"}, entrypointFlake, ""},
		{"flake.nix wins over shell.nix", []string{"flake.nix", "shell.nix"}, entrypointFlake, ""},
		{"shell.nix only", []string{"shell.nix"}, entrypointLegacy, "shell.nix"},
		{"default.nix only", []string{"default.nix"}, entrypointLegacy, "default.nix"},
		{"shell.nix wins over default.nix", []string{"shell.nix", "default.nix"}, entrypointLegacy, "shell.nix"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range c.files {
				writeFile(t, dir, f)
			}
			ep, err := discoverEntrypoint(dir)
			if err != nil {
				t.Fatalf("discoverEntrypoint(%s): %v", dir, err)
			}
			if ep.kind != c.wantKind {
				t.Errorf("kind = %v, want %v", ep.kind, c.wantKind)
			}
			if c.wantKind == entrypointLegacy {
				want := filepath.Join(dir, c.wantLegacy)
				if ep.legacyPath != want {
					t.Errorf("legacyPath = %q, want %q", ep.legacyPath, want)
				}
			}
		})
	}
}

// TestDiscoverEntrypoint_FileFlagDirEmpty covers -f pointing at a directory with none of the three files.
func TestDiscoverEntrypoint_FileFlagDirEmpty(t *testing.T) {
	if _, err := discoverEntrypoint(t.TempDir()); err == nil {
		t.Fatal("expected an error for a -f directory with no entrypoint file, got nil")
	}
}

// TestDiscoverEntrypoint_CWD covers CWD autodiscovery priority flake.nix -> shell.nix -> default.nix (→ ADR-0032).
func TestDiscoverEntrypoint_CWD(t *testing.T) {
	cases := []struct {
		name       string
		files      []string
		wantKind   entrypointKind
		wantLegacy string
		wantErr    bool
	}{
		{"flake.nix only", []string{"flake.nix"}, entrypointFlake, "", false},
		{"flake.nix wins over shell.nix", []string{"flake.nix", "shell.nix"}, entrypointFlake, "", false},
		{"shell.nix only", []string{"shell.nix"}, entrypointLegacy, "shell.nix", false},
		{"default.nix only", []string{"default.nix"}, entrypointLegacy, "default.nix", false},
		{"shell.nix wins over default.nix", []string{"shell.nix", "default.nix"}, entrypointLegacy, "shell.nix", false},
		{"none", nil, 0, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range c.files {
				writeFile(t, dir, f)
			}
			t.Chdir(dir)

			ep, err := discoverEntrypoint("")
			if c.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("discoverEntrypoint(\"\"): %v", err)
			}
			if ep.kind != c.wantKind {
				t.Errorf("kind = %v, want %v", ep.kind, c.wantKind)
			}
			if c.wantKind == entrypointLegacy {
				want := filepath.Join(dir, c.wantLegacy)
				if ep.legacyPath != want {
					t.Errorf("legacyPath = %q, want %q", ep.legacyPath, want)
				}
			}
		})
	}
}

// TestEntrypointInstallableArgs locks in the flake `<ep>#layat.<system>.<name><suffix>` form vs. the
// legacy `-f <ep> layat.<name><suffix>` form (no per-system dimension; → ADR-0032).
func TestEntrypointInstallableArgs(t *testing.T) {
	flakeEp := &entrypoint{kind: entrypointFlake, flakeRef: "/proj"}
	if got, want := flakeEp.installableArgs("x86_64-linux", "docs", ".rootKind"), []string{"/proj#layat.x86_64-linux.docs.rootKind"}; !reflect.DeepEqual(got, want) {
		t.Errorf("flake installableArgs = %v, want %v", got, want)
	}
	if got, want := flakeEp.namespaceArgs("x86_64-linux"), []string{"/proj#layat.x86_64-linux"}; !reflect.DeepEqual(got, want) {
		t.Errorf("flake namespaceArgs = %v, want %v", got, want)
	}
	if got, want := flakeEp.label("x86_64-linux", "docs"), "layat.x86_64-linux.docs"; got != want {
		t.Errorf("flake label = %q, want %q", got, want)
	}

	legacyEp := &entrypoint{kind: entrypointLegacy, legacyPath: "/proj/shell.nix"}
	if got, want := legacyEp.installableArgs("x86_64-linux", "docs", ".rootKind"), []string{"-f", "/proj/shell.nix", "layat.docs.rootKind"}; !reflect.DeepEqual(got, want) {
		t.Errorf("legacy installableArgs = %v, want %v", got, want)
	}
	if got, want := legacyEp.namespaceArgs("x86_64-linux"), []string{"-f", "/proj/shell.nix", "layat"}; !reflect.DeepEqual(got, want) {
		t.Errorf("legacy namespaceArgs = %v, want %v", got, want)
	}
	if got, want := legacyEp.label("x86_64-linux", "docs"), "layat.docs"; got != want {
		t.Errorf("legacy label = %q, want %q", got, want)
	}
}

// stubNix puts a fake nix first on PATH that prints stderr to its stderr and exits 1.
func stubNix(t *testing.T, stderr string) {
	t.Helper()
	bin := t.TempDir()
	script := "#!/bin/sh\nprintf '%s' '" + stderr + "' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "nix"), []byte(script), 0o755); err != nil {
		t.Fatalf("write nix stub: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// discovered returns a Generator that has discovered a flake entrypoint in a temp dir.
func discovered(t *testing.T, w io.Writer, debug bool) (*Generator, string) {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "flake.nix")
	g := New(w, debug)
	if err := g.Discover(dir); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return g, dir
}

// TestDiscoverFailureIsDiscoverStage pins a discovery failure as a nix-tagged generator.Error of
// Stage discover whose text is the pre-extraction message and whose cause still reaches
// fs.ErrNotExist (the CLI classifies discover failures by the cause · → ADR-0055 §7).
func TestDiscoverFailureIsDiscoverStage(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.nix")
	err := New(io.Discard, false).Discover(missing)
	var ge *generator.Error
	if !errors.As(err, &ge) {
		t.Fatalf("Discover error = %v, want a *generator.Error", err)
	}
	if ge.Generator != generator.NameNix || ge.Stage != generator.StageDiscover || ge.Kind != generator.KindFailed {
		t.Errorf("Generator/Stage/Kind = %q/%q/%q, want nix/discover/Failed", ge.Generator, ge.Stage, ge.Kind)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("errors.Is(err, fs.ErrNotExist) = false: %v", err)
	}
	if want := "layat: -f path not found (" + missing + "): "; !strings.HasPrefix(err.Error(), want) {
		t.Errorf("Error() = %q, want prefix %q", err.Error(), want)
	}
}

// TestRootsCommandFailure pins a failed nix eval as Stage roots carrying the pre-extraction
// message (the failed command + nix's stderr) and the captured stderr.
func TestRootsCommandFailure(t *testing.T) {
	stubNix(t, "error: boom")
	g, _ := discovered(t, io.Discard, false)

	_, err := g.Roots("web")
	var ge *generator.Error
	if !errors.As(err, &ge) {
		t.Fatalf("Roots error = %v, want a *generator.Error", err)
	}
	if ge.Stage != generator.StageRoots || ge.Kind != generator.KindFailed {
		t.Errorf("Stage/Kind = %q/%q, want roots/Failed", ge.Stage, ge.Kind)
	}
	if want := "layat: nix eval failed:\nerror: boom"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	if ge.Stderr != "error: boom" {
		t.Errorf("Stderr = %q, want the captured stderr", ge.Stderr)
	}
}

// TestRootsMissingConfigMessage pins the "config not found" lead line put before nix's error.
func TestRootsMissingConfigMessage(t *testing.T) {
	stubNix(t, "error: flake does not provide attribute layat")
	g, _ := discovered(t, io.Discard, false)

	_, err := g.Roots("web")
	want := "layat: " + g.ep.label(g.system, "web") + " not found in the entrypoint (check the config name)\n" +
		"layat: nix eval failed:\nerror: flake does not provide attribute layat"
	if err == nil || err.Error() != want {
		t.Errorf("Roots error = %v, want %q", err, want)
	}
}

// TestBuildCommandFailureIsBuildStage pins a failed nix build (Build and DryBuild) as Stage build.
func TestBuildCommandFailureIsBuildStage(t *testing.T) {
	stubNix(t, "error: boom")
	g, _ := discovered(t, io.Discard, false)

	for _, op := range []struct {
		name string
		run  func() (string, error)
	}{
		{"Build", func() (string, error) { return g.Build("web", filepath.Join(t.TempDir(), ".pending")) }},
		{"DryBuild", func() (string, error) { return g.DryBuild("web") }},
	} {
		_, err := op.run()
		var ge *generator.Error
		if !errors.As(err, &ge) || ge.Stage != generator.StageBuild {
			t.Errorf("%s error = %v, want a Stage build *generator.Error", op.name, err)
		}
	}
}

// TestDebugDisclosesCommandsToWriter pins the --debug disclosure line on the writer given to New.
func TestDebugDisclosesCommandsToWriter(t *testing.T) {
	stubNix(t, "")
	var buf bytes.Buffer
	g, dir := discovered(t, &buf, true)

	_, _ = g.Roots("web")
	want := "layat: + nix eval " + dir + "#layat." + g.system + ".web.rootKind --raw\n"
	if buf.String() != want {
		t.Errorf("debug output = %q, want %q", buf.String(), want)
	}
}
