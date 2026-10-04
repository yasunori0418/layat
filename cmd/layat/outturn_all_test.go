package main

// Conformance and shape tests for the --all paths' multiple SubjectResult output. N=1 and N>1
// documents differ only in the length of results[].

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
	"github.com/yasunori0418/layat/internal/planner"
)

// checkConformance fails the test unless buf holds a document outturn's checker accepts (schema
// with format assertions + the schema-external lint MUSTs).
func checkConformance(t *testing.T, buf *bytes.Buffer) {
	t.Helper()
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings:\n%s\ndocument: %s", strings.Join(findings, "\n"), buf.String())
	}
}

// subjectResults decodes the envelope and returns its results[] keyed by subject name, with the
// ordered names.
func subjectResults(t *testing.T, buf *bytes.Buffer) (map[string]map[string]any, []string) {
	t.Helper()
	doc := decodeEnvelope(t, buf)
	byName := map[string]map[string]any{}
	var order []string
	for _, r := range doc["results"].([]any) {
		sr := r.(map[string]any)
		name := sr["subject"].(map[string]any)["name"].(string)
		if _, dup := byName[name]; dup {
			t.Fatalf("subject %q appears twice in results[]", name)
		}
		byName[name] = sr
		order = append(order, name)
	}
	return byName, order
}

// statusAndErrors returns the SubjectResult's status and its (possibly absent) errors[].
func statusAndErrors(t *testing.T, sr map[string]any) (string, []any) {
	t.Helper()
	errs, _ := sr["errors"].([]any)
	return sr["status"].(string), errs
}

// placedResult is one config's successful apply with a single placed entry and its own profile.
func placedResult(name string) *engine.Result {
	return &engine.Result{
		Profile: "/state/nix/profiles/layat/" + name + "/profile",
		Entries: []manifest.Entry{{Target: "t/" + name}},
		Placed:  []string{"t/" + name},
	}
}

// TestApplyAllPartialFailureKeepsEverySubject: with one config failing, the envelope is error yet
// carries every succeeded config's SubjectResult. The failure stays on its own subject, and the
// top-level errors[] stays absent.
func TestApplyAllPartialFailureKeepsEverySubject(t *testing.T) {
	run, buf := newApplyTestRun()
	applied, skipped, failures := aggregateApply(run, []string{"a", "b", "c"}, 4, func(name string) (*engine.Result, error) {
		if name == "b" {
			return nil, io.ErrUnexpectedEOF
		}
		return placedResult(name), nil
	})
	if applied != 2 || skipped != 0 || failures != 1 {
		t.Fatalf("counts = (%d, %d, %d), want (2, 0, 1)", applied, skipped, failures)
	}
	// main emits with the aggregate error the exit-code path builds from the same counts.
	cmdErr := &exitCodeError{code: applyAllExitCode(failures > 0, false), msg: "layat: apply --all: 1 config(s) failed"}
	if err := run.emit(cmdErr); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" {
		t.Errorf("aggregate status = %v, want error (one subject failed)", doc["status"])
	}
	if topErrs, ok := doc["errors"]; ok {
		t.Errorf("top-level errors = %v, want absent — the failure belongs to subject b", topErrs)
	}
	if cmdErr.ExitCode() == 0 {
		t.Error("exit code = 0, want non-zero alongside the error status")
	}

	byName, order := subjectResults(t, buf)
	if want := []string{"a", "b", "c"}; !slices.Equal(order, want) {
		t.Errorf("results order = %v, want %v (selection order)", order, want)
	}
	for _, name := range []string{"a", "c"} {
		status, errs := statusAndErrors(t, byName[name])
		if status != "success" {
			t.Errorf("subject %s status = %s, want success (its own apply succeeded)", name, status)
		}
		if len(errs) != 0 {
			t.Errorf("subject %s errors = %v, want none — b's failure must not colour it", name, errs)
		}
		items := byName[name]["result"].(map[string]any)["items"].([]any)
		if len(items) != 1 {
			t.Errorf("subject %s items = %v, want the placed entry (the succeeded result is kept whole)", name, items)
		}
	}
	status, errs := statusAndErrors(t, byName["b"])
	if status != "error" {
		t.Errorf("subject b status = %s, want error", status)
	}
	if len(errs) != 1 {
		t.Fatalf("subject b errors = %v, want exactly one (subject-borne: no engine result exists)", errs)
	}
}

