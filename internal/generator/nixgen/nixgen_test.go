package nixgen

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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
	stubNixExit(t, "", stderr, 1)
}

// stubNixExit puts a fake nix first on PATH that prints stdout / stderr and exits with code.
func stubNixExit(t *testing.T, stdout, stderr string, code int) {
	t.Helper()
	stubNixScript(t, "printf '%s' '"+stdout+"'\nprintf '%s' '"+stderr+"' >&2\nexit "+strconv.Itoa(code))
}

// stubNixScript puts a fake nix first on PATH whose body is the given sh script.
func stubNixScript(t *testing.T, body string) {
	t.Helper()
	bin := t.TempDir()
	script := "#!/bin/sh\n" + body + "\n"
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

// TestRootsCommandFailure pins a failed nix eval as Stage roots whose one-line Message names the
// failed subcommand, with nix's stderr kept in Stderr instead (→ ADR-0055 §6).
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
	if want := "layat: nix eval failed: exit status 1"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	if ge.Stderr != "error: boom" {
		t.Errorf("Stderr = %q, want the captured stderr", ge.Stderr)
	}
}

// TestRootsMissingConfigMessage pins the "config not found" summary and its guidance.
func TestRootsMissingConfigMessage(t *testing.T) {
	stubNix(t, "error: flake does not provide attribute layat")
	g, _ := discovered(t, io.Discard, false)

	_, err := g.Roots("web")
	var ge *generator.Error
	if !errors.As(err, &ge) {
		t.Fatalf("Roots error = %v, want a *generator.Error", err)
	}
	if want := "layat: " + g.ep.label(g.system, "web") + " not found in the entrypoint (nix eval failed)"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	if want := "check the config name"; ge.Guidance != want {
		t.Errorf("Guidance = %q, want %q", ge.Guidance, want)
	}
}

// TestAllRootsMissingNamespaceMessage pins a failed batch eval as Stage roots carrying the
// namespace "not found" summary and the "no configs found" guidance.
func TestAllRootsMissingNamespaceMessage(t *testing.T) {
	stubNix(t, "error: flake does not provide attribute layat")
	g, _ := discovered(t, io.Discard, false)

	_, err := g.AllRoots()
	var ge *generator.Error
	if !errors.As(err, &ge) || ge.Stage != generator.StageRoots {
		t.Fatalf("AllRoots error = %v, want a Stage roots *generator.Error", err)
	}
	label := g.ep.namespaceLabel(g.system)
	if want := "layat: " + label + " not found in the entrypoint (nix eval failed)"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	if want := "no configs found; define configs under " + label; ge.Guidance != want {
		t.Errorf("Guidance = %q, want %q", ge.Guidance, want)
	}
}

// TestDryBuildCommandFailure pins a failed read-only nix build as Stage build carrying the
// one-line Message and the captured stderr.
func TestDryBuildCommandFailure(t *testing.T) {
	stubNix(t, "error: boom")
	g, _ := discovered(t, io.Discard, false)

	_, err := g.DryBuild("web")
	var ge *generator.Error
	if !errors.As(err, &ge) || ge.Stage != generator.StageBuild {
		t.Fatalf("DryBuild error = %v, want a Stage build *generator.Error", err)
	}
	if want := "layat: nix build failed: exit status 1"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	if ge.Stderr != "error: boom" {
		t.Errorf("Stderr = %q, want the captured stderr", ge.Stderr)
	}
}

