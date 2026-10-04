// Package planner computes the place/replace/remove plan from the previous and new manifests,
// the root, and an FS prober, as pure logic. A stale symlink is scheduled for removal only when
// the previous generation recorded it and it still points to the recorded destination.
package planner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yasunori0418/layat/internal/manifest"
)

// FS abstracts the lstat/readlink/readdir probes the planner needs, so classification can be
// tested with a fake FS.
type FS interface {
	Lstat(path string) (os.FileInfo, error)
	Readlink(path string) (string, error)
	// ReadDir lists path's immediate children, to classify an occupying real directory.
	ReadDir(path string) ([]os.DirEntry, error)
}

// osFS is the real-filesystem FS used at engine runtime.
type osFS struct{}

func (osFS) Lstat(path string) (os.FileInfo, error)     { return os.Lstat(path) }
func (osFS) Readlink(path string) (string, error)       { return os.Readlink(path) }
func (osFS) ReadDir(path string) ([]os.DirEntry, error) { return os.ReadDir(path) }

// OSFS probes the real filesystem (engine runtime).
var OSFS FS = osFS{}

// PlaceKind classifies how a new-generation entry maps onto the current FS.
type PlaceKind int

const (
	// PlaceNew creates a new symlink when target is absent.
	PlaceNew PlaceKind = iota
	// PlaceReplace silently re-links a symlink recorded by this profile's own previous-generation manifest.
	PlaceReplace
	// PlaceForeign last-wins replaces an unrecorded symlink (foreign) with a warning.
	PlaceForeign
)

// PlaceAction is a symlink to materialize at TargetAbs pointing to Dest.
type PlaceAction struct {
	Entry     manifest.Entry
	TargetAbs string
	Dest      string // LinkDest(Entry): <src>/<subpath>
	Kind      PlaceKind
}

// CopyAction is a place-once copy of Src (= <src>/<subpath>) into an absent TargetAbs.
type CopyAction struct {
	Entry     manifest.Entry
	TargetAbs string
	Src       string // LinkDest(Entry): <src>/<subpath> (copy source)
}

// RemoveKind distinguishes what a RemoveAction removes: an entry-recorded symlink (Unlink) or an
// empty directory (Rmdir).
type RemoveKind int

const (
	// RemoveUnlink removes a stale symlink recorded by Entry.
	RemoveUnlink RemoveKind = iota
	// RemoveRmdir removes an empty directory that occupies a placement target or is left
	// empty by a migration. Entry is the zero value; the engine re-checks emptiness instead.
	RemoveRmdir
)

// RemoveAction is a stale filesystem object that satisfies the conservative invariant at plan
// time; the engine re-verifies it before removal. Slices list children before their parents, so
// walking them front-to-back removes bottom-up.
type RemoveAction struct {
	Kind      RemoveKind
	Entry     manifest.Entry
	TargetAbs string
}

// BackupAction renames a foreign occupant aside (TargetAbs → BackupAbs = <target>.<suffix>)
// under apply --backup. It runs after PreRemove and before Place/Copies, which then see an
// absent target.
type BackupAction struct {
	Entry     manifest.Entry
	TargetAbs string
	BackupAbs string // <target>.<suffix>
}

// Conflict is a placement target the engine must stop on: occupied by a non-symlink
// (regular file / directory) or nested under a symlinked ancestor.
type Conflict struct {
	Entry     manifest.Entry
	TargetAbs string
	Reason    string
	Kind      ConflictKind
}

// ConflictKind classifies a Conflict for caller-side guidance, independent of the Reason text.
type ConflictKind int

