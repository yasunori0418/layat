package main

// Tests for apply --all's parallel build stage (→ ADR-0039, issue #153): the worker pool never runs
// more builds than --jobs at once, a config whose stage-1 build fails is settled on its own subject
// and never reaches stage 2, --jobs' value range (negative is an error, 0 resolves to the CPU count,
// it is an --all modifier), the --debug disclosure prefix, and that --jobs 1 and --jobs N aggregate
// to the same result. Builds and applies are injected, so nothing here runs nix.

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yasunori0418/layat/internal/engine"
)

// TestPrebuildAllBoundsConcurrency pins the pool's ceiling: with more configs than --jobs, the
// in-flight builds reach --jobs (the builds do run in parallel) and never exceed it. The builds
// hold at a barrier until min(jobs, n) of them are in flight, so reaching the ceiling does not
// depend on timing; a pool that runs fewer workers fails on the barrier's timeout instead. The
// barrier stays closed for a grace window after the ceiling is reached, so a surplus worker has
// time to enter and show up in the peak.
func TestPrebuildAllBoundsConcurrency(t *testing.T) {
	selected := []string{"a", "b", "c", "d", "e", "f", "g"}
	for _, jobs := range []int{1, 3, 16} {
		t.Run(fmt.Sprintf("jobs=%d", jobs), func(t *testing.T) {
			want := min(jobs, len(selected))
			var inFlight, peak atomic.Int32
			var once sync.Once
			release := make(chan struct{})
			var timedOut atomic.Bool
			built := prebuildAll(selected, jobs, func(name string) (string, error) {
				n := inFlight.Add(1)
				defer inFlight.Add(-1)
				for {
					p := peak.Load()
					if n <= p || peak.CompareAndSwap(p, n) {
						break
					}
				}
				if int(n) == want {
					once.Do(func() { time.AfterFunc(50*time.Millisecond, func() { close(release) }) })
				}
				select {
				case <-release:
				case <-time.After(2 * time.Second):
					timedOut.Store(true)
				}
				return "/nix/store/" + name, nil
			})
			if timedOut.Load() {
				t.Fatalf("builds never reached %d in flight (peak %d): the pool does not run --jobs workers", want, peak.Load())
			}
			if got := int(peak.Load()); got != want {
				t.Errorf("peak in-flight builds = %d, want %d (min(--jobs, configs))", got, want)
			}
			for _, name := range selected {
				r, ok := built[name]
				if !ok || r.err != nil || r.storePath != "/nix/store/"+name {
					t.Errorf("built[%s] = %+v (present %v), want its store path and no error", name, r, ok)
				}
			}
		})
	}
}

// stagedApply drives both stages the way runApplyAll composes them — stage 1 with the injected
// build, stage 2 through aggregateApply with stage-1 failures short-circuited — and records the
// configs stage 2 was actually called for.
func stagedApply(run *applyRun, selected []string, jobs int, build func(string) (string, error)) (applied, skipped, failures int, stage2 []string) {
	built := prebuildAll(selected, jobs, build)
	applied, skipped, failures = aggregateApply(run, selected, skipFailedPrebuilds(built, func(name string) (*engine.Result, error) {
		stage2 = append(stage2, name)
		if name == "skip" {
			return nil, engine.ErrSkipped
		}
		return placedResult(name), nil
	}))
	return applied, skipped, failures, stage2
}

// failBuild fails the stage-1 build of the configs in failing and succeeds for the rest.
func failBuild(failing ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		if slices.Contains(failing, name) {
			return "", errors.New("stub build of " + name + " failed")
		}
		return "/nix/store/" + name, nil
	}
}

