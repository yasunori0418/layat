// Package paths computes the on-disk profile layout the engine operates on
// (→ ADR-0005, ADR-0013, ADR-0022, ADR-0024, ADR-0025).
//
// profileDir is the per-config dedicated directory (= flock key). The profile
// link sits inside it as profile, generations as profile-N-link, the build
// out-link as .pending, and the backref .root at the <roothash> level
// (→ docs/spec.md "on-disk layout of the profile").
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

	"github.com/yasunori0418/nput/internal/manifest"
)

// rootHashLen is the hex character count of the roothash (128 bit; fixed length;
// FS-safe; → ADR-0013). It matches the digit count of lib.mkManifest's anchorName
// (the first 32 hex of sha256).
const rootHashLen = 32

// backrefName is the file at the <roothash> level recording the original
// root's absolute path (→ ADR-0013). Its presence is what makes a directory
// under the base a <roothash> series rather than a <name>-keyed profileDir.
const backrefName = ".root"

// StateDir returns the base <state> for the profiles. $XDG_STATE_HOME if set,
// otherwise $HOME/.local/state (consistent with nix's own profile default; → ADR-0022).
func StateDir() (string, error) {
	if s := os.Getenv("XDG_STATE_HOME"); s != "" {
		return s, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("nput: cannot resolve $HOME: %w", err)
	}
	return filepath.Join(home, ".local", "state"), nil
}

// RootHash returns the truncated hex of the sha256 of the resolved absolute root path (→ ADR-0013).
func RootHash(absRoot string) string {
	sum := sha256.Sum256([]byte(absRoot))
	return hex.EncodeToString(sum[:])[:rootHashLen]
}

// Base returns the base <state>/nix/profiles/nput for the profiles. The home
// (no --root) profileDir is <name> directly under it; the roothash series is
// <roothash>/<name> (→ ADR-0024).
func Base(stateDir string) string {
	return filepath.Join(stateDir, "nix", "profiles", "nput")
}

// GenerationLink returns the path of the generation link
// <profileDir>/profile-<gen>-link that nix-env creates as a sibling of the
// profile link (→ docs/spec.md on-disk layout; ADR-0025).
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
	// Names are the <name> profileDirs under the series, in whatever order
	// os.ReadDir returned them; nothing here depends on that order. Nil for a
	// series that holds nothing but the backref, and for one whose NamesErr is set
	// (Names being empty therefore does not by itself mean the series has no
	// <name> profile — check NamesErr first).
	Names []string
	// BackrefErr is the reason the backref could not be turned into a root path.
	// The series is still returned so the caller can report it rather than
	// silently drop it.
	BackrefErr error
	// NamesErr is the reason the <name> profileDirs could not be listed. The
	// series is returned the same way, with Names left nil.
	NamesErr error
}

// ReadBackref reads the backref <hashDir>/.root and returns the absolute root
// path it records. The engine writes it as root + "\n" (→ engine.go), so the
// surrounding whitespace is trimmed: the verdict on a root must not depend on a
// trailing newline. A backref that is missing, empty, or not an absolute path
// is an error — no root path can be decided from it, and a relative path would
// otherwise resolve against the cwd (→ ADR-0034, DSG-096dc893-21f4-45e3-9347-986e9275b4d1).
func ReadBackref(hashDir string) (string, error) {
	backref := filepath.Join(hashDir, backrefName)
	b, err := os.ReadFile(backref)
	if err != nil {
		return "", fmt.Errorf("nput: cannot read backref (%s): %w", backref, err)
	}
	root := strings.TrimSpace(string(b))
	if root == "" {
		return "", fmt.Errorf("nput: backref is empty (%s)", backref)
	}
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("nput: backref is not an absolute path (%s): %q", backref, root)
	}
	return root, nil
}

// ListRootHashSeries lists the <roothash> series directly under base — the
// directories that hold a backref .root, that is the series of project mode,
// fixed root, and --root override (→ ADR-0024, ADR-0034). A directory without
// .root is the <name>-keyed profileDir of home mode and system mode and is not
// a series, so it is not returned.
//
// A series whose backref cannot be read is returned with BackrefErr set, and one
// whose contents cannot be listed with NamesErr set, rather than dropped: the
// caller reports the reason per series and the rest of the base still comes back
// (→ REQ-c44433a1-7ee7-459a-9aae-7cc42166876f). Only a failure of the base itself
// is returned as an error. This is the plain FS read both nput prune (→ #133) and
// nput status (→ #198) enumerate with; it applies no policy of its own beyond the
// .root test.
//
// base is a finished profile base — Base(stateDir) for the user state, or the
// system base as-is (it does not go through Base()).
//
// A base that does not exist is an error matching fs.ErrNotExist, not an empty
// listing: whether a missing base is normal is the caller's call, and swallowing
// it here would make a base that could not be listed indistinguishable from one
// holding no series (→ REQ-c44433a1-7ee7-459a-9aae-7cc42166876f).
func ListRootHashSeries(base string) ([]RootHashSeries, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("nput: cannot list profile base (%s): %w", base, err)
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
			// Whether there is a backref cannot be decided. Return the directory
			// as a series carrying the reason rather than dropping it silently or
			// abandoning the rest of the base.
			s.BackrefErr = fmt.Errorf("nput: cannot stat backref (%s): %w", backref, err)
		default:
			s.Root, s.BackrefErr = ReadBackref(hashDir)
		}

		if names, err := os.ReadDir(hashDir); err != nil {
			// The series is still reported; only its <name> profiles are unknown.
			s.NamesErr = fmt.Errorf("nput: cannot list series (%s): %w", hashDir, err)
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

// Resolve determines the profile layout from the state base, config name,
// rootKind, resolved absolute root, and whether --root was overridden
// (→ docs/spec.md "root resolution" table; ADR-0024, ADR-0025).
//
//   - home (no --root)               : <state>/nix/profiles/nput/<name> (no backref)
//   - project / fixed / --root override : <state>/nix/profiles/nput/<roothash>/<name>
//     (backref .root at the <roothash> level)
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
