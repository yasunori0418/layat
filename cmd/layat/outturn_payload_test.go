package main

// Tests for the mutation payloads (apply / reset / rollback): the engine-result → outturn mapping,
// the partial-failure partition, the error-layer placement, and conformance of every shape.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/yasunori0418/outturn/go"
	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/planner"
)

func ip(n int) *int { return &n }

// mustItemID resolves the entry item id or fails the test.
func mustItemID(t *testing.T, target string) string {
	t.Helper()
	id, err := entryItemID(target)
	if err != nil {
		t.Fatalf("entryItemID(%q): %v", target, err)
	}
	return id
}

// findItem returns the item whose info.target matches, failing when absent.
func findItem(t *testing.T, items []layatItem, target string) layatItem {
	t.Helper()
	for _, it := range items {
		if it.Info != nil && it.Info.Target == target {
			return it
		}
	}
	t.Fatalf("no item for target %q in %+v", target, items)
	return layatItem{}
}

// changesFor returns every change whose itemId belongs to target.
func changesFor(t *testing.T, changes []layatChange, target string) []layatChange {
	t.Helper()
	id := mustItemID(t, target)
	var out []layatChange
	for _, c := range changes {
		if c.ItemID == id {
			out = append(out, c)
		}
	}
	return out
}

// emitPayloadDoc runs the payload through the real emit path, checks conformance, and returns the
// decoded document. newRun is the command's test-run constructor (newApplyTestRun, ...).
func emitPayloadDoc[TInfo, TEnvInfo any](t *testing.T, newRun func() (*outturnRun[TInfo, TEnvInfo], *bytes.Buffer), p *outturnPayload[TInfo], cmdErr error) map[string]any {
	t.Helper()
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}
	r, buf := newRun()
	r.beginSubject("default").setPayload(p)
	if err := r.emit(cmdErr); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings:\n%s\ndocument: %s", strings.Join(findings, "\n"), buf.String())
	}
	return decodeEnvelope(t, buf)
}

// assertNoInfoKeys fails when the document carries an info key at the envelope level or inside any
// result. "info":{} is schema-valid, so the conformance checker cannot catch it.
func assertNoInfoKeys(t *testing.T, doc map[string]any) {
	t.Helper()
	if v, ok := doc["info"]; ok {
		t.Errorf("envelope info = %v, want the key absent (the seat type is a nil pointer)", v)
	}
	results, _ := doc["results"].([]any)
	for i, r := range results {
		res, _ := r.(map[string]any)["result"].(map[string]any)
		if v, ok := res["info"]; ok {
			t.Errorf("results[%d].result.info = %v, want the key absent (the seat type is a nil pointer)", i, v)
		}
	}
}

// TestMutationSeatInfoKeysStayAbsent: apply / reset / rollback emit no info key at either level,
// with a payload and on the payload-less failure path.
func TestMutationSeatInfoKeysStayAbsent(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{{SrcKind: "store", Src: "/nix/store/a", Target: "a", Method: "symlink"}},
		Placed:  []string{"a"},
	}
	resetRes := &engine.ResetResult{
		Entries:         []manifest.Entry{{SrcKind: "store", Src: "/nix/store/s", Target: "s", Method: "symlink"}},
		RemovedSymlinks: []string{"s"},
	}

	t.Run("apply with payload", func(t *testing.T) {
		p, err := mutationPayload[*applyResultInfo](res, nil)
		if err != nil {
			t.Fatalf("mutationPayload: %v", err)
		}
		assertNoInfoKeys(t, emitPayloadDoc(t, newApplyTestRun, p, nil))
	})
	t.Run("rollback with payload", func(t *testing.T) {
		p, err := mutationPayload[*rollbackResultInfo](res, nil)
		if err != nil {
			t.Fatalf("mutationPayload: %v", err)
		}
		assertNoInfoKeys(t, emitPayloadDoc(t, newRollbackTestRun, p, nil))
	})
	t.Run("reset with payload", func(t *testing.T) {
		p, err := resetPayload[*resetResultInfo](resetRes, nil)
		if err != nil {
			t.Fatalf("resetPayload: %v", err)
		}
		assertNoInfoKeys(t, emitPayloadDoc(t, newResetTestRun, p, nil))
	})
	// Payload-less path: the command failed before any engine result, so Result.Info stays zero.
	t.Run("apply without payload", func(t *testing.T) {
		r, buf := newApplyTestRun()
		r.beginSubject("default")
		if err := r.emit(errors.New("layat: no entrypoint found")); err != nil {
			t.Fatalf("emit: %v", err)
		}
		assertNoInfoKeys(t, decodeEnvelope(t, buf))
	})
}

