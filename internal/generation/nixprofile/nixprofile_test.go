package nixprofile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
// reporting no generations, and names the op and the profileDir.
func TestListReadDirError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := New().List(filepath.Join(file, "profile"))
	if err == nil {
		t.Fatal("expected an error for an unreadable profileDir, got nil")
	}
	if want := "cannot list the generations: "; !strings.HasPrefix(err.Error(), want) || !strings.Contains(err.Error(), file) {
		t.Errorf("error %q, want prefix %q and the profileDir %s", err, want, file)
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

// fakeNixEnv puts a nix-env at the head of PATH that writes stderr and exits 1.
func fakeNixEnv(t *testing.T, stderr string) {
	t.Helper()
	bin := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' '" + stderr + "' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "nix-env"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// TestCommandErrorCapturesStderr pins that a failed nix-env comes back as a CommandError
// naming the op and the profile, with nix-env's stderr in Stderr and not in Error().
func TestCommandErrorCapturesStderr(t *testing.T) {
	const diag = "error: fake nix-env failure"
	fakeNixEnv(t, diag)
	profile := filepath.Join(t.TempDir(), "profile")

	cases := []struct {
		name string
		run  func() error
		want string
	}{
		{"commit", func() error { return New().Commit(profile, "/farm") },
			"cannot run nix-env --set (" + profile + "): exit status 1"},
		{"switch", func() error { return New().Switch(profile, 3) },
			"cannot run nix-env --switch-generation 3 (" + profile + "): exit status 1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.run()
			var ce *CommandError
			if !errors.As(err, &ce) {
				t.Fatalf("got %v, want a CommandError", err)
			}
			if ce.Profile != profile {
				t.Errorf("Profile = %q, want %q", ce.Profile, profile)
			}
			if !strings.Contains(ce.Stderr, diag) {
				t.Errorf("Stderr = %q, want it to contain %q", ce.Stderr, diag)
			}
			if got := err.Error(); got != c.want {
				t.Errorf("Error() = %q, want %q", got, c.want)
			}
			var ee *exec.ExitError
			if !errors.As(err, &ee) {
				t.Errorf("%v does not unwrap to *exec.ExitError", err)
			}
		})
	}
}

// TestNixEnvNotOnPath pins that a missing nix-env is reported with the op and the profile by
// both Commit and Switch.
func TestNixEnvNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	profile := filepath.Join(t.TempDir(), "profile")
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"commit", New().Commit(profile, "/farm"), "cannot run nix-env --set (" + profile + "): "},
		{"switch", New().Switch(profile, 3), "cannot run nix-env --switch-generation 3 (" + profile + "): "},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !errors.Is(c.err, exec.ErrNotFound) {
				t.Fatalf("got %v, want exec.ErrNotFound", c.err)
			}
			if !strings.HasPrefix(c.err.Error(), c.want) {
				t.Errorf("Error() = %q, want prefix %q", c.err.Error(), c.want)
			}
		})
	}
}
