package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yasunori0418/layat/internal/generation"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// testRecord returns the record location for a home-keyed config under state.
func testRecord(state, name string) paths.Record {
	return paths.ResolveRecord(state, name, manifest.RootKindHome, "/home/me", false)
}

// putRecord writes a last-applied record with the given link-farm and manifest.
func putRecord(t *testing.T, rec paths.Record, linkFarm string, m manifest.Manifest) {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRecord(rec, "/home/me", linkFarm, raw, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

// putProfile points <dir>/profile at linkFarm, as a committed generation would.
func putProfile(t *testing.T, dir, linkFarm string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(dir, "profile")
	if err := os.Symlink(linkFarm, profile); err != nil {
		t.Fatal(err)
	}
	return profile
}

// entryTargets lists the targets of a manifest's entries ("<nil>" for an empty baseline).
func entryTargets(m *manifest.Manifest) string {
	if m == nil {
		return "<nil>"
	}
	var ts []string
	for _, e := range m.Entries {
		ts = append(ts, e.Target)
	}
	return strings.Join(ts, ",")
}

func TestReadRecordAbsentIsNoRecord(t *testing.T) {
	rec, m, err := readRecord(testRecord(realTempDir(t), "c").File)
	if err != nil || rec != nil || m != nil {
		t.Fatalf("readRecord(absent) = %v, %v, %v; want nil, nil, nil", rec, m, err)
	}
}

func TestWriteRecordRoundTrip(t *testing.T) {
	state := realTempDir(t)
	root := realTempDir(t)
	rec := paths.ResolveRecord(state, "c", manifest.RootKindProject, root, false)
	m := projectManifest(storeEntry("/nix/store/x", ".", "a"))
	raw, _ := json.Marshal(m)
	at := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	if err := writeRecord(rec, root, "/nix/store/lf", raw, at); err != nil {
		t.Fatal(err)
	}

	got, gotM, err := readRecord(rec.File)
	if err != nil {
		t.Fatal(err)
	}
	if got.RecordVersion != 1 || got.LinkFarm != "/nix/store/lf" || got.AppliedAt != at.Format(time.RFC3339) {
		t.Errorf("record = %+v", got)
	}
	if entryTargets(gotM) != "a" {
		t.Errorf("manifest entries = %s, want a", entryTargets(gotM))
	}
	// Every user can read the record, so a read-only operation does not need the writer's privileges.
	if fi, err := os.Stat(rec.File); err != nil {
		t.Error(err)
	} else if fi.Mode().Perm() != 0o644 {
		t.Errorf("record mode = %v, want 0644", fi.Mode().Perm())
	}
	if b, err := os.ReadFile(rec.Backref); err != nil || string(b) != root+"\n" {
		t.Errorf("backref = %q, %v; want %q", b, err, root+"\n")
	}
	// Only the record remains in its directory: the temp file was renamed into place.
	ents, _ := os.ReadDir(rec.Dir)
	if len(ents) != 1 || ents[0].Name() != "last-applied.json" {
		t.Errorf("record dir holds %v, want only last-applied.json", ents)
	}
}

func TestRemoveRecord(t *testing.T) {
	rec := testRecord(realTempDir(t), "c")
	putRecord(t, rec, "/nix/store/lf", homeManifest())
	if err := removeRecord(rec.File); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(rec.File); !os.IsNotExist(err) {
		t.Errorf("record still exists: %v", err)
	}
	if err := removeRecord(rec.File); err != nil {
		t.Errorf("removing an absent record = %v, want nil", err)
	}
}

// TestReadRecordRejectsInvalid verifies an unreadable record is an error, never "no record".
func TestReadRecordRejectsInvalid(t *testing.T) {
	const m = `{"schemaVersion":1,"root":{"rootKind":"home"},"entries":[]}`
	for _, tc := range []struct{ desc, body string }{
		{"broken JSON", `{"recordVersion":1,`},
		{"recordVersion missing", `{"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"recordVersion 2", `{"recordVersion":2,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"linkFarm missing", `{"recordVersion":1,"appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"linkFarm wrong type", `{"recordVersion":1,"linkFarm":3,"appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"appliedAt missing", `{"recordVersion":1,"linkFarm":"/lf","manifest":` + m + `}`},
		{"manifest missing", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z"}`},
		{"manifest null", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":null}`},
		{"manifest wrong type", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":"x"}`},
		{"unknown field", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","extra":true,"manifest":` + m + `}`},
		{"manifest without rootKind", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":{"schemaVersion":1,"root":{},"entries":[]}}`},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			file := filepath.Join(realTempDir(t), "last-applied.json")
			if err := os.WriteFile(file, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if rec, _, err := readRecord(file); err == nil {
				t.Errorf("readRecord = %+v, nil; want an error", rec)
			}
		})
	}
}

// TestLoadBaselinePrecedence verifies the baseline order record → profile → empty, with a
// warning on a record/profile mismatch and on the empty baseline.
func TestLoadBaselinePrecedence(t *testing.T) {
	recLF := writeLinkFarm(t, homeManifest(storeEntry("/nix/store/x", ".", "from-record")))
	profLF := writeLinkFarm(t, homeManifest(storeEntry("/nix/store/x", ".", "from-profile")))

	for _, tc := range []struct {
		desc              string
		record, profile   bool
		recordLinkFarm    string
		want, wantWarning string
	}{
		{desc: "record only", record: true, recordLinkFarm: recLF, want: "from-record"},
		{desc: "profile only", profile: true, want: "from-profile"},
		{desc: "neither", want: "<nil>", wantWarning: "no last-applied record"},
		{desc: "both agree", record: true, profile: true, recordLinkFarm: profLF, want: "from-record"},
		{desc: "both disagree", record: true, profile: true, recordLinkFarm: recLF, want: "from-record", wantWarning: "disagrees"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			state := realTempDir(t)
			rec := testRecord(state, "c")
			profile := filepath.Join(state, "profiles", "c", "profile")
			if tc.record {
				putRecord(t, rec, tc.recordLinkFarm, homeManifest(storeEntry("/nix/store/x", ".", "from-record")))
			}
			if tc.profile {
				putProfile(t, filepath.Dir(profile), profLF)
			}
			var warns []string
			b, err := loadBaseline(rec.File, profile, collectFormatted(&warns))
			if err != nil {
				t.Fatal(err)
			}
			if got := entryTargets(b.manifest); got != tc.want {
				t.Errorf("baseline = %s, want %s", got, tc.want)
			}
			switch {
			case tc.wantWarning == "" && len(warns) != 0:
				t.Errorf("warnings = %q, want none", warns)
			case tc.wantWarning != "" && (len(warns) != 1 || !strings.Contains(warns[0], tc.wantWarning)):
				t.Errorf("warnings = %q, want one mentioning %q", warns, tc.wantWarning)
			}
		})
	}
}

// TestLoadBaselineRejectsBrokenRecord verifies a broken record stops the baseline read even
// when a profile is there to fall back on.
func TestLoadBaselineRejectsBrokenRecord(t *testing.T) {
	state := realTempDir(t)
	rec := testRecord(state, "c")
	if err := os.MkdirAll(rec.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rec.File, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := putProfile(t, filepath.Join(state, "profiles", "c"), writeLinkFarm(t, homeManifest()))
	if _, err := loadBaseline(rec.File, profile, func(string, ...any) {}); err == nil {
		t.Error("loadBaseline with a broken record = nil error, want an error")
	}
}

// --- apply / rollback / reset keep the record ------------------------------

// projectRecord returns the record of the project-mode config applyOnce places under root.
func projectRecord(state, root string) paths.Record {
	return paths.ResolveRecord(state, "c", manifest.RootKindProject, root, true)
}

// mustReadRecord reads the record at rec, failing the test when it is absent or invalid.
func mustReadRecord(t *testing.T, rec paths.Record) (*lastApplied, *manifest.Manifest) {
	t.Helper()
	got, m, err := readRecord(rec.File)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatalf("no last-applied record at %s", rec.File)
	}
	return got, m
}

func TestApplyWritesRecord(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	lf := writeLinkFarm(t, projectManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	before := time.Now().Add(-time.Second)
	applyOnce(t, lf, "c", root, state, &generation.Fake{}, nil)

	got, _ := mustReadRecord(t, projectRecord(state, root))
	if got.LinkFarm != lf {
		t.Errorf("linkFarm = %q, want %q", got.LinkFarm, lf)
	}
	want, err := os.ReadFile(filepath.Join(lf, manifest.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Manifest) != string(want) {
		t.Errorf("manifest = %s, want manifest.json verbatim %s", got.Manifest, want)
	}
	if at, err := time.Parse(time.RFC3339, got.AppliedAt); err != nil || at.Before(before) {
		t.Errorf("appliedAt = %q (%v), want the apply time", got.AppliedAt, err)
	}
}

// TestApplySystemWritesRecordUnderSystemStateDir verifies that a root system apply keeps its
// record under <SystemStateDir>/layat with the profile's key, not under the user state.
func TestApplySystemWritesRecordUnderSystemStateDir(t *testing.T) {
	root, state, sys, sysState := realTempDir(t), realTempDir(t), realTempDir(t), realTempDir(t)
	lf := writeLinkFarm(t, systemManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	if _, err := Apply(Options{
		LinkFarm: lf, Name: "sys", RootOverride: root, StateDir: state, SystemStateDir: sysState,
		Backend: &generation.Fake{System: sys}, Geteuid: euid(0),
	}); err != nil {
		t.Fatalf("root system apply: %v", err)
	}

	hashDir := filepath.Join(sysState, "layat", paths.RootHash(root))
	got, _ := mustReadRecord(t, paths.Record{File: filepath.Join(hashDir, "sys", "last-applied.json")})
	if got.LinkFarm != lf {
		t.Errorf("linkFarm = %q, want %q", got.LinkFarm, lf)
	}
	if b, err := os.ReadFile(filepath.Join(hashDir, ".root")); err != nil || string(b) != root+"\n" {
		t.Errorf("record backref = %q (%v), want %q", b, err, root+"\n")
	}
	assertEmptyDirs(t, state)
}

// TestRecordForSystemDefaultBase verifies that an empty SystemStateDir puts the system record
// under /var/lib/layat with the profile's key.
func TestRecordForSystemDefaultBase(t *testing.T) {
	for _, tc := range []struct {
		desc, root string
		override   bool
		want       string
	}{
		{"no --root", "/", false, "/var/lib/layat/sys/last-applied.json"},
		{"--root", "/tmp/sandbox", true, "/var/lib/layat/39063e070b845ac426247c9b241f1698/sys/last-applied.json"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			rec, err := recordFor("/state", "", "sys", manifest.RootKindSystem, tc.root, tc.override)
			if err != nil {
				t.Fatal(err)
			}
			if rec.File != tc.want {
				t.Errorf("File = %q, want %q", rec.File, tc.want)
			}
		})
	}
}

func TestApplyFSFailureWritesNoRecord(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	blockWrite(t, filepath.Join(root, "ro"))
	lf := writeLinkFarm(t, projectManifest(storeEntry(makeSrc(t, "x"), ".", "ro/a")))
	fake := &generation.Fake{}
	if _, err := Apply(Options{LinkFarm: lf, Name: "c", RootOverride: root, StateDir: state, Backend: fake}); err == nil {
		t.Fatal("Apply = nil error, want the placement failure")
	}
	if _, err := os.Lstat(projectRecord(state, root).File); !os.IsNotExist(err) {
		t.Errorf("record exists after a failed placement: %v", err)
	}
}

func TestApplyRecordWriteFailureUnwinds(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	blockWrite(t, filepath.Join(state, "layat"))
	lf := writeLinkFarm(t, projectManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	fake := &generation.Fake{}
	res, err := Apply(Options{LinkFarm: lf, Name: "c", RootOverride: root, StateDir: state, Backend: fake})
	if err == nil || !strings.Contains(err.Error(), "last-applied record") {
		t.Fatalf("Apply = %v, want the record write failure", err)
	}
	if !res.Unwound {
		t.Error("Unwound = false, want the placement unwound")
	}
	if _, err := os.Lstat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Errorf("placed symlink survived the failed record write: %v", err)
	}
	if len(fake.Commits) != 0 {
		t.Errorf("commit calls = %d, want 0 (the record is written before the commit)", len(fake.Commits))
	}
}

// TestGenerationSkipNeedsRecordAndProfile verifies a generation skip only when the new link-farm
// matches both the profile and the record. After a failed commit the record is ahead of the
// profile; re-applying the profile's config then takes the normal path with the record as baseline.
func TestGenerationSkipNeedsRecordAndProfile(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	srcA, srcC := makeSrc(t, "x"), makeSrc(t, "x")
	lf1 := writeLinkFarm(t, projectManifest(storeEntry(srcA, ".", "a")))
	lf2 := writeLinkFarm(t, projectManifest(storeEntry(srcA, ".", "a"), storeEntry(srcC, ".", "c")))
	fake := &generation.Fake{}
	applyOnce(t, lf1, "c", root, state, fake, nil)

	failing := &generation.Fake{OnCommit: func(string, string) error { return os.ErrPermission }}
	if _, err := Apply(Options{LinkFarm: lf2, Name: "c", RootOverride: root, StateDir: state, Backend: failing}); err == nil {
		t.Fatal("Apply with a failing commit = nil error")
	}

	var warns []string
	res := applyOnce(t, lf1, "c", root, state, fake, &warns)
	if res.GenerationSkipped {
		t.Error("GenerationSkipped = true while the record (lf2) disagrees with the profile (lf1)")
	}
	if _, err := os.Lstat(filepath.Join(root, "c")); !os.IsNotExist(err) {
		t.Errorf("c placed by the uncommitted apply was not stale-removed: %v", err)
	}
	if len(fake.Commits) != 2 {
		t.Errorf("commit calls = %d, want 2", len(fake.Commits))
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "disagrees") {
		t.Errorf("warnings = %q, want one record/profile mismatch", warns)
	}

	if res := applyOnce(t, lf1, "c", root, state, fake, nil); !res.GenerationSkipped {
		t.Error("GenerationSkipped = false once the record and the profile agree")
	}
}

func TestGenerationSkipUpdatesRecord(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	m := projectManifest(storeEntry(makeSrc(t, "x"), ".", "a"))
	lf := writeLinkFarm(t, m)
	fake := &generation.Fake{}
	applyOnce(t, lf, "c", root, state, fake, nil)
	rec := projectRecord(state, root)
	putRecord(t, rec, lf, m)

	if res := applyOnce(t, lf, "c", root, state, fake, nil); !res.GenerationSkipped {
		t.Fatal("GenerationSkipped = false, want a skip")
	}
	if got, _ := mustReadRecord(t, rec); strings.HasPrefix(got.AppliedAt, "2000-") {
		t.Errorf("appliedAt = %q, want it refreshed by the skipped apply", got.AppliedAt)
	}
}

// TestRollbackWritesRecordBeforeSwitch verifies the record holds generation N-1 once the FS has
// converged, even when the profile pointer move then fails.
func TestRollbackWritesRecordBeforeSwitch(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	lf1 := writeLinkFarm(t, homeManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	lf2 := writeLinkFarm(t, homeManifest(storeEntry(makeSrc(t, "x"), ".", "b")))
	fake := &generation.Fake{}
	for _, lf := range []string{lf1, lf2} {
		if _, err := Apply(Options{LinkFarm: lf, Name: "vim", RootOverride: root, StateDir: state, Backend: fake}); err != nil {
			t.Fatal(err)
		}
	}
	rec := paths.ResolveRecord(state, "vim", manifest.RootKindHome, root, true)
	opts := RollbackOptions{Name: "vim", RootKind: manifest.RootKindHome, RootOverride: root, StateDir: state, Backend: fake}

	fake.OnSwitch = func(string, int) error { return os.ErrPermission }
	if _, err := Rollback(opts); err == nil {
		t.Fatal("Rollback with a failing switch = nil error")
	}
	got, m := mustReadRecord(t, rec)
	if got.LinkFarm != lf1 || entryTargets(m) != "a" {
		t.Errorf("record after the failed switch = %s / %s, want %s / a", got.LinkFarm, entryTargets(m), lf1)
	}

	fake.OnSwitch = nil
	if _, err := Rollback(opts); err != nil {
		t.Fatal(err)
	}
	if got, m := mustReadRecord(t, rec); got.LinkFarm != lf1 || entryTargets(m) != "a" {
		t.Errorf("record after rollback = %s / %s, want %s / a", got.LinkFarm, entryTargets(m), lf1)
	}
}

// TestResetUpdatesRecord verifies reset drops only the entries it removed from the record, keeps
// the record's link-farm, and deletes the record once nothing is left.
func TestResetUpdatesRecord(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	srcA, srcB, srcC := makeSrc(t, "x"), makeSrc(t, "x"), makeSrc(t, "x")
	applyForReset(t, root, state, homeManifest(
		storeEntry(srcA, ".", "a"), storeEntry(srcB, ".", "b"), storeEntry(srcC, ".", "c"),
		copyEntry(makeSrc(t, "x"), "x", "d"),
	))
	rec := paths.ResolveRecord(state, "cfg", manifest.RootKindHome, root, true)
	lf, _ := mustReadRecord(t, rec)

	// c is re-pointed by another tool: reset keeps it as foreign.
	c := filepath.Join(root, "c")
	if err := os.Remove(c); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(srcA, c); err != nil {
		t.Fatal(err)
	}

	steps := []struct {
		targets []string
		want    string
	}{
		{[]string{"a"}, "b,c,d"},
		{nil, "c"},
	}
	for _, s := range steps {
		if _, err := Reset(resetOpts(root, state, s.targets, false, nil)); err != nil {
			t.Fatal(err)
		}
		got, m := mustReadRecord(t, rec)
		if entryTargets(m) != s.want || got.LinkFarm != lf.LinkFarm {
			t.Errorf("reset %v: record = %s / %s, want %s / %s", s.targets, entryTargets(m), got.LinkFarm, s.want, lf.LinkFarm)
		}
	}

	if err := os.Remove(c); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(srcC, c); err != nil {
		t.Fatal(err)
	}
	if _, err := Reset(resetOpts(root, state, nil, false, nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(rec.File); !os.IsNotExist(err) {
		t.Errorf("record survived a full reset: %v", err)
	}
}

// TestLoadBaselineUnreadableProfile verifies a profile whose manifest cannot be read is reported
// without a record, leaving the baseline empty.
func TestLoadBaselineUnreadableProfile(t *testing.T) {
	state := realTempDir(t)
	lf := realTempDir(t)
	if err := os.WriteFile(filepath.Join(lf, manifest.FileName), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := putProfile(t, filepath.Join(state, "profiles", "c"), lf)
	b, err := loadBaseline(testRecord(state, "c").File, profile, func(string, ...any) {})
	if err != nil {
		t.Fatal(err)
	}
	if b.manifest != nil || !b.profileUnreadable() {
		t.Errorf("baseline = %s, unreadable = %v; want an empty baseline reported unreadable", entryTargets(b.manifest), b.profileUnreadable())
	}
}

// failCommit is a backend whose commits fail, leaving the record ahead of the profile.
func failCommit() *generation.Fake {
	return &generation.Fake{OnCommit: func(string, string) error { return os.ErrPermission }}
}

// TestGenerationSkipNeedsProfile verifies a record without a profile (a failed first commit)
// never skips, so the commit is retried.
func TestGenerationSkipNeedsProfile(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	lf := writeLinkFarm(t, projectManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	if _, err := Apply(Options{LinkFarm: lf, Name: "c", RootOverride: root, StateDir: state, Backend: failCommit()}); err == nil {
		t.Fatal("Apply with a failing commit = nil error")
	}
	fake := &generation.Fake{}
	if res := applyOnce(t, lf, "c", root, state, fake, nil); res.GenerationSkipped || len(fake.Commits) != 1 {
		t.Errorf("GenerationSkipped = %v, commits = %d; want the commit retried", res.GenerationSkipped, len(fake.Commits))
	}
}

// TestGenerationSkipWithoutRecord verifies a profile matching the new link-farm skips when there
// is no record, and the skip writes the record.
func TestGenerationSkipWithoutRecord(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	lf := writeLinkFarm(t, projectManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	fake := &generation.Fake{}
	applyOnce(t, lf, "c", root, state, fake, nil)
	rec := projectRecord(state, root)
	if err := removeRecord(rec.File); err != nil {
		t.Fatal(err)
	}
	if res := applyOnce(t, lf, "c", root, state, fake, nil); !res.GenerationSkipped {
		t.Error("GenerationSkipped = false, want a skip on the profile alone")
	}
	if got, _ := mustReadRecord(t, rec); got.LinkFarm != lf {
		t.Errorf("record linkFarm = %q, want %q", got.LinkFarm, lf)
	}
}

// isRecordErr reports whether err comes from the last-applied record.
func isRecordErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "last-applied record")
}

// lockRecordDir makes the record directory read-only so the next record write fails.
func lockRecordDir(t *testing.T, rec paths.Record) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("running as root: directory write-permission checks are bypassed")
	}
	if err := os.Chmod(rec.Dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(rec.Dir, 0o755) })
}

// TestRecordWriteFailureFailsOperation verifies a record that cannot be written fails the
// generation skip (unwinding its repair), rollback (before the pointer move), and reset.
func TestRecordWriteFailureFailsOperation(t *testing.T) {
	t.Run("generation skip", func(t *testing.T) {
		root, state := realTempDir(t), realTempDir(t)
		lf := writeLinkFarm(t, projectManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
		fake := &generation.Fake{}
		applyOnce(t, lf, "c", root, state, fake, nil)
		if err := os.Remove(filepath.Join(root, "a")); err != nil {
			t.Fatal(err)
		}
		lockRecordDir(t, projectRecord(state, root))
		if _, err := Apply(Options{LinkFarm: lf, Name: "c", RootOverride: root, StateDir: state, Backend: fake}); !isRecordErr(err) {
			t.Fatalf("Apply = %v, want the record write failure", err)
		}
		if _, err := os.Lstat(filepath.Join(root, "a")); !os.IsNotExist(err) {
			t.Errorf("drift repair survived the failed record write: %v", err)
		}
	})
	t.Run("rollback", func(t *testing.T) {
		root, state := realTempDir(t), realTempDir(t)
		fake := &generation.Fake{}
		for _, target := range []string{"a", "b"} {
			lf := writeLinkFarm(t, homeManifest(storeEntry(makeSrc(t, "x"), ".", target)))
			if _, err := Apply(Options{LinkFarm: lf, Name: "vim", RootOverride: root, StateDir: state, Backend: fake}); err != nil {
				t.Fatal(err)
			}
		}
		lockRecordDir(t, paths.ResolveRecord(state, "vim", manifest.RootKindHome, root, true))
		if _, err := Rollback(RollbackOptions{Name: "vim", RootKind: manifest.RootKindHome, RootOverride: root, StateDir: state, Backend: fake}); !isRecordErr(err) {
			t.Fatalf("Rollback = %v, want the record write failure", err)
		}
		if len(fake.Switches) != 0 {
			t.Errorf("switches = %v, want none", fake.Switches)
		}
		if _, err := os.Lstat(filepath.Join(root, "b")); err != nil {
			t.Errorf("b was not restored by the unwind: %v", err)
		}
		if _, err := os.Lstat(filepath.Join(root, "a")); !os.IsNotExist(err) {
			t.Errorf("a survived the unwind: %v", err)
		}
	})
	t.Run("reset", func(t *testing.T) {
		root, state := realTempDir(t), realTempDir(t)
		applyForReset(t, root, state, homeManifest(storeEntry(makeSrc(t, "x"), ".", "a"), storeEntry(makeSrc(t, "x"), ".", "b")))
		lockRecordDir(t, paths.ResolveRecord(state, "cfg", manifest.RootKindHome, root, true))
		if _, err := Reset(resetOpts(root, state, []string{"a"}, false, nil)); !isRecordErr(err) {
			t.Fatalf("Reset = %v, want the record write failure", err)
		}
		if _, err := os.Lstat(filepath.Join(root, "a")); !os.IsNotExist(err) {
			t.Errorf("a was not removed: %v", err)
		}
		if _, err := os.Lstat(filepath.Join(root, "b")); err != nil {
			t.Errorf("b outside the targets was touched: %v", err)
		}
	})
}

// TestRollbackReadsRecordFirst verifies rollback tears down what the record says was placed, not
// the profile's current generation, after a failed commit left the record ahead.
func TestRollbackReadsRecordFirst(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	fake := &generation.Fake{}
	var lfs []string
	for _, target := range []string{"a", "b", "c"} {
		lfs = append(lfs, writeLinkFarm(t, homeManifest(storeEntry(makeSrc(t, "x"), ".", target))))
	}
	for i, lf := range lfs {
		backend := fake
		if i == 2 {
			backend = failCommit()
		}
		if _, err := Apply(Options{LinkFarm: lf, Name: "vim", RootOverride: root, StateDir: state, Backend: backend}); (err != nil) != (i == 2) {
			t.Fatalf("Apply %d = %v, want only the third commit to fail", i+1, err)
		}
	}
	if _, err := Rollback(RollbackOptions{Name: "vim", RootKind: manifest.RootKindHome, RootOverride: root, StateDir: state, Backend: fake, Warnf: func(string, ...any) {}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(root, "c")); !os.IsNotExist(err) {
		t.Errorf("c from the uncommitted apply survived the rollback: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, "a")); err != nil {
		t.Errorf("a was not placed by the rollback: %v", err)
	}
}

// TestResetRecordOnly verifies reset tears down a record-only placement (a failed first commit,
// no profile) and deletes the record, preparing the profileDir and its backref like apply.
func TestResetRecordOnly(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	lf := writeLinkFarm(t, homeManifest(storeEntry(makeSrc(t, "x"), ".", "a")))
	if _, err := Apply(Options{LinkFarm: lf, Name: "cfg", RootOverride: root, StateDir: state, Backend: failCommit()}); err == nil {
		t.Fatal("Apply with a failing commit = nil error")
	}
	prof := paths.Resolve(fakeBase(state), "cfg", manifest.RootKindHome, root, true)
	if err := os.RemoveAll(prof.BackrefDir); err != nil {
		t.Fatal(err)
	}
	if _, err := Reset(resetOpts(root, state, nil, false, nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Errorf("a survived the reset: %v", err)
	}
	if b, err := os.ReadFile(prof.Backref); err != nil || string(b) != root+"\n" {
		t.Errorf("profile backref = %q, %v; want %q", b, err, root+"\n")
	}
	if _, err := os.Lstat(paths.ResolveRecord(state, "cfg", manifest.RootKindHome, root, true).File); !os.IsNotExist(err) {
		t.Errorf("record survived a full reset: %v", err)
	}
}

// TestResetRecordFromProfile verifies reset without a record writes one from the profile's
// current generation, keeping that link-farm.
func TestResetRecordFromProfile(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	applyForReset(t, root, state, homeManifest(storeEntry(makeSrc(t, "x"), ".", "a"), storeEntry(makeSrc(t, "x"), ".", "b")))
	rec := paths.ResolveRecord(state, "cfg", manifest.RootKindHome, root, true)
	lf, _ := mustReadRecord(t, rec)
	if err := removeRecord(rec.File); err != nil {
		t.Fatal(err)
	}
	if _, err := Reset(resetOpts(root, state, []string{"a"}, false, nil)); err != nil {
		t.Fatal(err)
	}
	if got, m := mustReadRecord(t, rec); got.LinkFarm != lf.LinkFarm || entryTargets(m) != "b" {
		t.Errorf("record = %s / %s, want %s / b", got.LinkFarm, entryTargets(m), lf.LinkFarm)
	}
}

// TestResetPartialFailureKeepsRecord verifies a reset stopped mid-way still rewrites the record,
// keeping exactly the selected entries whose target is still on disk.
func TestResetPartialFailureKeepsRecord(t *testing.T) {
	for _, tc := range []struct {
		desc, frozen string
		entries      func(src string) []manifest.Entry
	}{
		{"symlink stage", "ro", func(src string) []manifest.Entry {
			return []manifest.Entry{storeEntry(src, ".", "a"), storeEntry(src, ".", "ro/s")}
		}},
		{"copy stage", "boxed", func(src string) []manifest.Entry {
			return []manifest.Entry{storeEntry(src, ".", "a"), copyEntry(src, "x", "boxed/c1"), copyEntry(src, "x", "later/c2")}
		}},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			root, state := realTempDir(t), realTempDir(t)
			entries := tc.entries(makeSrc(t, "x"))
			applyForReset(t, root, state, homeManifest(entries...))
			rec := paths.ResolveRecord(state, "cfg", manifest.RootKindHome, root, true)
			lf, _ := mustReadRecord(t, rec)
			if os.Geteuid() == 0 {
				t.Skip("running as root: directory write-permission checks are bypassed")
			}
			frozen := filepath.Join(root, tc.frozen)
			if err := os.Chmod(frozen, 0o555); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(frozen, 0o755) })

			if _, err := Reset(resetOpts(root, state, nil, false, nil)); err == nil {
				t.Fatal("Reset = nil error, want the frozen removal to fail")
			}
			var want []string
			for _, e := range entries {
				if _, err := os.Lstat(filepath.Join(root, e.Target)); err == nil {
					want = append(want, e.Target)
				}
			}
			got, m := mustReadRecord(t, rec)
			if len(want) == 0 || entryTargets(m) != strings.Join(want, ",") || got.LinkFarm != lf.LinkFarm {
				t.Errorf("record = %s / %s, want the entries still on disk %v / %s", entryTargets(m), got.LinkFarm, want, lf.LinkFarm)
			}
		})
	}
}

// TestResetPartialFailureJoinsRecordError verifies a reset stopped mid-way whose record rewrite
// also fails returns both the removal error and the record error.
func TestResetPartialFailureJoinsRecordError(t *testing.T) {
	root, state := realTempDir(t), realTempDir(t)
	src := makeSrc(t, "x")
	applyForReset(t, root, state, homeManifest(storeEntry(src, ".", "a"), copyEntry(src, "x", "boxed/c1")))
	lockRecordDir(t, paths.ResolveRecord(state, "cfg", manifest.RootKindHome, root, true))
	frozen := filepath.Join(root, "boxed")
	if err := os.Chmod(frozen, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(frozen, 0o755) })

	_, err := Reset(resetOpts(root, state, nil, false, nil))
	if err == nil || !strings.Contains(err.Error(), "cannot remove copy target") || !isRecordErr(err) {
		t.Fatalf("Reset = %v, want the removal error joined with the record write failure", err)
	}
}
