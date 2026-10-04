package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/engine"
)

// withPruneTestState snapshots the engine seam and flags prune's tests overwrite and restores them
// when t finishes.
func withPruneTestState(t *testing.T) {
	t.Helper()
	fn := pruneFn
	yes, jsonMode, verbose := flagYes, flagJSON, flagVerbose
	t.Cleanup(func() {
		pruneFn = fn
		flagYes, flagJSON, flagVerbose = yes, jsonMode, verbose
	})
}

// pruneFixture is a two-series result, one under each scan base, so "every root path" needs two
// lines.
func pruneFixture() *engine.PruneResult {
	return &engine.PruneResult{
		Removed: []engine.PruneSeries{
			{
				RootHash: "aaaa",
				Root:     "/home/u/src/gone",
				Names:    []string{"default", "docs"},
				Dir:      "/state/nix/profiles/layat/aaaa",
			},
			{
				RootHash: "bbbb",
				Root:     "/mnt/removable/proj",
				Names:    nil,
				Dir:      "/nix/var/nix/profiles/layat/bbbb",
			},
		},
		Skipped: []engine.PruneSkipped{
			{
				Series: engine.PruneSeries{
					RootHash: "cccc",
					Root:     "/home/u/src/busy",
					Names:    []string{"default"},
					Dir:      "/state/nix/profiles/layat/cccc",
				},
				Reason: engine.PruneSkipLocked,
				Detail: "layat: profile directory is locked by another process",
			},
		},
	}
}

// TestPrunePromptAllowed pins prune's prompt gate: a prompt needs a TTY, and --json forbids it.
func TestPrunePromptAllowed(t *testing.T) {
	cases := []struct {
		name        string
		interactive bool
		jsonMode    bool
		want        bool
	}{
		{"TTY without --json prompts", true, false, true},
		{"TTY with --json never prompts", true, true, false},
		{"non-TTY without --json cannot prompt", false, false, false},
		{"non-TTY with --json cannot prompt", false, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := prunePromptAllowed(c.interactive, c.jsonMode); got != c.want {
				t.Errorf("prunePromptAllowed(%v, %v) = %v, want %v", c.interactive, c.jsonMode, got, c.want)
			}
		})
	}
}

// TestPruneJSONWithoutYesFailsFast: with --json and no --yes, the policy refuses before any scan,
// even on a TTY.
func TestPruneJSONWithoutYesFailsFast(t *testing.T) {
	needPrompt, err := confirmPolicy(false, prunePromptAllowed(true, true), "prune")
	if err == nil {
		t.Fatal("--json without --yes must fail fast, got no error")
	}
	if needPrompt {
		t.Error("the refused policy must not ask for a prompt")
	}
	// The refusal names prune, not the command the shared policy was written for.
	if !strings.Contains(err.Error(), "prune") {
		t.Errorf("refusal = %q, want it to name prune", err)
	}
	// --yes restores the run, still without a prompt: the machine path never asks.
	needPrompt, err = confirmPolicy(true, prunePromptAllowed(true, true), "prune")
	if err != nil {
		t.Fatalf("--json with --yes must proceed, got %v", err)
	}
	if needPrompt {
		t.Error("--yes must skip the prompt")
	}
}

