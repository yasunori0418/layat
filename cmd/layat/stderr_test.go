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

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/gitutil"
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

// TestPrintFailureHints: each hint is written once however many errors call for it — the skew hint
// for manifest.ErrSchemaVersionUnsupported, the --root hint for gitutil.ErrNotInRepo — and an
// unrelated error gets none.
func TestPrintFailureHints(t *testing.T) {
	var skew, notInRepo bytes.Buffer
	printSchemaSkewHint(&skew)
	printNotInRepoHint(&notInRepo)
	inRepo := fmt.Errorf("apply a: %w", gitutil.ErrNotInRepo)
	skewed := fmt.Errorf("cannot read the manifest: %w", manifest.ErrSchemaVersionUnsupported)
	cases := []struct {
		name                string
		errs                []error
		wantSkew, wantNotIn int
	}{
		{"none", []error{errors.New("boom"), nil}, 0, 0},
		{"not in repo once", []error{inRepo, fmt.Errorf("apply b: %w", gitutil.ErrNotInRepo)}, 0, 1},
		{"skew once", []error{skewed, skewed}, 1, 0},
		{"both", []error{inRepo, skewed}, 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			printFailureHints(&buf, c.errs...)
			out := buf.String()
			if got := strings.Count(out, skew.String()); got != c.wantSkew {
				t.Errorf("skew hints = %d, want %d:\n%s", got, c.wantSkew, out)
			}
			if got := strings.Count(out, notInRepo.String()); got != c.wantNotIn {
				t.Errorf("not-in-repo hints = %d, want %d:\n%s", got, c.wantNotIn, out)
			}
			if c.wantSkew == 0 && c.wantNotIn == 0 && out != "" {
				t.Errorf("output = %q, want none", out)
			}
		})
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
		{"exit code", &exitError{code: 1, msg: "apply --all: 1 config(s) failed"}, "layat: apply --all: 1 config(s) failed\n"},
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
	if !strings.HasPrefix(stderr, "layat: warning: ") && !strings.Contains(stderr, "\nlayat: warning: ") {
		t.Errorf("stderr has no engine warning headed by \"layat: warning: \":\n%s", stderr)
	}
}

// cliEngineSetup isolates an apply / rollback / reset of config c, driven through the CLI, on a
// temp root with the Fake backend and a stub nix reporting rootKind=home. It returns the root.
func cliEngineSetup(t *testing.T) string {
	t.Helper()
	withFlakeEntrypoint(t)
	stubNixRootKind(t, manifest.RootKindHome)
	installFakeBackend(t, "")
	origManifest, origVerbose := flagManifest, flagVerbose
	t.Cleanup(func() { flagManifest, flagVerbose = origManifest, origVerbose })
	flagVerbose = false
	root := t.TempDir()
	flagRoot = root
	return root
}

// applyManifest runs apply c --manifest over a link-farm of entries and returns stderr and the
// error.
func applyManifest(t *testing.T, entries ...manifest.Entry) (string, error) {
	t.Helper()
	origFile := flagFile
	flagFile = ""
	flagManifest = []string{writeTestLinkFarm(t, entries...)}
	var err error
	_, stderr := captureOutErr(t, func() {
		run, _ := newApplyTestRun()
		err = runApply(run, "c")
	})
	flagFile, flagManifest = origFile, nil
	return stderr, err
}