const (
	// ConflictUnspecified is the zero value, never produced by Compute; callers fall back to
	// generic guidance for it.
	ConflictUnspecified ConflictKind = iota
	// ConflictForeignEntity is a regular file/directory occupying a symlink target.
	ConflictForeignEntity
	// ConflictForeignAncestor is a symlinked ancestor component not recorded by this profile's
	// own previous generation (unrecorded / mismatched dest / no previous generation).
	ConflictForeignAncestor
	// ConflictSelfContradictoryAncestor is a symlinked ancestor still kept by the new generation
	// while a descendant entry also targets beneath it.
	ConflictSelfContradictoryAncestor
	// ConflictCopyStructureMismatch is a copy entry whose src structure (dir/file) mismatches the
	// existing target kind.
	ConflictCopyStructureMismatch
	// ConflictDirMigrationFailed is a real directory occupying a symlink target where at least one
	// leaf beneath it (any depth) is not safely migratable — a regular file, a foreign or
	// record-mismatched symlink, or a self-contradictory kept symlink.
	ConflictDirMigrationFailed
	// ConflictBackupTargetExists is apply --backup's rename-aside destination (<target>.<suffix>)
	// already occupied, e.g. by a previous backup, which is not overwritten.
	ConflictBackupTargetExists
)

// WarnKind enumerates non-fatal conditions the planner surfaces to the user.
type WarnKind int

const (
	// WarnForeignReplace overwrites an unrecorded symlink (place; last-wins).
	WarnForeignReplace WarnKind = iota
	// WarnStaleMismatch keeps a stale target because its symlink mismatches the record.
	WarnStaleMismatch
	// WarnStaleNonSymlink keeps a stale target because it is not a symlink (regular file, etc.).
	WarnStaleNonSymlink
	// WarnCopyOrphan is the orphan of a vanished copy entry (not removed; cleared by reset).
	WarnCopyOrphan
	// WarnCopyForeign skips a copy target under place-once because an unrecorded real file exists there.
	WarnCopyForeign
)

// Warning is a non-fatal condition surfaced to the user for a given target.
type Warning struct {
	Kind   WarnKind
	Target string
}

// Plan is the computed plan plus non-fatal warnings and fatal conflicts. The engine executes
// PreRemove → Backup → Place / Copies → Remove; a non-empty Conflicts means apply must stop.
type Plan struct {
	Place  []PlaceAction
	Copies []CopyAction
	Remove []RemoveAction
	// PreRemove removes self-recorded stale symlinks and migratable directories occupying
	// placement paths before placement; foreign or still-kept ancestors stay Conflicts.
	PreRemove []RemoveAction
	// Backup renames foreign occupants aside under apply --backup, in place of their Conflict or
	// copy skip. Ancestor symlink conflicts are never backed up.
	Backup    []BackupAction
	Conflicts []Conflict
	Warnings  []Warning
}

// Options configures Compute's optional behaviors. The zero value is normal apply.
type Options struct {
	// Backup enables apply --backup: a foreign occupant is renamed aside to "<target>.<Suffix>"
	// and the entry is placed fresh.
	Backup bool
	// Suffix is the backup rename suffix (Backup's target becomes "<target>.<Suffix>").
	// Empty defaults to "layat-backup".
	Suffix string
}

// backupSuffix returns opts.Suffix, defaulting to "layat-backup" when empty.
func backupSuffix(opts Options) string {
	if opts.Suffix == "" {
		return "layat-backup"
	}
	return opts.Suffix
}

// LinkDest returns the destination the entry's symlink should point to (<src>/<subpath>).
func LinkDest(e manifest.Entry) string {
	if e.Subpath == "" || e.Subpath == "." {
		return e.Src
	}
	return filepath.Join(e.Src, e.Subpath)
}