// TestPruneOutputStreams pins prune's streams: the --dryrun plan goes to stdout tab-separated, and
// the confirmation listing goes to stderr.
func TestPruneOutputStreams(t *testing.T) {
	res := pruneFixture()

	t.Run("printPrunePlan owns stdout exclusively", func(t *testing.T) {
		withPruneTestState(t)
		flagJSON = false

		out, errOut := captureOutErr(t, func() { printPrunePlan(res) })
		for _, want := range []string{
			"remove-series\taaaa\t/home/u/src/gone\tdefault,docs\n",
			"remove-series\tbbbb\t/mnt/removable/proj\t\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("plan missing %q in stdout: %q", want, out)
			}
		}
		if strings.Contains(errOut, "remove-series") {
			t.Errorf("plan should not write the machine-readable lines to stderr: %q", errOut)
		}
	})

	t.Run("printPrunePlan is silent under --json", func(t *testing.T) {
		withPruneTestState(t)
		flagJSON = true

		out, _ := captureOutErr(t, func() { printPrunePlan(res) })
		if out != "" {
			t.Errorf("--json must leave stdout to the envelope alone, got %q", out)
		}
	})

	t.Run("printPrunePlan keeps the stderr notice with nothing to delete", func(t *testing.T) {
		// stdout stays empty with nothing to list, while the stderr notice survives --json.
		withPruneTestState(t)

		for _, jsonMode := range []bool{false, true} {
			flagJSON = jsonMode
			out, errOut := captureOutErr(t, func() { printPrunePlan(&engine.PruneResult{}) })
			if out != "" {
				t.Errorf("flagJSON=%v: stdout = %q, want empty for an empty plan", jsonMode, out)
			}
			if !strings.Contains(errOut, "nothing to delete") {
				t.Errorf("flagJSON=%v: stderr = %q, want the nothing-to-delete notice", jsonMode, errOut)
			}
		}
	})

	t.Run("reportPruneTargets lists every root path on stderr", func(t *testing.T) {
		out, errOut := captureOutErr(t, func() { reportPruneTargets(res) })
		if out != "" {
			t.Errorf("the confirmation listing pollutes stdout: %q", out)
		}
		// Every root path has to be shown. The expectations are literals so they cannot agree with the
		// value under test by construction.
		for _, root := range []string{"/home/u/src/gone", "/mnt/removable/proj"} {
			if !strings.Contains(errOut, root) {
				t.Errorf("root %q missing from the confirmation listing: %q", root, errOut)
			}
		}
		// The header count has to match what is listed.
		if got := strings.Count(errOut, "  root "); got != 2 {
			t.Errorf("root lines = %d, want 2: %q", got, errOut)
		}
		if !strings.Contains(errOut, "delete 2 orphan profile series") {
			t.Errorf("the header does not announce the 2 series it listed: %q", errOut)
		}
	})

	t.Run("reportPruneResult owns stderr exclusively", func(t *testing.T) {
		out, errOut := captureOutErr(t, func() { reportPruneResult(res) })
		if out != "" {
			t.Errorf("the result report pollutes stdout: %q", out)
		}
		if !strings.Contains(errOut, "removed-series") || !strings.Contains(errOut, "/home/u/src/gone") {
			t.Errorf("result report not emitted to stderr: %q", errOut)
		}

		// A run that deleted nothing says so rather than reporting an empty list.
		out, errOut = captureOutErr(t, func() { reportPruneResult(&engine.PruneResult{}) })
		if out != "" {
			t.Errorf("the empty result report pollutes stdout: %q", out)
		}
		if !strings.Contains(errOut, "no-op") {
			t.Errorf("an empty result must report no-op: %q", errOut)
		}
	})
}

// TestPruneConfirmShowsThePreview: the callback runPrune hands the engine lists the preview it is
// given, not the returned result (which is empty on an abort).
func TestPruneConfirmShowsThePreview(t *testing.T) {
	restore := withStdin(t, "y\n")
	defer restore()

	preview := pruneFixture()
	var errOut string
	var ok bool
	var err error
	_, errOut = captureOutErr(t, func() { ok, err = prunePrompt(preview) })
	if err != nil {
		t.Fatalf("prunePrompt: %v", err)
	}
	if !ok {
		t.Error("an explicit yes must confirm")
	}
	for _, root := range []string{"/home/u/src/gone", "/mnt/removable/proj"} {
		if !strings.Contains(errOut, root) {
			t.Errorf("root %q of the preview missing from the prompt: %q", root, errOut)
		}
	}
	if !strings.Contains(errOut, "Continue?") {
		t.Errorf("the prompt itself is missing: %q", errOut)
	}
}

