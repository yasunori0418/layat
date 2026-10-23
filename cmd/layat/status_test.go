package main

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
)

// stubStatus replaces statusFn with one returning series, and returns the options it was called with.
func stubStatus(t *testing.T, series []engine.StatusSeries) *engine.StatusOptions {
	t.Helper()
	orig := statusFn
	t.Cleanup(func() { statusFn = orig })
	got := &engine.StatusOptions{}
	statusFn = func(opts engine.StatusOptions) ([]engine.StatusSeries, error) {
		*got = opts
		return series, nil
	}
	return got
}

// withJSONFlag sets flagJSON for the test and restores it afterwards.
func withJSONFlag(t *testing.T, v bool) {
	t.Helper()
	orig := flagJSON
	t.Cleanup(func() { flagJSON = orig })
	flagJSON = v
}

// statusFixture covers every subject-name shape: a <name>-keyed record, a <roothash> series
// record, a record under an unreadable backref, and a series reported as a whole.
func statusFixture() []engine.StatusSeries {
	return []engine.StatusSeries{
		{
			Mode: "home", Name: "docs", RecordFile: "/s/layat/docs/last-applied.json",
			AppliedAt: "2026-10-23T10:00:00+09:00", LinkFarm: "/nix/store/aaa-layat-docs",
			Manifest: &manifest.Manifest{
				SchemaVersion: 1,
				Root:          manifest.Root{RootKind: "home"},
				Entries: []manifest.Entry{
					{SrcKind: manifest.SrcKindStore, Src: "/nix/store/bbb-skills", Subpath: "nix", Target: ".claude/skills/nix", Method: manifest.MethodSymlink},
					{SrcKind: manifest.SrcKindRootRelative, Src: "dotfiles/vimrc", Subpath: ".", Target: ".vimrc", Method: manifest.MethodCopy},
				},
			},
		},
		{
			Mode: "project", Name: "docs", Root: "/home/u/src/proj", RootHash: "1111",
			RecordFile: "/s/layat/1111/docs/last-applied.json",
			AppliedAt:  "2026-10-23T11:00:00+09:00", LinkFarm: "/nix/store/ccc-layat-docs",
			Manifest: &manifest.Manifest{SchemaVersion: 1, Root: manifest.Root{RootKind: "project"}},
		},
		{
			Mode: "home", Name: "web", RootHash: "2222", RecordFile: "/s/layat/2222/web/last-applied.json",
			Err: errors.New("layat: cannot read the backref (/s/layat/2222/.root)"),
		},
		{Mode: "system", RootHash: "3333", Err: errors.New("layat: cannot list the series (/var/lib/layat/3333)")},
	}
}

// TestStatusSubjectNames pins the subject-name shape of every kind of series.
func TestStatusSubjectNames(t *testing.T) {
	for _, tc := range []struct {
		s    engine.StatusSeries
		want string
	}{
		{engine.StatusSeries{Mode: "home", Name: "docs"}, "home:docs"},
		{engine.StatusSeries{Mode: "system", Name: "etc"}, "system:etc"},
		{engine.StatusSeries{Mode: "project", Name: "docs", Root: "/p", RootHash: "1111"}, "project:docs@/p"},
		{engine.StatusSeries{Mode: "home", Name: "web", RootHash: "2222"}, "home:web@2222"},
		{engine.StatusSeries{Mode: "system", RootHash: "3333"}, "system:@3333"},
	} {
		if got := statusSubjectName(tc.s); got != tc.want {
			t.Errorf("statusSubjectName(%+v) = %q, want %q", tc.s, got, tc.want)
		}
	}
}

