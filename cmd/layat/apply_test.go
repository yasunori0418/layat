package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

// TestExitCodeFor: only layat's own exitError sets the exit code; a wrapped child process's
// *exec.ExitError exits 1, not with the child's code.
func TestExitCodeFor(t *testing.T) {
	childErr := exec.Command("sh", "-c", "exit 3").Run()
	var xe *exec.ExitError
	if !errors.As(childErr, &xe) || xe.ExitCode() != 3 {
		t.Fatalf("sh -c 'exit 3' = %v, want an *exec.ExitError with code 3", childErr)
	}
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"plain error", errors.New("boom"), 1},
		{"wrapped child exit", fmt.Errorf("cannot run nix build: %w", childErr), 1},
		{"exitError", &exitError{code: 2}, 2},
		{"wrapped exitError", fmt.Errorf("apply: %w", &exitError{code: 2}), 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := exitCodeFor(c.err); got != c.want {
				t.Errorf("exitCodeFor(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

// TestApplyAllExitCode: priority error(1) > conflict(2) > 0, not the plain maximum.
func TestApplyAllExitCode(t *testing.T) {
	cases := []struct {
		name              string
		anyError, anyConf bool
		want              int
	}{
		{"none", false, false, 0},
		{"error only", true, false, 1},
		{"conflict only", false, true, 2},
		{"error wins over conflict", true, true, 1}, // the maximum would be 2, but do not hide error
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := applyAllExitCode(c.anyError, c.anyConf); got != c.want {
				t.Errorf("applyAllExitCode(%v, %v) = %d, want %d", c.anyError, c.anyConf, got, c.want)
			}
		})
	}
}

// captureStdout captures and returns stdout produced while f runs (verifies that the dryrun plan owns stdout).
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	_ = w.Close()
	os.Stdout = old
	return <-done
}

// captureStderr is captureStdout's stderr counterpart.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	_ = w.Close()
	os.Stderr = old
	return <-done
}

// TestAggregateDryRun: the exit code is error(1) > conflict(2) > 0 and the plan goes to stdout,
// with the apply injected.
func TestAggregateDryRun(t *testing.T) {
	clean := func(name string, _, _ func(string, ...any)) (*engine.Result, error) {
		return &engine.Result{Placed: []string{"/p/" + name}}, nil
	}
	conflict := func(name string, _, _ func(string, ...any)) (*engine.Result, error) {
		return &engine.Result{Conflicts: []planner.Conflict{
			{Entry: manifest.Entry{Target: "/c/" + name}, Reason: "occupied"},
		}}, nil
	}
	fail := func(name string, _, _ func(string, ...any)) (*engine.Result, error) {
		return nil, io.EOF
	}

	cases := []struct {
		name     string
		apply    configApplyFn
		selected []string
		wantCode int
	}{
		{"all clean → 0", clean, []string{"a", "b"}, 0},
		{"conflict → 2", func(n string, _, _ func(string, ...any)) (*engine.Result, error) {
			if n == "b" {
				return conflict(n, nil, nil)
			}
			return clean(n, nil, nil)
		}, []string{"a", "b"}, 2},
		{"error → 1", func(n string, _, _ func(string, ...any)) (*engine.Result, error) {
			if n == "b" {
				return fail(n, nil, nil)
			}
			return clean(n, nil, nil)
		}, []string{"a", "b"}, 1},
		{"error wins over conflict → 1", func(n string, _, _ func(string, ...any)) (*engine.Result, error) {
			switch n {
			case "a":
				return conflict(n, nil, nil)
			case "b":
				return fail(n, nil, nil)
			}
			return clean(n, nil, nil)
		}, []string{"a", "b", "c"}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, _ := newApplyTestRun()
			var got int
			out := captureStdout(t, func() { got = aggregateDryRun(run, c.selected, 4, c.apply) })
			if got != c.wantCode {
				t.Errorf("aggregateDryRun code = %d, want %d", got, c.wantCode)
			}
			// A successful config's plan goes to stdout.
			if strings.Contains(c.name, "all clean") && !strings.Contains(out, "place\t/p/a") {
				t.Errorf("plan not emitted to stdout: %q", out)
			}
		})
	}
}

