package engine

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yasunori0418/layat/internal/lock"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// statusAt is the appliedAt every record written by putStatusRecord carries.
var statusAt = time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)

// putStatusRecord writes the last-applied record of config name under state, keyed like apply
// would key it, and returns its location.
func putStatusRecord(t *testing.T, state, name, root string, rootOverride bool, linkFarm string, m manifest.Manifest) paths.Record {
	t.Helper()
	rec := paths.ResolveRecord(state, name, m.Root.RootKind, root, rootOverride)
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRecord(rec, root, linkFarm, raw, statusAt); err != nil {
		t.Fatal(err)
	}
	return rec
}

// writeStatusFile writes data to path, creating its parent directories.
func writeStatusFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// chmodStatus changes the mode of path and restores 0755 on cleanup; it skips as root, where
// permission checks are bypassed.
func chmodStatus(t *testing.T, path string, mode fs.FileMode) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("running as root: directory permission checks are bypassed")
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
}

// runStatus calls Status on the two state dirs and fails the test on error.
func runStatus(t *testing.T, state, sysState string, warnings *[]string) []StatusSeries {
	t.Helper()
	got, err := Status(StatusOptions{StateDir: state, SystemStateDir: sysState, Warnf: collectFormatted(warnings)})
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	return got
}

// statusKey is the identity and placement of one listed series, ignoring Err and Manifest.
type statusKey struct {
	Mode, Name, Root, RootHash, RecordFile string
	HasErr                                 bool
}

func statusKeys(ss []StatusSeries) []statusKey {
	var ks []statusKey
	for _, s := range ss {
		ks = append(ks, statusKey{s.Mode, s.Name, s.Root, s.RootHash, s.RecordFile, s.Err != nil})
	}
	return ks
}

func assertStatusKeys(t *testing.T, got []StatusSeries, want []statusKey) {
	t.Helper()
	if gk := statusKeys(got); !reflect.DeepEqual(gk, want) {
		t.Fatalf("Status keys:\n got  %+v\n want %+v", gk, want)
	}
}

// sortedRoots returns the two roots ordered by their roothash.
func sortedRoots(a, b string) (string, string) {
	if paths.RootHash(a) > paths.RootHash(b) {
		return b, a
	}
	return a, b
}