// Compute diffs the previous-generation manifest (prev; nil means first apply) against next,
// relative to root and FS state, and returns the plan without side effects.
func Compute(prev, next *manifest.Manifest, root string, fs FS, opts Options) (Plan, error) {
	var plan Plan

	// --- place / replace side: classify each new-generation entry against the current FS ---
	prevByTarget := byTarget(prev)
	nextByTarget := byTarget(next)
	// preRemoved dedups targets scheduled for pre-removal, so each is removed once.
	preRemoved := map[string]bool{}
	for _, e := range entriesOf(next) {
		targetAbs := filepath.Join(root, filepath.Clean(e.Target))

		// Nesting under a symlinked ancestor is a conflict, unless the ancestor is recorded by the
		// previous generation and dropped by the new one: then pre-remove it and place the child.
		offenderAbs, offenderRel, err := ancestorSymlink(root, e.Target, fs)
		if err != nil {
			return Plan{}, err
		}
		if offenderAbs != "" {
			// Matching offenderRel against the manifests assumes canonical targets; a non-canonical
			// one fails to match and degrades safely to a conflict.
			_, keptInNext := nextByTarget[offenderRel]
			if !keptInNext && recordedLink(offenderRel, offenderAbs, prevByTarget, fs) {
				if !preRemoved[offenderRel] {
					preRemoved[offenderRel] = true
					plan.PreRemove = append(plan.PreRemove, RemoveAction{Entry: prevByTarget[offenderRel], TargetAbs: offenderAbs})
				}
				// An lstat would resolve through the ancestor symlink into store content; after
				// PreRemove the child is absent, so place it as new without probing the FS.
				if err := appendAbsentPlacement(&plan, e, targetAbs); err != nil {
					return Plan{}, err
				}
				continue
			}
			// A foreign ancestor, or one the new generation still keeps, cannot be removed.
			kind := ConflictForeignAncestor
			if keptInNext {
				kind = ConflictSelfContradictoryAncestor
			}
			plan.Conflicts = append(plan.Conflicts, Conflict{
				Entry:     e,
				TargetAbs: targetAbs,
				Reason:    fmt.Sprintf("ancestor %q is a symlink; cannot nest beneath it (→ ADR-0015)", offenderAbs),
				Kind:      kind,
			})
			continue
		}

		// Branch the place-once / re-link classification per method.
		if e.Method == manifest.MethodCopy {
			if err := classifyCopy(&plan, e, targetAbs, prevByTarget, preRemoved, fs, opts); err != nil {
				return Plan{}, err
			}
			continue
		}
		if e.Method != manifest.MethodSymlink {
			return Plan{}, fmt.Errorf("layat: unknown method: %q (target: %s)", e.Method, e.Target)
		}

		info, err := fs.Lstat(targetAbs)
		switch {
		case err == nil && info.Mode()&os.ModeSymlink != 0:
			kind := PlaceForeign
			if recordedLink(e.Target, targetAbs, prevByTarget, fs) {
				kind = PlaceReplace
			} else {
				plan.Warnings = append(plan.Warnings, Warning{Kind: WarnForeignReplace, Target: e.Target})
			}
			plan.Place = append(plan.Place, PlaceAction{Entry: e, TargetAbs: targetAbs, Dest: LinkDest(e), Kind: kind})
		case err == nil && info.IsDir():
			// A real directory occupies the target (→ classifyRealDirTarget).
			if err := classifyRealDirTarget(&plan, e, targetAbs, prevByTarget, nextByTarget, preRemoved, fs, opts); err != nil {
				return Plan{}, err
			}
		case err == nil:
			// A regular file is not overwritten, unless apply --backup renames it aside.
			if err := appendBackupOrConflict(&plan, e, targetAbs, "target already has an existing file/directory (will not overwrite)", ConflictForeignEntity, fs, opts); err != nil {
				return Plan{}, err
			}
		case os.IsNotExist(err):
			plan.Place = append(plan.Place, PlaceAction{Entry: e, TargetAbs: targetAbs, Dest: LinkDest(e), Kind: PlaceNew})
		default:
			return Plan{}, fmt.Errorf("layat: cannot lstat target (%s): %w", targetAbs, err)
		}
	}

	// --- remove side: compute stale entries (prev ∖ next) under the conservative invariant ---
	// On first apply (prev == nil) nothing is removed.
	if prev != nil {
		for _, pe := range prev.Entries {
			if _, kept := nextByTarget[pe.Target]; kept {
				continue
			}
			if preRemoved[pe.Target] {
				// Already scheduled for pre-removal; do not remove it twice.
				continue
			}
			if pe.Method == manifest.MethodCopy {
				// copy is user-owned data: not removed, warn as orphan.
				plan.Warnings = append(plan.Warnings, Warning{Kind: WarnCopyOrphan, Target: pe.Target})
				continue
			}

			targetAbs := filepath.Join(root, filepath.Clean(pe.Target))
			info, err := fs.Lstat(targetAbs)
			switch {
			case err != nil && os.IsNotExist(err):
				continue // already gone = no-op (no warning).
			case err != nil:
				return Plan{}, fmt.Errorf("layat: cannot lstat stale target (%s): %w", targetAbs, err)
			case info.Mode()&os.ModeSymlink == 0:
				// A regular file / directory is left untouched.
				plan.Warnings = append(plan.Warnings, Warning{Kind: WarnStaleNonSymlink, Target: pe.Target})
				continue
			}

			onDisk, err := fs.Readlink(targetAbs)
			if err != nil || onDisk != LinkDest(pe) {
				// Record and reality mismatch (foreign / user-replaced) → not removed, warn.
				plan.Warnings = append(plan.Warnings, Warning{Kind: WarnStaleMismatch, Target: pe.Target})
				continue
			}
			plan.Remove = append(plan.Remove, RemoveAction{Entry: pe, TargetAbs: targetAbs})
		}
	}

	return plan, nil
}