// TestApplyAllPartialFailureItemBorne: when the engine returns a partial result with the error,
// the failure is item-borne, so the subject's errors[] stays empty while its status is error.
func TestApplyAllPartialFailureItemBorne(t *testing.T) {
	run, buf := newApplyTestRun()
	failure := errors.New("layat: symlink t/b: permission denied")
	_, _, failures := aggregateApply(run, []string{"a", "b", "c"}, 4, func(name string) (*engine.Result, error) {
		if name == "b" {
			// The engine's partial result: the entry it stopped on, plus a planned entry it never reached.
			res := placedResult(name)
			res.Placed = nil
			res.Entries = append(res.Entries, manifest.Entry{Target: "t/" + name + "-later"})
			res.FailedTarget = "t/" + name
			res.Unreached = []string{"t/" + name + "-later"}
			return res, failure
		}
		return placedResult(name), nil
	})
	if failures != 1 {
		t.Fatalf("failures = %d, want 1", failures)
	}
	if err := run.emit(&exitCodeError{code: 1, msg: "layat: apply --all: 1 config(s) failed"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" {
		t.Errorf("aggregate status = %v, want error", doc["status"])
	}
	if topErrs, ok := doc["errors"]; ok {
		t.Errorf("top-level errors = %v, want absent", topErrs)
	}
	byName, _ := subjectResults(t, buf)
	status, errs := statusAndErrors(t, byName["b"])
	if status != "error" {
		t.Errorf("failing subject status = %s, want error (its failed item makes it error)", status)
	}
	if len(errs) != 0 {
		t.Errorf("failing subject errors = %v, want none — the failure is item-borne (outturn §2)", errs)
	}
	// The reached-state partition has to survive onto the items, not just the status.
	byTarget := map[string]map[string]any{}
	for _, it := range byName["b"]["result"].(map[string]any)["items"].([]any) {
		item := it.(map[string]any)
		byTarget[item["label"].(string)] = item
	}
	failed := byTarget["t/b"]
	if failed == nil || failed["status"] != "failed" {
		t.Fatalf("item t/b = %v, want status failed", failed)
	}
	if itemErr, ok := failed["error"].(map[string]any); !ok || itemErr["message"] != failure.Error() {
		t.Errorf("item t/b error = %v, want the config's own failure carried on the item", failed["error"])
	}
	if unreached := byTarget["t/b-later"]; unreached == nil || unreached["status"] != "skipped" {
		t.Errorf("item t/b-later = %v, want status skipped (planned but never attempted)", unreached)
	}
	// The sibling configs still carry their own successful inventories.
	for _, name := range []string{"a", "c"} {
		if s, e := statusAndErrors(t, byName[name]); s != "success" || len(e) != 0 {
			t.Errorf("subject %s = (%s, %v), want (success, none)", name, s, e)
		}
	}
}

// TestApplyAllSharedTargetKeepsItemIDsResultScoped: item ids derive from the target alone, so two
// configs with the same target share an item.id across results[]. The document still conforms,
// since references resolve within one SubjectResult.
func TestApplyAllSharedTargetKeepsItemIDsResultScoped(t *testing.T) {
	run, buf := newApplyTestRun()
	shared := manifest.Entry{Target: ".config/shared"}
	aggregateApply(run, []string{"a", "b"}, 4, func(name string) (*engine.Result, error) {
		res := placedResult(name)
		res.Entries = []manifest.Entry{shared}
		res.Placed = []string{shared.Target}
		return res, nil
	})
	if err := run.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	// The whole point: a document carrying the same item id twice, in two results, still conforms.
	checkConformance(t, buf)

	byName, _ := subjectResults(t, buf)
	ids := map[string]string{}
	for _, name := range []string{"a", "b"} {
		items := byName[name]["result"].(map[string]any)["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("subject %s items = %v, want the shared entry", name, items)
		}
		item := items[0].(map[string]any)
		ids[name] = item["id"].(string)
		// Each result's changes must reference its own result's item, never a sibling's.
		changes := byName[name]["result"].(map[string]any)["changes"].([]any)
		if len(changes) != 1 {
			t.Fatalf("subject %s changes = %v, want the placement", name, changes)
		}
		if got := changes[0].(map[string]any)["itemId"]; got != ids[name] {
			t.Errorf("subject %s change itemId = %v, want its own item %s", name, got, ids[name])
		}
	}
	if ids["a"] != ids["b"] {
		t.Errorf("item ids = %v, want equal — the id derives from the target alone, not the config", ids)
	}
}

// TestApplyAllSkipIsNotAFailure: a try-lock skip settles its subject as success, and the aggregate
// stays success.
func TestApplyAllSkipIsNotAFailure(t *testing.T) {
	run, buf := newApplyTestRun()
	applied, skipped, failures := aggregateApply(run, []string{"a", "b"}, 4, func(name string) (*engine.Result, error) {
		if name == "b" {
			// The engine returns a result with no entries alongside ErrSkipped, so a payload is attached too.
			return &engine.Result{Profile: placedResult(name).Profile, Skipped: true}, engine.ErrSkipped
		}
		return placedResult(name), nil
	})
	if applied != 1 || skipped != 1 || failures != 0 {
		t.Fatalf("counts = (%d, %d, %d), want (1, 1, 0)", applied, skipped, failures)
	}
	if err := run.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	if got := decodeEnvelope(t, buf)["status"]; got != "success" {
		t.Errorf("aggregate status = %v, want success (a skip is not a failure)", got)
	}
	byName, _ := subjectResults(t, buf)
	if status, errs := statusAndErrors(t, byName["b"]); status != "success" || len(errs) != 0 {
		t.Errorf("skipped subject = (%s, %v), want (success, none)", status, errs)
	}
}

// TestApplyAllEmptySelectionEmitsEmptyResults: an --all matching no config emits results: [] with
// status success.
func TestApplyAllEmptySelectionEmitsEmptyResults(t *testing.T) {
	run, buf := newApplyTestRun()
	if applied, skipped, failures := aggregateApply(run, nil, 4, func(string) (*engine.Result, error) {
		t.Fatal("apply must not run for an empty selection")
		return nil, nil
	}); applied+skipped+failures != 0 {
		t.Fatalf("counts = (%d, %d, %d), want all zero", applied, skipped, failures)
	}
	if err := run.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "success" {
		t.Errorf("status = %v, want success (nothing selected is not a failure)", doc["status"])
	}
	results, ok := doc["results"].([]any)
	if !ok {
		t.Fatalf("results = %v, want the key present as an array", doc["results"])
	}
	if len(results) != 0 {
		t.Errorf("results = %v, want []", results)
	}
}

// TestApplyAllDryRunConflictIsItemBorne: a conflicting entry is a failed item with
// E_LAYAT_COLLISION and is not repeated in errors[]; subject and aggregate are error, and the exit
// code is conflict 2.
func TestApplyAllDryRunConflictIsItemBorne(t *testing.T) {
	run, buf := newApplyTestRun()
	run.dryRun = true
	var code int
	captureStdout(t, func() {
		code = aggregateDryRun(run, []string{"a", "b"}, 4, func(name string) (*engine.Result, error) {
			if name == "b" {
				conflicted := placedResult(name)
				conflicted.Placed = nil
				conflicted.Conflicts = []planner.Conflict{
					{Entry: manifest.Entry{Target: "t/" + name}, Reason: "occupied by a foreign entity"},
				}
				return conflicted, nil
			}
			return placedResult(name), nil
		})
	})
	if want := applyAllExitCode(false, true); code != want {
		t.Fatalf("aggregateDryRun code = %d, want %d (conflict, no error)", code, want)
	}
	// main turns that code into the exitError it exits with; emit sees it as the command error.
	if err := run.emit(&exitError{code: code}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" {
		t.Errorf("aggregate status = %v, want error (a conflicting config is in error)", doc["status"])
	}
	if doc["dryRun"] != true {
		t.Errorf("dryRun = %v, want true", doc["dryRun"])
	}
	if topErrs, ok := doc["errors"]; ok {
		t.Errorf("top-level errors = %v, want absent (the conflict belongs to subject b's item)", topErrs)
	}

	byName, _ := subjectResults(t, buf)
	if status, errs := statusAndErrors(t, byName["a"]); status != "success" || len(errs) != 0 {
		t.Errorf("clean subject = (%s, %v), want (success, none)", status, errs)
	}
	status, errs := statusAndErrors(t, byName["b"])
	if status != "error" {
		t.Errorf("conflicting subject status = %s, want error", status)
	}
	if len(errs) != 0 {
		t.Errorf("conflicting subject errors = %v, want none — the conflict is item-borne (outturn §2)", errs)
	}
	items := byName["b"]["result"].(map[string]any)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("conflicting subject items = %v, want the conflicting entry", items)
	}
	item := items[0].(map[string]any)
	if item["status"] != "failed" {
		t.Errorf("conflicting item status = %v, want failed", item["status"])
	}
	itemErr, ok := item["error"].(map[string]any)
	if !ok {
		t.Fatalf("conflicting item = %v, want an error object", item)
	}
	if itemErr["code"] != "E_LAYAT_COLLISION" {
		t.Errorf("conflicting item error code = %v, want E_LAYAT_COLLISION", itemErr["code"])
	}
}