// readOnlyDir creates dir with no write permission, restored at cleanup so TempDir can remove it.
// It skips the test under root, whom a read-only directory does not stop.
func readOnlyDir(t *testing.T, dir string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("a read-only directory does not stop root")
	}
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
		root := cliEngineSetup(t)
		if err := os.WriteFile(filepath.Join(root, ".a"), []byte("foreign"), 0o644); err != nil {
			t.Fatal(err)
		}
		stderr, err := applyManifest(t, entry(".a"))
		if err == nil {
			t.Fatal("apply succeeded, want a conflict stop")
		}
		assertFailureDetailsNotWarnings(t, stderr, "conflict: ", "  → ")
		assertEngineWarning(t, stderr)
	})

	t.Run("apply undo", func(t *testing.T) {
		root := cliEngineSetup(t)
		readOnlyDir(t, filepath.Join(root, "ro"))
		stderr, err := applyManifest(t, entry("ro/leaf"))
		if err == nil {
			t.Fatal("apply succeeded, want an FS failure")
		}
		assertFailureDetailsNotWarnings(t, stderr, "apply failed; ")
		assertEngineWarning(t, stderr)
	})

	// rollbackFrom applies [.a, ro/leaf] then [.b], so rolling back re-places .a and ro/leaf.
	rollbackFrom := func(t *testing.T, prepare func()) (string, error) {
		t.Helper()
		if stderr, err := applyManifest(t, entry(".a"), entry("ro/leaf")); err != nil {
			t.Fatalf("first apply: %v\n%s", err, stderr)
		}
		if stderr, err := applyManifest(t, entry(".b")); err != nil {
			t.Fatalf("second apply: %v\n%s", err, stderr)
		}
		prepare()
		var err error
		_, stderr := captureOutErr(t, func() {
			run, _ := newRollbackTestRun()
			err = runRollback(run, "c")
		})
		return stderr, err
	}

	t.Run("rollback conflict", func(t *testing.T) {
		root := cliEngineSetup(t)
		stderr, err := rollbackFrom(t, func() {
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
		root := cliEngineSetup(t)
		stderr, err := rollbackFrom(t, func() { readOnlyDir(t, filepath.Join(root, "ro")) })
		if err == nil {
			t.Fatal("rollback succeeded, want an FS failure")
		}
		assertFailureDetailsNotWarnings(t, stderr, "apply failed; ")
	})
}

// TestEngineWarningsReachStderrAsWarnings drives reset, status and prune through the CLI and checks
// that a warning the engine emits through Warnf reaches stderr as a "layat: warning: <text>" line.
func TestEngineWarningsReachStderrAsWarnings(t *testing.T) {
	for name, dryrun := range map[string]bool{"reset --dryrun": true, "reset": false} {
		t.Run(name, func(t *testing.T) {
			cliEngineSetup(t)
			src := t.TempDir()
			for i, target := range []string{".a", ".b"} {
				entry := manifest.Entry{SrcKind: "store", Src: src, Subpath: ".", Target: target, Method: "symlink"}
				if stderr, err := applyManifest(t, entry); err != nil {
					t.Fatalf("apply %d: %v\n%s", i+1, err, stderr)
				}
			}
			// Point the profile back at generation 1 so the record (generation 2) disagrees with it.
			profile := findProfileLink(t, os.Getenv("XDG_STATE_HOME"))
			if err := os.Remove(profile); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("profile-1-link", profile); err != nil {
				t.Fatal(err)
			}
			origYes := flagYes
			t.Cleanup(func() { flagYes = origYes })
			flagYes = true
			var err error
			_, stderr := captureOutErr(t, func() {
				run, _ := newResetTestRun()
				err = runReset(run, "c", nil, dryrun)
			})
			if err != nil {
				t.Fatalf("runReset: %v\n%s", err, stderr)
			}
			assertEngineWarning(t, stderr)
		})
	}

	t.Run("status", func(t *testing.T) {
		orig := statusFn
		t.Cleanup(func() { statusFn = orig })
		statusFn = func(opts engine.StatusOptions) ([]engine.StatusSeries, error) {
			if opts.Warnf == nil {
				t.Fatal("the CLI passed no Warnf to the engine")
			}
			opts.Warnf("skipping the system record base (%s): %v", "/s", errors.New("boom"))
			return nil, nil
		}
		_, stderr := captureOutErr(t, func() {
			run, _ := newStatusTestRun()
			if err := runStatus(run, nil); err != nil {
				t.Errorf("runStatus: %v", err)
			}
		})
		assertEngineWarning(t, stderr)
	})

	t.Run("prune", func(t *testing.T) {
		withPruneTestState(t)
		pruneFn = func(opts engine.PruneOptions) (*engine.PruneResult, error) {
			if opts.Warnf == nil {
				t.Fatal("the CLI passed no Warnf to the engine")
			}
			opts.Warnf("prune: skipped series %s (%s): %v", "/s/x", "locked", errors.New("busy"))
			return &engine.PruneResult{}, nil
		}
		_, stderr := captureOutErr(t, func() {
			run, _ := newPruneTestRun()
			if err := runPrune(run, true, false); err != nil {
				t.Errorf("runPrune: %v", err)
			}
		})
		assertEngineWarning(t, stderr)
	})
}

// findProfileLink returns the one profile link (a symlink named "profile") under base.
func findProfileLink(t *testing.T, base string) string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "profile" && d.Type()&fs.ModeSymlink != 0 {
			found = append(found, path)
		}
		return nil
	})
	if err != nil || len(found) != 1 {
		t.Fatalf("profile links under %s = %v (err %v), want one", base, found, err)
	}
	return found[0]
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