func entriesOf(m *manifest.Manifest) []manifest.Entry {
	if m == nil {
		return nil
	}
	return m.Entries
}

func byTarget(m *manifest.Manifest) map[string]manifest.Entry {
	if m == nil {
		return nil
	}
	out := make(map[string]manifest.Entry, len(m.Entries))
	for _, e := range m.Entries {
		out[e.Target] = e
	}
	return out
}

// recordedLink reports whether the previous generation has an entry for target and the on-disk
// symlink points to its recorded destination.
func recordedLink(target, targetAbs string, prevByTarget map[string]manifest.Entry, fs FS) bool {
	pe, ok := prevByTarget[target]
	if !ok {
		return false
	}
	onDisk, err := fs.Readlink(targetAbs)
	if err != nil {
		return false
	}
	return onDisk == LinkDest(pe)
}

// classifyCopy classifies a copy entry under place-once semantics:
//
//	target absent                     → CopyAction (new place-once copy)
//	target is a self-recorded stale symlink (method changed symlink→copy) → PreRemove(Unlink) + CopyAction
//	target exists, structure mismatch → conflict, or backup + CopyAction under apply --backup
//	target exists, recorded           → no-op (placed by layat in a previous generation; place-once leaves it untouched)
//	target exists, foreign            → skip + WarnCopyForeign, or backup + CopyAction under apply --backup
//
// apply --recopy bypasses this in the engine.
func classifyCopy(plan *Plan, e manifest.Entry, targetAbs string, prevByTarget map[string]manifest.Entry, preRemoved map[string]bool, fs FS, opts Options) error {
	info, err := fs.Lstat(targetAbs)
	switch {
	case err == nil && info.Mode()&os.ModeSymlink != 0 && prevByTarget[e.Target].Method == manifest.MethodSymlink && recordedLink(e.Target, targetAbs, prevByTarget, fs):
		// Method changed symlink→copy: pre-remove the recorded symlink and place a fresh copy.
		// A drifted symlink falls through to the foreign handling below.
		if !preRemoved[e.Target] {
			preRemoved[e.Target] = true
			plan.PreRemove = append(plan.PreRemove, RemoveAction{Kind: RemoveUnlink, Entry: prevByTarget[e.Target], TargetAbs: targetAbs})
		}
		plan.Copies = append(plan.Copies, CopyAction{Entry: e, TargetAbs: targetAbs, Src: LinkDest(e)})
		return nil
	case err == nil:
		// target exists: check whether the src structure and kind match (a symlink counts as a
		// non-directory).
		mismatch, err := copyStructureMismatch(e, info, fs)
		if err != nil {
			return err
		}
		if mismatch {
			return appendBackupOrConflict(plan, e, targetAbs, "copy src structure and target kind mismatch (dir↔file; will not overwrite)", ConflictCopyStructureMismatch, fs, opts)
		}
		// place-once: leave a copy recorded by the previous generation untouched. An unrecorded real file gets a foreign warning.
		if pe, ok := prevByTarget[e.Target]; ok && pe.Method == manifest.MethodCopy {
			return nil
		}
		if opts.Backup {
			return appendBackup(plan, e, targetAbs, fs, opts)
		}
		plan.Warnings = append(plan.Warnings, Warning{Kind: WarnCopyForeign, Target: e.Target})
		return nil
	case os.IsNotExist(err):
		plan.Copies = append(plan.Copies, CopyAction{Entry: e, TargetAbs: targetAbs, Src: LinkDest(e)})
		return nil
	default:
		return fmt.Errorf("layat: cannot lstat copy target (%s): %w", targetAbs, err)
	}
}

