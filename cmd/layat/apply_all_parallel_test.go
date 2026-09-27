package main

// Tests for apply --all's parallel placement stage (→ ADR-0039, issue #154): the configs are applied
// on a worker pool, so they complete in any order, yet results[], the -v report, the dryrun plan and
// the per-config stderr lines stay in lexical order, the counts stay exact under -race, a partial
// failure keeps every succeeded subject, and a try-lock skip is still not a failure. Applies are
// injected, so nothing here runs nix.

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

// reverseCompletion wraps fn so each config waits for its lexical successor to finish before it
// returns: the configs complete in reverse lexical order, which is reachable only when all of them
// run at once (call it with jobs >= len(selected)). completed returns the observed completion order,
// so a test can show the reversal actually happened before asserting the output is still lexical.
func reverseCompletion(t *testing.T, selected []string, fn func(string) (*engine.Result, error)) (wrapped func(string) (*engine.Result, error), completed func() []string) {
	t.Helper()
	done := make(map[string]chan struct{}, len(selected))
	next := make(map[string]string, len(selected))
	for i, name := range selected {
		done[name] = make(chan struct{})
		if i+1 < len(selected) {
			next[name] = selected[i+1]
		}
	}
	var mu sync.Mutex
	var order []string
	wrapped = func(name string) (*engine.Result, error) {
		if succ, ok := next[name]; ok {
			select {
			case <-done[succ]:
			case <-time.After(2 * time.Second):
				t.Errorf("%s waited on %s for 2s: the configs are not applied in parallel", name, succ)
			}
		}
		res, err := fn(name)
		mu.Lock()
		order = append(order, name)
		mu.Unlock()
		close(done[name])
		return res, err
	}
	completed = func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(order)
	}
	return wrapped, completed
}

// assertInOrder fails unless every needle appears in out, each after the previous one.
func assertInOrder(t *testing.T, out string, needles ...string) {
	t.Helper()
	pos := -1
	for _, n := range needles {
		i := strings.Index(out, n)
		if i < 0 {
			t.Errorf("output lacks %q:\n%s", n, out)
			return
		}
		if i < pos {
			t.Errorf("%q is out of lexical order:\n%s", n, out)
			return
		}
		pos = i
	}
}

// withVerbose turns -v on for the test and restores it afterwards.
func withVerbose(t *testing.T) {
	t.Helper()
	old := flagVerbose
	flagVerbose = true
	t.Cleanup(func() { flagVerbose = old })
}

// TestApplyAllParallelOutputIsLexicalDespiteCompletionOrder: with the completion order reversed,
// results[] and the -v report still follow the lexical selection order.
func TestApplyAllParallelOutputIsLexicalDespiteCompletionOrder(t *testing.T) {
	withVerbose(t)
	selected := []string{"a", "b", "c"}
	run, buf := newApplyTestRun()
	applyFn, completed := reverseCompletion(t, selected, func(name string) (*engine.Result, error) {
		return placedResult(name), nil
	})
	var applied, skipped, failures int
	stderr := captureStderr(t, func() {
		applied, skipped, failures = aggregateApply(run, selected, len(selected), applyFn)
	})
	if got, want := completed(), []string{"c", "b", "a"}; !slices.Equal(got, want) {
		t.Fatalf("completion order = %v, want %v (the reversal did not happen)", got, want)
	}
	if applied != 3 || skipped != 0 || failures != 0 {
		t.Fatalf("counts = (%d, %d, %d), want (3, 0, 0)", applied, skipped, failures)
	}
	assertInOrder(t, stderr, "layat: apply a done", "layat: apply b done", "layat: apply c done")

	if err := run.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)
	if _, order := subjectResults(t, buf); !slices.Equal(order, selected) {
		t.Errorf("results order = %v, want %v", order, selected)
	}
}