// subjectResultOf digs results[0] out of a decoded envelope.
func subjectResultOf(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	results := doc["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}
	return results[0].(map[string]any)
}

// TestMutationPayloadFullInventory pins the success mapping: new and stale-removed entries as
// items, diff-only changes, warnings on the warned item, and the generation (before omitted).
func TestMutationPayloadFullInventory(t *testing.T) {
	res := &engine.Result{
		Profile: "/state/layat/default/profile",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/aaa", Subpath: "conf", Target: ".config/tool", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/bbb", Target: ".config/relinked", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/ccc", Target: ".local/copy", Method: "copy"},
			{SrcKind: "store", Src: "/nix/store/ddd", Target: ".local/recopied", Method: "copy"},
			{SrcKind: "store", Src: "/nix/store/eee", Target: ".config/noop", Method: "symlink"},
		},
		RemovalEntries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/old", Subpath: "sub", Target: ".config/old", Method: "symlink"},
		},
		Placed:        []string{".config/tool"},
		Replaced:      []string{".config/relinked"},
		ReplacedDests: map[string]string{".config/relinked": "/nix/store/prev-bbb"},
		Copied:        []string{".local/copy"},
		Recopied:      []string{".local/recopied"},
		Removed:       []string{".config/old"},
		GenAfter:      ip(1),
		Warnings:      []planner.Warning{{Kind: planner.WarnForeignReplace, Target: ".config/relinked"}},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}

	if len(p.items) != 6 {
		t.Fatalf("items = %d, want 6 (5 new entries + 1 stale-removed old entry)", len(p.items))
	}
	for _, it := range p.items {
		if it.Status != outturn.ItemSuccess {
			t.Errorf("item %s status = %s, want success", it.Info.Target, it.Status)
		}
	}
	old := findItem(t, p.items, ".config/old")
	if old.Info.Method != "symlink" || old.Info.Subpath != "sub" || old.Label != ".config/old" {
		t.Errorf("stale-removed old entry item = %+v, want method/subpath from the previous generation", old.Info)
	}

	if len(p.changes) != 5 {
		t.Fatalf("changes = %d, want 5 (the no-op entry has none)", len(p.changes))
	}
	assertChange := func(target string, kind outturn.ChangeKind, reversible bool, old, new string) {
		t.Helper()
		cs := changesFor(t, p.changes, target)
		if len(cs) != 1 {
			t.Fatalf("changes for %s = %d, want 1", target, len(cs))
		}
		c := cs[0]
		if c.Kind != kind || c.Reversible != reversible {
			t.Errorf("%s change = kind %s reversible %v, want %s/%v", target, c.Kind, c.Reversible, kind, reversible)
		}
		gotOld, gotNew := "", ""
		if c.Info != nil {
			gotOld, gotNew = c.Info.Old, c.Info.New
		}
		if gotOld != old || gotNew != new {
			t.Errorf("%s change info = {old:%q new:%q}, want {old:%q new:%q}", target, gotOld, gotNew, old, new)
		}
	}
	assertChange(".config/tool", outturn.ChangeAdd, true, "", "/nix/store/aaa/conf")
	assertChange(".config/relinked", outturn.ChangeModify, true, "/nix/store/prev-bbb", "/nix/store/bbb")
	assertChange(".local/copy", outturn.ChangeAdd, true, "", "/nix/store/ccc")
	assertChange(".local/recopied", outturn.ChangeModify, false, "", "/nix/store/ddd")
	assertChange(".config/old", outturn.ChangeRemove, true, "/nix/store/old/sub", "")

	relinked := findItem(t, p.items, ".config/relinked")
	if len(relinked.Warnings) != 1 || relinked.Warnings[0].Code != "W_LAYAT_FOREIGN_SYMLINK" {
		t.Errorf("relinked item warnings = %+v, want one W_LAYAT_FOREIGN_SYMLINK", relinked.Warnings)
	}
	if len(p.warnings) != 0 {
		t.Errorf("subject warnings = %+v, want none (the warned target is an item)", p.warnings)
	}

	doc := emitPayloadDoc(t, newApplyTestRun, p, nil)
	sr := subjectResultOf(t, doc)
	gen, ok := sr["generation"].(map[string]any)
	if !ok {
		t.Fatalf("generation missing in %v", sr)
	}
	if _, hasBefore := gen["before"]; hasBefore {
		t.Errorf("generation.before = %v, want omitted on the first apply", gen["before"])
	}
	if gen["after"] != json.Number("1") || gen["profile"] != res.Profile {
		t.Errorf("generation = %v, want after=1 profile=%s", gen, res.Profile)
	}
}

// TestMutationPayloadMethodChangeCoalesces: a symlink→copy method change yields one item and one
// modify change from the old symlink dest to the new copy source.
func TestMutationPayloadMethodChangeCoalesces(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/new", Target: ".config/x", Method: "copy"},
		},
		RemovalEntries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/old", Target: ".config/x", Method: "symlink"},
		},
		Removed: []string{".config/x"},
		Copied:  []string{".config/x"},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if len(p.items) != 1 {
		t.Fatalf("items = %d, want 1 (the new entry shadows the removed old one)", len(p.items))
	}
	if got := findItem(t, p.items, ".config/x").Info.Method; got != "copy" {
		t.Errorf("item method = %s, want the new entry's copy", got)
	}
	if len(p.changes) != 1 {
		t.Fatalf("changes = %+v, want the coalesced single modify", p.changes)
	}
	c := p.changes[0]
	if c.Kind != outturn.ChangeModify || !c.Reversible || c.Info.Old != "/nix/store/old" || c.Info.New != "/nix/store/new" {
		t.Errorf("change = %+v info %+v, want reversible modify old→new", c, c.Info)
	}
}