// TestPruneRunDrivesTheEngine pins runPrune's orchestration against a stubbed engine: what reaches
// the prompt and the envelope, when confirm is passed, what -v reports, and refused / declined runs.
func TestPruneRunDrivesTheEngine(t *testing.T) {
	// Each subtest restores its own state via withPruneTestState; this is the outer net.
	withPruneTestState(t)
	flagJSON, flagVerbose = false, false
	// go test's stdin is never a TTY, so interactivity is passed in.
	const interactive = true

	// The preview and the returned result use disjoint series, so a prompt or envelope built from the
	// wrong one fails.
	previewRoot, resultRoot := "/home/u/src/candidate", "/home/u/src/deleted"
	preview := &engine.PruneResult{Removed: []engine.PruneSeries{
		{RootHash: "cand", Root: previewRoot, Dir: "/state/nix/profiles/layat/cand"},
	}}
	result := &engine.PruneResult{Removed: []engine.PruneSeries{
		{RootHash: "done", Root: resultRoot, Dir: "/state/nix/profiles/layat/done"},
	}}

	t.Run("the prompt sees the preview and the envelope sees the result", func(t *testing.T) {
		withPruneTestState(t)
		flagYes = false
		restore := withStdin(t, "y\n")
		defer restore()

		pruneFn = func(opts engine.PruneOptions) (*engine.PruneResult, error) {
			if opts.Confirm == nil {
				t.Fatal("a run without --yes must pass a confirmation callback")
			}
			if _, err := opts.Confirm(preview); err != nil {
				return nil, err
			}
			return result, nil
		}

		run, buf := newPruneTestRun()
		_, errOut := captureOutErr(t, func() {
			if err := runPrune(run, false, interactive); err != nil {
				t.Errorf("runPrune: %v", err)
			}
		})
		if !strings.Contains(errOut, previewRoot) {
			t.Errorf("the prompt did not list the preview's root: %q", errOut)
		}
		if strings.Contains(errOut, resultRoot) {
			t.Errorf("the prompt listed the returned result instead of the preview: %q", errOut)
		}
		// The inventory is the other way round: it reports what was deleted, not what was judged.
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		info := decodeEnvelope(t, buf)["info"].(map[string]any)
		removed := info["removed"].([]any)
		if len(removed) != 1 || removed[0].(map[string]any)["root"] != resultRoot {
			t.Errorf("info.removed = %v, want the one series the run deleted", removed)
		}
	})

	t.Run("--yes passes no callback", func(t *testing.T) {
		withPruneTestState(t)
		flagYes = true
		var sawConfirm bool
		pruneFn = func(opts engine.PruneOptions) (*engine.PruneResult, error) {
			sawConfirm = opts.Confirm != nil
			return result, nil
		}
		run, _ := newPruneTestRun()
		_, errOut := captureOutErr(t, func() {
			if err := runPrune(run, false, interactive); err != nil {
				t.Errorf("runPrune: %v", err)
			}
		})
		if sawConfirm {
			t.Error("--yes must skip the prompt entirely (no callback)")
		}
		// --yes skips the listing with the prompt, and the run stays quiet without -v.
		if strings.Contains(errOut, resultRoot) {
			t.Errorf("--yes must not print the series listing: %q", errOut)
		}
	})

	t.Run("--verbose reports what was deleted", func(t *testing.T) {
		withPruneTestState(t)
		flagYes = true
		pruneFn = func(engine.PruneOptions) (*engine.PruneResult, error) { return result, nil }

		flagVerbose = true
		run, _ := newPruneTestRun()
		out, errOut := captureOutErr(t, func() {
			if err := runPrune(run, false, interactive); err != nil {
				t.Errorf("runPrune: %v", err)
			}
		})
		if out != "" {
			t.Errorf("the -v report pollutes stdout: %q", out)
		}
		if !strings.Contains(errOut, resultRoot) {
			t.Errorf("-v must report the deleted series: %q", errOut)
		}
	})

	t.Run("a mid-deletion failure still reports what is already gone", func(t *testing.T) {
		// A run that failed partway still reports the series already deleted.
		withPruneTestState(t)
		flagYes = true
		failure := errors.New("layat: cannot remove series done")
		pruneFn = func(engine.PruneOptions) (*engine.PruneResult, error) { return result, failure }

		run, buf := newPruneTestRun()
		var runErr error
		_, _ = captureOutErr(t, func() { runErr = runPrune(run, false, interactive) })
		if runErr == nil {
			t.Fatal("a mid-deletion failure must surface as an error")
		}
		if err := run.emit(runErr); err != nil {
			t.Fatalf("emit: %v", err)
		}
		doc := decodeEnvelope(t, buf)
		if doc["status"] != "error" {
			t.Errorf("status = %v, want error", doc["status"])
		}
		removed := doc["info"].(map[string]any)["removed"].([]any)
		if len(removed) != 1 || removed[0].(map[string]any)["root"] != resultRoot {
			t.Errorf("info.removed = %v, want the series deleted before the failure", removed)
		}
	})

	t.Run("a refused policy never reaches the engine", func(t *testing.T) {
		// Non-interactive without --yes is refused before the scan: nothing is read or deleted.
		withPruneTestState(t)
		flagYes = false

		var called bool
		pruneFn = func(engine.PruneOptions) (*engine.PruneResult, error) {
			called = true
			return result, nil
		}
		run, buf := newPruneTestRun()
		var runErr error
		_, _ = captureOutErr(t, func() { runErr = runPrune(run, false, false) })
		if runErr == nil {
			t.Fatal("a non-interactive run without --yes must fail")
		}
		if called {
			t.Error("the refusal must come before the engine is driven")
		}
		// Nothing was scanned, so the envelope carries no inventory.
		if err := run.emit(runErr); err != nil {
			t.Fatalf("emit: %v", err)
		}
		assertNoInfoKeys(t, decodeEnvelope(t, buf))
	})

	t.Run("an aborted run reports on stderr and succeeds", func(t *testing.T) {
		withPruneTestState(t)
		flagYes = true
		pruneFn = func(engine.PruneOptions) (*engine.PruneResult, error) {
			// A declined run as the engine reports it: Aborted, with the judged candidates left in Removed.
			return &engine.PruneResult{Removed: preview.Removed, Aborted: true}, nil
		}
		run, buf := newPruneTestRun()
		var runErr error
		out, errOut := captureOutErr(t, func() { runErr = runPrune(run, false, interactive) })
		if runErr != nil {
			// Declining is not a failure: the exit code stays 0.
			t.Errorf("an aborted run must not fail: %v", runErr)
		}
		if out != "" {
			t.Errorf("the abort notice pollutes stdout: %q", out)
		}
		if !strings.Contains(errOut, "prune aborted") {
			t.Errorf("the abort notice is missing from stderr: %q", errOut)
		}
		// An aborted run must not reach the -v report.
		if strings.Contains(errOut, "removed-series") {
			t.Errorf("an aborted run reported deletions: %q", errOut)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		doc := decodeEnvelope(t, buf)
		if doc["status"] != "success" {
			t.Errorf("status = %v, want success (declining is not an error)", doc["status"])
		}
		// Nothing was deleted, so no inventory is emitted.
		assertNoInfoKeys(t, doc)
	})
}

// TestPruneInfoShape pins the --json payload: removed / skipped arrays, the skip reason verbatim
// with its detail, and the field named removed rather than pruned.
func TestPruneInfoShape(t *testing.T) {
	info := pruneInfoFrom(pruneFixture())

	if len(info.Removed) != 2 {
		t.Fatalf("removed = %v, want the two series", info.Removed)
	}
	first := info.Removed[0]
	if first.RootHash != "aaaa" || first.Root != "/home/u/src/gone" ||
		first.Dir != "/state/nix/profiles/layat/aaaa" {
		t.Errorf("removed[0] = %+v, want the state-base series", first)
	}
	if len(first.Names) != 2 || first.Names[0] != "default" || first.Names[1] != "docs" {
		t.Errorf("removed[0].names = %v, want [default docs]", first.Names)
	}
	// A series with only the backref still carries names as an array.
	if second := info.Removed[1]; second.Names == nil || len(second.Names) != 0 {
		t.Errorf("removed[1].names = %v, want []", second.Names)
	}
	if len(info.Skipped) != 1 {
		t.Fatalf("skipped = %v, want the one locked series", info.Skipped)
	}
	sk := info.Skipped[0]
	if sk.Reason != string(engine.PruneSkipLocked) {
		t.Errorf("skipped[0].reason = %q, want %q", sk.Reason, engine.PruneSkipLocked)
	}
	if sk.Detail == "" {
		t.Error("skipped[0].detail is empty; the reason alone does not say which failure it was")
	}
	if sk.Series.RootHash != "cccc" || sk.Series.Root != "/home/u/src/busy" {
		t.Errorf("skipped[0] series = %+v, want the cccc series", sk.Series)
	}
}

// TestPruneInfoSkipReasonsAreTheEngineVocabulary: every engine skip reason reaches the document
// verbatim.
func TestPruneInfoSkipReasonsAreTheEngineVocabulary(t *testing.T) {
	reasons := []engine.PruneSkipReason{
		engine.PruneSkipLocked,
		engine.PruneSkipBackrefUnreadable,
		engine.PruneSkipSeriesUnreadable,
		engine.PruneSkipRootStatFailed,
		engine.PruneSkipPermissionDenied,
	}
	res := &engine.PruneResult{}
	for _, r := range reasons {
		res.Skipped = append(res.Skipped, engine.PruneSkipped{
			Series: engine.PruneSeries{RootHash: string(r), Root: "/gone", Dir: "/state/" + string(r)},
			Reason: r,
			Detail: "detail",
		})
	}
	info := pruneInfoFrom(res)
	if len(info.Skipped) != len(reasons) {
		t.Fatalf("skipped = %d entries, want %d", len(info.Skipped), len(reasons))
	}
	for i, r := range reasons {
		if info.Skipped[i].Reason != string(r) {
			t.Errorf("skipped[%d].reason = %q, want %q", i, info.Skipped[i].Reason, r)
		}
	}
}

// TestPruneJSONEnvelope pins prune's envelope shape: results stays [] and the inventory rides in the
// envelope-wide info. The document is conformant.
func TestPruneJSONEnvelope(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}

	r, buf := newPruneTestRun()
	r.setEnvelopeInfo(pruneInfoFrom(pruneFixture()))
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings: %v\ndocument: %s", findings, buf.String())
	}

	doc := decodeEnvelope(t, buf)
	if results := doc["results"].([]any); len(results) != 0 {
		t.Errorf("results = %v, want [] (prune has no subject)", results)
	}
	info := doc["info"].(map[string]any)
	if _, ok := info["pruned"]; ok {
		t.Error(`info carries a "pruned" key; the deleted series are "removed" (pruned already means the rmdir-ed empty ancestors)`)
	}
	removed := info["removed"].([]any)
	if len(removed) != 2 {
		t.Fatalf("info.removed = %v, want two series", removed)
	}
	if removed[0].(map[string]any)["root"] != "/home/u/src/gone" {
		t.Errorf("info.removed[0].root = %v", removed[0])
	}
	skipped := info["skipped"].([]any)
	if len(skipped) != 1 || skipped[0].(map[string]any)["reason"] != string(engine.PruneSkipLocked) {
		t.Errorf("info.skipped = %v, want the one locked series", skipped)
	}
}

