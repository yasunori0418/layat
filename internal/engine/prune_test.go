package engine

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yasunori0418/nput/internal/lock"
	"github.com/yasunori0418/nput/internal/paths"
)

// --- fixture helpers -------------------------------------------------------
//
// The layout Prune walks is built by hand rather than by running Apply: the
// verdict only reads the backref .root and the root's existence, so a fixture
// made of directories and files covers every branch and keeps a series whose
// root never existed expressible (→ TC-bc653f83-70ce-4920-b7cd-9b70a4cd2cae).

// pruneBases returns a state dir (whose paths.Base is created) and a system
// base (created as-is), both under tmpdirs. The system base is a tmpdir so no
// test ever reaches the real /nix/var/nix/profiles (→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1).
func pruneBases(t *testing.T) (stateDir, systemDir string) {
	t.Helper()
	stateDir = realTempDir(t)
	if err := os.MkdirAll(paths.Base(stateDir), 0o755); err != nil {
		t.Fatal(err)
	}
	systemDir = filepath.Join(realTempDir(t), "system")
	if err := os.MkdirAll(systemDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return stateDir, systemDir
}

// seriesSpec describes one <roothash> series to lay down under a base.
type seriesSpec struct {
	hash    string
	backref string   // content written to .root ("" = no .root file at all)
	names   []string // <name> profileDirs to create under the series
	pending []string // names that get a .pending entry
	gens    []string // names that get a profile + profile-1-link generation link
}

// writeSeries lays down one series under base and returns the series directory.
func writeSeries(t *testing.T, base string, s seriesSpec) string {
	t.Helper()
	hashDir := filepath.Join(base, s.hash)
	if err := os.MkdirAll(hashDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range s.names {
		if err := os.MkdirAll(filepath.Join(hashDir, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range s.pending {
		if err := os.Symlink(filepath.Join(hashDir, n, "pending-target"), filepath.Join(hashDir, n, ".pending")); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range s.gens {
		gen := filepath.Join(hashDir, n, "profile-1-link")
		if err := os.MkdirAll(gen, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(gen, filepath.Join(hashDir, n, "profile")); err != nil {
			t.Fatal(err)
		}
	}
	if s.backref != "" {
		// Written with the trailing newline the engine writes (→ engine.go).
		if err := os.WriteFile(filepath.Join(hashDir, ".root"), []byte(s.backref+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return hashDir
}

// orphanSeries lays down a series whose recorded root does not exist.
func orphanSeries(t *testing.T, base, hash string, names ...string) string {
	t.Helper()
	root := filepath.Join(realTempDir(t), "gone")
	return writeSeries(t, base, seriesSpec{
		hash:    hash,
		backref: root,
		names:   names,
		pending: names,
	})
}

// liveSeries lays down a series whose recorded root exists.
func liveSeries(t *testing.T, base, hash string, names ...string) (hashDir, root string) {
	t.Helper()
	root = realTempDir(t)
	hashDir = writeSeries(t, base, seriesSpec{
		hash:    hash,
		backref: root,
		names:   names,
		gens:    names,
	})
	return hashDir, root
}

// warnRecorder collects the warnings Prune emits.
type warnRecorder struct {
	mu   sync.Mutex
	msgs []string
}

func (w *warnRecorder) warnf(format string, args ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.msgs = append(w.msgs, format)
	_ = args
}

func (w *warnRecorder) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.msgs)
}

// pruneOpts builds options driving both tmpdir bases with a warning recorder.
func pruneOpts(stateDir, systemDir string, w *warnRecorder) PruneOptions {
	return PruneOptions{StateDir: stateDir, SystemDir: systemDir, Warnf: w.warnf}
}

func hashes(series []PruneSeries) []string {
	out := make([]string, 0, len(series))
	for _, s := range series {
		out = append(out, s.RootHash)
	}
	return out
}

func skipHashes(skipped []PruneSkipped) []string {
	out := make([]string, 0, len(skipped))
	for _, s := range skipped {
		out = append(out, s.Series.RootHash)
	}
	return out
}

// findSkipped returns the skip record for a series, or a zero value.
func findSkipped(t *testing.T, res *PruneResult, hash string) PruneSkipped {
	t.Helper()
	for _, s := range res.Skipped {
		if s.Series.RootHash == hash {
			return s
		}
	}
	t.Fatalf("series %q not in Skipped (got %v)", hash, skipHashes(res.Skipped))
	return PruneSkipped{}
}

func mustNotExist(t *testing.T, path, what string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("%s should be gone (%s), lstat err = %v", what, path, err)
	}
}

func mustExist(t *testing.T, path, what string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("%s should remain (%s), lstat err = %v", what, path, err)
	}
}

// --- target accuracy (→ TC-bc653f83-70ce-4920-b7cd-9b70a4cd2cae) -----------

func TestPruneRemovesOrphanSeriesWhole(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg", "other")
	// A generation link under one of the names, so "everything under the
	// series is gone" covers generation links too.
	gen := filepath.Join(hashDir, "cfg", "profile-1-link")
	if err := os.MkdirAll(gen, 0o755); err != nil {
		t.Fatal(err)
	}

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustNotExist(t, hashDir, "the orphan series directory")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa]", got)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty", skipHashes(res.Skipped))
	}
	names := res.Removed[0].Names
	if len(names) != 2 {
		t.Errorf("Removed[0].Names = %v, want the two <name> profiles", names)
	}
	if res.Removed[0].Root == "" {
		t.Error("Removed[0].Root is empty, want the recorded root path")
	}
}

func TestPruneRemovesSeriesWithoutGenerations(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// The majority shape in the wild: .pending out-links only, no generation
	// link anywhere (→ DSG-096dc893 「世代を持たない系列も対象」).
	pendingOnly := writeSeries(t, base, seriesSpec{
		hash:    "bbbb",
		backref: filepath.Join(realTempDir(t), "gone"),
		names:   []string{"cfg"},
		pending: []string{"cfg"},
	})
	// A series holding nothing but the backref.
	bareRoot := writeSeries(t, base, seriesSpec{
		hash:    "cccc",
		backref: filepath.Join(realTempDir(t), "gone"),
	})

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustNotExist(t, pendingOnly, "the .pending-only series")
	mustNotExist(t, bareRoot, "the .root-only series")
	if got := hashes(res.Removed); len(got) != 2 {
		t.Errorf("Removed = %v, want both series", got)
	}
}

func TestPruneKeepsLiveSeries(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir, root := liveSeries(t, base, "aaaa", "cfg")
	// A file placed under the live root: prune must not touch placed artifacts.
	placed := filepath.Join(root, "placed")
	if err := os.WriteFile(placed, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the live series")
	mustExist(t, placed, "the placed artifact under the live root")
	if len(res.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", hashes(res.Removed))
	}
	// A live series is not a failure, so it is not reported as skipped either.
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty", skipHashes(res.Skipped))
	}
	if w.count() != 0 {
		t.Errorf("Warnf called %d times, want 0", w.count())
	}
}

func TestPruneIgnoresDirWithoutBackref(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// The <name>-keyed profileDir of home mode: no .root, so no root path can
	// be derived and the directory is structurally out of scope.
	homeDir := writeSeries(t, base, seriesSpec{hash: "cfg", names: []string{"profile-1-link"}})
	// The same shape under the system base.
	sysDir := writeSeries(t, system, seriesSpec{hash: "cfg", names: []string{"profile-1-link"}})

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, homeDir, "the <name>-keyed dir under the state base")
	mustExist(t, sysDir, "the <name>-keyed dir under the system base")
	if len(res.Removed) != 0 || len(res.Skipped) != 0 {
		t.Errorf("Removed = %v, Skipped = %v, want both empty", hashes(res.Removed), skipHashes(res.Skipped))
	}
}

func TestPruneScansBothBases(t *testing.T) {
	state, system := pruneBases(t)
	stateOrphan := orphanSeries(t, paths.Base(state), "aaaa", "cfg")
	systemOrphan := orphanSeries(t, system, "bbbb", "cfg")

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustNotExist(t, stateOrphan, "the orphan under the state base")
	mustNotExist(t, systemOrphan, "the orphan under the system base")
	if got := hashes(res.Removed); len(got) != 2 {
		t.Errorf("Removed = %v, want one series from each base", got)
	}
}

func TestPruneTreatsMissingBaseAsEmpty(t *testing.T) {
	// Neither base exists: normal for an environment that never used system
	// mode, so it is not an error and produces no warning.
	state := realTempDir(t)
	system := filepath.Join(realTempDir(t), "absent")

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 0 || len(res.Skipped) != 0 {
		t.Errorf("Removed = %v, Skipped = %v, want both empty", hashes(res.Removed), skipHashes(res.Skipped))
	}
	if w.count() != 0 {
		t.Errorf("Warnf called %d times for absent bases, want 0", w.count())
	}
}

func TestPruneWarnsOnUnlistableBase(t *testing.T) {
	state, system := pruneBases(t)
	stateOrphan := orphanSeries(t, paths.Base(state), "aaaa", "cfg")
	// A regular file where the system base should be: ReadDir fails with
	// ENOTDIR, which root cannot bypass either (→ TP-deb05610 の横断規約).
	if err := os.RemoveAll(system); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(system, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	// The other base is still processed.
	mustNotExist(t, stateOrphan, "the orphan under the readable base")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa]", got)
	}
	if w.count() == 0 {
		t.Error("Warnf not called for the unlistable base, want a warning naming it")
	}
	// A base-level failure names no series, so it cannot be a Skipped entry.
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty (a base failure names no series)", skipHashes(res.Skipped))
	}
}