// TestApplyAllParallelPartialFailureAndSkip: a failure, a try-lock skip and a success completing in
// reverse keep their stderr lines in lexical order, count as (1 applied, 1 skipped, 1 failed), and
// settle each subject on its own outcome — the succeeded config keeps its whole result, and the
// skip is not a failure.
func TestApplyAllParallelPartialFailureAndSkip(t *testing.T) {
	withVerbose(t)
	selected := []string{"a", "b", "c"}
	run, buf := newApplyTestRun()
	applyFn, completed := reverseCompletion(t, selected, func(name string) (*engine.Result, error) {
		switch name {
		case "a":
			return nil, errors.New("stub apply of a failed")
		case "b":
			return &engine.Result{Profile: placedResult(name).Profile, Skipped: true}, engine.ErrSkipped
		}
		return placedResult(name), nil
	})
	var applied, skipped, failures int
	stderr := captureStderr(t, func() {
		applied, skipped, failures = aggregateApply(run, selected, len(selected), applyFn)
	})
	if got, want := completed(), []string{"c", "b", "a"}; !slices.Equal(got, want) {
		t.Fatalf("completion order = %v, want %v (the reversal did not happen)", got, want)
	}
	if applied != 1 || skipped != 1 || failures != 1 {
		t.Fatalf("counts = (%d, %d, %d), want (1, 1, 1)", applied, skipped, failures)
	}
	assertInOrder(t, stderr,
		"layat: apply a failed: stub apply of a failed",
		"layat: skipped apply b (another apply is in progress)",
		"layat: apply c done")

	if err := run.emit(&exitCodeError{code: applyAllExitCode(failures > 0, false), msg: "layat: apply --all: 1 config(s) failed"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)
	byName, order := subjectResults(t, buf)
	if !slices.Equal(order, selected) {
		t.Errorf("results order = %v, want %v", order, selected)
	}
	if status, errs := statusAndErrors(t, byName["a"]); status != "error" || len(errs) != 1 {
		t.Errorf("subject a = (%s, %v), want error with exactly its own failure", status, errs)
	}
	if status, errs := statusAndErrors(t, byName["b"]); status != "success" || len(errs) != 0 {
		t.Errorf("subject b = (%s, %v), want (success, none): a skip is not a failure", status, errs)
	}
	status, _ := statusAndErrors(t, byName["c"])
	items, _ := byName["c"]["result"].(map[string]any)["items"].([]any)
	if status != "success" || len(items) == 0 {
		t.Errorf("subject c = (%s, %d items), want success with its placed items", status, len(items))
	}
}

// TestApplyAllDryRunParallelPlanIsLexical: the dryrun plan on stdout follows the lexical order even
// when the read-only applies complete in reverse, and a conflict in one config still yields exit 2.
func TestApplyAllDryRunParallelPlanIsLexical(t *testing.T) {
	selected := []string{"a", "b", "c"}
	run, buf := newApplyTestRun()
	run.dryRun = true
	applyDry, completed := reverseCompletion(t, selected, func(name string) (*engine.Result, error) {
		res := placedResult(name)
		if name == "b" {
			res.Placed = nil
			res.Conflicts = []planner.Conflict{{Entry: manifest.Entry{Target: "t/" + name}, Reason: "occupied"}}
		}
		return res, nil
	})
	var code int
	stdout := captureStdout(t, func() {
		code = aggregateDryRun(run, selected, len(selected), applyDry)
	})
	if got, want := completed(), []string{"c", "b", "a"}; !slices.Equal(got, want) {
		t.Fatalf("completion order = %v, want %v (the reversal did not happen)", got, want)
	}
	if code != 2 {
		t.Errorf("aggregateDryRun code = %d, want 2 (conflict, no error)", code)
	}
	assertInOrder(t, stdout, "place\t"+placedResult("a").Placed[0], "conflict\tt/b: occupied", "place\t"+placedResult("c").Placed[0])

	if err := run.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)
	if _, order := subjectResults(t, buf); !slices.Equal(order, selected) {
		t.Errorf("results order = %v, want %v", order, selected)
	}
}

// TestApplyAllDryRunParallelFailureLinesAreLexical: the dryrun's per-config failure lines go to
// stderr in lexical order even when the failing configs complete in reverse, and an error decides
// exit 1.
func TestApplyAllDryRunParallelFailureLinesAreLexical(t *testing.T) {
	selected := []string{"a", "b", "c"}
	run, _ := newApplyTestRun()
	run.dryRun = true
	applyDry, completed := reverseCompletion(t, selected, func(name string) (*engine.Result, error) {
		if name == "c" {
			return placedResult(name), nil
		}
		return nil, errors.New("stub build of " + name + " failed")
	})
	var code int
	stderr := captureStderr(t, func() {
		captureStdout(t, func() { code = aggregateDryRun(run, selected, len(selected), applyDry) })
	})
	if got, want := completed(), []string{"c", "b", "a"}; !slices.Equal(got, want) {
		t.Fatalf("completion order = %v, want %v (the reversal did not happen)", got, want)
	}
	if code != 1 {
		t.Errorf("aggregateDryRun code = %d, want 1 (error)", code)
	}
	assertInOrder(t, stderr,
		"layat: apply a --dryrun failed: stub build of a failed",
		"layat: apply b --dryrun failed: stub build of b failed")
}

// TestApplyAllParallelCountsAreExact runs many configs with mixed outcomes on a smaller pool: the
// counts must come out exact (a racy counter shows up under -race and as a lost update), the pool
// never exceeds jobs, and results[] keeps the selection order.
func TestApplyAllParallelCountsAreExact(t *testing.T) {
	const n, jobs = 60, 8
	selected := make([]string, n)
	for i := range selected {
		selected[i] = fmt.Sprintf("c%02d", i)
	}
	var inFlight, peak atomic.Int32
	run, buf := newApplyTestRun()
	var applied, skipped, failures int
	captureStderr(t, func() {
		applied, skipped, failures = aggregateApply(run, selected, jobs, func(name string) (*engine.Result, error) {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)
			for {
				p := peak.Load()
				if cur <= p || peak.CompareAndSwap(p, cur) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			switch slices.Index(selected, name) % 3 {
			case 1:
				return nil, engine.ErrSkipped
			case 2:
				return nil, errors.New("stub failure")
			}
			return placedResult(name), nil
		})
	})
	if applied != 20 || skipped != 20 || failures != 20 {
		t.Errorf("counts = (%d, %d, %d), want (20, 20, 20)", applied, skipped, failures)
	}
	if got := peak.Load(); got > jobs {
		t.Errorf("peak in-flight applies = %d, want at most %d (--jobs)", got, jobs)
	}
	if err := run.emit(&exitCodeError{code: 1, msg: "layat: apply --all: 20 config(s) failed"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if _, order := subjectResults(t, buf); !slices.Equal(order, selected) {
		t.Errorf("results order = %v, want the selection order", order)
	}
}