// TestApplyAllDryRunMixedErrorAndConflict: an outright failure carries a subject-level error, a
// conflict carries none, and the exit code is error(1).
func TestApplyAllDryRunMixedErrorAndConflict(t *testing.T) {
	run, buf := newApplyTestRun()
	run.dryRun = true
	buildFailure := errors.New("layat: nix build failed")
	var code int
	captureStdout(t, func() {
		code = aggregateDryRun(run, []string{"broken", "clashing", "clean"}, 4, func(name string) (*engine.Result, error) {
			switch name {
			case "broken":
				return nil, buildFailure
			case "clashing":
				res := placedResult(name)
				res.Placed = nil
				res.Conflicts = []planner.Conflict{
					{Entry: manifest.Entry{Target: "t/" + name}, Reason: "occupied by a foreign entity"},
				}
				return res, nil
			}
			return placedResult(name), nil
		})
	})
	if want := applyAllExitCode(true, true); code != want {
		t.Fatalf("code = %d, want %d (error must win over conflict)", code, want)
	}
	if err := run.emit(&exitError{code: code}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checkConformance(t, buf)

	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" {
		t.Errorf("aggregate status = %v, want error", doc["status"])
	}
	if topErrs, ok := doc["errors"]; ok {
		t.Errorf("top-level errors = %v, want absent (both failures belong to their configs)", topErrs)
	}
	byName, _ := subjectResults(t, buf)
	// The outright failure is subject-borne: no engine result exists, so nothing else can carry it.
	status, errs := statusAndErrors(t, byName["broken"])
	if status != "error" || len(errs) != 1 {
		t.Fatalf("failed subject = (%s, %v), want (error, exactly one subject-borne error)", status, errs)
	}
	if got := errs[0].(map[string]any)["code"]; got != "E_LAYAT_FAILED" {
		t.Errorf("failed subject error code = %v, want E_LAYAT_FAILED (the build error is unclassified here)", got)
	}
	// The conflict is item-borne: same error status, but errors[] stays empty.
	status, errs = statusAndErrors(t, byName["clashing"])
	if status != "error" {
		t.Errorf("conflicting subject status = %s, want error", status)
	}
	if len(errs) != 0 {
		t.Errorf("conflicting subject errors = %v, want none — its item carries the conflict", errs)
	}
	if status, errs := statusAndErrors(t, byName["clean"]); status != "success" || len(errs) != 0 {
		t.Errorf("clean subject = (%s, %v), want (success, none)", status, errs)
	}
}

