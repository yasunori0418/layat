package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/yasunori0418/niface/go/conformance"

	"github.com/yasunori0418/nput/internal/engine"
)

// pruneTestState snapshots the package-level state prune's tests reach into — the engine seam and
// the flags the run reads — and restores it when t finishes. Each subtest calls it before overwriting
// anything, so a case added later cannot leak its overrides into the ones after it.
func pruneTestState(t *testing.T) {
	t.Helper()
	fn := pruneFn
	yes, jsonMode, verbose := flagYes, flagJSON, flagVerbose
	t.Cleanup(func() {
		pruneFn = fn
		flagYes, flagJSON, flagVerbose = yes, jsonMode, verbose
	})
}

// pruneFixture is the two-series result the output tests drive: one series under the user state
// base and one under the system base, so an assertion about "every root path" cannot be
// satisfied by a single line (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed の root パス一覧).
func pruneFixture() *engine.PruneResult {
	return &engine.PruneResult{
		Removed: []engine.PruneSeries{
			{
				RootHash: "aaaa",
				Root:     "/home/u/src/gone",
				Names:    []string{"default", "docs"},
				Dir:      "/state/nix/profiles/nput/aaaa",
			},
			{
				RootHash: "bbbb",
				Root:     "/mnt/removable/proj",
				Names:    nil,
				Dir:      "/nix/var/nix/profiles/nput/bbbb",
			},
		},
		Skipped: []engine.PruneSkipped{
			{
				Series: engine.PruneSeries{
					RootHash: "cccc",
					Root:     "/home/u/src/busy",
					Names:    []string{"default"},
					Dir:      "/state/nix/profiles/nput/cccc",
				},
				Reason: engine.PruneSkipLocked,
				Detail: "nput: profile directory is locked by another process",
			},
		},
	}
}

// TestPrunePromptAllowed pins prune's prompt gate: a prompt needs a TTY, and --json forbids it
// unconditionally so the confirmPolicy refuse path fires without --yes (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed,
// the same shape as reset's REQ-2a613337-7646-4ced-8807-e43bca18acf3).
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

// TestPruneJSONWithoutYesFailsFast pins the fail-fast composition the two seams produce: with
// --json and no --yes, the policy refuses before anything is scanned rather than prompting
// (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed). The TTY case is the load-bearing one — a
// non-TTY would refuse even without --json.
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

