package main

// Tests for apply --all's cross-config target conflict preflight (→ ADR-0038, issue #152): the
// bucket rules (per rootKind, per fixed root value, one bucket under --root), the interplay with
// the root filter's selection, the non-check of a named apply, and the --json top-level errors[].

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectCrossConfigConflictsBuckets pins the bucket rules on the pure detector, without nix.
func TestDetectCrossConfigConflictsBuckets(t *testing.T) {
	cases := []struct {
		name         string
		roots        map[string]rootInfo
		selected     []string
		rootOverride string
		wantConflict bool
	}{
		{
			name: "same rootKind shares a bucket",
			roots: map[string]rootInfo{
				"a": {RootKind: "home", Targets: []string{".config/x", ".config/y"}},
				"b": {RootKind: "home", Targets: []string{".config/y"}},
			},
			selected:     []string{"a", "b"},
			wantConflict: true,
		},
		{
			name: "different rootKinds are different buckets",
			roots: map[string]rootInfo{
				"a": {RootKind: "home", Targets: []string{".config/y"}},
				"b": {RootKind: "project", Targets: []string{".config/y"}},
			},
			selected: []string{"a", "b"},
		},
		{
			name: "fixed roots with the same value share a bucket",
			roots: map[string]rootInfo{
				"a": {RootKind: "fixed", Root: "/srv/app", Targets: []string{"conf"}},
				"b": {RootKind: "fixed", Root: "/srv/app", Targets: []string{"conf"}},
			},
			selected:     []string{"a", "b"},
			wantConflict: true,
		},
		{
			name: "fixed roots with different values are different buckets",
			roots: map[string]rootInfo{
				"a": {RootKind: "fixed", Root: "/srv/app", Targets: []string{"conf"}},
				"b": {RootKind: "fixed", Root: "/srv/other", Targets: []string{"conf"}},
			},
			selected: []string{"a", "b"},
		},
		{
			// The fixed root value is compared, not the kind: a fixed root that happens to equal the
			// project root is a different bucket (an undetectable case left to the runtime; → ADR-0038).
			name: "fixed and project are different buckets",
			roots: map[string]rootInfo{
				"a": {RootKind: "fixed", Root: "/srv/app", Targets: []string{"conf"}},
				"b": {RootKind: "project", Targets: []string{"conf"}},
			},
			selected: []string{"a", "b"},
		},
		{
			name: "--root collapses every config into one bucket",
			roots: map[string]rootInfo{
				"a": {RootKind: "home", Targets: []string{"conf"}},
				"b": {RootKind: "fixed", Root: "/srv/app", Targets: []string{"conf"}},
			},
			selected:     []string{"a", "b"},
			rootOverride: "/tmp/override",
			wantConflict: true,
		},
		{
			name: "a conflict outside the selection is ignored",
			roots: map[string]rootInfo{
				"a": {RootKind: "home", Targets: []string{"conf"}},
				"b": {RootKind: "project", Targets: []string{"conf"}},
				"c": {RootKind: "project", Targets: []string{"conf"}},
			},
			// --home-root selects a alone; b and c collide but are not selected.
			selected: []string{"a"},
		},
		{
			name: "no shared target",
			roots: map[string]rootInfo{
				"a": {RootKind: "home", Targets: []string{".config/x"}},
				"b": {RootKind: "home", Targets: []string{".config/y"}},
			},
			selected: []string{"a", "b"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := detectCrossConfigConflicts(tc.roots, tc.selected, tc.rootOverride)
			if tc.wantConflict && err == nil {
				t.Fatal("want a conflict error, got nil")
			}
			if !tc.wantConflict && err != nil {
				t.Fatalf("want no conflict, got %v", err)
			}
		})
	}
}

// TestDetectCrossConfigConflictsMessage checks the error names the target and both configs.
func TestDetectCrossConfigConflictsMessage(t *testing.T) {
	roots := map[string]rootInfo{
		"docs":  {RootKind: "project", Targets: []string{".claude/skills/nix"}},
		"tools": {RootKind: "project", Targets: []string{".claude/skills/nix", ".config/other"}},
	}
	err := detectCrossConfigConflicts(roots, []string{"docs", "tools"}, "")
	if err == nil {
		t.Fatal("want a conflict error, got nil")
	}
	for _, want := range []string{".claude/skills/nix", "docs", "tools"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
	if strings.Contains(err.Error(), ".config/other") {
		t.Errorf("error %q mentions a target that does not collide", err)
	}
}

// TestDetectCrossConfigConflictsJSON checks the preflight error lands on the top-level errors[]
// as E_LAYAT_FAILED: it fails before any subject is registered (→ ADR-0043 §6).
func TestDetectCrossConfigConflictsJSON(t *testing.T) {
	roots := map[string]rootInfo{
		"a": {RootKind: "home", Targets: []string{"conf"}},
		"b": {RootKind: "home", Targets: []string{"conf"}},
	}
	run, buf := newApplyTestRun()
	if err := run.emit(detectCrossConfigConflicts(roots, []string{"a", "b"}, "")); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" {
		t.Errorf("status = %v, want error", doc["status"])
	}
	if results, _ := doc["results"].([]any); len(results) != 0 {
		t.Errorf("results = %v, want none (the preflight precedes subject registration)", results)
	}
	errs, _ := doc["errors"].([]any)
	if len(errs) != 1 {
		t.Fatalf("top-level errors = %v, want exactly one", doc["errors"])
	}
	if code := errs[0].(map[string]any)["code"]; code != "E_LAYAT_FAILED" {
		t.Errorf("code = %v, want E_LAYAT_FAILED", code)
	}
}