// TestApplyAllSingleConfigMatchesNamedApply: a one-config --all and a named apply of the same config
// emit byte-identical documents once the clocks agree.
func TestApplyAllSingleConfigMatchesNamedApply(t *testing.T) {
	res := placedResult("default")

	// The named path: one subject, the payload attached, the command error carried by emit.
	named, namedBuf := newApplyTestRun()
	attachMutationPayload(named.beginSubject("default"), res, nil)
	if err := named.emit(nil); err != nil {
		t.Fatalf("emit named: %v", err)
	}

	// The --all path at N=1: the same subject, settled by the aggregator itself.
	all, allBuf := newApplyTestRun()
	aggregateApply(all, []string{"default"}, 4, func(string) (*engine.Result, error) { return res, nil })
	if err := all.emit(nil); err != nil {
		t.Fatalf("emit all: %v", err)
	}

	if allBuf.String() != namedBuf.String() {
		t.Errorf("apply --all at N=1 diverged from the named apply\n--all: %s\nnamed: %s", allBuf, namedBuf)
	}
	checkConformance(t, allBuf)
}

// TestApplyAllSingleConfigFailureMatchesNamedApply: on failure too, the named apply (settled by
// emit) and --all (settled by the aggregator) emit the same document.
func TestApplyAllSingleConfigFailureMatchesNamedApply(t *testing.T) {
	failure := errors.New("layat: generation commit (nix-env --set) failed")
	// A commit failure has a result but no failed target, so it is subject-borne on both paths.
	newRes := func() *engine.Result {
		res := placedResult("default")
		res.Placed = nil
		return res
	}

	named, namedBuf := newApplyTestRun()
	attachMutationPayload(named.beginSubject("default"), newRes(), failure)
	if err := named.emit(failure); err != nil {
		t.Fatalf("emit named: %v", err)
	}

	all, allBuf := newApplyTestRun()
	aggregateApply(all, []string{"default"}, 4, func(string) (*engine.Result, error) { return newRes(), failure })
	// --all's aggregate error must not reach the already settled subject.
	if err := all.emit(&exitCodeError{code: 1, msg: "layat: apply --all: 1 config(s) failed"}); err != nil {
		t.Fatalf("emit all: %v", err)
	}

	if allBuf.String() != namedBuf.String() {
		t.Errorf("failing apply --all at N=1 diverged from the named apply\n--all: %s\nnamed: %s", allBuf, namedBuf)
	}
	checkConformance(t, allBuf)
	// Guard the equality against being satisfied by both sides losing the error.
	byName, _ := subjectResults(t, allBuf)
	if status, errs := statusAndErrors(t, byName["default"]); status != "error" || len(errs) != 1 {
		t.Fatalf("subject = (%s, %v), want (error, exactly one subject-borne error)", status, errs)
	}
}

