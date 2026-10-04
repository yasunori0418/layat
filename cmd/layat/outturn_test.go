package main

// Conformance tests for the --json outturn envelope: every emitted document passes outturn's
// conformance checker, and item ids match outturn's id-vectors byte-for-byte.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yasunori0418/outturn/go"
	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/lock"
)

// fixedClock returns a clock that yields t0 and advances one second per call, so
// startedAt / finishedAt are deterministic yet distinct.
func fixedClock(t0 time.Time) func() time.Time {
	n := 0
	return func() time.Time {
		t := t0.Add(time.Duration(n) * time.Second)
		n++
		return t
	}
}

// newTestRun returns an outturnRun with a pinned clock and buffer sink, already begun for command.
// Prefer the per-command wrappers below, which take the production run aliases' type pairs.
func newTestRun[TInfo, TEnvInfo any](command string) (*outturnRun[TInfo, TEnvInfo], *bytes.Buffer) {
	var buf bytes.Buffer
	r := &outturnRun[TInfo, TEnvInfo]{
		now: fixedClock(time.Date(2026, 7, 19, 12, 0, 0, 0, time.FixedZone("JST", 9*3600))),
		out: &buf,
	}
	r.begin(command)
	return r, &buf
}

// The per-command test runs, each typed by its production alias so a changed info pair breaks
// compilation. Command strings are literals because the envelope's command field is part of the
// output contract; TestBeginRunPublishesEveryCommand checks every subcommand has a case.
func newApplyTestRun() (*applyRun, *bytes.Buffer) {
	return newTestRun[*applyResultInfo, *applyEnvInfo]("apply")
}
func newResetTestRun() (*resetRun, *bytes.Buffer) {
	return newTestRun[*resetResultInfo, *resetEnvInfo]("reset")
}
func newRollbackTestRun() (*rollbackRun, *bytes.Buffer) {
	return newTestRun[*rollbackResultInfo, *rollbackEnvInfo]("rollback")
}
func newListGenerationsTestRun() (*listGenerationsRun, *bytes.Buffer) {
	return newTestRun[*generationsInfo, *struct{}]("list-generations")
}
func newGitignoreTestRun() (*gitignoreRun, *bytes.Buffer) {
	return newTestRun[*gitignoreInfo, *struct{}]("gitignore")
}
func newInitTestRun() (*initRun, *bytes.Buffer) {
	return newTestRun[*struct{}, *initInfo]("init")
}
func newPruneTestRun() (*pruneRun, *bytes.Buffer) {
	return newTestRun[*struct{}, *pruneInfo]("prune")
}

