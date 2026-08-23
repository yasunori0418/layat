package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/yasunori0418/niface/go/conformance"

	"github.com/yasunori0418/nput/internal/engine"
)

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
		origJSON := flagJSON
		flagJSON = false
		defer func() { flagJSON = origJSON }()

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
		origJSON := flagJSON
		flagJSON = true
		defer func() { flagJSON = origJSON }()

		out, _ := captureOutErr(t, func() { printPrunePlan(res) })
		if out != "" {
			t.Errorf("--json must leave stdout to the envelope alone, got %q", out)
		}
	})

	t.Run("reportPruneTargets lists every root path on stderr", func(t *testing.T) {
		out, errOut := captureOutErr(t, func() { reportPruneTargets(res) })
		if out != "" {
			t.Errorf("the confirmation listing pollutes stdout: %q", out)
		}
		// The root path list is the only guard against an unmounted root being taken for a
		// deleted one (→ ADR-0034 §2), so every one of them has to be shown.
		for _, s := range res.Removed {
			if !strings.Contains(errOut, s.Root) {
				t.Errorf("root %q missing from the confirmation listing: %q", s.Root, errOut)
			}
		}
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
	if opts.Warnf == nil {
		t.Error("Warnf must be wired so a skipped series is reported even in dryrun")
	}

	confirm := func(*engine.PruneResult) (bool, error) { return true, nil }
	real := pruneOptions(false, confirm)
	if real.DryRun {
		t.Error("the non-dryrun path must not set DryRun")
	}
	if real.Confirm == nil {
		t.Error("the non-dryrun path must pass the confirmation callback through")
	}
}