// TestPruneOutputStreams pins prune's stream discipline (→ REQ-fea038de-55eb-45ac-87fc-ec3a7287592a):
// the --dryrun plan owns stdout in the tab-separated machine-readable form, and the confirmation
// listing goes to stderr without polluting it.
func TestPruneOutputStreams(t *testing.T) {
	res := pruneFixture()

	t.Run("printPrunePlan owns stdout exclusively", func(t *testing.T) {
		pruneTestState(t)
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
		pruneTestState(t)
		flagJSON = true

		out, _ := captureOutErr(t, func() { printPrunePlan(res) })
		if out != "" {
			t.Errorf("--json must leave stdout to the envelope alone, got %q", out)
		}
	})

	t.Run("printPrunePlan keeps the stderr notice with nothing to delete", func(t *testing.T) {
		// The empty quadrants of the contract: stdout stays empty under both contracts (nothing
		// to list), while the stderr notice survives --json — human diagnostics coexist with the
		// envelope (→ ADR-0043 §2).
		pruneTestState(t)

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
		// The root path list is the only guard against an unmounted root being taken for a
		// deleted one (→ ADR-0034 §2), so every one of them has to be shown. The expectations are
		// literals rather than a loop over res.Removed: reading the wanted roots back out of the
		// value under test would pass for any listing that agreed with itself.
		for _, root := range []string{"/home/u/src/gone", "/mnt/removable/proj"} {
			if !strings.Contains(errOut, root) {
				t.Errorf("root %q missing from the confirmation listing: %q", root, errOut)
			}
		}
		// The count in the header has to match what is actually listed: a listing that showed one
		// series while announcing two would still contain both roots if the header carried them.
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

// TestPruneConfirmShowsThePreview pins the wiring TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9 left to
// the cli-json 区分: the callback runPrune hands the engine lists the *preview* it is given. The
// engine's preview is a separate value from the result Prune returns (whose Removed means
// "actually deleted", and is empty on an abort), so a callback that reported the returned result
// instead would show "0 series" to the user about to confirm a deletion — the root path list that
// ADR-0034 §2 calls the only guard would be blank while the deletion still went ahead.
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

// TestPruneRunDrivesTheEngine pins runPrune's own orchestration against a stubbed engine: which
// value reaches the prompt, and what an aborted run does. The stub stands in for the state dir so
// the CLI's decisions are observable without one on disk.
func TestPruneRunDrivesTheEngine(t *testing.T) {
	// Every subtest reassigns some of these, so each one restores its own state through
	// pruneTestState below; this is the outer net for whatever a subtest leaves behind.
	pruneTestState(t)
	flagJSON, flagVerbose = false, false
	// go test's stdin is never a TTY, so interactivity is passed in rather than detected. The
	// TTY判定 itself is TestIsInteractiveNonTTY's (reset_test.go) subject.
	const interactive = true

	// The two values the CLI must keep apart. The preview is what the engine hands the callback
	// (the judged candidates); the returned result is what the run actually did. They are given
	// disjoint series here so no assertion can be satisfied by the wrong one: a prompt that
	// listed the returned result would show the deleted-root, and an envelope built from the
	// preview would report the candidate-root as deleted.
	previewRoot, resultRoot := "/home/u/src/candidate", "/home/u/src/deleted"
	preview := &engine.PruneResult{Removed: []engine.PruneSeries{
		{RootHash: "cand", Root: previewRoot, Dir: "/state/nix/profiles/nput/cand"},
	}}
	result := &engine.PruneResult{Removed: []engine.PruneSeries{
		{RootHash: "done", Root: resultRoot, Dir: "/state/nix/profiles/nput/done"},
	}}

	t.Run("the prompt sees the preview and the envelope sees the result", func(t *testing.T) {
		pruneTestState(t)
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
		pruneTestState(t)
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
		// Skipping the prompt skips the listing with it (→ ADR-0034 §2), and silent-on-success
		// keeps the run quiet without -v.
		if strings.Contains(errOut, resultRoot) {
			t.Errorf("--yes must not print the series listing: %q", errOut)
		}
	})

	t.Run("--verbose reports what was deleted", func(t *testing.T) {
		pruneTestState(t)
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

	t.Run("a refused policy never reaches the engine", func(t *testing.T) {
		// Non-interactive without --yes: the refusal has to happen before the scan, so no state
		// dir is read and nothing is deleted (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed).
		pruneTestState(t)
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
		// Nothing was scanned, so the envelope must not carry an inventory that reads as a
		// completed scan that found nothing.
		if err := run.emit(runErr); err != nil {
			t.Fatalf("emit: %v", err)
		}
		assertNoInfoKeys(t, decodeEnvelope(t, buf))
	})

	t.Run("an aborted run reports on stderr and succeeds", func(t *testing.T) {
		pruneTestState(t)
		flagYes = true
		pruneFn = func(engine.PruneOptions) (*engine.PruneResult, error) {
			// A declined run as the engine reports it: Aborted on the returned value, and the
			// judged candidates left behind in Removed. The CLI must not turn those candidates
			// into a report of what was deleted — that is what the abort branch is for.
			return &engine.PruneResult{Removed: preview.Removed, Aborted: true}, nil
		}
		run, buf := newPruneTestRun()
		var runErr error
		out, errOut := captureOutErr(t, func() { runErr = runPrune(run, false, interactive) })
		if runErr != nil {
			// Declining is not a failure: the exit code stays 0 (→ reset's aborted path).
			t.Errorf("an aborted run must not fail: %v", runErr)
		}
		if out != "" {
			t.Errorf("the abort notice pollutes stdout: %q", out)
		}
		if !strings.Contains(errOut, "prune aborted") {
			t.Errorf("the abort notice is missing from stderr: %q", errOut)
		}
		// The -v report is the thing that must not run: an aborted run that fell through to it
		// would print "removed-series" for series that are still on disk.
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
		// Nothing was deleted, so no inventory is emitted at all: an envelope carrying the judged
		// candidates in removed would report a deletion the user declined.
		assertNoInfoKeys(t, doc)
	})
}

// TestPruneInfoShape pins the --json payload (→ issue #134): removed / skipped arrays, the skip
// reason carried verbatim from the engine vocabulary with its detail, and the field name being
// removed rather than pruned (Result.Pruned already means the rmdir-ed empty ancestors ·
// → REQ-8409db86-a1ba-4053-86dc-588985cc1ca7).
func TestPruneInfoShape(t *testing.T) {
	info := pruneInfoFrom(pruneFixture())

	if len(info.Removed) != 2 {
		t.Fatalf("removed = %v, want the two series", info.Removed)
	}
	first := info.Removed[0]
	if first.RootHash != "aaaa" || first.Root != "/home/u/src/gone" ||
		first.Dir != "/state/nix/profiles/nput/aaaa" {
		t.Errorf("removed[0] = %+v, want the state-base series", first)
	}
	if len(first.Names) != 2 || first.Names[0] != "default" || first.Names[1] != "docs" {
		t.Errorf("removed[0].names = %v, want [default docs]", first.Names)
	}
	// A series holding nothing but the backref still carries names as an array, so a consumer
	// never has to distinguish null from empty.
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

// TestPruneInfoSkipReasonsAreTheEngineVocabulary pins that every engine skip reason survives
// into the document verbatim: a CLI-side re-spelling would give consumers a vocabulary that
// drifts from the engine's (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1 の 5 値).
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

// TestPruneJSONEnvelope pins prune's envelope shape: prune names no config, so results stays []
// and the inventory rides in the envelope-wide info (the init shape · niface ADR-0018). The
// document is conformant.
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

// TestPruneJSONEnvelopeInfoAbsentBeforeScan pins the failure boundary that forces pruneInfo to
// be carried as a pointer (→ issue #196 §4): the --json-without---yes refusal happens before the
// scan exists, so the envelope's info must stay absent rather than emit an empty inventory that
// reads as "scanned, found nothing".
func TestPruneJSONEnvelopeInfoAbsentBeforeScan(t *testing.T) {
	r, buf := newPruneTestRun()
	if err := r.emit(errors.New("nput: refusing destructive prune without --yes in a non-interactive context")); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertNoInfoKeys(t, decodeEnvelope(t, buf))
}

// TestPruneDryrunDoesNotDelete pins that the CLI's dryrun path drives the engine read-only: the
// options it builds carry DryRun and no Confirm, so nothing can prompt or delete
// (→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed の副作用ゼロ).
func TestPruneDryrunDoesNotDelete(t *testing.T) {
	opts := pruneOptions(true, nil)
	if !opts.DryRun {
		t.Error("the dryrun path must set DryRun")
	}
	if opts.Confirm != nil {
		t.Error("the dryrun path must pass no Confirm (a preview never asks)")
	}
	// The scan bases stay at their defaults: the CLI names neither, so a run cannot be pointed at
	// a base the engine did not resolve itself (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1 の層分け).
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