// TestPruneJSONEmptyResultKeepsArrays pins the zero-series boundary: a run that found no orphan
// still emits removed / skipped as empty arrays, so a consumer never has to handle null.
func TestPruneJSONEmptyResultKeepsArrays(t *testing.T) {
	r, buf := newPruneTestRun()
	r.setEnvelopeInfo(pruneInfoFrom(&engine.PruneResult{}))
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	info := decodeEnvelope(t, buf)["info"].(map[string]any)
	for _, key := range []string{"removed", "skipped"} {
		arr, ok := info[key].([]any)
		if !ok {
			t.Fatalf("info[%q] = %v, want an array", key, info[key])
		}
		if len(arr) != 0 {
			t.Errorf("info[%q] = %v, want []", key, arr)
		}
	}
}

// TestPruneJSONEnvelopeInfoAbsentBeforeScan: the --json-without---yes refusal happens before the
// scan, so the envelope's info stays absent.
func TestPruneJSONEnvelopeInfoAbsentBeforeScan(t *testing.T) {
	r, buf := newPruneTestRun()
	if err := r.emit(errors.New("layat: refusing destructive prune without --yes in a non-interactive context")); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertNoInfoKeys(t, decodeEnvelope(t, buf))
}

// TestPruneDryrunDoesNotDelete: the dryrun path builds options with DryRun and no Confirm.
func TestPruneDryrunDoesNotDelete(t *testing.T) {
	// Clear the system base override so the assertions do not depend on the caller's env.
	t.Setenv(systemBaseEnv, "")

	opts := pruneOptions(true, nil)
	if !opts.DryRun {
		t.Error("the dryrun path must set DryRun")
	}
	if opts.Confirm != nil {
		t.Error("the dryrun path must pass no Confirm (a preview never asks)")
	}
	// With the override unset, the CLI leaves both scan bases to the engine.
	if opts.StateDir != "" || opts.SystemDir != "" {
		t.Errorf("the CLI must leave the scan bases to the engine, got %+v", opts)
	}

	confirm := func(*engine.PruneResult) (bool, error) { return true, nil }
	real := pruneOptions(false, confirm)
	if real.DryRun {
		t.Error("the non-dryrun path must not set DryRun")
	}
	if real.Confirm == nil {
		t.Error("the non-dryrun path must pass the confirmation callback through")
	}
	if real.StateDir != "" || real.SystemDir != "" {
		t.Errorf("the CLI must leave the scan bases to the engine, got %+v", real)
	}
}

