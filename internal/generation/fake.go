package generation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/yasunori0418/layat/internal/paths"
)

// Fake is the file-system test double of Backend. It keeps the same profileDir layout as the
// real backend (profile → profile-<N>-link → link-farm) with plain symlinks, so the engine's
// layout reads work unchanged while no external command runs. The zero value is ready to use.
type Fake struct {
	// System is what SystemBase returns. Tests point it at a tmpdir.
	System string
	// OnCommit / OnList / OnSwitch run before the operation when set; a non-nil error fails the
	// operation before it touches the layout.
	OnCommit func(profileLink, linkFarm string) error
	OnList   func(profileLink string) error
	OnSwitch func(profileLink string, gen int) error

	mu sync.Mutex
	// Commits records every Commit call (also the failed ones) in order.
	Commits []FakeCommit
	// Switches records the generation of every Switch call in order.
	Switches []int
}

// FakeCommit is one recorded Commit call.
type FakeCommit struct {
	ProfileLink string
	LinkFarm    string
}

// fakeDateLayout is the timestamp layout List reports for a generation link's mtime.
const fakeDateLayout = "2006-01-02 15:04:05"

// Commit creates <profileLink>-<N>-link → linkFarm with N one above the highest existing
// generation and re-points the profile link at it with a relative sibling symlink.
func (f *Fake) Commit(profileLink, linkFarm string) error {
	f.mu.Lock()
	f.Commits = append(f.Commits, FakeCommit{ProfileLink: profileLink, LinkFarm: linkFarm})
	f.mu.Unlock()
	if f.OnCommit != nil {
		if err := f.OnCommit(profileLink, linkFarm); err != nil {
			return err
		}
	}
	gens, err := fakeGenerationLinks(profileLink)
	if err != nil {
		return err
	}
	next := 1
	if len(gens) > 0 {
		next = gens[len(gens)-1] + 1
	}
	genLink := paths.GenerationLink(profileLink, next)
	if err := os.Symlink(linkFarm, genLink); err != nil {
		return err
	}
	return fakeRepoint(profileLink, genLink)
}

// List scans the profile's sibling generation links. The current generation is the one the
// profile link names, or failing that the one resolving to the same path as the profile.
func (f *Fake) List(profileLink string) ([]Generation, error) {
	if f.OnList != nil {
		if err := f.OnList(profileLink); err != nil {
			return nil, err
		}
	}
	nums, err := fakeGenerationLinks(profileLink)
	if err != nil {
		return nil, err
	}
	current := fakeCurrent(profileLink, nums)
	gens := make([]Generation, 0, len(nums))
	for _, n := range nums {
		date := ""
		if fi, err := os.Lstat(paths.GenerationLink(profileLink, n)); err == nil {
			date = fi.ModTime().Format(fakeDateLayout)
		}
		gens = append(gens, Generation{Number: n, Date: date, Current: n == current})
	}
	return gens, nil
}

// Switch re-points the profile link at the existing generation gen.
func (f *Fake) Switch(profileLink string, gen int) error {
	f.mu.Lock()
	f.Switches = append(f.Switches, gen)
	f.mu.Unlock()
	if f.OnSwitch != nil {
		if err := f.OnSwitch(profileLink, gen); err != nil {
			return err
		}
	}
	genLink := paths.GenerationLink(profileLink, gen)
	if _, err := os.Lstat(genLink); err != nil {
		return fmt.Errorf("generation %d does not exist (%s): %w", gen, genLink, err)
	}
	return fakeRepoint(profileLink, genLink)
}

// UserBase keys the profileDirs directly under <state>/profiles.
func (f *Fake) UserBase(stateDir string) string {
	return filepath.Join(stateDir, "profiles")
}

// SystemBase returns System.
func (f *Fake) SystemBase() string {
	return f.System
}

// fakeGenerationNumber parses the generation number out of a sibling name of the form
// <profile>-<N>-link; ok is false for any other name.
func fakeGenerationNumber(profileLink, name string) (n int, ok bool) {
	prefix := filepath.Base(profileLink) + "-"
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, "-link") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, prefix), "-link"))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// fakeGenerationLinks returns the generation numbers of the <profile>-<N>-link siblings in
// ascending order. A missing profileDir holds no generations.
func fakeGenerationLinks(profileLink string) ([]int, error) {
	entries, err := os.ReadDir(filepath.Dir(profileLink))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var nums []int
	for _, e := range entries {
		if n, ok := fakeGenerationNumber(profileLink, e.Name()); ok {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums, nil
}

// fakeCurrent returns the current generation number among nums, or -1 when the profile link is
// absent or points at none of them.
func fakeCurrent(profileLink string, nums []int) int {
	dest, err := os.Readlink(profileLink)
	if err != nil {
		return -1
	}
	if n, ok := fakeGenerationNumber(profileLink, filepath.Base(dest)); ok {
		return n
	}
	// The profile names something else (a link-farm directly): match by resolved path, taking
	// the newest generation that resolves to the same place.
	resolved, err := filepath.EvalSymlinks(profileLink)
	if err != nil {
		return -1
	}
	current := -1
	for _, n := range nums {
		if r, err := filepath.EvalSymlinks(paths.GenerationLink(profileLink, n)); err == nil && r == resolved {
			current = n
		}
	}
	return current
}

// fakeRepoint replaces profileLink with a relative symlink to its sibling genLink.
func fakeRepoint(profileLink, genLink string) error {
	if err := os.Remove(profileLink); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Symlink(filepath.Base(genLink), profileLink)
}