// TestMutationPayloadNoopRelinkSuppressed: a re-link to the recorded dest produces no change, while
// a re-link to a different dest does.
func TestMutationPayloadNoopRelinkSuppressed(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/same", Target: ".same", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/new", Target: ".moved", Method: "symlink"},
		},
		Replaced: []string{".same", ".moved"},
		ReplacedDests: map[string]string{
			".same":  "/nix/store/same", // re-linked back to the recorded dest = noop
			".moved": "/nix/store/old",  // genuinely moved
		},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if cs := changesFor(t, p.changes, ".same"); len(cs) != 0 {
		t.Errorf(".same changes = %+v, want none (noop re-link must be suppressed)", cs)
	}
	if cs := changesFor(t, p.changes, ".moved"); len(cs) != 1 || cs[0].Kind != outturn.ChangeModify {
		t.Errorf(".moved changes = %+v, want one modify", cs)
	}
	if it := findItem(t, p.items, ".same"); it.Status != outturn.ItemSuccess {
		t.Errorf(".same item status = %s, want success (still in the inventory)", it.Status)
	}
}

// TestMutationPayloadOrphanSubjectWarning: a warning whose target is outside the inventory lands
// in subjectResult.warnings.
func TestMutationPayloadOrphanSubjectWarning(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/a", Target: "a", Method: "symlink"},
		},
		Warnings: []planner.Warning{{Kind: planner.WarnCopyOrphan, Target: ".gone/copy"}},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if len(p.warnings) != 1 || p.warnings[0].Code != "W_LAYAT_COPY_ORPHAN" || p.warnings[0].Detail["target"] != ".gone/copy" {
		t.Fatalf("subject warnings = %+v, want one W_LAYAT_COPY_ORPHAN carrying the target", p.warnings)
	}
	if it := findItem(t, p.items, "a"); len(it.Warnings) != 0 {
		t.Errorf("item warnings = %+v, want none (the orphan is not this item's)", it.Warnings)
	}
	doc := emitPayloadDoc(t, newApplyTestRun, p, nil)
	sr := subjectResultOf(t, doc)
	warns, ok := sr["warnings"].([]any)
	if !ok || len(warns) != 1 {
		t.Fatalf("subjectResult.warnings = %v, want the orphan warning", sr["warnings"])
	}
}