// stubNixForApply puts a fake nix first on PATH that logs every invocation to the returned file.
// `nix eval … --apply …` prints evalAll (the batch eval), any other eval prints "home" (evalRoot's
// rootKind), and `nix build` fails. The log tells whether build was reached and how many evals ran.
func stubNixForApply(t *testing.T, evalAll string) string {
	t.Helper()
	bin := t.TempDir()
	log := filepath.Join(bin, "nix.log")
	data := filepath.Join(bin, "eval-all.json")
	if err := os.WriteFile(data, []byte(evalAll), 0o644); err != nil {
		t.Fatalf("write eval-all.json: %v", err)
	}
	script := "#!/bin/sh\n" +
		"echo \"$*\" >> '" + log + "'\n" +
		"case \"$1\" in\n" +
		"eval) case \" $* \" in *' --apply '*) cat '" + data + "';; *) printf home;; esac;;\n" +
		"*) echo 'error: stub build failed' >&2; exit 1;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(bin, "nix"), []byte(script), 0o755); err != nil {
		t.Fatalf("write nix stub: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return log
}

// nixCalls reads the stub's log as one line per nix invocation.
func nixCalls(t *testing.T, log string) []string {
	t.Helper()
	b, err := os.ReadFile(log)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read nix log: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

// withFlakeEntrypoint points -f at a temp dir holding a flake.nix and isolates the state / home dirs.
func withFlakeEntrypoint(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "flake.nix"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write flake.nix: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	origFile, origDryrun, origRoot, origHome := flagFile, flagDryrun, flagRoot, flagHomeRoot
	t.Cleanup(func() { flagFile, flagDryrun, flagRoot, flagHomeRoot = origFile, origDryrun, origRoot, origHome })
	flagFile, flagRoot, flagHomeRoot = dir, "", false
}

// conflictMarker is a phrase only the preflight's error carries, so a build failure's aggregate
// error ("N config(s) failed") cannot pass for it.
const conflictMarker = "claimed by both"

// TestRunApplyAllPreflight drives runApplyAll through a stub nix: a conflict stops the run after
// the batch eval, so no build (hence no placement) is ever started, with or without --dryrun.
// The selection and --root reach the detector from runApplyAll itself: a conflict outside the root
// filter's selection goes on to build, and --root makes configs of different rootKinds collide.
func TestRunApplyAllPreflight(t *testing.T) {
	const sameKind = `{"a":{"rootKind":"home","targets":["dup"]},"b":{"rootKind":"home","targets":["dup"]}}`
	const unselected = `{"a":{"rootKind":"home","targets":["dup"]},"b":{"rootKind":"project","targets":["dup"]},"c":{"rootKind":"project","targets":["dup"]}}`
	const mixedKinds = `{"a":{"rootKind":"home","targets":["dup"]},"b":{"rootKind":"project","targets":["dup"]}}`
	cases := []struct {
		name     string
		evalAll  string
		dryrun   bool
		homeRoot bool
		root     string
		wantStop bool
	}{
		{name: "apply", evalAll: sameKind, wantStop: true},
		{name: "dryrun", evalAll: sameKind, dryrun: true, wantStop: true},
		{name: "conflict outside the --home-root selection", evalAll: unselected, homeRoot: true},
		{name: "--root collapses rootKinds", evalAll: mixedKinds, root: "/tmp/override", wantStop: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withFlakeEntrypoint(t)
			flagDryrun, flagHomeRoot, flagRoot = tc.dryrun, tc.homeRoot, tc.root
			log := stubNixForApply(t, tc.evalAll)
			run, buf := newApplyTestRun()
			err := runApplyAll(run)
			calls := nixCalls(t, log)
			if !tc.wantStop {
				// The stub build fails, so the run errors — but on the build, not on the preflight.
				if err != nil && strings.Contains(err.Error(), conflictMarker) {
					t.Fatalf("runApplyAll = %v, want no conflict (only the selected configs are checked)", err)
				}
				if len(calls) < 2 {
					t.Errorf("nix calls = %q, want the run to go on to build", calls)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), conflictMarker) {
				t.Fatalf("runApplyAll = %v, want the conflict error", err)
			}
			if len(calls) != 1 || !strings.HasPrefix(calls[0], "eval ") {
				t.Errorf("nix calls = %q, want the batch eval alone (no build)", calls)
			}
			// No subject is registered before the preflight, so the error lands on the top level.
			if err := run.emit(err); err != nil {
				t.Fatalf("emit: %v", err)
			}
			doc := decodeEnvelope(t, buf)
			if results, _ := doc["results"].([]any); len(results) != 0 {
				t.Errorf("results = %v, want none", results)
			}
			if errs, _ := doc["errors"].([]any); len(errs) != 1 {
				t.Errorf("top-level errors = %v, want exactly one", doc["errors"])
			}
		})
	}
}

// TestRunApplyNamedIsNotChecked drives a named apply through the same stub: it runs evalRoot's
// single eval and goes straight on to build — no batch eval is added for a preflight.
func TestRunApplyNamedIsNotChecked(t *testing.T) {
	withFlakeEntrypoint(t)
	flagDryrun = false
	log := stubNixForApply(t, `{}`)
	run, _ := newApplyTestRun()
	if err := runApply(run, "a"); err == nil {
		t.Fatal("runApply must surface the stub build failure")
	}
	var evals, builds int
	for _, c := range nixCalls(t, log) {
		switch {
		case strings.Contains(c, "--apply"):
			t.Errorf("named apply ran a batch eval: %q", c)
		case strings.HasPrefix(c, "eval "):
			evals++
		case strings.HasPrefix(c, "build "):
			builds++
		}
	}
	if evals != 1 || builds != 1 {
		t.Errorf("evals = %d, builds = %d, want 1 and 1 (evalRoot then the in-lock build)", evals, builds)
	}
}
