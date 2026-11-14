package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/manifest"
)

// TestStderrLineForms pins the line form of each kind of stderr output.
func TestStderrLineForms(t *testing.T) {
	cases := []struct {
		name  string
		write func(*bytes.Buffer)
		want  string
	}{
		{"notice", func(b *bytes.Buffer) { noticef(b, "reset %s done (root=%s)", "a", "/r") },
			"layat: reset a done (root=/r)\n"},
		{"warning", func(b *bytes.Buffer) { warnf(b, "skipping the system profile base (%s): %v", "/s", errors.New("boom")) },
			"layat: warning: skipping the system profile base (/s): boom\n"},
		{"failure", func(b *bytes.Buffer) { printFailure(b, "apply a failed", errors.New("boom")) },
			"layat: apply a failed: boom\n"},
		{"nix disclosure", func(b *bytes.Buffer) { discloseNix(b, []string{"build", "--no-link", ".#layat"}) },
			"layat: + nix build --no-link .#layat\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			c.write(&buf)
			if buf.String() != c.want {
				t.Errorf("output = %q, want %q", buf.String(), c.want)
			}
		})
	}
}

// TestPrintSchemaSkewHint pins the schemaVersion skew hint: a blank line, then the hint headed by
// the prefix.
func TestPrintSchemaSkewHint(t *testing.T) {
	var buf bytes.Buffer
	printSchemaSkewHint(&buf)
	want := "\nlayat: the layat version pinned by the CLI (engine) and by the flake may be out of sync.\n" +
		"  The flake's layat input is generating a manifest newer than the CLI.\n" +
		"  Update the CLI, or lower the flake's layat input to match the CLI so both versions align.\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// TestPrintErrorPrefix pins the command failure's form: "layat: " heads each element of an
// errors.Join, a newline inside a message gets none, and a generator failure is its summary and
// guidance.
func TestPrintErrorPrefix(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"single", errors.New("unknown flag: --foo"), "layat: unknown flag: --foo\n"},
		{"wrapped", fmt.Errorf("cannot read the manifest: %w", errors.New("boom")), "layat: cannot read the manifest: boom\n"},
		{"multi-line message", errors.New("apply --all: same target:\n  t: claimed by both a and b"),
			"layat: apply --all: same target:\n  t: claimed by both a and b\n"},
		{"joined", errors.Join(errors.New("cannot remove a"), errors.New("cannot write the record")),
			"layat: cannot remove a\nlayat: cannot write the record\n"},
		{"several %w", fmt.Errorf("x: %w, %w", errors.New("a"), errors.New("b")), "layat: x: a, b\n"},
		{"exit code", &exitCodeError{code: 1, msg: "apply --all: 1 config(s) failed"}, "layat: apply --all: 1 config(s) failed\n"},
		{"generator", fmt.Errorf("ctx: %w", generator.NewError(generator.NameNix, generator.StageBuild, generator.KindFailed,
			"nix build failed: exit status 1", "see above", "error: boom", nil)), "layat: nix build failed: exit status 1\nsee above\n"},
	}
	withDebug(t, false)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			printError(&buf, c.err)
			if buf.String() != c.want {
				t.Errorf("output = %q, want %q", buf.String(), c.want)
			}
		})
	}
}

// TestPrintFailureJoined pins a failure line over an errors.Join: the head carries the first
// element, and each other element starts its own "layat: " line.
func TestPrintFailureJoined(t *testing.T) {
	var buf bytes.Buffer
	printFailure(&buf, "status home:a", errors.Join(errors.New("cannot read the backref"), errors.New("cannot read the record")))
	want := "layat: status home:a: cannot read the backref\nlayat: cannot read the record\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// failureDetailMarkers are the texts that open the engine's conflict listing and undo report lines.
var failureDetailMarkers = []string{"conflict: ", "  → ", "apply failed; "}

// assertFailureDetailsNotWarnings checks the engine output of an apply / rollback on stderr: the
// conflict listing and the undo report are "layat: <text>" lines, never warnings.
func assertFailureDetailsNotWarnings(t *testing.T, stderr string, wantMarkers ...string) {
	t.Helper()
	lines := strings.Split(stderr, "\n")
	for _, l := range lines {
		for _, m := range failureDetailMarkers {
			if strings.HasPrefix(l, "layat: warning: "+m) || strings.HasPrefix(l, "layat: warning:"+m) {
				t.Errorf("failure detail written as a warning: %q", l)
			}
		}
	}
	for _, m := range wantMarkers {
		if !strings.Contains(stderr, "\nlayat: "+m) && !strings.HasPrefix(stderr, "layat: "+m) {
			t.Errorf("stderr has no %q line:\n%s", "layat: "+m, stderr)
		}
	}
}

// assertEngineWarning checks that an ordinary engine warning reaches stderr as a
// "layat: warning: <text>" line.
func assertEngineWarning(t *testing.T, stderr string) {
	t.Helper()
	if !strings.HasPrefix(stderr, "layat: warning: no last-applied record") {
		t.Errorf("stderr has no engine warning headed by \"layat: warning: \":\n%s", stderr)
	}
}

// failureDetailsSetup isolates an apply / rollback of config c on a temp root with the Fake
// backend and a stub nix reporting rootKind=home. It returns the root and the flake directory.
func failureDetailsSetup(t *testing.T) (root, flakeDir string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("a read-only directory does not stop root")
	}
	withFlakeEntrypoint(t)
	stubNixRootKind(t, manifest.RootKindHome)
	installFakeBackend(t, "")
	origManifest, origVerbose := flagManifest, flagVerbose
	t.Cleanup(func() { flagManifest, flagVerbose = origManifest, origVerbose })
	flagVerbose = false
	root = t.TempDir()
	flagRoot = root
	return root, flagFile
}

