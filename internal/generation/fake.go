package generation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	gens, err := generationLinks(profileLink)
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

// List scans the profile's sibling generation links with ListLinks.
func (f *Fake) List(profileLink string) ([]Generation, error) {
	if f.OnList != nil {
		if err := f.OnList(profileLink); err != nil {
			return nil, err
		}
	}
	return ListLinks(profileLink)
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

// fakeRepoint replaces profileLink with a relative symlink to its sibling genLink.
func fakeRepoint(profileLink, genLink string) error {
	if err := os.Remove(profileLink); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Symlink(filepath.Base(genLink), profileLink)
}
