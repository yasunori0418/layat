package main

// The partial-failure matrix of apply --all (→ ADR-0024, ADR-0039, issue #155): every non-empty
// combination of the per-config outcomes — a stage-1 build failure, a stage-2 failure (item-borne or
// subject-borne), a success, a try-lock skip, and on --dryrun a conflict — is driven through both
// stages the way runApplyAll composes them, at --jobs 1 and at full parallelism. Each combination
// must settle the exit code by priority error(1) > conflict(2) > 0, the applied / skipped / failed
// counts, and every subject's status and error layer. Builds and applies are injected, so nothing
// here runs nix; run it with -race to cover the parallel aggregation. The composition and the
// non-dryrun exit-code decision are mirrored here rather than reached through runApplyAll, so its
// wiring is out of scope (the stage-1 wiring is TestRunApplyAllStageOneWiring's; e2e 09-apply-all
// drives the whole command).

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

// Each config is named after its scripted outcome, so a combination sorted lexically is also the
// expected results[] order.
const (
	outcomeBuild    = "build"    // the stage-1 build fails: stage 2 never runs for it
	outcomeItem     = "item"     // the engine stops on an entry and returns a partial result
	outcomePlace    = "place"    // the engine fails before producing any result
	outcomeOK       = "ok"       // applied
	outcomeSkip     = "skip"     // a try-lock skip: not a failure
	outcomeDryFail  = "dryfail"  // the read-only apply fails
	outcomeConflict = "conflict" // the plan conflicts: exit 2 unless an error outranks it
)

// matrixLayer is where a config's failure must settle in the envelope.
type matrixLayer struct {
	subjectBorne bool   // exactly one E_LAYAT_FAILED on the subject's errors[]
	itemCode     string // otherwise, a failed item carrying this code (subject errors[] absent)
}

// The failing outcomes and their layers; every other outcome settles as success with no errors.
var matrixLayers = map[string]matrixLayer{
	outcomeBuild:    {subjectBorne: true},
	outcomePlace:    {subjectBorne: true},
	outcomeDryFail:  {subjectBorne: true},
	outcomeItem:     {itemCode: "E_LAYAT_FAILED"},
	outcomeConflict: {itemCode: "E_LAYAT_COLLISION"},
}

// combinations returns every non-empty subset of outcomes, each sorted lexically.
func combinations(outcomes ...string) [][]string {
	var out [][]string
	for mask := 1; mask < 1<<len(outcomes); mask++ {
		var set []string
		for i, o := range outcomes {
			if mask&(1<<i) != 0 {
				set = append(set, o)
			}
		}
		slices.Sort(set)
		out = append(out, set)
	}
	return out
}

// matrixStageTwo is stage 2's injected (read-only when dry) apply. It fails the test if it is ever
// reached for a config whose stage-1 build failed.
func matrixStageTwo(t *testing.T) func(string) (*engine.Result, error) {
	return func(name string) (*engine.Result, error) {
		switch name {
		case outcomeBuild:
			t.Errorf("stage 2 ran for %s, whose stage-1 build failed", name)
		case outcomeItem:
			res := placedResult(name)
			res.Placed = nil
			res.FailedTarget = "t/" + name
			return res, errors.New("stub symlink t/item: permission denied")
		case outcomePlace:
			return nil, errors.New("stub apply of place failed")
		case outcomeDryFail:
			return nil, errors.New("stub dryrun of dryfail failed")
		case outcomeSkip:
			return &engine.Result{Profile: placedResult(name).Profile, Skipped: true}, engine.ErrSkipped
		case outcomeConflict:
			res := placedResult(name)
			res.Placed = nil
			res.Conflicts = []planner.Conflict{
				{Entry: manifest.Entry{Target: "t/" + name}, Reason: "occupied by a foreign entity"},
			}
			return res, nil
		}
		return placedResult(name), nil
	}
}

// assertMatrixEnvelope checks the emitted document: conformance, the aggregate status, results[] in
// lexical order, no top-level errors[] (every failure belongs to a subject), and each subject's
// status and error layer.
func assertMatrixEnvelope(t *testing.T, run *applyRun, buf *bytes.Buffer, cmdErr error, combo []string) {
	t.Helper()
	if err := run.emit(cmdErr); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)
	doc := decodeEnvelope(t, buf)
	anyFailing := slices.ContainsFunc(combo, func(n string) bool { _, ok := matrixLayers[n]; return ok })
	if want := map[bool]string{true: "error", false: "success"}[anyFailing]; doc["status"] != want {
		t.Errorf("aggregate status = %v, want %s", doc["status"], want)
	}
	if topErrs, ok := doc["errors"]; ok {
		t.Errorf("top-level errors = %v, want absent (every failure belongs to its config)", topErrs)
	}
	byName, order := subjectResults(t, buf)
	if !slices.Equal(order, combo) {
		t.Errorf("results order = %v, want %v", order, combo)
	}
	for _, name := range combo {
		status, errs := statusAndErrors(t, byName[name])
		layer, failing := matrixLayers[name]
		if want := map[bool]string{true: "error", false: "success"}[failing]; status != want {
			t.Errorf("subject %s status = %s, want %s", name, status, want)
		}
		if layer.subjectBorne {
			if len(errs) != 1 || errs[0].(map[string]any)["code"] != "E_LAYAT_FAILED" {
				t.Errorf("subject %s errors = %v, want exactly one subject-borne E_LAYAT_FAILED", name, errs)
			}
			continue
		}
		if len(errs) != 0 {
			t.Errorf("subject %s errors = %v, want none", name, errs)
		}
		if layer.itemCode != "" && !hasFailedItem(byName[name], layer.itemCode) {
			t.Errorf("subject %s result = %v, want a failed item carrying %s", name, byName[name]["result"], layer.itemCode)
		}
	}
}