// TestMutationPayloadRollbackGeneration: the From→To transition rides generation.before/after, and
// a failed rollback observes before == after == current.
func TestMutationPayloadRollbackGeneration(t *testing.T) {
	rr := &engine.RollbackResult{
		Result: engine.Result{
			Profile: "/p",
			Entries: []manifest.Entry{
				{SrcKind: "store", Src: "/nix/store/a", Target: "a", Method: "symlink"},
			},
			Replaced:      []string{"a"},
			ReplacedDests: map[string]string{"a": "/nix/store/newer"},
			GenBefore:     ip(5),
			GenAfter:      ip(4),
		},
		From: 5, To: 4,
	}
	p, err := mutationPayload[*rollbackResultInfo](&rr.Result, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	doc := emitPayloadDoc(t, newRollbackTestRun, p, nil)
	sr := subjectResultOf(t, doc)
	gen := sr["generation"].(map[string]any)
	if gen["before"] != json.Number("5") || gen["after"] != json.Number("4") {
		t.Errorf("generation = %v, want the 5→4 rollback transition", gen)
	}

	// A failed rollback pins the pointer at the unmoved current generation.
	rr.GenBefore, rr.GenAfter = ip(5), ip(5)
	p, err = mutationPayload[*rollbackResultInfo](&rr.Result, errors.New("layat: failed to move the profile pointer"))
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if p.generation.Before == nil || *p.generation.Before != 5 || p.generation.After == nil || *p.generation.After != 5 {
		t.Errorf("failed rollback generation = %+v, want the unmoved 5→5 observation", p.generation)
	}
}

// TestOutturnWarningMapping pins every planner WarnKind → W_LAYAT_* code pair, plus the
// defensive fallback for an unknown kind.
func TestOutturnWarningMapping(t *testing.T) {
	cases := []struct {
		kind planner.WarnKind
		code string
	}{
		{planner.WarnForeignReplace, "W_LAYAT_FOREIGN_SYMLINK"},
		{planner.WarnStaleMismatch, "W_LAYAT_STALE_MISMATCH"},
		{planner.WarnStaleNonSymlink, "W_LAYAT_STALE_NON_SYMLINK"},
		{planner.WarnCopyOrphan, "W_LAYAT_COPY_ORPHAN"},
		{planner.WarnCopyForeign, "W_LAYAT_COPY_FOREIGN"},
		{planner.WarnKind(99), "W_LAYAT_WARNING"},
	}
	for _, c := range cases {
		w := outturnWarning(planner.Warning{Kind: c.kind, Target: "t"})
		if w.Code != c.code {
			t.Errorf("kind %v: code = %s, want %s", c.kind, w.Code, c.code)
		}
		if w.Message == "" {
			t.Errorf("kind %v: message must not be empty", c.kind)
		}
		if w.Detail["target"] != "t" {
			t.Errorf("kind %v: detail = %v, want {target: t}", c.kind, w.Detail)
		}
	}
}

// TestMutationPayloadPartialFailure pins the reached-state partition: the failed entry carries the
// error, unreached entries are skipped, completed ones keep their changes, W_LAYAT_UNWOUND is on the
// subject, and the error is not duplicated in subjectResult.errors[].
func TestMutationPayloadPartialFailure(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/a", Target: "a", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/b", Target: "b", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/c", Target: "c", Method: "symlink"},
		},
		Placed:       []string{"a"},
		FailedTarget: "b",
		Unreached:    []string{"c"},
		Unwound:      true,
		GenBefore:    ip(3),
		GenAfter:     ip(3),
	}
	cmdErr := &os.PathError{Op: "symlink", Path: "/root/b", Err: fs.ErrPermission}
	p, err := mutationPayload[*applyResultInfo](res, cmdErr)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if !p.itemBorne {
		t.Error("itemBorne = false, want true for an entry-scoped failure")
	}

	if it := findItem(t, p.items, "a"); it.Status != outturn.ItemSuccess {
		t.Errorf("completed item status = %s, want success", it.Status)
	}
	failed := findItem(t, p.items, "b")
	if failed.Status != outturn.ItemFailed || failed.Error == nil || failed.Error.Code != "E_PERMISSION" {
		t.Errorf("failed item = status %s error %+v, want failed + E_PERMISSION", failed.Status, failed.Error)
	}
	if it := findItem(t, p.items, "c"); it.Status != outturn.ItemSkipped {
		t.Errorf("unreached item status = %s, want skipped", it.Status)
	}
	if cs := changesFor(t, p.changes, "a"); len(cs) != 1 || cs[0].Kind != outturn.ChangeAdd {
		t.Errorf("completed entry changes = %+v, want its add present despite the failure", cs)
	}
	var unwound bool
	for _, w := range p.warnings {
		if w.Code == "W_LAYAT_UNWOUND" {
			unwound = true
		}
	}
	if !unwound {
		t.Errorf("subject warnings = %+v, want W_LAYAT_UNWOUND for the unwound run", p.warnings)
	}

	doc := emitPayloadDoc(t, newApplyTestRun, p, cmdErr)
	if doc["status"] != "error" {
		t.Errorf("status = %v, want error", doc["status"])
	}
	sr := subjectResultOf(t, doc)
	if sr["status"] != "error" {
		t.Errorf("subject status = %v, want error", sr["status"])
	}
	if errList, ok := sr["errors"]; ok {
		t.Errorf("subjectResult.errors = %v, want absent (the failure is item-borne)", errList)
	}
	gen := sr["generation"].(map[string]any)
	if gen["before"] != json.Number("3") || gen["after"] != json.Number("3") {
		t.Errorf("generation = %v, want an unmoved 3→3 observation", gen)
	}
}

// TestMutationPayloadSubjectBorneFailure: a commit / build failure has no failed item, keeps the
// changes, and puts the error in subjectResult.errors[].
func TestMutationPayloadSubjectBorneFailure(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/a", Target: "a", Method: "symlink"},
		},
		Placed:   []string{"a"},
		GenAfter: ip(2), GenBefore: ip(2),
	}
	cmdErr := errors.New("layat: generation commit (nix-env --set) failed: exit status 1")
	p, err := mutationPayload[*applyResultInfo](res, cmdErr)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if p.itemBorne {
		t.Error("itemBorne = true, want false for a commit failure")
	}
	if it := findItem(t, p.items, "a"); it.Status != outturn.ItemSuccess {
		t.Errorf("placed item status = %s, want success (the commit, not the entry, failed)", it.Status)
	}
	doc := emitPayloadDoc(t, newApplyTestRun, p, cmdErr)
	sr := subjectResultOf(t, doc)
	errList, ok := sr["errors"].([]any)
	if !ok || len(errList) != 1 {
		t.Fatalf("subjectResult.errors = %v, want exactly one subject-borne error", sr["errors"])
	}
	if code := errList[0].(map[string]any)["code"]; code != "E_LAYAT_FAILED" {
		t.Errorf("error code = %v, want the generic fallback for a commit failure", code)
	}
	items := sr["result"].(map[string]any)["items"].([]any)
	if len(items) != 1 {
		t.Errorf("items = %v, want the full inventory alongside the subject error", items)
	}
}