// appendBackupOrConflict appends the Conflict the caller describes, or under apply --backup
// schedules a rename-aside of targetAbs and places the entry fresh.
func appendBackupOrConflict(plan *Plan, e manifest.Entry, targetAbs, reason string, kind ConflictKind, fs FS, opts Options) error {
	if !opts.Backup {
		plan.Conflicts = append(plan.Conflicts, Conflict{Entry: e, TargetAbs: targetAbs, Reason: reason, Kind: kind})
		return nil
	}
	return appendBackup(plan, e, targetAbs, fs, opts)
}

// appendBackup schedules targetAbs to be renamed aside to "<targetAbs>.<suffix>" and appends the
// entry's fresh placement. An occupied rename destination is a ConflictBackupTargetExists.
func appendBackup(plan *Plan, e manifest.Entry, targetAbs string, fs FS, opts Options) error {
	backupAbs := targetAbs + "." + backupSuffix(opts)
	if _, err := fs.Lstat(backupAbs); err == nil {
		plan.Conflicts = append(plan.Conflicts, Conflict{
			Entry:     e,
			TargetAbs: targetAbs,
			Reason:    fmt.Sprintf("backup destination already exists (%s); remove or move it manually before re-running --backup", backupAbs),
			Kind:      ConflictBackupTargetExists,
		})
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("layat: cannot lstat backup destination (%s): %w", backupAbs, err)
	}
	plan.Backup = append(plan.Backup, BackupAction{Entry: e, TargetAbs: targetAbs, BackupAbs: backupAbs})
	return appendAbsentPlacement(plan, e, targetAbs)
}

// appendAbsentPlacement records a new symlink or place-once copy for an entry whose target is
// known to be absent at execution time, without probing the FS.
func appendAbsentPlacement(plan *Plan, e manifest.Entry, targetAbs string) error {
	switch e.Method {
	case manifest.MethodSymlink:
		plan.Place = append(plan.Place, PlaceAction{Entry: e, TargetAbs: targetAbs, Dest: LinkDest(e), Kind: PlaceNew})
	case manifest.MethodCopy:
		plan.Copies = append(plan.Copies, CopyAction{Entry: e, TargetAbs: targetAbs, Src: LinkDest(e)})
	default:
		return fmt.Errorf("layat: unknown method: %q (target: %s)", e.Method, e.Target)
	}
	return nil
}

// copyStructureMismatch reports whether the dir/file kind of src (<src>/<subpath>) disagrees
// with the kind of the existing target. A symlink target counts as a file.
func copyStructureMismatch(e manifest.Entry, targetInfo os.FileInfo, fs FS) (bool, error) {
	srcInfo, err := fs.Lstat(LinkDest(e))
	if err != nil {
		return false, fmt.Errorf("layat: cannot lstat copy src (%s): %w", LinkDest(e), err)
	}
	return srcInfo.IsDir() != targetInfo.IsDir(), nil
}

// classifyRealDirTarget classifies a symlink-method entry whose target is a real directory. A
// fully migratable tree is pre-removed and the symlink placed; otherwise it is a Conflict, or
// under apply --backup the whole directory is renamed aside.
func classifyRealDirTarget(plan *Plan, e manifest.Entry, targetAbs string, prevByTarget, nextByTarget map[string]manifest.Entry, preRemoved map[string]bool, fs FS, opts Options) error {
	dirActions, reason, err := classifyDirMigration(filepath.Clean(e.Target), targetAbs, prevByTarget, nextByTarget, fs)
	if err != nil {
		return err
	}
	if reason != "" {
		// Entries recorded beneath a backed-up directory leave with it, so keep the remove-side
		// loop off them. Without --backup they stay ordinary Remove candidates.
		if opts.Backup {
			markDirEntriesPreRemoved(filepath.Clean(e.Target), prevByTarget, preRemoved)
		}
		return appendBackupOrConflict(plan, e, targetAbs, fmt.Sprintf("target directory cannot be fully migrated: %s (→ ADR-0047)", reason), ConflictDirMigrationFailed, fs, opts)
	}
	plan.PreRemove = append(plan.PreRemove, dirActions...)
	plan.PreRemove = append(plan.PreRemove, RemoveAction{Kind: RemoveRmdir, TargetAbs: targetAbs})
	for _, da := range dirActions {
		if da.Kind == RemoveUnlink {
			preRemoved[da.Entry.Target] = true
		}
	}
	plan.Place = append(plan.Place, PlaceAction{Entry: e, TargetAbs: targetAbs, Dest: LinkDest(e), Kind: PlaceNew})
	return nil
}

