package paths

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/yasunori0418/nput/internal/manifest"
)

func TestStateDirUsesXDGStateHome(t *testing.T) {
	// $XDG_STATE_HOME takes precedence regardless of $HOME.
	t.Setenv("XDG_STATE_HOME", "/xdg/state")
	t.Setenv("HOME", "/home/me")
	got, err := StateDir()
	if err != nil {
		t.Fatalf("StateDir() error = %v", err)
	}
	if got != "/xdg/state" {
		t.Errorf("StateDir() = %q, want %q", got, "/xdg/state")
	}
}

func TestStateDirFallsBackToHome(t *testing.T) {
	// Without $XDG_STATE_HOME, fall back to $HOME/.local/state.
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "/home/me")
	got, err := StateDir()
	if err != nil {
		t.Fatalf("StateDir() error = %v", err)
	}
	want := filepath.Join("/home/me", ".local", "state")
	if got != want {
		t.Errorf("StateDir() = %q, want %q", got, want)
	}
}

func TestStateDirErrorsWhenHomeUnresolvable(t *testing.T) {
	// With neither $XDG_STATE_HOME nor $HOME set, os.UserHomeDir fails and the
	// error is surfaced (no fallback path is returned).
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "")
	got, err := StateDir()
	if err == nil {
		t.Fatalf("StateDir() error = nil, want non-nil (got %q)", got)
	}
	if got != "" {
		t.Errorf("StateDir() = %q on error, want empty string", got)
	}
}

func TestGenerationLinkFormat(t *testing.T) {
	profileLink := filepath.Join("/state", "nix", "profiles", "nput", "vim", "profile")
	for _, gen := range []int{0, 1, 42} {
		got := GenerationLink(profileLink, gen)
		want := fmt.Sprintf("%s-%d-link", profileLink, gen)
		if got != want {
			t.Errorf("GenerationLink(%q, %d) = %q, want %q", profileLink, gen, got, want)
		}
	}
}

func TestRootHashDeterministicAndFixedLen(t *testing.T) {
	a := RootHash("/home/me/proj")
	b := RootHash("/home/me/proj")
	if a != b {
		t.Errorf("RootHash not deterministic: %q != %q", a, b)
	}
	if len(a) != rootHashLen {
		t.Errorf("RootHash len = %d, want %d", len(a), rootHashLen)
	}
	if c := RootHash("/home/me/other"); c == a {
		t.Errorf("distinct roots collided: %q", c)
	}
	// FS-safe (hex only).
	for _, r := range a {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			t.Errorf("RootHash has non-hex char %q in %q", r, a)
		}
	}
}

func TestResolveProjectUsesRootHash(t *testing.T) {
	state := "/state"
	root := "/home/me/proj"
	p := Resolve(state, "skills", manifest.RootKindProject, root, false)

	wantDir := filepath.Join(state, "nix", "profiles", "nput", RootHash(root), "skills")
	if p.Dir != wantDir {
		t.Errorf("Dir = %q, want %q", p.Dir, wantDir)
	}
	if p.Profile != filepath.Join(wantDir, "profile") {
		t.Errorf("Profile = %q", p.Profile)
	}
	if p.Pending != filepath.Join(wantDir, ".pending") {
		t.Errorf("Pending = %q", p.Pending)
	}
	wantBackref := filepath.Join(state, "nix", "profiles", "nput", RootHash(root), ".root")
	if p.Backref != wantBackref {
		t.Errorf("Backref = %q, want %q", p.Backref, wantBackref)
	}
}

func TestResolveHomeUsesNameKey(t *testing.T) {
	state := "/state"
	p := Resolve(state, "vim", manifest.RootKindHome, "/home/me", false)
	wantDir := filepath.Join(state, "nix", "profiles", "nput", "vim")
	if p.Dir != wantDir {
		t.Errorf("Dir = %q, want %q", p.Dir, wantDir)
	}
	if p.Backref != "" {
		t.Errorf("home (no --root) should have no backref, got %q", p.Backref)
	}
}