// hasFailedItem reports whether the subject's result has a failed item whose error carries code.
func hasFailedItem(sr map[string]any, code string) bool {
	for _, it := range sr["result"].(map[string]any)["items"].([]any) {
		item := it.(map[string]any)
		if e, ok := item["error"].(map[string]any); ok && item["status"] == "failed" && e["code"] == code {
			return true
		}
	}
	return false
}

// TestApplyAllFailureMatrix: every combination of stage-1 build failure × stage-2 failure (item- and
// subject-borne) × success × try-lock skip settles the counts, exit code 1 exactly when something
// failed (a skip alone is 0), and each subject's status and error layer.
func TestApplyAllFailureMatrix(t *testing.T) {
	for _, combo := range combinations(outcomeBuild, outcomeItem, outcomePlace, outcomeOK, outcomeSkip) {
		for _, jobs := range []int{1, len(combo)} {
			t.Run(fmt.Sprintf("%s/jobs=%d", strings.Join(combo, "+"), jobs), func(t *testing.T) {
				run, buf := newApplyTestRun()
				var applied, skipped, failures int
				stderr := captureStderr(t, func() {
					built := prebuildAll(combo, jobs, failBuild(outcomeBuild))
					applied, skipped, failures = aggregateApply(run, combo, jobs, skipFailedPrebuilds(built, matrixStageTwo(t)))
				})
				var wantApplied, wantSkipped, wantFailures int
				for _, name := range combo {
					switch name {
					case outcomeOK:
						wantApplied++
					case outcomeSkip:
						wantSkipped++
					default:
						wantFailures++
						if !strings.Contains(stderr, "layat: apply "+name+" failed: ") {
							t.Errorf("stderr lacks the failure of %s:\n%s", name, stderr)
						}
					}
				}
				if applied != wantApplied || skipped != wantSkipped || failures != wantFailures {
					t.Fatalf("counts = (%d, %d, %d), want (%d, %d, %d)", applied, skipped, failures, wantApplied, wantSkipped, wantFailures)
				}
				// The exit-code decision runApplyAll makes: non-dryrun --all yields only error / 0.
				code := applyAllExitCode(failures > 0, false)
				if want := map[bool]int{true: 1, false: 0}[wantFailures > 0]; code != want {
					t.Errorf("exit code = %d, want %d", code, want)
				}
				var cmdErr error
				if code != 0 {
					cmdErr = &exitCodeError{code: code, msg: fmt.Sprintf("layat: apply --all: %d config(s) failed", failures)}
				}
				assertMatrixEnvelope(t, run, buf, cmdErr, combo)
			})
		}
	}
}

// TestApplyAllDryRunFailureMatrix: every combination of stage-1 build failure × read-only apply
// failure × conflict × success on --dryrun settles the exit code by priority error(1) > conflict(2)
// > 0 — a conflict never masks an error — and each subject's status and error layer.
func TestApplyAllDryRunFailureMatrix(t *testing.T) {
	for _, combo := range combinations(outcomeBuild, outcomeDryFail, outcomeConflict, outcomeOK) {
		for _, jobs := range []int{1, len(combo)} {
			t.Run(fmt.Sprintf("%s/jobs=%d", strings.Join(combo, "+"), jobs), func(t *testing.T) {
				run, buf := newApplyTestRun()
				run.dryRun = true
				var code int
				captureStdout(t, func() {
					captureStderr(t, func() {
						built := prebuildAll(combo, jobs, failBuild(outcomeBuild))
						code = aggregateDryRun(run, combo, jobs, skipFailedPrebuilds(built, matrixStageTwo(t)))
					})
				})
				want := 0
				switch {
				case slices.Contains(combo, outcomeBuild) || slices.Contains(combo, outcomeDryFail):
					want = 1
				case slices.Contains(combo, outcomeConflict):
					want = 2
				}
				if code != want {
					t.Errorf("exit code = %d, want %d (error > conflict > 0)", code, want)
				}
				var cmdErr error
				if code != 0 {
					cmdErr = &exitError{code: code}
				}
				assertMatrixEnvelope(t, run, buf, cmdErr, combo)
			})
		}
	}
}