// stubNixEnvListGenerations puts a fake nix-env first on PATH: it prints one generation for every
// profile except failFor, for which it exits non-zero.
func stubNixEnvListGenerations(t *testing.T, failFor string) {
	t.Helper()
	bin := t.TempDir()
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do case \"$prev\" in --profile) prof=$a;; esac; prev=$a; done\n" +
		"case \"$prof\" in\n" +
		"*/" + failFor + "/profile) echo 'error: not a valid profile' >&2; exit 1;;\n" +
		"esac\n" +
		"printf '   1   2026-07-19 12:00:00   (current)\\n'\n"
	if err := os.WriteFile(filepath.Join(bin, "nix-env"), []byte(script), 0o755); err != nil {
		t.Fatalf("write nix-env stub: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// makeHomeProfiles creates <state>/nix/profiles/layat/<name>/profile links, the home-mode layout
// runListAllGenerations scans.
func makeHomeProfiles(t *testing.T, names ...string) {
	t.Helper()
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	for _, name := range names {
		dir := filepath.Join(paths.Base(state), name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		// The scan only lstats the link, so the destination need not exist.
		if err := os.Symlink("profile-1-link", filepath.Join(dir, "profile")); err != nil {
			t.Fatalf("symlink profile: %v", err)
		}
	}
}

// TestListGenerationsAllWiring drives the real runListAllGenerations, which registers and settles
// one subject per scanned config.
func TestListGenerationsAllWiring(t *testing.T) {
	origJSON := flagJSON
	defer func() { flagJSON = origJSON }()
	flagJSON = true

	t.Run("every scanned config becomes its own SubjectResult", func(t *testing.T) {
		makeHomeProfiles(t, "home", "work")
		stubNixEnvListGenerations(t, "")
		run, buf := newListGenerationsTestRun()
		if err := runListAllGenerations(run); err != nil {
			t.Fatalf("runListAllGenerations: %v", err)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		checkConformance(t, buf)

		byName, order := subjectResults(t, buf)
		want := []string{"home", "work"}
		if !slices.Equal(order, want) {
			t.Fatalf("results order = %v, want %v (lexical scan order)", order, want)
		}
		for _, name := range want {
			if status, errs := statusAndErrors(t, byName[name]); status != "success" || len(errs) != 0 {
				t.Errorf("subject %s = (%s, %v), want (success, none)", name, status, errs)
			}
			if got := infoArray(t, byName[name], "generations"); len(got) != 1 {
				t.Errorf("subject %s generations = %v, want its own listing", name, got)
			}
		}
	})

	t.Run("a mid-enumeration failure lands on its own subject", func(t *testing.T) {
		// "work" sorts after "home", so the scan lists home, then fails on work and stops.
		makeHomeProfiles(t, "home", "work", "zzz")
		stubNixEnvListGenerations(t, "work")
		run, buf := newListGenerationsTestRun()
		err := runListAllGenerations(run)
		if err == nil {
			t.Fatal("runListAllGenerations must fail when a profile cannot be listed")
		}
		if err := run.emit(err); err != nil {
			t.Fatalf("emit: %v", err)
		}
		checkConformance(t, buf)

		doc := decodeEnvelope(t, buf)
		if doc["status"] != "error" {
			t.Errorf("aggregate status = %v, want error", doc["status"])
		}
		// The failure belongs to its config's subject; the top-level errors[] stays empty.
		if topErrs, ok := doc["errors"]; ok {
			t.Errorf("top-level errors = %v, want absent (the failure belongs to subject work)", topErrs)
		}
		byName, order := subjectResults(t, buf)
		// The scan stops at the failure, so the untouched config never becomes a subject at all.
		if want := []string{"home", "work"}; !slices.Equal(order, want) {
			t.Fatalf("results = %v, want %v — the scan stops at the failure and zzz is never reached", order, want)
		}
		if status, errs := statusAndErrors(t, byName["home"]); status != "success" || len(errs) != 0 {
			t.Errorf("already-listed subject = (%s, %v), want (success, none) — it keeps its result", status, errs)
		}
		if status, errs := statusAndErrors(t, byName["work"]); status != "error" || len(errs) != 1 {
			t.Errorf("failing subject = (%s, %v), want (error, exactly one subject-borne error)", status, errs)
		}
	})
}

// TestGitignoreAllWiring drives the real enumerateGitignoreAll: a failure stops the enumeration,
// its config carries it, and the configs after it never appear.
func TestGitignoreAllWiring(t *testing.T) {
	origJSON := flagJSON
	defer func() { flagJSON = origJSON }()
	flagJSON = true

	t.Run("every selected config becomes its own SubjectResult", func(t *testing.T) {
		run, buf := newGitignoreTestRun()
		shared := ".layat-out/shared"
		err := enumerateGitignoreAll(run, []string{"docs", "web"}, func(name string) ([]string, error) {
			if name == "docs" {
				return []string{".claude/skills/nix", shared}, nil
			}
			return []string{shared}, nil
		})
		if err != nil {
			t.Fatalf("enumerateGitignoreAll: %v", err)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		checkConformance(t, buf)

		byName, order := subjectResults(t, buf)
		if want := []string{"docs", "web"}; !slices.Equal(order, want) {
			t.Fatalf("results order = %v, want %v", order, want)
		}
		// The shared path stays under both configs: --json does not dedup across them.
		if got := infoArray(t, byName["docs"], "paths"); len(got) != 2 {
			t.Errorf("docs paths = %v, want both of its own targets", got)
		}
		if got := infoArray(t, byName["web"], "paths"); len(got) != 1 || got[0] != gitignoreAnchor(shared) {
			t.Errorf("web paths = %v, want the shared path kept here too", got)
		}
	})

	t.Run("a mid-enumeration failure lands on its own subject", func(t *testing.T) {
		run, buf := newGitignoreTestRun()
		failure := errors.New("layat: nix build failed")
		err := enumerateGitignoreAll(run, []string{"docs", "broken", "later"}, func(name string) ([]string, error) {
			if name == "broken" {
				return nil, failure
			}
			return []string{".layat-out/" + name}, nil
		})
		if err == nil {
			t.Fatal("enumerateGitignoreAll must propagate the config's failure")
		}
		if err := run.emit(err); err != nil {
			t.Fatalf("emit: %v", err)
		}
		checkConformance(t, buf)

		doc := decodeEnvelope(t, buf)
		if doc["status"] != "error" {
			t.Errorf("aggregate status = %v, want error", doc["status"])
		}
		if topErrs, ok := doc["errors"]; ok {
			t.Errorf("top-level errors = %v, want absent (the failure belongs to subject broken)", topErrs)
		}
		byName, order := subjectResults(t, buf)
		if want := []string{"docs", "broken"}; !slices.Equal(order, want) {
			t.Fatalf("results = %v, want %v — the enumeration stops and later is never reached", order, want)
		}
		if status, errs := statusAndErrors(t, byName["docs"]); status != "success" || len(errs) != 0 {
			t.Errorf("already-listed subject = (%s, %v), want (success, none)", status, errs)
		}
		if status, errs := statusAndErrors(t, byName["broken"]); status != "error" || len(errs) != 1 {
			t.Errorf("failing subject = (%s, %v), want (error, exactly one subject-borne error)", status, errs)
		}
	})
}

// infoArray reads an array out of one SubjectResult's result.info, failing if the key is missing
// or not an array.
func infoArray(t *testing.T, sr map[string]any, key string) []any {
	t.Helper()
	info, ok := sr["result"].(map[string]any)["info"].(map[string]any)
	if !ok {
		t.Fatalf("result.info = %v, want an object holding %s", sr["result"], key)
	}
	arr, ok := info[key].([]any)
	if !ok {
		t.Fatalf("info.%s = %v, want the key present as an array", key, info[key])
	}
	return arr
}