// applyManifest runs apply c --manifest over a link-farm of entries and returns the error and
// stderr.
func applyManifest(t *testing.T, entries ...manifest.Entry) (error, string) {
	t.Helper()
	origFile := flagFile
	t.Cleanup(func() { flagFile = origFile })
	flagFile = ""
	flagManifest = []string{writeTestLinkFarm(t, entries...)}
	var err error
	_, stderr := captureOutErr(t, func() {
		run, _ := newApplyTestRun()
		err = runApply(run, "c")
	})
	flagFile, flagManifest = origFile, nil
	return err, stderr
}

// readOnlyDir creates dir with no write permission, restored at cleanup so TempDir can remove it.
func readOnlyDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

// TestFailureDetailsAreNotWarnings drives apply and rollback through the CLI to a conflict stop and
// to an FS failure, checking the conflict listing and the undo report reach stderr through the
// engine's Reportf, not as warnings.
func TestFailureDetailsAreNotWarnings(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := func(target string) manifest.Entry {
		return manifest.Entry{SrcKind: "store", Src: src, Subpath: "f", Target: target, Method: "symlink"}
	}

	t.Run("apply conflict", func(t *testing.T) {
		root, _ := failureDetailsSetup(t)
		if err := os.WriteFile(filepath.Join(root, ".a"), []byte("foreign"), 0o644); err != nil {
			t.Fatal(err)
		}
		err, stderr := applyManifest(t, entry(".a"))
		if err == nil {
			t.Fatal("apply succeeded, want a conflict stop")
		}
		assertFailureDetailsNotWarnings(t, stderr, "conflict: ", "  → ")
		assertEngineWarning(t, stderr)
	})

	t.Run("apply undo", func(t *testing.T) {
		root, _ := failureDetailsSetup(t)
		readOnlyDir(t, filepath.Join(root, "ro"))
		err, stderr := applyManifest(t, entry("ro/leaf"))
		if err == nil {
			t.Fatal("apply succeeded, want an FS failure")
		}
		assertFailureDetailsNotWarnings(t, stderr, "apply failed; ")
		assertEngineWarning(t, stderr)
	})

	// rollbackFrom applies [.a, ro/leaf] then [.b], so rolling back re-places .a and ro/leaf.
	rollbackFrom := func(t *testing.T, root, flakeDir string, prepare func()) (error, string) {
		t.Helper()
		if err, stderr := applyManifest(t, entry(".a"), entry("ro/leaf")); err != nil {
			t.Fatalf("first apply: %v\n%s", err, stderr)
		}
		if err, stderr := applyManifest(t, entry(".b")); err != nil {
			t.Fatalf("second apply: %v\n%s", err, stderr)
		}
		prepare()
		flagFile = flakeDir
		var err error
		_, stderr := captureOutErr(t, func() {
			run, _ := newRollbackTestRun()
			err = runRollback(run, "c")
		})
		return err, stderr
	}

	t.Run("rollback conflict", func(t *testing.T) {
		root, flakeDir := failureDetailsSetup(t)
		err, stderr := rollbackFrom(t, root, flakeDir, func() {
			if err := os.WriteFile(filepath.Join(root, ".a"), []byte("foreign"), 0o644); err != nil {
				t.Fatal(err)
			}
		})
		if err == nil {
			t.Fatal("rollback succeeded, want a conflict stop")
		}
		assertFailureDetailsNotWarnings(t, stderr, "conflict: ", "  → ")
	})

	t.Run("rollback undo", func(t *testing.T) {
		root, flakeDir := failureDetailsSetup(t)
		err, stderr := rollbackFrom(t, root, flakeDir, func() { readOnlyDir(t, filepath.Join(root, "ro")) })
		if err == nil {
			t.Fatal("rollback succeeded, want an FS failure")
		}
		assertFailureDetailsNotWarnings(t, stderr, "apply failed; ")
	})
}

// TestPrefixLiteralsOnlyInStderrFile checks that no string literal in the non-test Go of internal/
// and cmd/layat/ holds "layat: ", except in stderr.go: the prefix is added only where stderr is
// written.
func TestPrefixLiteralsOnlyInStderrFile(t *testing.T) {
	fset := token.NewFileSet()
	var found []string
	for _, dir := range []string{"../../internal", "."} {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
				(dir == "." && filepath.Base(path) == "stderr.go") {
				return nil
			}
			f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				s, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Errorf("%s: cannot unquote %s: %v", fset.Position(lit.Pos()), lit.Value, err)
					return true
				}
				if strings.Contains(s, "layat: ") {
					found = append(found, fmt.Sprintf("%s: %s", fset.Position(lit.Pos()), lit.Value))
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	for _, f := range found {
		t.Errorf("string literal holds \"layat: \" outside stderr.go: %s", f)
	}
}