// TestStatusSystemStateDirOverride: LAYAT_SYSTEM_STATE_DIR reaches engine.Status as
// SystemStateDir, and its absence passes "" so the engine keeps its default.
func TestStatusSystemStateDirOverride(t *testing.T) {
	withJSONFlag(t, false)
	for _, want := range []string{"/tmp/isolated-state", ""} {
		t.Setenv(systemStateEnv, want)
		got := stubStatus(t, nil)
		run, _ := newStatusTestRun()
		captureOutErr(t, func() {
			if err := runStatus(run, nil); err != nil {
				t.Errorf("runStatus: %v", err)
			}
		})
		if got.SystemStateDir != want {
			t.Errorf("SystemStateDir = %q, want %q", got.SystemStateDir, want)
		}
		if got.StateDir != "" {
			t.Errorf("StateDir = %q, want empty (the engine resolves it)", got.StateDir)
		}
	}
}

// TestStatusTextOutput pins the heading / entry lines, the blank separator, and that an error
// subject goes to stderr while the others stay on stdout with exit 1.
func TestStatusTextOutput(t *testing.T) {
	withJSONFlag(t, false)
	stubStatus(t, statusFixture())
	run, _ := newStatusTestRun()

	var err error
	out, errOut := captureOutErr(t, func() { err = runStatus(run, nil) })

	want := "# home:docs (applied 2026-10-23T10:00:00+09:00)\n" +
		".claude/skills/nix\tsymlink\t/nix/store/bbb-skills/nix\n" +
		".vimrc\tcopy\tdotfiles/vimrc\n" +
		"\n" +
		"# project:docs@/home/u/src/proj (applied 2026-10-23T11:00:00+09:00)\n"
	if out != want {
		t.Errorf("stdout =\n%q\nwant\n%q", out, want)
	}
	for _, subject := range []string{"home:web@2222", "system:@3333"} {
		if !strings.Contains(errOut, subject) {
			t.Errorf("stderr = %q, want the error of %s", errOut, subject)
		}
		if strings.Contains(out, subject) {
			t.Errorf("stdout = %q, want no line for the error subject %s", out, subject)
		}
	}
	var ec exitCodeX
	if !errors.As(err, &ec) || ec.ExitCode() != 1 {
		t.Errorf("runStatus error = %v, want an exit code 1 error", err)
	}
}

// TestStatusFilter: [name] matches the <name> part exactly across modes and roots, and keeps a
// series reported as a whole since the name may be inside it.
func TestStatusFilter(t *testing.T) {
	withJSONFlag(t, true)
	for _, tc := range []struct {
		name string
		want []string
	}{
		{"docs", []string{"home:docs", "project:docs@/home/u/src/proj", "system:@3333"}},
		{"web", []string{"home:web@2222", "system:@3333"}},
		{"doc", []string{"system:@3333"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stubStatus(t, statusFixture())
			run, buf := newStatusTestRun()
			var err error
			out, errOut := captureOutErr(t, func() { err = runStatus(run, []string{tc.name}) })
			if out != "" {
				t.Errorf("stdout = %q, want empty under --json", out)
			}
			if strings.Contains(errOut, "no applied config") {
				t.Errorf("stderr = %q, want no zero-match notice while an error subject is listed", errOut)
			}
			if err == nil {
				t.Errorf("runStatus = nil, want exit 1 (the whole-series error subject is listed)")
			}
			if emitErr := run.emit(err); emitErr != nil {
				t.Fatalf("emit: %v", emitErr)
			}
			_, names := subjectResults(t, buf)
			if !slices.Equal(names, tc.want) {
				t.Errorf("subjects = %v, want %v", names, tc.want)
			}
		})
	}
}