// decodeEnvelope asserts buf holds exactly one JSON document with a trailing newline and
// returns it decoded (UseNumber, so nothing degrades to float64).
func decodeEnvelope(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	s := buf.String()
	if !strings.HasSuffix(s, "\n") {
		t.Fatalf("envelope must end with a newline, got %q", s)
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var extra any
	if err := dec.Decode(&extra); err == nil {
		t.Fatalf("stdout must hold exactly one JSON document, found a second: %v", extra)
	}
	return doc
}

// TestOutturnEnvelopeConformance emits success with / without a subject, pre-subject and
// subject-borne failures, and a dryrun conflict, and checks each against outturn's checker.
func TestOutturnEnvelopeConformance(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}

	cases := []struct {
		name        string
		subject     string // "" = no subject registered
		cmdErr      error
		wantStatus  string
		wantResults int
		wantTopErrs bool // error attached at the top level (vs the subject)
		wantCode    string
	}{
		{name: "success with subject", subject: "default", wantStatus: "success", wantResults: 1},
		{name: "success without subject", wantStatus: "success", wantResults: 0},
		{name: "pre-subject failure", cmdErr: errors.New("layat: no entrypoint found"),
			wantStatus: "error", wantResults: 0, wantTopErrs: true, wantCode: "E_LAYAT_FAILED"},
		{name: "subject-borne failure", subject: "web", cmdErr: errors.New("layat: build failed"),
			wantStatus: "error", wantResults: 1, wantCode: "E_LAYAT_FAILED"},
		{name: "lock failure", subject: "web", cmdErr: lock.ErrLocked,
			wantStatus: "error", wantResults: 1, wantCode: "E_LOCK"},
		{name: "dryrun conflict", subject: "default", cmdErr: &exitError{code: 2},
			wantStatus: "error", wantResults: 1, wantCode: "E_LAYAT_COLLISION"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, buf := newApplyTestRun()
			if c.subject != "" {
				r.beginSubject(c.subject)
			}
			if err := r.emit(c.cmdErr); err != nil {
				t.Fatalf("emit: %v", err)
			}

			if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
				t.Fatalf("conformance findings:\n%s\ndocument: %s", strings.Join(findings, "\n"), buf.String())
			}

			doc := decodeEnvelope(t, buf)
			if doc["status"] != c.wantStatus {
				t.Errorf("status = %v, want %v", doc["status"], c.wantStatus)
			}
			results := doc["results"].([]any)
			if len(results) != c.wantResults {
				t.Fatalf("results length = %d, want %d", len(results), c.wantResults)
			}
			tool := doc["tool"].(map[string]any)
			if tool["name"] != "layat" || tool["version"] != version {
				t.Errorf("tool = %v, want name=layat version=%s (main.version)", tool, version)
			}
			if doc["command"] != "apply" {
				t.Errorf("command = %v, want apply", doc["command"])
			}

			if c.subject != "" {
				sr := results[0].(map[string]any)
				if sr["subject"].(map[string]any)["name"] != c.subject {
					t.Errorf("subject = %v, want %s", sr["subject"], c.subject)
				}
				if sr["status"] != c.wantStatus {
					t.Errorf("subject status = %v, want %v", sr["status"], c.wantStatus)
				}
				// A SubjectResult without a payload carries an empty items array.
				items := sr["result"].(map[string]any)["items"].([]any)
				if len(items) != 0 {
					t.Errorf("items = %v, want empty in the minimal envelope", items)
				}
			}

			if c.cmdErr != nil {
				var errList []any
				if c.wantTopErrs {
					errList, _ = doc["errors"].([]any)
				} else {
					errList, _ = results[0].(map[string]any)["errors"].([]any)
				}
				if len(errList) != 1 {
					t.Fatalf("error layer mismatch (wantTopErrs=%v): %v", c.wantTopErrs, doc)
				}
				e := errList[0].(map[string]any)
				if e["code"] != c.wantCode {
					t.Errorf("error code = %v, want %s", e["code"], c.wantCode)
				}
				if e["message"] == "" {
					t.Error("error message must not be empty")
				}
			}
		})
	}
}

// TestOutturnEnvelopeDryRunFlag pins the envelope's dryRun field to the run's captured --dryrun
// value (begin snapshots flagDryrun; tests set the field directly).
func TestOutturnEnvelopeDryRunFlag(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}
	for _, dryRun := range []bool{false, true} {
		r, buf := newApplyTestRun()
		r.dryRun = dryRun
		r.beginSubject("default")
		if err := r.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
			t.Fatalf("conformance findings: %v", findings)
		}
		doc := decodeEnvelope(t, buf)
		if doc["dryRun"] != dryRun {
			t.Errorf("dryRun = %v, want %v", doc["dryRun"], dryRun)
		}
	}
}

// failingWriter simulates a broken stdout (EPIPE etc.) for the emit write-failure path.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

// TestOutturnEmitWriteFailure: a failed envelope write is an error from emit, so main exits
// non-zero. applyRun stands in for any command.
func TestOutturnEmitWriteFailure(t *testing.T) {
	r := &applyRun{now: fixedClock(time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)), out: failingWriter{}}
	r.begin("apply")
	r.beginSubject("default")
	if err := r.emit(nil); err == nil {
		t.Fatal("emit must propagate the writer failure")
	}
}