func TestResolveHomeWithOverrideUsesRootHash(t *testing.T) {
	// With --root explicit, even home uses the roothash key (→ ADR-0023).
	state := "/state"
	root := "/tmp/sandbox"
	p := Resolve(state, "vim", manifest.RootKindHome, root, true)
	wantDir := filepath.Join(state, "nix", "profiles", "nput", RootHash(root), "vim")
	if p.Dir != wantDir {
		t.Errorf("Dir = %q, want %q", p.Dir, wantDir)
	}
	if p.Backref == "" {
		t.Error("override should produce a backref")
	}
}

func TestResolveFixedUsesRootHash(t *testing.T) {
	// A fixed root (no --root) also uses the roothash key (→ ADR-0024).
	state := "/state"
	root := "/opt/x"
	p := Resolve(state, "c", manifest.RootKindFixed, root, false)
	if p.Backref == "" {
		t.Error("fixed root should produce a backref")
	}
	wantDir := filepath.Join(state, "nix", "profiles", "nput", RootHash(root), "c")
	if p.Dir != wantDir {
		t.Errorf("Dir = %q, want %q", p.Dir, wantDir)
	}
}

// writeBackref writes a .root under hashDir exactly as the engine does
// (the recorded root plus a trailing newline is the shape on disk).
func writeBackref(t *testing.T, hashDir, content string) {
	t.Helper()
	if err := os.MkdirAll(hashDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", hashDir, err)
	}
	if err := os.WriteFile(filepath.Join(hashDir, backrefName), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q/.root) error = %v", hashDir, err)
	}
}

func TestReadBackrefTrimsTrailingNewline(t *testing.T) {
	// The engine writes root + "\n"; the trailing newline must not survive
	// into the root path (it would break every stat of it).
	dir := t.TempDir()
	writeBackref(t, dir, "/home/me/proj\n")
	got, err := ReadBackref(dir)
	if err != nil {
		t.Fatalf("ReadBackref() error = %v", err)
	}
	if got != "/home/me/proj" {
		t.Errorf("ReadBackref() = %q, want %q", got, "/home/me/proj")
	}
}

func TestReadBackrefRejectsEmpty(t *testing.T) {
	// A .root holding nothing but whitespace decides no root path. The three
	// inputs take the same branch but are distinct equivalence classes, so an
	// implementation that only trims "\n" is caught by the third.
	for _, content := range []string{"", "\n", "  \n\t"} {
		t.Run(fmt.Sprintf("%q", content), func(t *testing.T) {
			dir := t.TempDir()
			writeBackref(t, dir, content)
			got, err := ReadBackref(dir)
			if err == nil {
				t.Fatalf("ReadBackref() error = nil, want non-nil (got %q)", got)
			}
			if got != "" {
				t.Errorf("ReadBackref() = %q on error, want empty string", got)
			}
		})
	}
}

func TestReadBackrefRejectsRelativePath(t *testing.T) {
	// A relative path must not reach os.Stat, where it would resolve against
	// the cwd and let an unrelated path decide the verdict.
	dir := t.TempDir()
	writeBackref(t, dir, "proj/sub\n")
	got, err := ReadBackref(dir)
	if err == nil {
		t.Fatalf("ReadBackref() error = nil, want non-nil (got %q)", got)
	}
	if got != "" {
		t.Errorf("ReadBackref() = %q on error, want empty string", got)
	}
}

func TestReadBackrefErrorsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	got, err := ReadBackref(dir)
	if err == nil {
		t.Fatalf("ReadBackref() error = nil, want non-nil (got %q)", got)
	}
	if got != "" {
		t.Errorf("ReadBackref() = %q on error, want empty string", got)
	}
}

// mkdirAll / writeFile keep the fixture building in the FS-backed tests short.
func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

