// Package paths computes the on-disk profile layout the engine operates on. profileDir is the
// per-config directory (= flock key) holding profile, profile-N-link, and .pending; the backref
// .root sits at the <roothash> level.
package paths

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yasunori0418/layat/internal/manifest"
)

// rootHashLen is the hex character count of the roothash (128 bit). It matches the digit count
// of lib.mkManifest's anchorName (the first 32 hex of sha256).
const rootHashLen = 32

// backrefName is the file at the <roothash> level recording the original root's absolute path.
// Its presence makes a directory under the base a <roothash> series rather than a profileDir.
const backrefName = ".root"

// StateDir returns the base <state> for the profiles. $XDG_STATE_HOME if set,
// otherwise $HOME/.local/state (consistent with nix's own profile default).
func StateDir() (string, error) {
	if s := os.Getenv("XDG_STATE_HOME"); s != "" {
		return s, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("layat: cannot resolve $HOME: %w", err)
	}
	return filepath.Join(home, ".local", "state"), nil
}

// RootHash returns the truncated hex of the sha256 of the resolved absolute root path.
func RootHash(absRoot string) string {
	sum := sha256.Sum256([]byte(absRoot))
	return hex.EncodeToString(sum[:])[:rootHashLen]
}

// Base returns the base <state>/nix/profiles/layat for the profiles. The home
// (no --root) profileDir is <name> directly under it; the roothash series is
// <roothash>/<name>.
func Base(stateDir string) string {
	return filepath.Join(stateDir, "nix", "profiles", "layat")
}

// GenerationLink returns the path of the generation link <profileDir>/profile-<gen>-link that
// nix-env creates as a sibling of the profile link.
func GenerationLink(profileLink string, gen int) string {
	return fmt.Sprintf("%s-%d-link", profileLink, gen)
}

// Profile is the full set of profile-layout paths for one config.
type Profile struct {
	// Dir is the profileDir (per-config directory; flock key).
	Dir string
	// Profile is the profile link (the target of nix-env --profile).
	Profile string
	// Pending is the output of nix build --out-link (a sibling that does not pass through the profile).
	Pending string
	// BackrefDir is the level where the backref .root is placed. Empty for home (no --root).
	BackrefDir string
	// Backref is the backref file recording the original root's absolute path. Empty for home.
	Backref string
}

// RootHashSeries is one <roothash> series found under a profile base: the
// series directory, the root its backref records, and the <name> profileDirs
// under it.
type RootHashSeries struct {
	// RootHash is the series directory name (<roothash>).
	RootHash string
	// Root is the absolute root path the backref .root records. Empty when BackrefErr is set.
	Root string
	// Names are the <name> profileDirs under the series, in os.ReadDir order. Nil when the
	// series holds only the backref or when NamesErr is set (check NamesErr first).
	Names []string
	// BackrefErr is the reason the backref could not be turned into a root path.
	BackrefErr error
	// NamesErr is the reason the <name> profileDirs could not be listed.
	NamesErr error
}

// ReadBackref returns the absolute root path the backref <hashDir>/.root records, with
// surrounding whitespace trimmed. A missing, empty, or non-absolute backref is an error.
func ReadBackref(hashDir string) (string, error) {
	backref := filepath.Join(hashDir, backrefName)
	b, err := os.ReadFile(backref)
	if err != nil {
		return "", fmt.Errorf("layat: cannot read backref (%s): %w", backref, err)
	}
	root := strings.TrimSpace(string(b))
	if root == "" {
		return "", fmt.Errorf("layat: backref is empty (%s)", backref)
	}
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("layat: backref is not an absolute path (%s): %q", backref, root)
	}
	return root, nil
}

// ListRootHashSeries lists the <roothash> series (directories holding a backref .root) directly
// under base. A series that fails to read is returned with BackrefErr / NamesErr set; only a
// failure of base itself, including fs.ErrNotExist, is returned as an error.
func ListRootHashSeries(base string) ([]RootHashSeries, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("layat: cannot list profile base (%s): %w", base, err)
	}

	var series []RootHashSeries
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		hashDir := filepath.Join(base, e.Name())
		backref := filepath.Join(hashDir, backrefName)
		s := RootHashSeries{RootHash: e.Name()}
		switch _, err := os.Lstat(backref); {
		case errors.Is(err, fs.ErrNotExist):
			// No backref: a <name>-keyed profileDir, not a series.
			continue
		case err != nil:
			// Whether there is a backref cannot be decided; report it as a series with the reason.
			s.BackrefErr = fmt.Errorf("layat: cannot stat backref (%s): %w", backref, err)
		default:
			s.Root, s.BackrefErr = ReadBackref(hashDir)
		}

		if names, err := os.ReadDir(hashDir); err != nil {
			// The series is still reported; only its <name> profiles are unknown.
			s.NamesErr = fmt.Errorf("layat: cannot list series (%s): %w", hashDir, err)
		} else {
			for _, n := range names {
				// The backref is not a <name> profile even where it is a directory.
				if n.IsDir() && n.Name() != backrefName {
					s.Names = append(s.Names, n.Name())
				}
			}
		}
		series = append(series, s)
	}
	return series, nil
}

// Resolve determines the profile layout for a config: <base>/<name> for home without --root,
// otherwise <base>/<roothash>/<name> with the backref .root at the <roothash> level.
func Resolve(stateDir, name, rootKind, absRoot string, rootOverride bool) Profile {
	base := Base(stateDir)

	// Only home (no --root) keys directly on <name>. Otherwise separate into an independent series per root.
	if rootKind == manifest.RootKindHome && !rootOverride {
		dir := filepath.Join(base, name)
		return Profile{
			Dir:     dir,
			Profile: filepath.Join(dir, "profile"),
			Pending: filepath.Join(dir, ".pending"),
		}
	}

	hashDir := filepath.Join(base, RootHash(absRoot))
	dir := filepath.Join(hashDir, name)
	return Profile{
		Dir:        dir,
		Profile:    filepath.Join(dir, "profile"),
		Pending:    filepath.Join(dir, ".pending"),
		BackrefDir: hashDir,
		Backref:    filepath.Join(hashDir, backrefName),
	}
}