// TestOutturnTimestampOffset pins the timestamp shape: RFC 3339, "T" separator, explicit offset
// (local zone renders its UTC offset, UTC renders "Z").
func TestOutturnTimestampOffset(t *testing.T) {
	jst := time.Date(2026, 7, 19, 12, 0, 0, 0, time.FixedZone("JST", 9*3600))
	if got, want := outturnTimestamp(jst), "2026-07-19T12:00:00+09:00"; got != want {
		t.Errorf("outturnTimestamp(JST) = %q, want %q", got, want)
	}
	utc := time.Date(2026, 7, 19, 3, 0, 0, 0, time.UTC)
	if got, want := outturnTimestamp(utc), "2026-07-19T03:00:00Z"; got != want {
		t.Errorf("outturnTimestamp(UTC) = %q, want %q", got, want)
	}
}

// TestEntryItemIDMatchesVectors: every outturn id-vector reproduces through outturn.DeriveID, and
// the entry-kind vectors also through entryItemID (kind="entry", key={target}).
func TestEntryItemIDMatchesVectors(t *testing.T) {
	var doc struct {
		Vectors []struct {
			Identity struct {
				Kind string          `json:"kind"`
				Key  json.RawMessage `json:"key"`
			} `json:"identity"`
			Expected string `json:"expected"`
		} `json:"vectors"`
	}
	dec := json.NewDecoder(bytes.NewReader(outturn.IDVectorsV1()))
	dec.UseNumber()
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode id-vectors: %v", err)
	}
	if len(doc.Vectors) == 0 {
		t.Fatal("id-vectors has no vectors")
	}

	entryVectors := 0
	for i, v := range doc.Vectors {
		var key any
		kd := json.NewDecoder(bytes.NewReader(v.Identity.Key))
		kd.UseNumber()
		if err := kd.Decode(&key); err != nil {
			t.Fatalf("vector %d: decode key: %v", i, err)
		}
		got, err := outturn.DeriveID(outturn.Identity{Kind: v.Identity.Kind, Key: key})
		if err != nil {
			t.Errorf("vector %d (%s): DeriveID: %v", i, v.Identity.Kind, err)
			continue
		}
		if got != v.Expected {
			t.Errorf("vector %d (%s): id = %s, want %s", i, v.Identity.Kind, got, v.Expected)
		}

		// The entry-kind, single-target-key vectors must also reproduce through layat's seam.
		if m, ok := key.(map[string]any); ok && v.Identity.Kind == "entry" && len(m) == 1 {
			if target, ok := m["target"].(string); ok {
				entryVectors++
				got, err := entryItemID(target)
				if err != nil {
					t.Errorf("entryItemID(%q): %v", target, err)
				} else if got != v.Expected {
					t.Errorf("entryItemID(%q) = %s, want %s", target, got, v.Expected)
				}
			}
		}
	}
	if entryVectors == 0 {
		t.Error("no entry-kind vectors exercised layat's entryItemID seam")
	}
}