func TestPruneKeepsSeriesWithUnusableBackref(t *testing.T) {
	// Empty, whitespace-only and relative backrefs are separate branches in
	// ReadBackref; an implementation checking only the ReadFile error passes
	// the last two, so each gets its own case.
	cases := []struct {
		name    string
		backref string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"relative", "relative/path"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, system := pruneBases(t)
			base := paths.Base(state)
			hashDir := filepath.Join(base, "aaaa")
			if err := os.MkdirAll(filepath.Join(hashDir, "cfg"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(hashDir, ".root"), []byte(tc.backref), 0o644); err != nil {
				t.Fatal(err)
			}

			var w warnRecorder
			res, err := Prune(pruneOpts(state, system, &w))
			if err != nil {
				t.Fatalf("Prune: %v", err)
			}

			mustExist(t, hashDir, "the series with an unusable backref")
			if got := findSkipped(t, res, "aaaa").Reason; got != PruneSkipBackrefUnreadable {
				t.Errorf("Reason = %q, want %q", got, PruneSkipBackrefUnreadable)
			}
			if w.count() == 0 {
				t.Error("Warnf not called for the skipped series")
			}
		})
	}
}

func TestPruneTrimsTrailingNewlineInBackref(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	root := realTempDir(t)
	// The engine writes root + "\n"; a verdict that does not trim reports a
	// live root as root-stat-failed and stops pruning genuine orphans.
	hashDir := filepath.Join(base, "aaaa")
	if err := os.MkdirAll(filepath.Join(hashDir, "cfg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hashDir, ".root"), []byte(root+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the series whose root exists")
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty (the backref is usable once trimmed)", skipHashes(res.Skipped))
	}
	if len(res.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", hashes(res.Removed))
	}
}

func TestPruneKeepsSeriesWithUnstatableRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	// The root sits under a directory stripped of search permission, so its
	// stat fails with something other than ErrNotExist.
	parent := filepath.Join(realTempDir(t), "sealed")
	root := filepath.Join(parent, "proj")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
	hashDir := writeSeries(t, base, seriesSpec{hash: "aaaa", backref: root, names: []string{"cfg"}})

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the series whose root cannot be statted")
	if got := findSkipped(t, res, "aaaa").Reason; got != PruneSkipRootStatFailed {
		t.Errorf("Reason = %q, want %q", got, PruneSkipRootStatFailed)
	}
	if w.count() == 0 {
		t.Error("Warnf not called for the skipped series")
	}
}

func TestPruneRemovesSeriesWithDanglingSymlinkRoot(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// The recorded root is reachable only through a symlink whose target is
	// gone: os.Stat follows it, so it falls out as not existing.
	link := filepath.Join(realTempDir(t), "link")
	if err := os.Symlink(filepath.Join(realTempDir(t), "gone"), link); err != nil {
		t.Fatal(err)
	}
	hashDir := writeSeries(t, base, seriesSpec{hash: "aaaa", backref: link, names: []string{"cfg"}})

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustNotExist(t, hashDir, "the series whose root is a dangling symlink")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa]", got)
	}
}