// listRootHashSeriesFixture builds a base holding every shape the enumeration
// has to tell apart, and returns the result keyed by <roothash>.
//
//   - h1: a series with two <name> profiles — one holding only .pending (no
//     generation link at all), one holding a generation link — plus a stray
//     file directly under the series
//   - home-name: a <name>-keyed profileDir with no backref (home / system mode)
//   - h2: a series holding nothing but .root
//   - h3: a series whose .root records a relative path
func listRootHashSeriesFixture(t *testing.T) map[string]RootHashSeries {
	t.Helper()
	base := t.TempDir()

	h1 := filepath.Join(base, "h1")
	writeBackref(t, h1, "/home/me/proj\n")
	mkdirAll(t, filepath.Join(h1, "a"))
	writeFile(t, filepath.Join(h1, "a", ".pending"))
	mkdirAll(t, filepath.Join(h1, "b"))
	writeFile(t, filepath.Join(h1, "b", "profile-1-link"))
	// A non-dir entry directly under the series is not a <name> profile.
	writeFile(t, filepath.Join(h1, "stray"))

	mkdirAll(t, filepath.Join(base, "home-name"))

	writeBackref(t, filepath.Join(base, "h2"), "/home/me/other\n")
	writeBackref(t, filepath.Join(base, "h3"), "proj/sub\n")

	got, err := ListRootHashSeries(base)
	if err != nil {
		t.Fatalf("ListRootHashSeries() error = %v", err)
	}
	byHash := map[string]RootHashSeries{}
	for _, s := range got {
		if _, dup := byHash[s.RootHash]; dup {
			t.Errorf("duplicate series %q", s.RootHash)
		}
		byHash[s.RootHash] = s
	}
	return byHash
}

func TestListRootHashSeries(t *testing.T) {
	byHash := listRootHashSeriesFixture(t)

	t.Run("only .root-bearing dirs are series", func(t *testing.T) {
		if len(byHash) != 3 {
			t.Errorf("got %d series (%v), want 3", len(byHash), byHash)
		}
		if _, ok := byHash["home-name"]; ok {
			t.Error("a dir without .root must not be returned")
		}
	})

	t.Run("names are the <name> profileDirs", func(t *testing.T) {
		// Both <name> dirs come back regardless of what they hold, and the
		// stray file under the series does not.
		names := append([]string(nil), byHash["h1"].Names...)
		sort.Strings(names)
		if len(names) != 2 || names[0] != "a" || names[1] != "b" {
			t.Errorf("h1 Names = %v, want [a b]", byHash["h1"].Names)
		}
	})

	t.Run("root comes from the backref", func(t *testing.T) {
		s := byHash["h1"]
		if s.Root != "/home/me/proj" {
			t.Errorf("h1 Root = %q, want %q", s.Root, "/home/me/proj")
		}
		if s.BackrefErr != nil {
			t.Errorf("h1 BackrefErr = %v, want nil", s.BackrefErr)
		}
	})

	t.Run("a series of .root alone has no names", func(t *testing.T) {
		s := byHash["h2"]
		if s.Root != "/home/me/other" {
			t.Errorf("h2 Root = %q, want %q", s.Root, "/home/me/other")
		}
		if s.Names != nil {
			t.Errorf("h2 Names = %v, want nil", s.Names)
		}
		if s.BackrefErr != nil {
			t.Errorf("h2 BackrefErr = %v, want nil", s.BackrefErr)
		}
	})

	t.Run("an unreadable backref is reported, not dropped", func(t *testing.T) {
		s := byHash["h3"]
		if s.BackrefErr == nil {
			t.Error("h3 BackrefErr = nil, want non-nil (relative .root)")
		}
		if s.Root != "" {
			t.Errorf("h3 Root = %q, want empty string", s.Root)
		}
	})
}

func TestListRootHashSeriesWithBackrefDir(t *testing.T) {
	// A .root that is a directory still marks a series, but it is not a <name>
	// profile: letting it into Names would have the engine lock and remove it
	// as one, breaking the <name> → .root deletion order.
	base := t.TempDir()
	h := filepath.Join(base, "h")
	mkdirAll(t, filepath.Join(h, backrefName))
	mkdirAll(t, filepath.Join(h, "a"))

	got, err := ListRootHashSeries(base)
	if err != nil {
		t.Fatalf("ListRootHashSeries() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d series (%v), want 1", len(got), got)
	}
	if len(got[0].Names) != 1 || got[0].Names[0] != "a" {
		t.Errorf("Names = %v, want [a]", got[0].Names)
	}
	if got[0].BackrefErr == nil {
		t.Error("BackrefErr = nil, want non-nil (.root is a directory)")
	}
}