// TestClassifyErrorCodes pins classifyError per code: generator roots / build failures (also
// re-wrapped) → E_LAYAT_BUILD, discover / prebuilt failures keep their cause's code, inputError →
// E_INPUT, fs sentinels beat E_IO, plus the residual-I/O and fallback arms.
func TestClassifyErrorCodes(t *testing.T) {
	notExist := &fs.PathError{Op: "stat", Path: "/x", Err: fs.ErrNotExist}
	genErr := func(name string, stage generator.Stage, cause error) error {
		return generator.NewError(name, stage, generator.KindFailed, "layat: generator failed", "", "", cause)
	}
	// The input rejections come from their real sources: the resolver and apply's --manifest checks.
	resolveErr := func(flag, env string) error {
		_, err := resolveGenerator(flag, env, "", "")
		return err
	}
	origManifest, origFile, origAll := flagManifest, flagFile, flagApplyAll
	t.Cleanup(func() { flagManifest, flagFile, flagApplyAll = origManifest, origFile, origAll })
	flagManifest, flagFile, flagApplyAll = "/nonexistent/link-farm", "/nonexistent/flake", false
	run, _ := newApplyTestRun()
	manifestWithFile := runApply(run, "default")
	flagFile, flagApplyAll = "", true
	run, _ = newApplyTestRun()
	manifestWithAll := runApplyAll(run)
	flagManifest, flagFile, flagApplyAll = origManifest, origFile, origAll
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"generator roots failure", genErr(generator.NameNix, generator.StageRoots, nil), "E_LAYAT_BUILD"},
		{"generator build failure", genErr(generator.NameNix, generator.StageBuild, nil), "E_LAYAT_BUILD"},
		{"build failure beats a not-found cause", genErr(generator.NameNix, generator.StageBuild, notExist), "E_LAYAT_BUILD"},
		{"marker survives a %w rewrap",
			fmt.Errorf("context: %w", genErr(generator.NameNix, generator.StageRoots, nil)), "E_LAYAT_BUILD"},
		{"discover failure keeps its not-found cause", genErr(generator.NameNix, generator.StageDiscover, notExist), "E_NOTFOUND"},
		{"discover failure without a cause falls back", genErr(generator.NameNix, generator.StageDiscover, nil), "E_LAYAT_FAILED"},
		{"prebuilt failure keeps its not-found cause", genErr(generator.NamePrebuilt, generator.StageRoots, notExist), "E_NOTFOUND"},
		{"input rejection", &inputError{err: errors.New("layat: --manifest cannot be combined with -f")}, "E_INPUT"},
		{"input marker beats its not-found cause", &inputError{err: notExist}, "E_INPUT"},
		{"unknown --generator value", resolveErr("bogus", ""), "E_INPUT"},
		{"unknown LAYAT_GENERATOR value", resolveErr("", "bogus"), "E_INPUT"},
		{"--manifest with -f", manifestWithFile, "E_INPUT"},
		{"--manifest with --all", manifestWithAll, "E_INPUT"},
		{"input marker survives a %w rewrap",
			fmt.Errorf("context: %w", &inputError{err: errors.New("layat: bad flags")}), "E_INPUT"},
		{"lock sentinel", lock.ErrLocked, "E_LOCK"},
		{"not-found beats the IO shape", &fs.PathError{Op: "stat", Path: "/x", Err: fs.ErrNotExist}, "E_NOTFOUND"},
		{"permission beats the IO shape", &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrPermission}, "E_PERMISSION"},
		{"residual IO PathError", &fs.PathError{Op: "rmdir", Path: "/x", Err: syscall.ENOTEMPTY}, "E_IO"},
		{"residual IO LinkError", &os.LinkError{Op: "symlink", Old: "/a", New: "/b", Err: syscall.EEXIST}, "E_IO"},
		{"unclassified fallback", errors.New("layat: generation commit (nix-env --set) failed"), "E_LAYAT_FAILED"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := classifyError(c.err)
			if e.Code != c.want {
				t.Errorf("classifyError(%v).Code = %s, want %s", c.err, e.Code, c.want)
			}
			if e.Message == "" {
				t.Error("message must not be empty")
			}
		})
	}
}

// TestJSONSuppressesLineOrientedStdout: every line-oriented printer emits nothing under --json and
// everything under the default contract.
func TestJSONSuppressesLineOrientedStdout(t *testing.T) {
	origJSON := flagJSON
	defer func() { flagJSON = origJSON }()

	applyRes := &engine.Result{Placed: []string{"a"}, Removed: []string{"b"}}
	resetRes := &engine.ResetResult{RemovedSymlinks: []string{"s"}, RemovedCopies: []string{"c"}, KeptForeign: []string{"k"}}
	gens := []engine.Generation{{Number: 1, Date: "2026-07-19", Current: true}}
	targets := []string{".claude/skills"}
	pruneRes := &engine.PruneResult{Removed: []engine.PruneSeries{{RootHash: "aaaa", Root: "/gone"}}}

	printers := []struct {
		name  string
		print func()
	}{
		{"printApplyPlan", func() { printApplyPlan(applyRes) }},
		{"printResetPlan", func() { printResetPlan(resetRes) }},
		{"printGenerations", func() { printGenerations(gens) }},
		{"printGitignore", func() { printGitignore(targets) }},
		{"printPrunePlan", func() { printPrunePlan(pruneRes) }},
	}
	for _, p := range printers {
		t.Run(p.name, func(t *testing.T) {
			flagJSON = false
			if out := captureStdout(t, p.print); out == "" {
				t.Errorf("%s must print under the default contract", p.name)
			}
			flagJSON = true
			if out := captureStdout(t, p.print); out != "" {
				t.Errorf("%s must print nothing under --json, got %q", p.name, out)
			}
		})
	}
}