// TestMutationPayloadConflicts: each conflicted entry is a failed item with E_LAYAT_COLLISION, the
// rest is skipped, and the command error is not duplicated at the subject.
func TestMutationPayloadConflicts(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/a", Target: "a", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/b", Target: "b", Method: "symlink"},
		},
		Conflicts: []planner.Conflict{
			{Entry: manifest.Entry{Target: "a"}, Reason: "a regular file occupies the symlink target", Kind: planner.ConflictForeignEntity},
		},
		Unreached: []string{"b"},
	}
	cmdErr := errors.New("layat: 1 conflict(s) detected; stopped without placing (see above)")
	p, err := mutationPayload[*applyResultInfo](res, cmdErr)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if !p.itemBorne {
		t.Error("itemBorne = false, want true for a conflict stop")
	}
	conflicted := findItem(t, p.items, "a")
	if conflicted.Status != outturn.ItemFailed || conflicted.Error == nil ||
		conflicted.Error.Code != "E_LAYAT_COLLISION" || conflicted.Error.Message != "a regular file occupies the symlink target" {
		t.Errorf("conflicted item = %+v error %+v, want failed + E_LAYAT_COLLISION with the planner reason", conflicted, conflicted.Error)
	}
	if it := findItem(t, p.items, "b"); it.Status != outturn.ItemSkipped {
		t.Errorf("non-conflicted item status = %s, want skipped (nothing ran)", it.Status)
	}
	if len(p.changes) != 0 {
		t.Errorf("changes = %+v, want none (the run stopped before any FS action)", p.changes)
	}
	doc := emitPayloadDoc(t, newApplyTestRun, p, cmdErr)
	sr := subjectResultOf(t, doc)
	if errList, ok := sr["errors"]; ok {
		t.Errorf("subjectResult.errors = %v, want absent (conflicts are item-borne)", errList)
	}
}

// TestResetPayload pins reset's mapping: selected entries as items, reversible symlink removes with
// the dest, irreversible copy removes without info, a kept-foreign warning on its item, and no
// generation.
func TestResetPayload(t *testing.T) {
	res := &engine.ResetResult{
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/s1", Target: "s1", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/s2", Target: "s2", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/c1", Target: "c1", Method: "copy"},
		},
		RemovedSymlinks: []string{"s1"},
		RemovedCopies:   []string{"c1"},
		KeptForeign:     []string{"s2"},
		Warnings:        []planner.Warning{{Kind: planner.WarnStaleMismatch, Target: "s2"}},
	}
	p, err := resetPayload[*resetResultInfo](res, nil)
	if err != nil {
		t.Fatalf("resetPayload: %v", err)
	}
	if p.generation != nil {
		t.Fatalf("generation = %+v, want none for reset", p.generation)
	}
	for _, target := range []string{"s1", "s2", "c1"} {
		if it := findItem(t, p.items, target); it.Status != outturn.ItemSuccess {
			t.Errorf("item %s status = %s, want success", target, it.Status)
		}
	}
	kept := findItem(t, p.items, "s2")
	if len(kept.Warnings) != 1 || kept.Warnings[0].Code != "W_LAYAT_STALE_MISMATCH" {
		t.Errorf("kept item warnings = %+v, want one W_LAYAT_STALE_MISMATCH", kept.Warnings)
	}
	if cs := changesFor(t, p.changes, "s1"); len(cs) != 1 || cs[0].Kind != outturn.ChangeRemove ||
		!cs[0].Reversible || cs[0].Info == nil || cs[0].Info.Old != "/nix/store/s1" {
		t.Errorf("symlink removal change = %+v, want reversible remove with the recorded dest", cs)
	}
	if cs := changesFor(t, p.changes, "c1"); len(cs) != 1 || cs[0].Kind != outturn.ChangeRemove ||
		cs[0].Reversible || cs[0].Info != nil {
		t.Errorf("copy removal change = %+v, want irreversible remove without info", cs)
	}
	if cs := changesFor(t, p.changes, "s2"); len(cs) != 0 {
		t.Errorf("kept target changes = %+v, want none (policy inaction is not a diff)", cs)
	}

	doc := emitPayloadDoc(t, newResetTestRun, p, nil)
	sr := subjectResultOf(t, doc)
	if gen, ok := sr["generation"]; ok {
		t.Errorf("generation = %v, want the slot absent for reset", gen)
	}
}