// TestStatusZeroMatch: with nothing listed, stdout stays empty, stderr carries one notice line
// and the command exits 0; --json emits results: [].
func TestStatusZeroMatch(t *testing.T) {
	series := statusFixture()[:2]
	for _, jsonMode := range []bool{false, true} {
		withJSONFlag(t, jsonMode)
		stubStatus(t, series)
		run, buf := newStatusTestRun()
		var err error
		out, errOut := captureOutErr(t, func() { err = runStatus(run, []string{"nope"}) })
		if err != nil {
			t.Errorf("json=%v: runStatus = %v, want nil (exit 0)", jsonMode, err)
		}
		if out != "" {
			t.Errorf("json=%v: stdout = %q, want empty", jsonMode, out)
		}
		if errOut != "layat: status: no applied config named nope\n" {
			t.Errorf("json=%v: stderr = %q, want one notice line", jsonMode, errOut)
		}
		if !jsonMode {
			continue
		}
		if emitErr := run.emit(err); emitErr != nil {
			t.Fatalf("emit: %v", emitErr)
		}
		checkConformance(t, buf)
		doc := decodeEnvelope(t, buf)
		if doc["status"] != "success" {
			t.Errorf("status = %v, want success", doc["status"])
		}
		results, ok := doc["results"].([]any)
		if !ok || len(results) != 0 {
			t.Errorf("results = %#v, want []", doc["results"])
		}
	}
}

// TestStatusJSONShape pins result.info {mode, root, appliedAt, linkFarm, entries}, entries as an
// array even when empty, items: [], no generation slot, errors[] on the error subjects, the error
// envelope status and conformance.
func TestStatusJSONShape(t *testing.T) {
	withJSONFlag(t, true)
	stubStatus(t, statusFixture())
	run, buf := newStatusTestRun()

	var err error
	out, _ := captureOutErr(t, func() { err = runStatus(run, nil) })
	if out != "" {
		t.Errorf("stdout = %q, want empty (the envelope owns stdout)", out)
	}
	if emitErr := run.emit(err); emitErr != nil {
		t.Fatalf("emit: %v", emitErr)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" {
		t.Errorf("status = %v, want error", doc["status"])
	}
	if doc["dryRun"] != false {
		t.Errorf("dryRun = %v, want false", doc["dryRun"])
	}
	byName, names := subjectResults(t, buf)
	want := []string{"home:docs", "project:docs@/home/u/src/proj", "home:web@2222", "system:@3333"}
	if !slices.Equal(names, want) {
		t.Fatalf("subjects = %v, want %v", names, want)
	}

	for _, name := range want[:2] {
		sr := byName[name]
		if sr["status"] != "success" {
			t.Errorf("%s: status = %v, want success", name, sr["status"])
		}
		if _, ok := sr["generation"]; ok {
			t.Errorf("%s: generation slot present, want absent", name)
		}
		res := sr["result"].(map[string]any)
		if items, ok := res["items"].([]any); !ok || len(items) != 0 {
			t.Errorf("%s: items = %#v, want []", name, res["items"])
		}
	}

	info := byName["home:docs"]["result"].(map[string]any)["info"].(map[string]any)
	if info["mode"] != "home" || info["root"] != "" || info["appliedAt"] != "2026-10-23T10:00:00+09:00" ||
		info["linkFarm"] != "/nix/store/aaa-layat-docs" {
		t.Errorf("home:docs info = %v", info)
	}
	entries := info["entries"].([]any)
	if len(entries) != 2 {
		t.Fatalf("home:docs entries = %v, want 2", entries)
	}
	first := entries[0].(map[string]any)
	if first["target"] != ".claude/skills/nix" || first["method"] != "symlink" || first["srcKind"] != "store" ||
		first["src"] != "/nix/store/bbb-skills" || first["subpath"] != "nix" {
		t.Errorf("entries[0] = %v", first)
	}

	projInfo := byName["project:docs@/home/u/src/proj"]["result"].(map[string]any)["info"].(map[string]any)
	if projInfo["mode"] != "project" || projInfo["root"] != "/home/u/src/proj" {
		t.Errorf("project info = %v", projInfo)
	}
	if e, ok := projInfo["entries"].([]any); !ok || len(e) != 0 {
		t.Errorf("project entries = %#v, want []", projInfo["entries"])
	}

	for _, name := range want[2:] {
		sr := byName[name]
		if sr["status"] != "error" {
			t.Errorf("%s: status = %v, want error", name, sr["status"])
		}
		if errs, ok := sr["errors"].([]any); !ok || len(errs) != 1 {
			t.Errorf("%s: errors = %#v, want one error", name, sr["errors"])
		}
	}
}