// TestJSONEmptyResetPlanKeepsStderrNotice: with nothing to remove, the stderr notice is printed
// under both contracts while stdout stays empty.
func TestJSONEmptyResetPlanKeepsStderrNotice(t *testing.T) {
	origJSON := flagJSON
	defer func() { flagJSON = origJSON }()

	empty := &engine.ResetResult{}
	for _, jsonMode := range []bool{false, true} {
		flagJSON = jsonMode
		out, errOut := captureOutErr(t, func() { printResetPlan(empty) })
		if out != "" {
			t.Errorf("flagJSON=%v: stdout = %q, want empty for an empty plan", jsonMode, out)
		}
		if !strings.Contains(errOut, "nothing to remove") {
			t.Errorf("flagJSON=%v: stderr = %q, want the nothing-to-remove notice", jsonMode, errOut)
		}
	}
}

// TestResetPromptAllowed: --json forbids prompting even on a TTY, so reset --json without --yes
// is refused.
func TestResetPromptAllowed(t *testing.T) {
	cases := []struct{ interactive, jsonMode, want bool }{
		{true, false, true},   // TTY, default contract → prompting allowed
		{true, true, false},   // TTY + --json → machine consumption never prompts
		{false, false, false}, // non-TTY → refuse path
		{false, true, false},
	}
	for _, c := range cases {
		if got := resetPromptAllowed(c.interactive, c.jsonMode); got != c.want {
			t.Errorf("resetPromptAllowed(%v, %v) = %v, want %v", c.interactive, c.jsonMode, got, c.want)
		}
	}
	// The composed contract: --json without --yes refuses; --json with --yes runs promptless.
	if _, err := confirmPolicy(false, resetPromptAllowed(true, true), "reset"); err == nil {
		t.Error("reset --json without --yes must refuse (fail fast)")
	}
	if needPrompt, err := confirmPolicy(true, resetPromptAllowed(true, true), "reset"); err != nil || needPrompt {
		t.Errorf("reset --json --yes: needPrompt=%v err=%v, want promptless success", needPrompt, err)
	}
}

// TestBeginRunPublishesEveryCommand: every begin<Cmd>Run begins its run and publishes it to
// outturnReport, which is what main emits.
func TestBeginRunPublishesEveryCommand(t *testing.T) {
	origReport := outturnReport
	defer func() { outturnReport = origReport }()

	// Each entry begins the command's run exactly as its RunE does.
	begins := map[string]func(string) emitter{
		"apply":            func(c string) emitter { return beginApplyRun(c) },
		"reset":            func(c string) emitter { return beginResetRun(c) },
		"rollback":         func(c string) emitter { return beginRollbackRun(c) },
		"list-generations": func(c string) emitter { return beginListGenerationsRun(c) },
		"gitignore":        func(c string) emitter { return beginGitignoreRun(c) },
		"init":             func(c string) emitter { return beginInitRun(c) },
		"prune":            func(c string) emitter { return beginPruneRun(c) },
	}
	for command, begin := range begins {
		t.Run(command, func(t *testing.T) {
			outturnReport = noopEmitter{}
			run := begin(command)
			if !run.began() {
				t.Errorf("%s: the returned run is not begun", command)
			}
			if outturnReport != emitter(run) {
				t.Fatalf("%s: outturnReport = %#v, want the run just begun (main emits what it finds here)", command, outturnReport)
			}
			if !outturnReport.began() {
				t.Errorf("%s: the published run is not begun; main's gate would skip the envelope", command)
			}
		})
	}

	// Every registered subcommand must be covered above.
	for _, cmd := range newRootCmd().Commands() {
		if cmd.RunE == nil {
			continue // cobra's own utility commands (help / completion) never begin a run
		}
		if _, ok := begins[cmd.Name()]; !ok {
			t.Errorf("subcommand %q has a RunE but no begin<Cmd>Run case here", cmd.Name())
		}
	}
}

