package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

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
	clean := func(name string) (*engine.Result, error) {
		return &engine.Result{Placed: []string{"/p/" + name}}, nil
	}
	conflict := func(name string) (*engine.Result, error) {
		return &engine.Result{Conflicts: []planner.Conflict{
			{Entry: manifest.Entry{Target: "/c/" + name}, Reason: "occupied"},
		}}, nil
	}
	fail := func(name string) (*engine.Result, error) {
		return nil, io.EOF
	}

	cases := []struct {
		name     string
		apply    func(string) (*engine.Result, error)
		selected []string
		wantCode int
	}{
		{"all clean → 0", clean, []string{"a", "b"}, 0},
		{"conflict → 2", func(n string) (*engine.Result, error) {
			if n == "b" {
				return conflict(n)
			}
			return clean(n)
		}, []string{"a", "b"}, 2},
		{"error → 1", func(n string) (*engine.Result, error) {
			if n == "b" {
				return fail(n)
			}
			return clean(n)
		}, []string{"a", "b"}, 1},
		{"error wins over conflict → 1", func(n string) (*engine.Result, error) {
			switch n {
			case "a":
				return conflict(n)
			case "b":
				return fail(n)
			}
			return clean(n)
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
	clean := func(name string) (*engine.Result, error) {
		return &engine.Result{Placed: []string{"/p/" + name}}, nil
	}
	skipped := func(string) (*engine.Result, error) {
		return nil, engine.ErrSkipped
	}
	failed := func(string) (*engine.Result, error) {
		return nil, io.EOF
	}

	cases := []struct {
		name         string
		apply        func(string) (*engine.Result, error)
		selected     []string
		wantApplied  int
		wantSkipped  int
		wantFailures int
	}{
		{"all succeed → 0 failures", clean, []string{"a", "b"}, 2, 0, 0},
		{"ErrSkipped counts as skip, not failure", func(n string) (*engine.Result, error) {
			if n == "b" {
				return skipped(n)
			}
			return clean(n)
		}, []string{"a", "b"}, 1, 1, 0},
		{"error counts as failure and continues", func(n string) (*engine.Result, error) {
			if n == "b" {
				return failed(n)
			}
			return clean(n)
		}, []string{"a", "b", "c"}, 2, 0, 1},
		{"mixed skip and failure both continue", func(n string) (*engine.Result, error) {
			switch n {
			case "a":
				return skipped(n)
			case "b":
				return failed(n)
			}
			return clean(n)
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