// TestResetPayloadPartialFailure: removed-so-far keeps its changes, the failing target carries the
// error, and the rest is skipped.
func TestResetPayloadPartialFailure(t *testing.T) {
	res := &engine.ResetResult{
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/s1", Target: "s1", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/c1", Target: "c1", Method: "copy"},
			{SrcKind: "store", Src: "/nix/store/c2", Target: "c2", Method: "copy"},
		},
		RemovedSymlinks: []string{"s1"},
		FailedTarget:    "c1",
		Unreached:       []string{"c2"},
	}
	cmdErr := &os.PathError{Op: "removeall", Path: "/root/c1", Err: fs.ErrPermission}
	p, err := resetPayload[*resetResultInfo](res, cmdErr)
	if err != nil {
		t.Fatalf("resetPayload: %v", err)
	}
	if !p.itemBorne {
		t.Error("itemBorne = false, want true")
	}
	if it := findItem(t, p.items, "s1"); it.Status != outturn.ItemSuccess {
		t.Errorf("removed item status = %s, want success", it.Status)
	}
	failed := findItem(t, p.items, "c1")
	if failed.Status != outturn.ItemFailed || failed.Error == nil || failed.Error.Code != "E_PERMISSION" {
		t.Errorf("failed item = %s / %+v, want failed + E_PERMISSION", failed.Status, failed.Error)
	}
	if it := findItem(t, p.items, "c2"); it.Status != outturn.ItemSkipped {
		t.Errorf("unreached item status = %s, want skipped", it.Status)
	}
	if cs := changesFor(t, p.changes, "s1"); len(cs) != 1 {
		t.Errorf("removed-so-far changes = %+v, want the remove present despite the failure", cs)
	}
	doc := emitPayloadDoc(t, newResetTestRun, p, cmdErr)
	if doc["status"] != "error" {
		t.Errorf("status = %v, want error", doc["status"])
	}
	sr := subjectResultOf(t, doc)
	if sr["status"] != "error" {
		t.Errorf("subject status = %v, want error", sr["status"])
	}
	if errList, ok := sr["errors"]; ok {
		t.Errorf("subjectResult.errors = %v, want absent (the failure is item-borne)", errList)
	}
}

// --- end-to-end through the real engine on a tmpdir ---

// genCommit fakes nix-env --set with a real generation-link layout, so observeGeneration
// reads a numeric generation the same way it does under nix.
func genCommit(gen int) engine.CommitFunc {
	return func(profileLink, linkFarm string) error {
		genLink := profileLink + "-" + strconv.Itoa(gen) + "-link"
		_ = os.Remove(genLink)
		if err := os.Symlink(linkFarm, genLink); err != nil {
			return err
		}
		_ = os.Remove(profileLink)
		return os.Symlink(filepath.Base(genLink), profileLink)
	}
}