func TestPruneKeepsSeriesWhoseRootIsAFile(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// The verdict is "does not exist" only; the kind is not looked at, so a
	// root replaced by a regular file falls on the existing side.
	root := filepath.Join(realTempDir(t), "was-a-dir")
	if err := os.WriteFile(root, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	hashDir := writeSeries(t, base, seriesSpec{hash: "aaaa", backref: root, names: []string{"cfg"}})

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the series whose root is now a regular file")
	if len(res.Removed) != 0 || len(res.Skipped) != 0 {
		t.Errorf("Removed = %v, Skipped = %v, want both empty", hashes(res.Removed), skipHashes(res.Skipped))
	}
}

func TestPruneKeepsSeriesWhoseNamesCannotBeListed(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	healthy := orphanSeries(t, base, "aaaa", "cfg")
	// The series directory is stripped of read permission after .root is
	// written: the backref is still statable through the path, but the <name>
	// profileDirs cannot be listed, so the set of lock keys is unknown.
	unlistable := writeSeries(t, base, seriesSpec{
		hash:    "bbbb",
		backref: filepath.Join(realTempDir(t), "gone"),
		names:   []string{"cfg"},
	})
	if err := os.Chmod(unlistable, 0o111); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unlistable, 0o755) })

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	// Names being nil must not be read as "a series with no <name>": that
	// path deletes without taking any lock.
	mustExist(t, unlistable, "the series whose names cannot be listed")
	if got := findSkipped(t, res, "bbbb").Reason; got != PruneSkipSeriesUnreadable {
		t.Errorf("Reason = %q, want %q", got, PruneSkipSeriesUnreadable)
	}
	if w.count() == 0 {
		t.Error("Warnf not called for the skipped series")
	}
	// One failing series does not stop the rest of the base.
	mustNotExist(t, healthy, "the healthy orphan series")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa]", got)
	}
}

// --- safety gates (→ TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9) --------------

func TestPruneDryRunChangesNothing(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg")

	var w warnRecorder
	opts := pruneOpts(state, system, &w)
	opts.DryRun = true
	res, err := Prune(opts)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the orphan series under dryrun")
	mustExist(t, filepath.Join(hashDir, "cfg"), "the <name> profileDir under dryrun")
	mustExist(t, filepath.Join(hashDir, ".root"), "the backref under dryrun")
	if !res.DryRun {
		t.Error("DryRun = false, want true")
	}
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want the planned series [aaaa]", got)
	}
}

func TestPruneDryRunTakesNoLock(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg")

	var w warnRecorder
	opts := pruneOpts(state, system, &w)
	opts.DryRun = true
	res, err := Prune(opts)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	// A dryrun that took the lock would block a real apply on the same series.
	l, err := lock.Acquire(filepath.Join(hashDir, "cfg"), false)
	if err != nil {
		t.Fatalf("lock.Acquire after dryrun: %v, want the lock to be free", err)
	}
	_ = l.Release()
	if got := hashes(res.Removed); len(got) != 1 {
		t.Errorf("Removed = %v, want the planned series", got)
	}
}

func TestPruneDryRunListsLockedSeries(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg")

	held, err := lock.Acquire(filepath.Join(hashDir, "cfg"), true)
	if err != nil {
		t.Fatalf("setup lock.Acquire: %v", err)
	}
	defer func() { _ = held.Release() }()

	var w warnRecorder
	opts := pruneOpts(state, system, &w)
	opts.DryRun = true
	res, err := Prune(opts)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	// Dryrun stops at the verdict, so a held lock does not change the preview.
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa] even while the series is locked", got)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty under dryrun", skipHashes(res.Skipped))
	}
}

func TestPruneConfirmReceivesRootPaths(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	first := orphanSeries(t, base, "aaaa", "cfg")
	second := orphanSeries(t, base, "bbbb", "cfg")

	var seen []PruneSeries
	var w warnRecorder
	opts := pruneOpts(state, system, &w)
	opts.Confirm = func(res *PruneResult) (bool, error) {
		seen = append([]PruneSeries(nil), res.Removed...)
		return true, nil
	}
	if _, err := Prune(opts); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustNotExist(t, first, "the first orphan series")
	mustNotExist(t, second, "the second orphan series")
	// The CLI builds its root path list from what Confirm receives, so every
	// candidate must carry a non-empty Root.
	if len(seen) != 2 {
		t.Fatalf("Confirm saw %d series, want 2", len(seen))
	}
	for _, s := range seen {
		if s.Root == "" {
			t.Errorf("Confirm saw series %q with an empty Root", s.RootHash)
		}
	}
}