// TestAggregateApply: the applied / skipped / failure counts and continue-on-partial-failure, with
// the apply injected.
func TestAggregateApply(t *testing.T) {
	clean := func(name string, _, _ func(string, ...any)) (*engine.Result, error) {
		return &engine.Result{Placed: []string{"/p/" + name}}, nil
	}
	skipped := func(string, func(string, ...any), func(string, ...any)) (*engine.Result, error) {
		return nil, engine.ErrSkipped
	}
	failed := func(string, func(string, ...any), func(string, ...any)) (*engine.Result, error) {
		return nil, io.EOF
	}

	cases := []struct {
		name         string
		apply        configApplyFn
		selected     []string
		wantApplied  int
		wantSkipped  int
		wantFailures int
	}{
		{"all succeed → 0 failures", clean, []string{"a", "b"}, 2, 0, 0},
		{"ErrSkipped counts as skip, not failure", func(n string, _, _ func(string, ...any)) (*engine.Result, error) {
			if n == "b" {
				return skipped(n, nil, nil)
			}
			return clean(n, nil, nil)
		}, []string{"a", "b"}, 1, 1, 0},
		{"error counts as failure and continues", func(n string, _, _ func(string, ...any)) (*engine.Result, error) {
			if n == "b" {
				return failed(n, nil, nil)
			}
			return clean(n, nil, nil)
		}, []string{"a", "b", "c"}, 2, 0, 1},
		{"mixed skip and failure both continue", func(n string, _, _ func(string, ...any)) (*engine.Result, error) {
			switch n {
			case "a":
				return skipped(n, nil, nil)
			case "b":
				return failed(n, nil, nil)
			}
			return clean(n, nil, nil)
		}, []string{"a", "b", "c"}, 1, 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, _ := newApplyTestRun()
			applied, skippedCount, failures := aggregateApply(run, c.selected, 4, c.apply)
			if applied != c.wantApplied || skippedCount != c.wantSkipped || failures != c.wantFailures {
				t.Errorf("aggregateApply() = (applied=%d, skipped=%d, failures=%d), want (applied=%d, skipped=%d, failures=%d)",
					applied, skippedCount, failures, c.wantApplied, c.wantSkipped, c.wantFailures)
			}
		})
	}
}

func TestSelectedRootFilter(t *testing.T) {
	defer func() { flagProjectRoot, flagHomeRoot, flagSystemRoot = false, false, false }()

	t.Run("none", func(t *testing.T) {
		flagProjectRoot, flagHomeRoot, flagSystemRoot = false, false, false
		got, err := selectedRootFilter()
		if err != nil || got != "" {
			t.Fatalf("got (%q, %v), want (\"\", nil)", got, err)
		}
	})
	t.Run("project", func(t *testing.T) {
		flagProjectRoot, flagHomeRoot, flagSystemRoot = true, false, false
		got, err := selectedRootFilter()
		if err != nil || got != "project" {
			t.Fatalf("got (%q, %v), want (\"project\", nil)", got, err)
		}
	})
	t.Run("multiple → error", func(t *testing.T) {
		flagProjectRoot, flagHomeRoot, flagSystemRoot = true, true, false
		if _, err := selectedRootFilter(); err == nil {
			t.Fatal("specifying multiple should be an error")
		}
	})
}

// TestPrintApplyPlanBackupLine: apply --dryrun's plan includes a "backup\t<target>" line.
func TestPrintApplyPlanBackupLine(t *testing.T) {
	out := captureStdout(t, func() {
		printApplyPlan(&engine.Result{BackedUp: []string{".config/foo"}})
	})
	if !strings.Contains(out, "backup\t.config/foo") {
		t.Errorf("printApplyPlan output = %q, want a line containing %q", out, "backup\t.config/foo")
	}
}

