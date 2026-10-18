package nixprofile

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/yasunori0418/layat/internal/generation"
)

// layGenerations lays out <dir>/profile-<N>-link → <dir>/farm-<N> for each n and points the
// profile link at profile-<current>-link. It returns the profile link.
func layGenerations(t *testing.T, dir string, nums []int, current int) string {
	t.Helper()
	profileLink := filepath.Join(dir, "profile")
	for _, n := range nums {
		farm := filepath.Join(dir, "farm-"+strconv.Itoa(n))
		if err := os.Mkdir(farm, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(farm, filepath.Join(dir, "profile-"+strconv.Itoa(n)+"-link")); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("profile-"+strconv.Itoa(current)+"-link", profileLink); err != nil {
		t.Fatal(err)
	}
	return profileLink
}

// TestListScansGenerationLinks pins that List enumerates the generation links in ascending
// order, marks the one the profile link names as current, and ignores non-generation siblings.
func TestListScansGenerationLinks(t *testing.T) {
	dir := t.TempDir()
	profileLink := layGenerations(t, dir, []int{10, 2, 1}, 2)
	for _, name := range []string{"profile-x-link", "profile-3", "other-4-link", "profile.lock"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gens, err := New().List(profileLink)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var nums []int
	for _, g := range gens {
		nums = append(nums, g.Number)
		if g.Current != (g.Number == 2) {
			t.Errorf("gen %d Current = %v", g.Number, g.Current)
		}
	}
	if len(nums) != 3 || nums[0] != 1 || nums[1] != 2 || nums[2] != 10 {
		t.Errorf("numbers = %v, want [1 2 10]", nums)
	}
}

// TestListDateIsLinkMtime pins that a generation's date is its link's own mtime (lstat, not the
// link-farm's) formatted with the shared layout in local time.
func TestListDateIsLinkMtime(t *testing.T) {
	dir := t.TempDir()
	profileLink := layGenerations(t, dir, []int{1}, 1)
	old := time.Date(2001, 2, 3, 4, 5, 6, 0, time.Local)
	if err := os.Chtimes(filepath.Join(dir, "farm-1"), old, old); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(filepath.Join(dir, "profile-1-link"))
	if err != nil {
		t.Fatal(err)
	}
	gens, err := New().List(profileLink)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(gens) != 1 {
		t.Fatalf("len = %d, want 1", len(gens))
	}
	if want := fi.ModTime().Format(generation.DateLayout); gens[0].Date != want {
		t.Errorf("Date = %q, want the link's mtime %q", gens[0].Date, want)
	}
	if gens[0].Date == old.Format(generation.DateLayout) {
		t.Errorf("Date = %q is the link-farm's mtime", gens[0].Date)
	}
}

// TestListMissingProfileDir pins that a profileDir that does not exist yet holds no generations.
func TestListMissingProfileDir(t *testing.T) {
	gens, err := New().List(filepath.Join(t.TempDir(), "absent", "profile"))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(gens) != 0 {
		t.Errorf("len = %d, want 0", len(gens))
	}
}

// TestListReadDirError pins that a profileDir that cannot be read fails List instead of
// reporting no generations.
func TestListReadDirError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := New().List(filepath.Join(file, "profile")); err == nil {
		t.Fatal("expected an error for an unreadable profileDir, got nil")
	}
}

// TestBases pins the two bases the nix profile backend owns: the user state base under <state>
// and the fixed system base.
func TestBases(t *testing.T) {
	b := New()
	if got, want := b.UserBase("/state"), filepath.Join("/state", "nix", "profiles", "layat"); got != want {
		t.Errorf("UserBase = %q, want %q", got, want)
	}
	if got, want := b.SystemBase(), "/nix/var/nix/profiles/layat"; got != want {
		t.Errorf("SystemBase = %q, want %q", got, want)
	}
}