func TestPruneConfirmFalseAborts(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg")

	var w warnRecorder
	opts := pruneOpts(state, system, &w)
	opts.Confirm = func(*PruneResult) (bool, error) { return false, nil }
	res, err := Prune(opts)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the orphan series after an aborted confirmation")
	if !res.Aborted {
		t.Error("Aborted = false, want true")
	}
}

func TestPruneConfirmErrorPropagates(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg")

	sentinel := errors.New("confirm failed")
	var w warnRecorder
	opts := pruneOpts(state, system, &w)
	opts.Confirm = func(*PruneResult) (bool, error) { return false, sentinel }
	if _, err := Prune(opts); !errors.Is(err, sentinel) {
		t.Errorf("Prune error = %v, want it to wrap the confirm error", err)
	}
	mustExist(t, hashDir, "the orphan series after a failed confirmation")
}

func TestPruneSkipsLockedSeries(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// Two <name> profileDirs, only one of them locked: the whole series is
	// skipped, all-or-nothing.
	locked := orphanSeries(t, base, "aaaa", "cfg", "other")
	free := orphanSeries(t, base, "bbbb", "cfg")

	held, err := lock.Acquire(filepath.Join(locked, "other"), true)
	if err != nil {
		t.Fatalf("setup lock.Acquire: %v", err)
	}
	defer func() { _ = held.Release() }()

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, locked, "the locked series")
	mustExist(t, filepath.Join(locked, "cfg"), "the unlocked <name> of the locked series")
	if got := findSkipped(t, res, "aaaa").Reason; got != PruneSkipLocked {
		t.Errorf("Reason = %q, want %q", got, PruneSkipLocked)
	}
	if w.count() == 0 {
		t.Error("Warnf not called for the locked series")
	}
	// One locked series does not stop the others.
	mustNotExist(t, free, "the unlocked orphan series")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "bbbb" {
		t.Errorf("Removed = %v, want [bbbb]", got)
	}
}

func TestPruneReleasesLocksTakenBeforeASkip(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// "cfg" sorts before "other", so the lock on "cfg" is taken and must be
	// released once "other" turns out to be held.
	hashDir := orphanSeries(t, base, "aaaa", "cfg", "other")
	held, err := lock.Acquire(filepath.Join(hashDir, "other"), true)
	if err != nil {
		t.Fatalf("setup lock.Acquire: %v", err)
	}
	defer func() { _ = held.Release() }()

	var w warnRecorder
	if _, err := Prune(pruneOpts(state, system, &w)); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	l, err := lock.Acquire(filepath.Join(hashDir, "cfg"), false)
	if err != nil {
		t.Fatalf("lock.Acquire on cfg: %v, want the partial lock to have been released", err)
	}
	_ = l.Release()
}

func TestPruneSkipsSeriesWhoseLockCannotBeAcquired(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	hashDir := orphanSeries(t, base, "aaaa", "cfg")
	// lock.Acquire opens the profileDir; with no read permission the open
	// fails with something other than ErrLocked, and an undecidable lock must
	// not fall through to deletion.
	if err := os.Chmod(filepath.Join(hashDir, "cfg"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(hashDir, "cfg"), 0o755) })

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, hashDir, "the series whose lock could not be acquired")
	if got := findSkipped(t, res, "aaaa").Reason; got != PruneSkipLocked {
		t.Errorf("Reason = %q, want %q", got, PruneSkipLocked)
	}
	if w.count() == 0 {
		t.Error("Warnf not called for the skipped series")
	}
}

func TestPruneHoldsLocksWhileDeleting(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// Two <name> profileDirs: the deletion window is the interval in which the
	// first is already gone and the second is not. Probing only inside that
	// interval is what makes this deterministic — a probe that ran before
	// Prune took its locks would find them free for a reason that is not a
	// defect.
	hashDir := orphanSeries(t, base, "aaaa", "first", "second")
	first := filepath.Join(hashDir, "first")
	second := filepath.Join(hashDir, "second")

	// An implementation that releases before removing would let another
	// process take the lock on the <name> still standing (→ the same harm as
	// RISK-2b17fefb-6e92-4513-9e7c-de21897c9cfe).
	tookLockMidDeletion := make(chan bool, 1)
	go func() {
		for {
			// Wait for the window to open: the first <name> gone.
			if _, err := os.Lstat(first); !errors.Is(err, fs.ErrNotExist) {
				continue
			}
			// The window closes when the second <name> goes too.
			if _, err := os.Lstat(second); errors.Is(err, fs.ErrNotExist) {
				tookLockMidDeletion <- false
				return
			}
			if l, err := lock.Acquire(second, false); err == nil {
				_ = l.Release()
				tookLockMidDeletion <- true
				return
			}
		}
	}()

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	if <-tookLockMidDeletion {
		t.Error("another holder took the lock on a <name> still standing, want the locks held through the deletion")
	}
	mustNotExist(t, hashDir, "the orphan series")
	if got := hashes(res.Removed); len(got) != 1 {
		t.Errorf("Removed = %v, want the deleted series", got)
	}
}