// writeTestLinkFarm writes a manifest.json-only link-farm for the engine's pre-built path.
func writeTestLinkFarm(t *testing.T, entries ...manifest.Entry) string {
	t.Helper()
	dir := t.TempDir()
	m := manifest.Manifest{
		SchemaVersion: 1,
		Root:          manifest.Root{RootKind: manifest.RootKindHome},
		Entries:       entries,
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifest.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestJSONEndToEndApplyAndResetPayload drives the real engine on a tmpdir: a first apply, a second
// apply dropping an entry, then reset, checking each emitted envelope.
func TestJSONEndToEndApplyAndResetPayload(t *testing.T) {
	root := t.TempDir()
	state := t.TempDir()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	keep := manifest.Entry{SrcKind: "store", Src: src, Subpath: "f", Target: ".keep", Method: "symlink"}
	drop := manifest.Entry{SrcKind: "store", Src: src, Subpath: "f", Target: ".drop", Method: "symlink"}
	apply := func(gen int, entries ...manifest.Entry) *engine.Result {
		t.Helper()
		res, err := engine.Apply(engine.Options{
			LinkFarm: writeTestLinkFarm(t, entries...), Name: "cfg",
			RootOverride: root, StateDir: state, Commit: genCommit(gen),
			Warnf: func(string, ...any) {},
		})
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
		return res
	}

	// First apply: two adds, no previous generation to observe.
	res := apply(1, keep, drop)
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	doc := emitPayloadDoc(t, newApplyTestRun, p, nil)
	sr := subjectResultOf(t, doc)
	gen := sr["generation"].(map[string]any)
	if _, hasBefore := gen["before"]; hasBefore || gen["after"] != json.Number("1") {
		t.Errorf("first apply generation = %v, want before omitted / after 1", gen)
	}
	if items := sr["result"].(map[string]any)["items"].([]any); len(items) != 2 {
		t.Errorf("items = %v, want both entries", items)
	}
	if changes := sr["result"].(map[string]any)["changes"].([]any); len(changes) != 2 {
		t.Errorf("changes = %v, want two adds", changes)
	}

	// Second apply drops .drop: its old entry is stale-removed and stays in the inventory.
	res = apply(2, keep)
	p, err = mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	doc = emitPayloadDoc(t, newApplyTestRun, p, nil)
	sr = subjectResultOf(t, doc)
	gen = sr["generation"].(map[string]any)
	if gen["before"] != json.Number("1") || gen["after"] != json.Number("2") {
		t.Errorf("second apply generation = %v, want 1→2", gen)
	}
	if len(p.items) != 2 {
		t.Fatalf("second apply items = %+v, want the kept entry + the stale-removed old entry", p.items)
	}
	if cs := changesFor(t, p.changes, ".drop"); len(cs) != 1 || cs[0].Kind != outturn.ChangeRemove ||
		cs[0].Info == nil || cs[0].Info.Old != filepath.Join(src, "f") {
		t.Errorf(".drop changes = %+v, want one remove with the recorded dest", cs)
	}
	if cs := changesFor(t, p.changes, ".keep"); len(cs) != 0 {
		t.Errorf(".keep changes = %+v, want none (unchanged entry)", cs)
	}

	// Reset tears the remaining placement down: a reversible remove, no generation slot.
	resetRes, err := engine.Reset(engine.ResetOptions{
		Name: "cfg", RootKind: manifest.RootKindHome, RootOverride: root, StateDir: state,
		Warnf: func(string, ...any) {},
	})
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	rp, err := resetPayload[*resetResultInfo](resetRes, nil)
	if err != nil {
		t.Fatalf("resetPayload: %v", err)
	}
	doc = emitPayloadDoc(t, newResetTestRun, rp, nil)
	sr = subjectResultOf(t, doc)
	if gen, ok := sr["generation"]; ok {
		t.Errorf("reset generation = %v, want the slot absent", gen)
	}
	if cs := changesFor(t, rp.changes, ".keep"); len(cs) != 1 || cs[0].Kind != outturn.ChangeRemove || !cs[0].Reversible {
		t.Errorf("reset changes = %+v, want one reversible remove for .keep", cs)
	}
}

// TestDryrunPayloadFirstPlanOmitsGenerationNumbers: a dryrun over a not-yet-created profile emits
// the generation with the profile path alone, no before / after keys.
func TestDryrunPayloadFirstPlanOmitsGenerationNumbers(t *testing.T) {
	res := &engine.Result{
		Profile: "/state/nix/profiles/layat/home/profile",
		DryRun:  true,
		Entries: []manifest.Entry{{SrcKind: "store", Src: "/nix/store/z", Target: ".zshrc", Method: "symlink"}},
		Placed:  []string{".zshrc"},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	doc := emitPayloadDoc(t, newApplyTestRun, p, nil)
	gen := subjectResultOf(t, doc)["generation"].(map[string]any)
	if gen["profile"] != res.Profile {
		t.Errorf("generation.profile = %v, want %s", gen["profile"], res.Profile)
	}
	for _, key := range []string{"before", "after"} {
		if v, ok := gen[key]; ok {
			t.Errorf("generation.%s present (= %v), want omitted before the first apply", key, v)
		}
	}
}

// TestDryrunPayloadConflictKeepsEnvelopeBesideExit2: with cmdErr nil, the conflicted entry is a
// failed item with E_LAYAT_COLLISION, and the envelope stays conformant with status error and
// dryRun true.
func TestDryrunPayloadConflictKeepsEnvelopeBesideExit2(t *testing.T) {
	res := &engine.Result{
		Profile: "/state/nix/profiles/layat/home/profile",
		DryRun:  true,
		Entries: []manifest.Entry{
			{SrcKind: "store", Src: "/nix/store/a", Target: ".zshrc", Method: "symlink"},
			{SrcKind: "store", Src: "/nix/store/b", Target: ".config/nvim", Method: "symlink"},
		},
		Placed: []string{".config/nvim"},
		Conflicts: []planner.Conflict{
			{Entry: manifest.Entry{Target: ".zshrc"}, Reason: "target already has an existing file/directory (will not overwrite)", Kind: planner.ConflictForeignEntity},
		},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil) // the dryrun wiring passes cmdErr nil (→ runApply)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	if !p.itemBorne {
		t.Error("itemBorne = false, want true (the conflict is fully represented by its item)")
	}
	if it := findItem(t, p.items, ".config/nvim"); it.Status != outturn.ItemSuccess {
		t.Errorf("sibling item status = %s, want success (a dryrun attempts nothing)", it.Status)
	}

	r, buf := newApplyTestRun()
	r.dryRun = true
	r.beginSubject("default").setPayload(p)
	if err := r.emit(&exitError{code: 2}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings:\n%s\ndocument: %s", strings.Join(findings, "\n"), buf.String())
	}
	doc := decodeEnvelope(t, buf)
	if doc["status"] != "error" || doc["dryRun"] != true {
		t.Errorf("status/dryRun = %v/%v, want error/true", doc["status"], doc["dryRun"])
	}
	sr := subjectResultOf(t, doc)
	if errList, ok := sr["errors"]; ok {
		t.Errorf("subjectResult.errors = %v, want absent (the failed item carries the collision)", errList)
	}
	item := findItem(t, mustDecodeItems(t, sr), ".zshrc")
	if item.Status != outturn.ItemFailed || item.Error == nil || item.Error.Code != "E_LAYAT_COLLISION" {
		t.Errorf("conflicted item = %+v, want failed with E_LAYAT_COLLISION", item)
	}
}

// mustDecodeItems re-decodes a subjectResult's items into typed layatItem values so the typed
// helpers (findItem) work on emitted documents too.
func mustDecodeItems(t *testing.T, sr map[string]any) []layatItem {
	t.Helper()
	raw, err := json.Marshal(sr["result"].(map[string]any)["items"])
	if err != nil {
		t.Fatalf("re-marshal items: %v", err)
	}
	var items []layatItem
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	return items
}

// TestDryrunPayloadRelinkNotSuppressed: a dryrun observes no pre-relink dest, so a planned re-link
// stays a modify.
func TestDryrunPayloadRelinkNotSuppressed(t *testing.T) {
	res := &engine.Result{
		Profile:  "/state/nix/profiles/layat/home/profile",
		DryRun:   true,
		Entries:  []manifest.Entry{{SrcKind: "store", Src: "/nix/store/same", Target: ".same", Method: "symlink"}},
		Replaced: []string{".same"},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	cs := changesFor(t, p.changes, ".same")
	if len(cs) != 1 || cs[0].Kind != outturn.ChangeModify || !cs[0].Reversible {
		t.Fatalf(".same changes = %+v, want one reversible modify (unobserved old dest ⇒ not provably a noop)", cs)
	}
	if cs[0].Info == nil || cs[0].Info.New != "/nix/store/same" || cs[0].Info.Old != "" {
		t.Errorf("change info = %+v, want new only (the old dest is unobserved in a dryrun)", cs[0].Info)
	}
}

// TestMutationPayloadKeptStaleSubjectWarning: a kept-stale target has no item, so its
// W_LAYAT_STALE_* warning lands on the subject.
func TestMutationPayloadKeptStaleSubjectWarning(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		DryRun:  true,
		Entries: []manifest.Entry{{SrcKind: "store", Src: "/nix/store/a", Target: ".zshrc", Method: "symlink"}},
		Warnings: []planner.Warning{
			{Kind: planner.WarnStaleMismatch, Target: ".config/drifted"},
			{Kind: planner.WarnStaleNonSymlink, Target: ".config/solidified"},
		},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	want := []struct{ code, target string }{
		{"W_LAYAT_STALE_MISMATCH", ".config/drifted"},
		{"W_LAYAT_STALE_NON_SYMLINK", ".config/solidified"},
	}
	if len(p.warnings) != len(want) {
		t.Fatalf("subject warnings = %+v, want the %d kept-stale warnings", p.warnings, len(want))
	}
	for i, w := range want {
		if p.warnings[i].Code != w.code || p.warnings[i].Detail["target"] != w.target {
			t.Errorf("subject warnings[%d] = %+v, want %s on %s", i, p.warnings[i], w.code, w.target)
		}
	}
	if it := findItem(t, p.items, ".zshrc"); len(it.Warnings) != 0 {
		t.Errorf("inventory item warnings = %+v, want none (the stale warnings are subject-borne)", it.Warnings)
	}
}

// TestMutationPayloadConflictWithWarningStaysConformant: a conflict-failed item can carry a warning,
// and the envelope stays conformant.
func TestMutationPayloadConflictWithWarningStaysConformant(t *testing.T) {
	res := &engine.Result{
		Profile: "/p",
		DryRun:  true,
		Entries: []manifest.Entry{{SrcKind: "store", Src: "/nix/store/a", Target: ".zshrc", Method: "symlink"}},
		Conflicts: []planner.Conflict{
			{Entry: manifest.Entry{Target: ".zshrc"}, Reason: "target already has an existing file", Kind: planner.ConflictForeignEntity},
		},
		Warnings: []planner.Warning{{Kind: planner.WarnForeignReplace, Target: ".zshrc"}},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	it := findItem(t, p.items, ".zshrc")
	if it.Status != outturn.ItemFailed || it.Error == nil || len(it.Warnings) != 1 {
		t.Fatalf("item = %+v, want failed with both error and the entry-borne warning", it)
	}
	doc := emitPayloadDoc(t, newApplyTestRun, p, &exitError{code: 2})
	if doc["status"] != "error" {
		t.Errorf("status = %v, want error", doc["status"])
	}
}

// TestMutationPayloadCopyForeignItemWarning: W_LAYAT_COPY_FOREIGN rides on its item, whose entry
// stays in the manifest.
func TestMutationPayloadCopyForeignItemWarning(t *testing.T) {
	res := &engine.Result{
		Profile:  "/p",
		DryRun:   true,
		Entries:  []manifest.Entry{{SrcKind: "store", Src: "/nix/store/c", Target: ".config/copydir", Method: "copy"}},
		Warnings: []planner.Warning{{Kind: planner.WarnCopyForeign, Target: ".config/copydir"}},
	}
	p, err := mutationPayload[*applyResultInfo](res, nil)
	if err != nil {
		t.Fatalf("mutationPayload: %v", err)
	}
	it := findItem(t, p.items, ".config/copydir")
	if len(it.Warnings) != 1 || it.Warnings[0].Code != "W_LAYAT_COPY_FOREIGN" {
		t.Errorf("item warnings = %+v, want W_LAYAT_COPY_FOREIGN on the inventory item", it.Warnings)
	}
	if len(p.warnings) != 0 {
		t.Errorf("subject warnings = %+v, want none", p.warnings)
	}
	if it.Status != outturn.ItemSuccess {
		t.Errorf("item status = %s, want success (a skip is policy inaction, not a failure)", it.Status)
	}
}