// TestJSONFlagRegistered pins --json as a root persistent flag accepted by every subcommand.
func TestJSONFlagRegistered(t *testing.T) {
	root := newRootCmd()
	f := root.PersistentFlags().Lookup("json")
	if f == nil {
		t.Fatal("--json is not registered as a persistent flag")
	}
	if f.Shorthand != "" {
		t.Errorf("--json shorthand = %q, want none", f.Shorthand)
	}
}

// TestJSONUtilityCommandsDoNotBegin: cobra's help / completion never begin an outturn run, since
// they own stdout with their own text.
func TestJSONUtilityCommandsDoNotBegin(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"completion", "bash"}} {
		origReport := outturnReport
		// Only a RunE of ours replaces the noop report, so a noop after Execute means nothing was emitted.
		outturnReport = noopEmitter{}
		root := newRootCmd()
		root.SetArgs(args)
		out := captureStdout(t, func() {
			if err := root.Execute(); err != nil {
				t.Errorf("%v: Execute: %v", args, err)
			}
		})
		if outturnReport.began() {
			t.Errorf("%v began an outturn run; utility commands must not emit an envelope", args)
		}
		// Also assert the observable output, not just the began() gate.
		if strings.Contains(out, `"specVersion"`) {
			t.Errorf("%v wrote an envelope to stdout; utility commands own it with their own text: %q", args, out)
		}
		outturnReport = origReport
	}
}

// TestJSONEndToEndSubjectBorneFailure runs apply in an entrypoint-less directory in-process: the
// envelope is conformant, status error, with the error on results[0] and nothing else on stdout.
func TestJSONEndToEndSubjectBorneFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	origReport := outturnReport
	origJSON, origDryrun := flagJSON, flagDryrun
	defer func() {
		outturnReport = origReport
		flagJSON, flagDryrun = origJSON, origDryrun
	}()

	outturnReport = noopEmitter{}

	// RunE emits to os.Stdout, so Execute and the main-style emit both run inside the capture.
	var execErr error
	out := captureStdout(t, func() {
		root := newRootCmd()
		root.SetArgs([]string{"apply", "--json"})
		execErr = root.Execute()
		if execErr == nil {
			return
		}
		if !outturnReport.began() {
			return
		}
		if err := outturnReport.emit(execErr); err != nil {
			t.Errorf("emit: %v", err)
		}
	})
	if execErr == nil {
		t.Fatal("apply in an entrypoint-less directory must fail")
	}
	if !outturnReport.began() {
		t.Fatal("apply's RunE did not publish a begun outturn run")
	}
	// Check for stray output first, so it is not reported as a JSON decode error.
	if !strings.HasPrefix(out, "{") {
		t.Fatalf("stdout must hold the envelope alone (the --json contract), got %q", out)
	}
	buf := bytes.NewBufferString(out)

	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings:\n%s\ndocument: %s", strings.Join(findings, "\n"), buf.String())
	}
	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" || doc["command"] != "apply" {
		t.Errorf("status/command = %v/%v, want error/apply", doc["status"], doc["command"])
	}
	results := doc["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results = %v, want the registered subject (default)", results)
	}
	sr := results[0].(map[string]any)
	if sr["subject"].(map[string]any)["name"] != "default" {
		t.Errorf("subject = %v, want default", sr["subject"])
	}
	if errList, _ := sr["errors"].([]any); len(errList) != 1 {
		t.Errorf("subject errors = %v, want exactly one (subject-borne)", sr["errors"])
	}
	if topErrs, ok := doc["errors"]; ok {
		t.Errorf("top-level errors present = %v, want the failure attached to the subject", topErrs)
	}
	// No payload exists, so both info slots stay nil and are omitted.
	assertNoInfoKeys(t, doc)
}