func TestPruneReleasesLocksAfterDeleting(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// The series is deleted, so the profileDir is gone and the lock cannot be
	// probed through it. Recreate the path and check the lock is takeable:
	// a leaked flock lives on the open fd, which Release closes.
	hashDir := orphanSeries(t, base, "aaaa", "cfg")
	profileDir := filepath.Join(hashDir, "cfg")

	var w warnRecorder
	if _, err := Prune(pruneOpts(state, system, &w)); err != nil {
		t.Fatalf("Prune: %v", err)
	}
	mustNotExist(t, profileDir, "the <name> profileDir")

	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	l, err := lock.Acquire(profileDir, false)
	if err != nil {
		t.Fatalf("lock.Acquire after prune: %v, want a takeable lock", err)
	}
	_ = l.Release()
}

func TestPruneRemovesBareSeriesWithoutLock(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	// No <name> profileDir means no lock key exists; the series is still
	// removed rather than skipped for want of a lock.
	hashDir := writeSeries(t, base, seriesSpec{hash: "aaaa", backref: filepath.Join(realTempDir(t), "gone")})

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustNotExist(t, hashDir, "the .root-only series")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa]", got)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty", skipHashes(res.Skipped))
	}
}

func TestPruneSkipsSeriesItCannotBeginToDelete(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	blocked := orphanSeries(t, base, "aaaa", "cfg")
	free := orphanSeries(t, base, "bbbb", "cfg")
	// The series directory is read-execute only: its entries can be listed
	// and locked, but nothing inside it can be unlinked, so the very first
	// removal fails with a permission error and nothing has been deleted.
	if err := os.Chmod(blocked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, 0o755) })

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	mustExist(t, blocked, "the series that could not be deleted")
	mustExist(t, filepath.Join(blocked, "cfg"), "the <name> profileDir of the blocked series")
	if got := findSkipped(t, res, "aaaa").Reason; got != PruneSkipPermissionDenied {
		t.Errorf("Reason = %q, want %q", got, PruneSkipPermissionDenied)
	}
	if w.count() == 0 {
		t.Error("Warnf not called for the skipped series")
	}
	// Other series keep going.
	mustNotExist(t, free, "the deletable orphan series")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "bbbb" {
		t.Errorf("Removed = %v, want [bbbb]", got)
	}
}