// TestReportResultBackupLineAndNoOp: the -v report shows "backedUp <target>" for a backup-only run
// and does not report it as "no-op".
func TestReportResultBackupLineAndNoOp(t *testing.T) {
	out := captureStderr(t, func() {
		reportResult(&engine.Result{Root: "/root", BackedUp: []string{".config/foo"}}, "c")
	})
	if !strings.Contains(out, "backedUp .config/foo") {
		t.Errorf("reportResult output = %q, want a line containing %q", out, "backedUp .config/foo")
	}
	if strings.Contains(out, "no-op") {
		t.Errorf("reportResult output = %q, must not report no-op when BackedUp is non-empty", out)
	}
}

// TestReportResultAllEmptyIsNoOp: a fully empty Result reports "no-op".
func TestReportResultAllEmptyIsNoOp(t *testing.T) {
	out := captureStderr(t, func() {
		reportResult(&engine.Result{Root: "/root"}, "c")
	})
	if !strings.Contains(out, "no-op") {
		t.Errorf("reportResult output = %q, want it to report no-op for a fully empty Result", out)
	}
}

// TestRunApplyAllSystemRootSelectsSystem drives runApplyAll through a stub nix with --system-root:
// only the systemRoot config reaches the build; the home / project ones are never built.
func TestRunApplyAllSystemRootSelectsSystem(t *testing.T) {
	withFlakeEntrypoint(t)
	origSystem := flagSystemRoot
	t.Cleanup(func() { flagSystemRoot = origSystem })
	flagDryrun, flagSystemRoot = false, true
	log := stubNixForApply(t, `{"h":{"rootKind":"home","targets":["a"]},"p":{"rootKind":"project","targets":["b"]},"s":{"rootKind":"system","targets":["etc/c"]}}`)
	run, _ := newApplyTestRun()
	captureStderr(t, func() {
		// The stub build fails, so the run errors on the selected config's build.
		if err := runApplyAll(run); err == nil {
			t.Error("runApplyAll must surface the stub build failure")
		}
	})
	var built []string
	for _, c := range nixCalls(t, log) {
		if !strings.HasPrefix(c, "build ") {
			continue
		}
		for _, f := range strings.Fields(c) {
			if strings.Contains(f, "#") {
				built = append(built, f[strings.LastIndex(f, ".")+1:])
			}
		}
	}
	if len(built) != 1 || built[0] != "s" {
		t.Errorf("built configs = %v, want [s] (--system-root selects systemRoot configs only)", built)
	}
}