func TestListRootHashSeriesReportsUnstatableBackref(t *testing.T) {
	// A directory the backref cannot even be stat'd under is neither "no backref"
	// nor a decided one, so it must come back carrying the reason — and the
	// healthy series beside it must still be listed rather than lost with it.
	// The only induction available is dropping traversal on the directory, since
	// a non-directory entry never reaches the backref check at all.
	if os.Geteuid() == 0 {
		t.Skip("running as root: a non-traversable directory is still traversable")
	}
	base := t.TempDir()
	broken := filepath.Join(base, "broken")
	writeBackref(t, broken, "/home/me/gone\n")
	if err := os.Chmod(broken, 0o000); err != nil {
		t.Fatalf("Chmod error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(broken, 0o755) })
	writeBackref(t, filepath.Join(base, "ok"), "/home/me/proj\n")

	got, err := ListRootHashSeries(base)
	if err != nil {
		t.Fatalf("ListRootHashSeries() error = %v", err)
	}
	byHash := map[string]RootHashSeries{}
	for _, s := range got {
		byHash[s.RootHash] = s
	}
	if len(byHash) != 2 {
		t.Fatalf("got %d series (%v), want 2 (the healthy one must survive)", len(byHash), byHash)
	}
	broke := byHash["broken"]
	if broke.BackrefErr == nil {
		t.Error("broken BackrefErr = nil, want non-nil")
	}
	// A directory that cannot be traversed cannot be listed either, so both
	// reasons are set — neither one short-circuits the other.
	if broke.NamesErr == nil {
		t.Error("broken NamesErr = nil, want non-nil")
	}
	if broke.Names != nil {
		t.Errorf("broken Names = %v, want nil", broke.Names)
	}
	if byHash["ok"].Root != "/home/me/proj" {
		t.Errorf("ok Root = %q, want %q", byHash["ok"].Root, "/home/me/proj")
	}
}

func TestListRootHashSeriesReportsUnlistableSeries(t *testing.T) {
	// A series whose contents cannot be listed is still reported, with Names left
	// nil and NamesErr saying why. Unlike the base and the backref, this failure
	// has no ENOTDIR construction — a readable backref needs a traversable series
	// directory — so it is induced by dropping read permission on it.
	if os.Geteuid() == 0 {
		t.Skip("running as root: an unreadable directory is still listable")
	}
	base := t.TempDir()
	h := filepath.Join(base, "h")
	writeBackref(t, h, "/home/me/proj\n")
	if err := os.Chmod(h, 0o300); err != nil { // -wx: traversable, not listable
		t.Fatalf("Chmod error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(h, 0o755) })

	got, err := ListRootHashSeries(base)
	if err != nil {
		t.Fatalf("ListRootHashSeries() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d series (%v), want 1", len(got), got)
	}
	if got[0].NamesErr == nil {
		t.Error("NamesErr = nil, want non-nil")
	}
	if got[0].Names != nil {
		t.Errorf("Names = %v, want nil", got[0].Names)
	}
	// The backref is still readable through the traversable directory, so the
	// root is decided even though the <name> profiles are not known.
	if got[0].Root != "/home/me/proj" {
		t.Errorf("Root = %q, want %q", got[0].Root, "/home/me/proj")
	}
}

func TestListRootHashSeriesOnMissingBase(t *testing.T) {
	// A base that was never created comes back as an fs.ErrNotExist error, not
	// an empty listing: whether that is normal is the caller's call, and it has
	// to stay distinguishable from a base that could not be listed.
	got, err := ListRootHashSeries(filepath.Join(t.TempDir(), "absent"))
	if err == nil {
		t.Fatalf("ListRootHashSeries() error = nil, want non-nil (got %v)", got)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ListRootHashSeries() error = %v, want a not-exist error", err)
	}
}

func TestListRootHashSeriesOnUnlistableBase(t *testing.T) {
	// A base that exists but cannot be listed (here a regular file, so ENOTDIR)
	// must not look like a missing one — the caller reports this and stays
	// silent about the missing case.
	base := filepath.Join(t.TempDir(), "notadir")
	writeFile(t, base)

	got, err := ListRootHashSeries(base)
	if err == nil {
		t.Fatalf("ListRootHashSeries() error = nil, want non-nil (got %v)", got)
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ListRootHashSeries() error = %v, want an error that is not fs.ErrNotExist", err)
	}
}

func TestListRootHashSeriesIgnoresFiles(t *testing.T) {
	// Only directories can be a series; a stray file under the base is skipped.
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "stray"))
	got, err := ListRootHashSeries(base)
	if err != nil {
		t.Fatalf("ListRootHashSeries() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListRootHashSeries() = %v, want empty", got)
	}
}