// TestBuildCommandFailureStreamsStderr pins a failed in-lock nix build as Stage build whose
// output streams to the generator's writer under the one-line Message.
func TestBuildCommandFailureStreamsStderr(t *testing.T) {
	stubNix(t, "error: boom")
	var buf bytes.Buffer
	g, _ := discovered(t, &buf, false)

	_, err := g.Build("web", filepath.Join(t.TempDir(), ".pending"))
	var ge *generator.Error
	if !errors.As(err, &ge) || ge.Stage != generator.StageBuild {
		t.Fatalf("Build error = %v, want a Stage build *generator.Error", err)
	}
	if want := "layat: nix build failed: exit status 1"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	if buf.String() != "error: boom" {
		t.Errorf("writer = %q, want nix's stderr streamed through", buf.String())
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

// TestIsExperimentalDisabled pins the three nix wordings of nix-command / flakes not being
// enabled, and that an unrelated failure is not taken for one (→ ADR-0025 §1).
func TestIsExperimentalDisabled(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		want   bool
	}{
		{"experimental Nix feature", "error: experimental Nix feature 'nix-command' is disabled; add '--extra-experimental-features nix-command' to enable it", true},
		{"experimental-features", "error: cannot use flakes: enable experimental-features first", true},
		{"flakes + disabled", "error: flakes are disabled", true},
		{"flakes alone", "error: flakes ok but evaluation failed", false},
		{"unrelated", "error: attribute 'web' missing", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isExperimentalDisabled(c.stderr); got != c.want {
				t.Errorf("isExperimentalDisabled(%q) = %v, want %v", c.stderr, got, c.want)
			}
		})
	}
}

// TestPrerequisiteMissingGuidance pins an experimental-features failure as Kind
// PrerequisiteMissing whose Guidance names both ways to enable them, with nix's raw stderr kept
// in Stderr (→ ADR-0055 §6, ADR-0025 §1).
func TestPrerequisiteMissingGuidance(t *testing.T) {
	const raw = "error: experimental Nix feature nix-command is disabled"
	stubNix(t, raw)
	g, _ := discovered(t, io.Discard, false)

	_, err := g.Roots("web")
	var ge *generator.Error
	if !errors.As(err, &ge) {
		t.Fatalf("Roots error = %v, want a *generator.Error", err)
	}
	if ge.Kind != generator.KindPrerequisiteMissing {
		t.Errorf("Kind = %q, want PrerequisiteMissing", ge.Kind)
	}
	if want := "layat: nix's experimental-features are not enabled (nix eval failed)"; ge.Message != want {
		t.Errorf("Message = %q, want %q", ge.Message, want)
	}
	for _, want := range []string{"nix.conf", "NIX_CONFIG"} {
		if !strings.Contains(ge.Guidance, want) {
			t.Errorf("Guidance lacks %q:\n%s", want, ge.Guidance)
		}
	}
	if ge.Stderr != raw {
		t.Errorf("Stderr = %q, want nix's raw stderr", ge.Stderr)
	}
}

// TestNotFoundKind pins a missing attribute as Kind NotFound on both the single and the batch
// eval (in both of nix's wordings) but not on build, and any other failure as Kind Failed.
func TestNotFoundKind(t *testing.T) {
	stubNix(t, "error: flake does not provide attribute layat")
	g, _ := discovered(t, io.Discard, false)

	var ge *generator.Error
	if _, err := g.Roots("web"); !errors.As(err, &ge) || ge.Kind != generator.KindNotFound {
		t.Errorf("Roots error = %v, want Kind NotFound", err)
	}
	if _, err := g.AllRoots(); !errors.As(err, &ge) || ge.Kind != generator.KindNotFound {
		t.Errorf("AllRoots error = %v, want Kind NotFound", err)
	}

	stubNix(t, "error: attribute web missing")
	if _, err := g.Roots("web"); !errors.As(err, &ge) || ge.Kind != generator.KindNotFound {
		t.Errorf("Roots error = %v, want Kind NotFound for nix's attribute-missing wording", err)
	}

	// Build runs after Roots found the config, so a missing attribute there is an evaluation
	// error inside the config, not a missing config name.
	stubNix(t, "error: flake does not provide attribute layat")
	if _, err := g.DryBuild("web"); !errors.As(err, &ge) || ge.Kind != generator.KindFailed {
		t.Errorf("DryBuild error = %v, want Kind Failed", err)
	}

	stubNix(t, "error: boom")
	if _, err := g.Roots("web"); !errors.As(err, &ge) || ge.Kind != generator.KindFailed {
		t.Errorf("Roots error = %v, want Kind Failed", err)
	}
}