func TestPruneErrorsWhenDeletionFailsForANonPermissionReason(t *testing.T) {
	state, system := pruneBases(t)
	base := paths.Base(state)
	completed := orphanSeries(t, base, "aaaa", "cfg")
	// A stray regular file in the series directory: every <name> and .root is
	// removed, and the final rmdir of the <roothash> then fails with
	// ENOTEMPTY. That is not a permission error, so it must not be folded
	// into Skipped{permission-denied} — an implementation that skips on any
	// pre-progress failure, or one that maps every failure to the permission
	// reason, would report "fix the permissions and it will go" for a series
	// that will never go. root cannot bypass ENOTEMPTY, so no Geteuid guard.
	partial := writeSeries(t, base, seriesSpec{
		hash:    "bbbb",
		backref: filepath.Join(realTempDir(t), "gone"),
		names:   []string{"cfg"},
	})
	if err := os.WriteFile(filepath.Join(partial, "stray"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err == nil {
		t.Fatal("Prune error = nil, want an error naming the series that could not be removed")
	}
	if errors.Is(err, fs.ErrPermission) {
		t.Errorf("Prune error = %v, want a non-permission error", err)
	}
	if !strings.Contains(err.Error(), "bbbb") {
		t.Errorf("Prune error = %v, want it to name the series bbbb", err)
	}
	if res == nil {
		t.Fatal("Prune returned a nil result alongside the error, want the partial result")
	}
	for _, s := range res.Skipped {
		if s.Series.RootHash == "bbbb" {
			t.Errorf("Skipped holds the series that failed to be removed (%q), want an error only", s.Reason)
		}
	}
	// The series completed before the failure is still reported as removed.
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want the series completed before the failure", got)
	}
	mustNotExist(t, completed, "the series completed before the failure")
	mustExist(t, partial, "the series that could not be removed")
}

func TestPruneErrorsWhenDeletionFailsPartway(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	// "aaaa" sorts before "bbbb", so the first series completes and lands in
	// Removed before the second one fails partway.
	completed := orphanSeries(t, base, "aaaa", "cfg")
	partial := writeSeries(t, base, seriesSpec{
		hash:    "bbbb",
		backref: filepath.Join(realTempDir(t), "gone"),
		names:   []string{"first", "second"},
	})
	// "first" is removable; "second" holds an entry that cannot be unlinked
	// because "second" itself is read-only, so its RemoveAll fails after
	// "first" is already gone.
	inner := filepath.Join(partial, "second", "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(partial, "second"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(partial, "second"), 0o755) })

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err == nil {
		t.Fatal("Prune error = nil, want an error naming the half-deleted series")
	}
	if !strings.Contains(err.Error(), "bbbb") {
		t.Errorf("Prune error = %v, want it to name the series bbbb", err)
	}
	if res == nil {
		t.Fatal("Prune returned a nil result alongside the error, want the partial result")
	}

	// The half-deleted series is neither removed nor skipped.
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want the series completed before the failure", got)
	}
	for _, s := range res.Skipped {
		if s.Series.RootHash == "bbbb" {
			t.Errorf("Skipped holds the half-deleted series bbbb (%q), want it reported as an error only", s.Reason)
		}
	}
	mustNotExist(t, completed, "the series completed before the failure")
	// .root survives so the series is still reachable on a re-run.
	mustExist(t, filepath.Join(partial, ".root"), "the backref of the failed series")
	mustNotExist(t, filepath.Join(partial, "first"), "the <name> removed before the failure")
}

func TestPruneErrorsWhenDeletionFailsPartwayOnPermission(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	// Same shape as above, driven to make explicit that a permission failure
	// after something was already removed is an error, not
	// Skipped{permission-denied}: an implementation checking errors.Is before
	// the progress state would pass every other cell.
	partial := writeSeries(t, base, seriesSpec{
		hash:    "aaaa",
		backref: filepath.Join(realTempDir(t), "gone"),
		names:   []string{"first", "second"},
	})
	inner := filepath.Join(partial, "second", "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(partial, "second"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(partial, "second"), 0o755) })

	var w warnRecorder
	res, err := Prune(pruneOpts(state, system, &w))
	if err == nil {
		t.Fatal("Prune error = nil, want an error for the half-deleted series")
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Prune error = %v, want it to wrap a permission error", err)
	}
	for _, s := range res.Skipped {
		if s.Series.RootHash == "aaaa" {
			t.Errorf("Skipped holds the half-deleted series (%q), want an error instead", s.Reason)
		}
	}
	if len(res.Removed) != 0 {
		t.Errorf("Removed = %v, want empty (the only series failed partway)", hashes(res.Removed))
	}
}

func TestPruneRemovesNamesBeforeBackref(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the permission denial this case needs")
	}
	state, system := pruneBases(t)
	base := paths.Base(state)
	// A series that fails partway must still be reachable afterwards, which
	// only holds when .root outlives the <name> profiles.
	partial := writeSeries(t, base, seriesSpec{
		hash:    "aaaa",
		backref: filepath.Join(realTempDir(t), "gone"),
		names:   []string{"first", "second"},
	})
	if err := os.MkdirAll(filepath.Join(partial, "second", "inner"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(partial, "second"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(partial, "second"), 0o755) })

	var w warnRecorder
	if _, err := Prune(pruneOpts(state, system, &w)); err == nil {
		t.Fatal("Prune error = nil, want the partial failure")
	}

	mustExist(t, filepath.Join(partial, ".root"), "the backref of the failed series")

	// A second run finds the same series again — the enumeration keys on .root.
	if err := os.Chmod(filepath.Join(partial, "second"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Prune(pruneOpts(state, system, &w))
	if err != nil {
		t.Fatalf("second Prune: %v", err)
	}
	mustNotExist(t, partial, "the series on the re-run")
	if got := hashes(res.Removed); len(got) != 1 || got[0] != "aaaa" {
		t.Errorf("Removed = %v, want [aaaa] on the re-run", got)
	}
}