// markDirEntriesPreRemoved marks every prevByTarget entry beneath dirRel (any depth) as
// preRemoved without emitting a RemoveAction.
func markDirEntriesPreRemoved(dirRel string, prevByTarget map[string]manifest.Entry, preRemoved map[string]bool) {
	for target := range prevByTarget {
		rel, err := filepath.Rel(dirRel, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		preRemoved[target] = true
	}
}

// classifyDirMigration walks an occupying real directory without following symlinks. It is
// migratable only when every leaf is a recorded, dropped symlink (RemoveUnlink) or an empty
// directory (RemoveRmdir); otherwise reason names the offending leaf. Actions are children-first.
func classifyDirMigration(dirRel, dirAbs string, prevByTarget, nextByTarget map[string]manifest.Entry, fs FS) (actions []RemoveAction, reason string, err error) {
	children, err := fs.ReadDir(dirAbs)
	if err != nil {
		return nil, "", fmt.Errorf("layat: cannot read directory (%s): %w", dirAbs, err)
	}
	for _, de := range children {
		childAbs := filepath.Join(dirAbs, de.Name())
		childRel := filepath.Join(dirRel, de.Name())

		info, lerr := fs.Lstat(childAbs)
		if lerr != nil {
			return nil, "", fmt.Errorf("layat: cannot lstat (%s): %w", childAbs, lerr)
		}

		switch {
		case info.Mode()&os.ModeSymlink != 0:
			if _, kept := nextByTarget[childRel]; kept {
				return nil, fmt.Sprintf("%q is a symlink the new generation still keeps (self-contradictory manifest)", childRel), nil
			}
			if !recordedLink(childRel, childAbs, prevByTarget, fs) {
				return nil, fmt.Sprintf("%q is a foreign or record-mismatched symlink", childRel), nil
			}
			actions = append(actions, RemoveAction{Kind: RemoveUnlink, Entry: prevByTarget[childRel], TargetAbs: childAbs})
		case info.IsDir():
			sub, subReason, serr := classifyDirMigration(childRel, childAbs, prevByTarget, nextByTarget, fs)
			if serr != nil {
				return nil, "", serr
			}
			if subReason != "" {
				return nil, subReason, nil
			}
			actions = append(actions, sub...)
			actions = append(actions, RemoveAction{Kind: RemoveRmdir, TargetAbs: childAbs})
		default:
			return nil, fmt.Sprintf("%q is a regular file", childRel), nil
		}
	}
	return actions, "", nil
}

// ancestorSymlink returns the first existing ancestor of target under root that is a symlink, as
// its absolute and root-relative paths. A non-existent ancestor stops the walk with "", "".
func ancestorSymlink(root, target string, fs FS) (abs, rel string, err error) {
	clean := filepath.Clean(target)
	comps := strings.Split(clean, string(os.PathSeparator))
	cur := root
	for i := 0; i < len(comps)-1; i++ {
		if comps[i] == "" {
			continue
		}
		cur = filepath.Join(cur, comps[i])
		if rel == "" {
			rel = comps[i]
		} else {
			rel = filepath.Join(rel, comps[i])
		}
		info, lerr := fs.Lstat(cur)
		if lerr != nil {
			if os.IsNotExist(lerr) {
				return "", "", nil
			}
			return "", "", fmt.Errorf("layat: cannot lstat ancestor (%s): %w", cur, lerr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return cur, rel, nil
		}
	}
	return "", "", nil
}