// TestPruneSystemBaseEnvOverridesTheScanBase: systemBaseEnv reaches PruneOptions.SystemDir verbatim,
// while the user state base is left to the engine. This keeps the destructive e2e path off the
// machine's real /nix/var/nix/profiles/layat.
func TestPruneSystemBaseEnvOverridesTheScanBase(t *testing.T) {
	isolated := filepath.Join(t.TempDir(), "system")
	t.Setenv(systemBaseEnv, isolated)

	for _, c := range []struct {
		name string
		opts engine.PruneOptions
	}{
		{"dryrun", pruneOptions(true, nil)},
		{"destructive", pruneOptions(false, func(*engine.PruneResult) (bool, error) { return true, nil })},
	} {
		t.Run(c.name, func(t *testing.T) {
			if c.opts.SystemDir != isolated {
				t.Errorf("SystemDir = %q, want the override %q", c.opts.SystemDir, isolated)
			}
			// Passed through verbatim, since the engine takes SystemDir as the finished base.
			if c.opts.StateDir != "" {
				t.Errorf("StateDir = %q, want it left to the engine", c.opts.StateDir)
			}
		})
	}

	// An empty value is not an override: it reaches the engine as an empty SystemDir.
	t.Setenv(systemBaseEnv, "")
	if got := pruneOptions(false, nil).SystemDir; got != "" {
		t.Errorf("SystemDir = %q with the override unset, want the engine default", got)
	}
}