// failingWriter fails every Write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// TestStderrCapturedWhenWriterFails pins that nix's stderr is still captured into Stderr when the
// diagnostics writer fails.
func TestStderrCapturedWhenWriterFails(t *testing.T) {
	stubNix(t, "error: boom\n")
	g, _ := discovered(t, failingWriter{}, false)

	_, err := g.Roots("web")
	var ge *generator.Error
	if !errors.As(err, &ge) {
		t.Fatalf("Roots error = %v, want a *generator.Error", err)
	}
	if ge.Stderr != "error: boom" {
		t.Errorf("Stderr = %q, want the captured stderr", ge.Stderr)
	}
}

// TestSuccessTeesStderrToWriter pins that nix's stderr reaches the writer on success too, on
// both the eval and the build path, while stdout stays the command's result (→ ADR-0055 §6).
func TestSuccessTeesStderrToWriter(t *testing.T) {
	const warn = "warning: Git tree is dirty\n"
	stubNixExit(t, "managed", warn, 0)
	var buf bytes.Buffer
	g, _ := discovered(t, &buf, false)

	root, err := g.Roots("web")
	if err != nil || root.RootKind != "managed" {
		t.Fatalf("Roots = %v, %v; want rootKind managed", root, err)
	}
	if buf.String() != warn {
		t.Errorf("writer after eval = %q, want %q", buf.String(), warn)
	}

	buf.Reset()
	stubNixExit(t, "/nix/store/x-link-farm", warn, 0)
	if _, err := g.DryBuild("web"); err != nil {
		t.Fatalf("DryBuild: %v", err)
	}
	if buf.String() != warn {
		t.Errorf("writer after build = %q, want %q", buf.String(), warn)
	}
}

// TestBuildFailureCapturesStderr pins that the in-lock build path, which streams nix's output to
// the writer, also keeps nix's stderr in Stderr on failure.
func TestBuildFailureCapturesStderr(t *testing.T) {
	stubNix(t, "error: boom")
	var buf bytes.Buffer
	g, _ := discovered(t, &buf, false)

	_, err := g.Build("web", filepath.Join(t.TempDir(), ".pending"))
	var ge *generator.Error
	if !errors.As(err, &ge) {
		t.Fatalf("Build error = %v, want a *generator.Error", err)
	}
	if ge.Stderr != "error: boom" {
		t.Errorf("Stderr = %q, want the captured stderr", ge.Stderr)
	}
	if buf.String() != "error: boom" {
		t.Errorf("writer = %q, want nix's stderr streamed through", buf.String())
	}
}

// writeLog records every Write call separately.
type writeLog struct{ writes []string }

func (l *writeLog) Write(b []byte) (int, error) {
	l.writes = append(l.writes, string(b))
	return len(b), nil
}

// TestTeeWritesWholeLines pins that nix's stderr reaches the writer in whole lines even when nix
// writes a line in pieces, with the unterminated tail written once when nix exits — so a
// line-prefixing writer shared by parallel generators never interleaves mid-line.
func TestTeeWritesWholeLines(t *testing.T) {
	stubNixScript(t, "printf 'warning: par' >&2\nsleep 0.2\nprintf 'tial\\nnext\\ntail' >&2\nexit 1")
	log := &writeLog{}
	g, _ := discovered(t, log, false)

	_, _ = g.DryBuild("web")
	if got, want := strings.Join(log.writes, ""), "warning: partial\nnext\ntail"; got != want {
		t.Fatalf("writer = %q, want %q", got, want)
	}
	for i, w := range log.writes[:len(log.writes)-1] {
		if !strings.HasSuffix(w, "\n") {
			t.Errorf("write %d = %q, want whole lines (only the last may be unterminated)", i, w)
		}
	}
	if last := log.writes[len(log.writes)-1]; last != "tail" {
		t.Errorf("last write = %q, want the unterminated tail alone", last)
	}
}