// TestApplyManifestDryRunIsReadOnly drives apply c --manifest <link-farm> --dryrun through the
// real engine on the Fake backend: it prints the plan and writes nothing under the root or the
// state directory, and a conflict exits 2 leaving the existing file as it was.
func TestApplyManifestDryRunIsReadOnly(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := func(target string) manifest.Entry {
		return manifest.Entry{SrcKind: "store", Src: src, Subpath: "f", Target: target, Method: "symlink"}
	}
	// dryrun returns stdout, the --json items of subject c and runApply's error.
	dryrun := func(t *testing.T, entries ...manifest.Entry) (string, []any, error) {
		t.Helper()
		origFile, origManifest, origDryrun := flagFile, flagManifest, flagDryrun
		t.Cleanup(func() { flagFile, flagManifest, flagDryrun = origFile, origManifest, origDryrun })
		flagFile, flagManifest, flagDryrun = "", []string{writeTestLinkFarm(t, entries...)}, true
		run, buf := newApplyTestRun()
		var err error
		stdout, _ := captureOutErr(t, func() { err = runApply(run, "c") })
		if emitErr := run.emit(err); emitErr != nil {
			t.Fatalf("emit: %v", emitErr)
		}
		byName, _ := subjectResults(t, buf)
		items, _ := byName["c"]["result"].(map[string]any)["items"].([]any)
		return stdout, items, err
	}
	assertNoState := func(t *testing.T) {
		t.Helper()
		state := os.Getenv("XDG_STATE_HOME")
		err := filepath.WalkDir(state, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Name() == "profile" || d.Name() == "last-applied.json" {
				t.Errorf("dryrun wrote %s", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("no conflict", func(t *testing.T) {
		root := cliEngineSetup(t)
		stdout, items, err := dryrun(t, entry(".a"))
		if err != nil {
			t.Fatalf("dryrun = %v, want nil", err)
		}
		if len(items) != 1 {
			t.Errorf("--json items = %v, want the one planned entry", items)
		}
		if !strings.Contains(stdout, "place\t") {
			t.Errorf("stdout has no place line:\n%s", stdout)
		}
		if _, err := os.Lstat(filepath.Join(root, ".a")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("dryrun placed %s (lstat: %v)", filepath.Join(root, ".a"), err)
		}
		assertNoState(t)
	})

	t.Run("conflict", func(t *testing.T) {
		root := cliEngineSetup(t)
		foreign := filepath.Join(root, ".a")
		if err := os.WriteFile(foreign, []byte("foreign"), 0o644); err != nil {
			t.Fatal(err)
		}
		stdout, items, err := dryrun(t, entry(".a"))
		var xe *exitError
		if !errors.As(err, &xe) || xe.code != 2 {
			t.Fatalf("dryrun = %v, want an exitError with code 2", err)
		}
		if !strings.Contains(stdout, "conflict\t") {
			t.Errorf("stdout has no conflict line:\n%s", stdout)
		}
		if len(items) != 1 {
			t.Errorf("--json items = %v, want the one conflicted entry", items)
		}
		if got, err := os.ReadFile(foreign); err != nil || string(got) != "foreign" {
			t.Errorf("%s = %q (err %v), want it unchanged", foreign, got, err)
		}
		assertNoState(t)
	})
}

// implicitInputProject moves into a fresh git repository holding flake.nix with -f, --generator,
// --root and --manifest cleared, on the Fake backend. A fake nix first on PATH answers the
// rootKind eval with kind, the batch eval with evalAll, and every build with a link-farm of
// rootKind kind and no entries. It returns the repository's path.
func implicitInputProject(t *testing.T, kind, evalAll string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	withFlakeEntrypoint(t)
	installFakeBackend(t, "")
	withManifestFlags(t, false)
	flagFile, flagGenerator, flagRoot = "", "", ""
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(repo, "flake.nix"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	farm := writeRootKindLinkFarm(t, kind)
	bin := t.TempDir()
	data := filepath.Join(bin, "eval-all.json")
	if err := os.WriteFile(data, []byte(evalAll), 0o644); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n" +
		"case \"$1\" in\n" +
		"eval) case \" $* \" in *' --apply '*) cat '" + data + "';; *) printf '" + kind + "';; esac;;\n" +
		"build) case \" $* \" in\n" +
		"  *' --out-link '*) prev=; for a; do [ \"$prev\" = --out-link ] && ln -s '" + farm + "' \"$a\"; prev=$a; done;;\n" +
		"  *) echo '" + farm + "';;\n" +
		"  esac;;\n" +
		"*) exit 1;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(bin, "nix"), []byte(script), 0o755); err != nil {
		t.Fatalf("write nix stub: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return repo
}

// envelopeInfo returns the envelope's top-level info and results[i].result.info, nil where absent.
func envelopeInfo(t *testing.T, doc map[string]any, i int) (top, result map[string]any) {
	t.Helper()
	top, _ = doc["info"].(map[string]any)
	results, _ := doc["results"].([]any)
	if i >= len(results) {
		t.Fatalf("results = %v, want an element at %d", results, i)
	}
	res, _ := results[i].(map[string]any)["result"].(map[string]any)
	result, _ = res["info"].(map[string]any)
	return top, result
}

// TestApplyImplicitInputRecord: apply records in info only the inputs it resolved itself. With
// -f, --generator and --root all given there is no info at either level; with all three omitted
// the top-level info holds the discovered entrypoint and the adopted generator and result.info
// holds the git toplevel; --manifest records neither entrypoint nor generator but still the root.
func TestApplyImplicitInputRecord(t *testing.T) {
	apply := func(t *testing.T) map[string]any {
		t.Helper()
		run, buf := newApplyTestRun()
		var err error
		captureOutErr(t, func() { err = runApply(run, "c") })
		if err != nil {
			t.Fatalf("runApply: %v", err)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		return decodeEnvelope(t, buf)
	}

	t.Run("all given", func(t *testing.T) {
		repo := implicitInputProject(t, manifest.RootKindProject, "")
		flagFile, flagGenerator, flagRoot = repo, "nix", t.TempDir()
		assertNoInfoKeys(t, apply(t))
	})

	t.Run("all omitted", func(t *testing.T) {
		repo := implicitInputProject(t, manifest.RootKindProject, "")
		top, result := envelopeInfo(t, apply(t), 0)
		if top["entrypoint"] != repo || top["generator"] != "nix" || len(top) != 2 {
			t.Errorf("info = %v, want {entrypoint: %s, generator: nix}", top, repo)
		}
		if result["root"] != repo || len(result) != 1 {
			t.Errorf("result.info = %v, want {root: %s}", result, repo)
		}
	})

	t.Run("dryrun with all omitted", func(t *testing.T) {
		repo := implicitInputProject(t, manifest.RootKindProject, "")
		flagDryrun = true
		top, result := envelopeInfo(t, apply(t), 0)
		if top["entrypoint"] != repo || top["generator"] != "nix" {
			t.Errorf("info = %v, want the discovered entrypoint and nix", top)
		}
		if result["root"] != repo {
			t.Errorf("result.info = %v, want {root: %s}", result, repo)
		}
	})

	t.Run("--manifest", func(t *testing.T) {
		repo := implicitInputProject(t, manifest.RootKindProject, "")
		flagManifest = []string{writeRootKindLinkFarm(t, manifest.RootKindProject)}
		doc := apply(t)
		if v, ok := doc["info"]; ok {
			t.Errorf("info = %v, want absent (--manifest discovers no entrypoint and selects no generator)", v)
		}
		if _, result := envelopeInfo(t, doc, 0); result["root"] != repo {
			t.Errorf("result.info = %v, want {root: %s}", result, repo)
		}
	})

	t.Run("-f given", func(t *testing.T) {
		repo := implicitInputProject(t, manifest.RootKindProject, "")
		flagFile = repo
		top, _ := envelopeInfo(t, apply(t), 0)
		if len(top) != 1 || top["generator"] != "nix" {
			t.Errorf("info = %v, want {generator: nix} (-f is explicit)", top)
		}
	})

	t.Run("--generator given", func(t *testing.T) {
		repo := implicitInputProject(t, manifest.RootKindProject, "")
		flagGenerator = "nix"
		top, _ := envelopeInfo(t, apply(t), 0)
		if len(top) != 1 || top["entrypoint"] != repo {
			t.Errorf("info = %v, want {entrypoint: %s} (--generator is explicit)", top, repo)
		}
	})

	t.Run("--root given on a home config", func(t *testing.T) {
		implicitInputProject(t, manifest.RootKindHome, "")
		flagRoot = t.TempDir()
		top, result := envelopeInfo(t, apply(t), 0)
		if _, ok := top["entrypoint"]; !ok {
			t.Errorf("info = %v, want entrypoint", top)
		}
		if result != nil {
			t.Errorf("result.info = %v, want absent (--root is explicit)", result)
		}
	})

	t.Run("home root", func(t *testing.T) {
		implicitInputProject(t, manifest.RootKindHome, "")
		if _, result := envelopeInfo(t, apply(t), 0); result["root"] != os.Getenv("HOME") {
			t.Errorf("result.info = %v, want {root: $HOME}", result)
		}
	})
}