// TestStatusListsBothBasesInOrder verifies the user base then the system base, each with the
// <name>-keyed records first (by <name>) and the <roothash> series after (by <roothash>, <name>),
// the same on every run, carrying the record's values.
func TestStatusListsBothBasesInOrder(t *testing.T) {
	state, sysState := realTempDir(t), realTempDir(t)
	r1, r2 := sortedRoots(realTempDir(t), realTempDir(t))
	h1, h2 := paths.RootHash(r1), paths.RootHash(r2)
	home := homeManifest(storeEntry("/nix/store/h", ".", ".config/h"))
	proj := projectManifest(storeEntry("/nix/store/p", ".", "p"))
	fixed := manifest.Manifest{SchemaVersion: 1, Root: manifest.Root{RootKind: manifest.RootKindFixed}}
	sys := systemManifest(storeEntry("/nix/store/s", ".", "etc/s"))

	// Written out of order so the listing order cannot come from creation order.
	putStatusRecord(t, state, "b", "/home/me", false, "/nix/store/lf-b", home)
	putStatusRecord(t, state, "a", "/home/me", false, "/nix/store/lf-a", home)
	putStatusRecord(t, state, "y", r2, false, "/nix/store/lf-y2", proj)
	putStatusRecord(t, state, "z", r1, false, "/nix/store/lf-z1", fixed)
	putStatusRecord(t, state, "x", r1, false, "/nix/store/lf-x1", proj)
	putStatusRecord(t, sysState, "s", "/", false, "/nix/store/lf-s", sys)
	putStatusRecord(t, sysState, "s", r1, true, "/nix/store/lf-s1", sys)
	// A <name> without a record (e.g. after reset removed it) is not listed in either stage.
	if err := os.MkdirAll(filepath.Join(state, "layat", "norecord"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(state, "layat", h1, "norecord"), 0o755); err != nil {
		t.Fatal(err)
	}

	user, system := filepath.Join(state, "layat"), filepath.Join(sysState, "layat")
	rf := func(base string, parts ...string) string {
		return filepath.Join(append(append([]string{base}, parts...), "last-applied.json")...)
	}
	want := []statusKey{
		{Mode: "home", Name: "a", RecordFile: rf(user, "a")},
		{Mode: "home", Name: "b", RecordFile: rf(user, "b")},
		{Mode: "project", Name: "x", Root: r1, RootHash: h1, RecordFile: rf(user, h1, "x")},
		{Mode: "fixed", Name: "z", Root: r1, RootHash: h1, RecordFile: rf(user, h1, "z")},
		{Mode: "project", Name: "y", Root: r2, RootHash: h2, RecordFile: rf(user, h2, "y")},
		{Mode: "system", Name: "s", RecordFile: rf(system, "s")},
		{Mode: "system", Name: "s", Root: r1, RootHash: h1, RecordFile: rf(system, h1, "s")},
	}

	var warnings []string
	got := runStatus(t, state, sysState, &warnings)
	assertStatusKeys(t, got, want)
	if len(warnings) != 0 {
		t.Errorf("warnings = %q; want none", warnings)
	}

	for _, s := range got {
		if s.Err != nil {
			t.Errorf("%s: Err = %v", s.RecordFile, s.Err)
			continue
		}
		if s.AppliedAt != statusAt.Format(time.RFC3339) {
			t.Errorf("%s: AppliedAt = %q", s.RecordFile, s.AppliedAt)
		}
		if s.Manifest == nil || s.Manifest.Root.RootKind == "" {
			t.Errorf("%s: Manifest = %+v; want the recorded manifest", s.RecordFile, s.Manifest)
		}
	}
	if got[0].LinkFarm != "/nix/store/lf-a" || entryTargets(got[0].Manifest) != ".config/h" {
		t.Errorf("a: LinkFarm = %q, entries = %q", got[0].LinkFarm, entryTargets(got[0].Manifest))
	}
	if got[4].LinkFarm != "/nix/store/lf-y2" || entryTargets(got[4].Manifest) != "p" {
		t.Errorf("y: LinkFarm = %q, entries = %q", got[4].LinkFarm, entryTargets(got[4].Manifest))
	}
	if got[6].LinkFarm != "/nix/store/lf-s1" || entryTargets(got[6].Manifest) != "etc/s" {
		t.Errorf("system s@root: LinkFarm = %q, entries = %q", got[6].LinkFarm, entryTargets(got[6].Manifest))
	}

	if again := runStatus(t, state, sysState, &warnings); !reflect.DeepEqual(again, got) {
		t.Errorf("second run differs:\n got  %+v\n want %+v", statusKeys(again), want)
	}
}

// TestStatusAbsentBases verifies a base that does not exist is empty, not an error.
func TestStatusAbsentBases(t *testing.T) {
	for name, state := range map[string]string{
		"empty state dir":   realTempDir(t),
		"missing state dir": filepath.Join(realTempDir(t), "missing"),
	} {
		t.Run(name, func(t *testing.T) {
			var warnings []string
			got := runStatus(t, state, filepath.Join(realTempDir(t), "missing"), &warnings)
			if len(got) != 0 || len(warnings) != 0 {
				t.Fatalf("Status = %+v, warnings %q; want empty and none", got, warnings)
			}
		})
	}
}

// TestStatusUserBaseUnreadableFails verifies a user base that exists but cannot be read fails.
func TestStatusUserBaseUnreadableFails(t *testing.T) {
	state, sysState := realTempDir(t), realTempDir(t)
	putStatusRecord(t, state, "a", "/home/me", false, "/nix/store/lf", homeManifest())
	putStatusRecord(t, sysState, "s", "/", false, "/nix/store/lf", systemManifest())
	chmodStatus(t, filepath.Join(state, "layat"), 0o000)

	got, err := Status(StatusOptions{StateDir: state, SystemStateDir: sysState, Warnf: func(string, ...any) {}})
	if err == nil || !strings.Contains(err.Error(), "cannot list the record base") {
		t.Fatalf("Status = %+v, %v; want an error naming the unreadable record base", statusKeys(got), err)
	}
}

// TestStatusSystemBaseUnreadableWarns verifies a system base that cannot be read is skipped
// with one warning naming it, and the user base is still listed.
func TestStatusSystemBaseUnreadableWarns(t *testing.T) {
	state, sysState := realTempDir(t), realTempDir(t)
	rec := putStatusRecord(t, state, "a", "/home/me", false, "/nix/store/lf", homeManifest())
	putStatusRecord(t, sysState, "s", "/", false, "/nix/store/lf", systemManifest())
	system := filepath.Join(sysState, "layat")
	chmodStatus(t, system, 0o000)

	var warnings []string
	got := runStatus(t, state, sysState, &warnings)
	assertStatusKeys(t, got, []statusKey{{Mode: "home", Name: "a", RecordFile: rec.File}})
	if len(warnings) != 1 || !strings.Contains(warnings[0], "cannot list the record base ("+system+")") {
		t.Fatalf("warnings = %q; want one line naming the record base %s", warnings, system)
	}
}

// TestStatusInvalidRecordListedWithErr verifies an unreadable or invalid record is an error on
// its series, placed among the others, and the listing continues.
func TestStatusInvalidRecordListedWithErr(t *testing.T) {
	state, sysState := realTempDir(t), realTempDir(t)
	root := realTempDir(t)
	h := paths.RootHash(root)
	user, system := filepath.Join(state, "layat"), filepath.Join(sysState, "layat")

	a := putStatusRecord(t, state, "a", "/home/me", false, "/nix/store/lf", homeManifest())
	writeStatusFile(t, filepath.Join(user, "b", "last-applied.json"), "{not json")
	writeStatusFile(t, filepath.Join(user, "c", "last-applied.json"),
		`{"recordVersion":1,"linkFarm":"/nix/store/lf","appliedAt":"x","manifest":{"schemaVersion":1,"root":{"rootKind":"home"},"entries":[]},"extra":1}`)
	x := putStatusRecord(t, state, "x", root, false, "/nix/store/lf", projectManifest())
	writeStatusFile(t, filepath.Join(user, h, "y", "last-applied.json"),
		`{"recordVersion":2,"linkFarm":"/nix/store/lf","appliedAt":"x","manifest":{"schemaVersion":1,"root":{"rootKind":"project"},"entries":[]}}`)
	writeStatusFile(t, filepath.Join(system, "s", "last-applied.json"), "")

	var warnings []string
	got := runStatus(t, state, sysState, &warnings)
	assertStatusKeys(t, got, []statusKey{
		{Mode: "home", Name: "a", RecordFile: a.File},
		{Mode: "home", Name: "b", RecordFile: filepath.Join(user, "b", "last-applied.json"), HasErr: true},
		{Mode: "home", Name: "c", RecordFile: filepath.Join(user, "c", "last-applied.json"), HasErr: true},
		{Mode: "project", Name: "x", Root: root, RootHash: h, RecordFile: x.File},
		{Mode: "home", Name: "y", Root: root, RootHash: h, RecordFile: filepath.Join(user, h, "y", "last-applied.json"), HasErr: true},
		{Mode: "system", Name: "s", RecordFile: filepath.Join(system, "s", "last-applied.json"), HasErr: true},
	})
	for _, i := range []int{1, 2, 4, 5} {
		if got[i].Manifest != nil {
			t.Errorf("%s: Manifest = %+v; want nil with Err", got[i].RecordFile, got[i].Manifest)
		}
	}
}

// TestStatusUnreadableSeries verifies the series-level errors: a backref that cannot be read,
// <name> profileDirs that cannot be listed, and a record that cannot be stat'd.
func TestStatusUnreadableSeries(t *testing.T) {
	t.Run("backref unreadable with records", func(t *testing.T) {
		state := realTempDir(t)
		root := realTempDir(t)
		h := paths.RootHash(root)
		hashDir := filepath.Join(state, "layat", h)
		n := putStatusRecord(t, state, "n", root, false, "/nix/store/lf", projectManifest())
		if err := os.MkdirAll(filepath.Join(hashDir, "norecord"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeStatusFile(t, filepath.Join(hashDir, ".root"), "relative/root\n")

		var warnings []string
		got := runStatus(t, state, realTempDir(t), &warnings)
		assertStatusKeys(t, got, []statusKey{{Mode: "home", RootHash: h, Name: "n", RecordFile: n.File, HasErr: true}})
	})

	t.Run("backref unreadable without records", func(t *testing.T) {
		state, sysState := realTempDir(t), realTempDir(t)
		hashDir := filepath.Join(sysState, "layat", "deadbeef")
		writeStatusFile(t, filepath.Join(hashDir, ".root"), "")
		if err := os.MkdirAll(filepath.Join(hashDir, "norecord"), 0o755); err != nil {
			t.Fatal(err)
		}

		var warnings []string
		got := runStatus(t, state, sysState, &warnings)
		assertStatusKeys(t, got, []statusKey{{Mode: "system", RootHash: "deadbeef", HasErr: true}})
	})

	t.Run("names unlistable", func(t *testing.T) {
		state := realTempDir(t)
		root := realTempDir(t)
		h := paths.RootHash(root)
		putStatusRecord(t, state, "n", root, false, "/nix/store/lf", projectManifest())
		chmodStatus(t, filepath.Join(state, "layat", h), 0o300)

		var warnings []string
		got := runStatus(t, state, realTempDir(t), &warnings)
		assertStatusKeys(t, got, []statusKey{{Mode: "home", RootHash: h, HasErr: true}})
	})

	t.Run("whole series sits at its roothash with both reasons", func(t *testing.T) {
		state := realTempDir(t)
		user := filepath.Join(state, "layat")
		root := realTempDir(t)
		h := paths.RootHash(root)
		n := putStatusRecord(t, state, "n", root, false, "/nix/store/lf", projectManifest())
		zz := putStatusRecord(t, state, "zz", "/home/me", false, "/nix/store/lf", homeManifest())
		// "0" sorts before any other roothash, "zz" and "zzz" after every roothash.
		for _, name := range []string{"0", "zzz"} {
			whole := filepath.Join(user, name)
			writeStatusFile(t, filepath.Join(whole, ".root"), root+"\n")
			chmodStatus(t, whole, 0o000)
		}

		var warnings []string
		got := runStatus(t, state, realTempDir(t), &warnings)
		assertStatusKeys(t, got, []statusKey{
			{Mode: "home", Name: "zz", RecordFile: zz.File},
			{Mode: "home", RootHash: "0", HasErr: true},
			{Mode: "project", Name: "n", Root: root, RootHash: h, RecordFile: n.File},
			{Mode: "home", RootHash: "zzz", HasErr: true},
		})
		if msg := got[1].Err.Error(); !strings.Contains(msg, "backref") || !strings.Contains(msg, "cannot list series") {
			t.Errorf("Err = %q; want both the backref and the listing reason", msg)
		}
	})

	t.Run("unsearchable name-keyed dir is listed once", func(t *testing.T) {
		state := realTempDir(t)
		putStatusRecord(t, state, "a", "/home/me", false, "/nix/store/lf", homeManifest())
		putStatusRecord(t, state, "x", "/home/me", false, "/nix/store/lf", homeManifest())
		chmodStatus(t, filepath.Join(state, "layat", "x"), 0o600)

		var warnings []string
		got := runStatus(t, state, realTempDir(t), &warnings)
		assertStatusKeys(t, got, []statusKey{
			{Mode: "home", Name: "a", RecordFile: filepath.Join(state, "layat", "a", "last-applied.json")},
			{Mode: "home", RootHash: "x", HasErr: true},
		})
	})

	t.Run("record cannot be stat'd", func(t *testing.T) {
		state := realTempDir(t)
		root := realTempDir(t)
		h := paths.RootHash(root)
		n := putStatusRecord(t, state, "n", root, false, "/nix/store/lf", projectManifest())
		chmodStatus(t, n.Dir, 0o600)

		var warnings []string
		got := runStatus(t, state, realTempDir(t), &warnings)
		assertStatusKeys(t, got, []statusKey{{Mode: "home", Name: "n", Root: root, RootHash: h, RecordFile: n.File, HasErr: true}})
	})
}

// statusSnapshot lists every entry under dirs with each regular file's content.
func statusSnapshot(t *testing.T, dirs ...string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			snap[path] = ""
			if d.Type().IsRegular() {
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				snap[path] = string(b)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return snap
}

// TestStatusIsReadOnly verifies Status leaves both state dirs (record and profile bases) as they
// were, rewrites no record, and does not wait on a profileDir lock held by another.
func TestStatusIsReadOnly(t *testing.T) {
	state, sysState := realTempDir(t), realTempDir(t)
	root := realTempDir(t)
	putStatusRecord(t, state, "a", "/home/me", false, "/nix/store/lf", homeManifest())
	putStatusRecord(t, state, "x", root, false, "/nix/store/lf", projectManifest())
	writeStatusFile(t, filepath.Join(state, "layat", "bad", "last-applied.json"), "{")
	putStatusRecord(t, sysState, "s", "/", false, "/nix/store/lf", systemManifest())
	// Profile base of the same configs, with no .pending or .root yet.
	putProfile(t, filepath.Join(fakeBase(state), "a"), "/nix/store/lf")
	profileDir := filepath.Join(fakeBase(state), paths.RootHash(root), "x")
	putProfile(t, profileDir, "/nix/store/lf")

	for _, dir := range []string{profileDir, filepath.Join(state, "layat", "a")} {
		l, err := lock.Acquire(dir, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = l.Release() })
	}

	before := statusSnapshot(t, state, sysState)
	done := make(chan []StatusSeries, 1)
	go func() {
		got, _ := Status(StatusOptions{StateDir: state, SystemStateDir: sysState, Warnf: func(string, ...any) {}})
		done <- got
	}()
	select {
	case got := <-done:
		if len(got) != 4 {
			t.Errorf("Status listed %d series; want 4: %+v", len(got), statusKeys(got))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Status waited on a held profileDir lock")
	}
	if after := statusSnapshot(t, state, sysState); !reflect.DeepEqual(after, before) {
		t.Errorf("state changed:\n before %q\n after  %q", before, after)
	}
}