// TestStageOneFailureSkipsStageTwo: a config whose build fails in stage 1 is not handed to stage 2,
// yet it still gets its own subject (in selection order) settled with the build error, counts as a
// failure, and is reported on stderr — while the succeeded configs keep their whole results.
func TestStageOneFailureSkipsStageTwo(t *testing.T) {
	run, buf := newApplyTestRun()
	var applied, skipped, failures int
	var stage2 []string
	stderr := captureStderr(t, func() {
		applied, skipped, failures, stage2 = stagedApply(run, []string{"a", "b", "c"}, 4, failBuild("b"))
	})
	if applied != 2 || skipped != 0 || failures != 1 {
		t.Fatalf("counts = (%d, %d, %d), want (2, 0, 1)", applied, skipped, failures)
	}
	if want := []string{"a", "c"}; !slices.Equal(stage2, want) {
		t.Errorf("stage 2 ran for %v, want %v (lexical order, the failed build skipped)", stage2, want)
	}
	if !strings.Contains(stderr, "layat: apply b failed: stub build of b failed") {
		t.Errorf("stderr = %q, want the stage-1 failure of b reported", stderr)
	}
	if err := run.emit(&exitCodeError{code: 1, msg: "layat: apply --all: 1 config(s) failed"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)
	byName, order := subjectResults(t, buf)
	if want := []string{"a", "b", "c"}; !slices.Equal(order, want) {
		t.Errorf("results order = %v, want %v (selection order, the failed build included)", order, want)
	}
	for _, name := range []string{"a", "c"} {
		if status, _ := statusAndErrors(t, byName[name]); status != "success" {
			t.Errorf("subject %s status = %s, want success", name, status)
		}
	}
	status, errs := statusAndErrors(t, byName["b"])
	if status != "error" || len(errs) != 1 {
		t.Errorf("subject b = (%s, %v), want error with exactly the build error", status, errs)
	}
}

// TestStageOneFailureSkipsDryRunStageTwo: --dryrun takes the same two stages, so a failed build is
// kept out of aggregateDryRun's read-only apply and still decides the error exit code.
func TestStageOneFailureSkipsDryRunStageTwo(t *testing.T) {
	run, _ := newApplyTestRun()
	built := prebuildAll([]string{"a", "b"}, 2, failBuild("a"))
	var stage2 []string
	var code int
	captureStdout(t, func() {
		captureStderr(t, func() {
			code = aggregateDryRun(run, []string{"a", "b"}, skipFailedPrebuilds(built, func(name string) (*engine.Result, error) {
				stage2 = append(stage2, name)
				return placedResult(name), nil
			}))
		})
	})
	if code != 1 {
		t.Errorf("aggregateDryRun code = %d, want 1 (the build error)", code)
	}
	if want := []string{"b"}; !slices.Equal(stage2, want) {
		t.Errorf("dryrun stage 2 ran for %v, want %v", stage2, want)
	}
}

// TestJobsAggregateIsOrderIndependent: the parallel build must not change the final state, so
// --jobs 1 and --jobs 8 settle the same counts, call stage 2 for the same configs in the same
// lexical order, and emit byte-identical envelopes (the test runs pin the clock). The builds take
// longer the earlier the config sorts, so under --jobs 8 they finish in reverse lexical order.
func TestJobsAggregateIsOrderIndependent(t *testing.T) {
	selected := []string{"a", "b", "c", "d", "e", "f", "skip"}
	failing := failBuild("b", "e")
	build := func(name string) (string, error) {
		time.Sleep(time.Duration(len(selected)-slices.Index(selected, name)) * 10 * time.Millisecond)
		return failing(name)
	}
	type outcome struct {
		applied, skipped, failures int
		stage2                     []string
		envelope                   string
	}
	runWith := func(jobs int) outcome {
		run, buf := newApplyTestRun()
		var o outcome
		captureStderr(t, func() {
			o.applied, o.skipped, o.failures, o.stage2 = stagedApply(run, selected, jobs, build)
		})
		if err := run.emit(&exitCodeError{code: 1, msg: "layat: apply --all: 2 config(s) failed"}); err != nil {
			t.Fatalf("emit: %v", err)
		}
		o.envelope = buf.String()
		return o
	}
	serial, parallel := runWith(1), runWith(8)
	if serial.applied != 4 || serial.skipped != 1 || serial.failures != 2 {
		t.Fatalf("--jobs 1 counts = (%d, %d, %d), want (4, 1, 2)", serial.applied, serial.skipped, serial.failures)
	}
	if parallel.applied != serial.applied || parallel.skipped != serial.skipped || parallel.failures != serial.failures {
		t.Errorf("--jobs 8 counts = (%d, %d, %d), want --jobs 1's (%d, %d, %d)",
			parallel.applied, parallel.skipped, parallel.failures, serial.applied, serial.skipped, serial.failures)
	}
	if want := []string{"a", "c", "d", "f", "skip"}; !slices.Equal(serial.stage2, want) || !slices.Equal(parallel.stage2, want) {
		t.Errorf("stage 2 order = %v (--jobs 1) / %v (--jobs 8), want %v", serial.stage2, parallel.stage2, want)
	}
	if parallel.envelope != serial.envelope {
		t.Errorf("--jobs 8 envelope differs from --jobs 1:\n%s\nvs\n%s", parallel.envelope, serial.envelope)
	}
}

// TestResolveApplyJobs pins --jobs' value range: negative is an error, 0 resolves to the logical
// CPU count, and a positive value is taken as-is.
func TestResolveApplyJobs(t *testing.T) {
	if _, err := resolveApplyJobs(-1); err == nil || !strings.Contains(err.Error(), "--jobs") {
		t.Errorf("resolveApplyJobs(-1) err = %v, want a --jobs range error", err)
	}
	if got, err := resolveApplyJobs(0); err != nil || got != runtime.NumCPU() {
		t.Errorf("resolveApplyJobs(0) = (%d, %v), want (%d, nil)", got, err, runtime.NumCPU())
	}
	if got, err := resolveApplyJobs(1); err != nil || got != 1 {
		t.Errorf("resolveApplyJobs(1) = (%d, %v), want (1, nil)", got, err)
	}
	if got, err := resolveApplyJobs(3); err != nil || got != 3 {
		t.Errorf("resolveApplyJobs(3) = (%d, %v), want (3, nil)", got, err)
	}
}

// executeApply runs the apply command with args, restoring the globals RunE touches.
func executeApply(t *testing.T, args ...string) error {
	t.Helper()
	origAll, origJobs, origReport := flagApplyAll, flagApplyJobs, nifaceReport
	t.Cleanup(func() { flagApplyAll, flagApplyJobs, nifaceReport = origAll, origJobs, origReport })
	cmd := newApplyCmd()
	cmd.SetArgs(args)
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	return cmd.Execute()
}

// TestJobsFlagIsAnAllModifier: --jobs is local to apply and only meaningful with --all, so a named
// apply rejects it — even an explicit --jobs 0 — before any entrypoint discovery or nix call.
func TestJobsFlagIsAnAllModifier(t *testing.T) {
	for _, args := range [][]string{{"--jobs", "2", "a"}, {"--jobs", "0"}} {
		err := executeApply(t, args...)
		if err == nil || !strings.Contains(err.Error(), "--jobs is a modifier for apply --all") {
			t.Errorf("apply %v err = %v, want the --all modifier error", args, err)
		}
	}
	if f := newApplyCmd().Flags().Lookup("jobs"); f == nil || f.DefValue != "0" {
		t.Errorf("apply --jobs flag = %+v, want a local flag defaulting to 0", f)
	}
	if newRootCmd().PersistentFlags().Lookup("jobs") != nil {
		t.Error("--jobs must not be a persistent flag")
	}
}

// TestJobsNegativeStopsApplyAll: a negative --jobs stops apply --all before the batch eval.
func TestJobsNegativeStopsApplyAll(t *testing.T) {
	withFlakeEntrypoint(t)
	log := stubNixForApply(t, `{}`)
	err := executeApply(t, "--all", "--jobs", "-1")
	if err == nil || !strings.Contains(err.Error(), "--jobs") {
		t.Fatalf("apply --all --jobs -1 err = %v, want the --jobs range error", err)
	}
	if calls := nixCalls(t, log); len(calls) != 0 {
		t.Errorf("nix calls = %q, want none (rejected before the batch eval)", calls)
	}
}

// TestRunApplyAllStageOneWiring drives runApplyAll through the stub nix, whose builds all fail,
// with --debug, with and without --dryrun. The stage-1 builds disclose their nix command lines
// prefixed with the config name, so interleaved lines stay attributable, while the batch eval
// before the pool keeps the plain disclosure. Every build is a stage-1 --no-link realize — a
// failed config never re-enters stage 2's build — and nix's own stderr surfaces only inside the
// failed config's error report.
func TestRunApplyAllStageOneWiring(t *testing.T) {
	for _, dryrun := range []bool{false, true} {
		t.Run(fmt.Sprintf("dryrun=%v", dryrun), func(t *testing.T) {
			withFlakeEntrypoint(t)
			origDebug, origJobs := flagDebug, flagApplyJobs
			t.Cleanup(func() { flagDebug, flagApplyJobs = origDebug, origJobs })
			flagDebug, flagApplyJobs, flagDryrun = true, 2, dryrun
			log := stubNixForApply(t, `{"a":{"rootKind":"home","targets":["x"]},"b":{"rootKind":"home","targets":["y"]}}`)
			run, _ := newApplyTestRun()
			stderr := captureStderr(t, func() {
				if err := runApplyAll(run); err == nil {
					t.Error("runApplyAll must surface the stub build failures")
				}
			})
			lines := strings.Split(stderr, "\n")
			for _, name := range []string{"a", "b"} {
				prefix := "[" + name + "] layat: + nix build "
				if !slices.ContainsFunc(lines, func(l string) bool {
					return strings.HasPrefix(l, prefix) && strings.Contains(l, "."+name+" --no-link")
				}) {
					t.Errorf("stderr has no %q line for %s:\n%s", prefix, name, stderr)
				}
			}
			if !slices.ContainsFunc(lines, func(l string) bool { return strings.HasPrefix(l, "layat: + nix eval ") }) {
				t.Errorf("stderr has no unprefixed batch eval disclosure:\n%s", stderr)
			}
			var builds int
			for _, c := range nixCalls(t, log) {
				if strings.HasPrefix(c, "build ") {
					builds++
					if !strings.Contains(c, "--no-link") {
						t.Errorf("build %q is not a stage-1 realize: a failed config reached stage 2", c)
					}
				}
			}
			// Under --dryrun stage 2's build is also --no-link, so there the count below is what
			// catches a failed config re-entering stage 2.
			if builds != 2 {
				t.Errorf("builds = %d, want 2 (one stage-1 realize per config, no stage-2 retry)", builds)
			}
			var nixStderr int
			for i, l := range lines {
				if l != "error: stub build failed" {
					continue
				}
				nixStderr++
				if i == 0 || !strings.HasPrefix(lines[i-1], "layat: apply ") || !strings.HasSuffix(lines[i-1], "failed:") {
					t.Errorf("nix stderr at line %d is not inside a failure report:\n%s", i, stderr)
				}
			}
			if nixStderr != 2 {
				t.Errorf("nix stderr lines = %d, want 2 (one inside each failed config's report):\n%s", nixStderr, stderr)
			}
		})
	}
}
