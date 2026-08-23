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
	if err := os.WriteFile(filepath.Join(hashDir, ".root"), []byte(content), 0o644); err != nil {
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
	// A .root holding nothing but whitespace decides no root path.
	for _, content := range []string{"", "\n", "  \n\t"} {
		dir := t.TempDir()
		writeBackref(t, dir, content)
		got, err := ReadBackref(dir)
		if err == nil {
			t.Errorf("ReadBackref() with %q error = nil, want non-nil (got %q)", content, got)
		}
		if got != "" {
			t.Errorf("ReadBackref() with %q = %q on error, want empty string", content, got)
		}
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

func TestListRootHashSeries(t *testing.T) {
	// A base holding: a full series (two <name> profiles), a <name>-keyed dir
	// without .root (home / system mode), a series with .root alone, and a
	// series whose .root is empty.
	base := t.TempDir()

	h1 := filepath.Join(base, "h1")
	writeBackref(t, h1, "/home/me/proj\n")
	if err := os.MkdirAll(filepath.Join(h1, "a"), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(h1, "a", ".pending"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(h1, "b"), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(h1, "b", "profile-1-link"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// home / system mode keys on <name> directly and has no .root.
	if err := os.MkdirAll(filepath.Join(base, "home-name"), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	h2 := filepath.Join(base, "h2")
	writeBackref(t, h2, "/home/me/other\n")

	h3 := filepath.Join(base, "h3")
	writeBackref(t, h3, "\n")

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
	if len(byHash) != 3 {
		t.Fatalf("ListRootHashSeries() returned %d series (%v), want 3", len(byHash), byHash)
	}
	if _, ok := byHash["home-name"]; ok {
		t.Error("a dir without .root must not be returned")
	}

	s1 := byHash["h1"]
	if s1.Root != "/home/me/proj" {
		t.Errorf("h1 Root = %q, want %q", s1.Root, "/home/me/proj")
	}
	if s1.BackrefErr != nil {
		t.Errorf("h1 BackrefErr = %v, want nil", s1.BackrefErr)
	}
	names := append([]string(nil), s1.Names...)
	sort.Strings(names)
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("h1 Names = %v, want [a b]", s1.Names)
	}

	s2 := byHash["h2"]
	if s2.Root != "/home/me/other" {
		t.Errorf("h2 Root = %q, want %q", s2.Root, "/home/me/other")
	}
	if len(s2.Names) != 0 {
		t.Errorf("h2 Names = %v, want empty", s2.Names)
	}
	if s2.BackrefErr != nil {
		t.Errorf("h2 BackrefErr = %v, want nil", s2.BackrefErr)
	}

	// An unreadable backref is reported on the series, not dropped from it.
	s3 := byHash["h3"]
	if s3.BackrefErr == nil {
		t.Error("h3 BackrefErr = nil, want non-nil (empty .root)")
	}
	if s3.Root != "" {
		t.Errorf("h3 Root = %q, want empty string", s3.Root)
	}
}

func TestListRootHashSeriesOnMissingBase(t *testing.T) {
	// A base that was never created holds no series; that is not an error here
	// (the caller decides whether a missing base is normal).
	got, err := ListRootHashSeries(filepath.Join(t.TempDir(), "absent"))
	if err == nil {
		t.Fatalf("ListRootHashSeries() error = nil, want non-nil (got %v)", got)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ListRootHashSeries() error = %v, want a not-exist error", err)
	}
}

func TestListRootHashSeriesIgnoresFiles(t *testing.T) {
	// Only directories can be a series; a stray file under the base is skipped.
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "stray"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	got, err := ListRootHashSeries(base)
	if err != nil {
		t.Fatalf("ListRootHashSeries() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListRootHashSeries() = %v, want empty", got)
	}
}
