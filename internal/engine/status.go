package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// StatusOptions are the inputs of Status.
type StatusOptions struct {
	// StateDir is the user <state> (empty = paths.StateDir()).
	StateDir string
	// SystemStateDir is the system-mode record base <state> (empty = /var/lib).
	SystemStateDir string
	// Warnf is the warning output sink (nil = stderr).
	Warnf func(format string, args ...any)
}

// StatusSeries is one last-applied record found by Status, or a series that could not be read.
type StatusSeries struct {
	// Mode is the record's rootKind; home / system by base for a <name>-keyed record or an Err.
	Mode string
	// Name is the <name>. Empty for a <roothash> series reported as a whole.
	Name string
	// Root is the backref .root of a <roothash> series. Empty for a <name>-keyed record, a
	// series whose backref cannot be read, and a series reported as a whole.
	Root string
	// RootHash is the <roothash> series directory name. Empty for a <name>-keyed record.
	RootHash string
	// RecordFile is the last-applied.json path. Empty for a series reported as a whole.
	RecordFile string
	AppliedAt  string
	LinkFarm   string
	Manifest   *manifest.Manifest
	// Err is why the record, the backref or the <name> entries could not be read; check it first.
	Err error
}

// Status lists the last-applied records under the user record base then the system record
// base, each with the <name>-keyed records first and the <roothash> series after, in name
// order. It only reads: no flock, directory, .pending, .root or record is created.
func Status(opts StatusOptions) ([]StatusSeries, error) {
	warnf := opts.Warnf
	if warnf == nil {
		warnf = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}
	stateDir := opts.StateDir
	if stateDir == "" {
		var err error
		if stateDir, err = paths.StateDir(); err != nil {
			return nil, err
		}
	}
	systemStateDir := opts.SystemStateDir
	if systemStateDir == "" {
		systemStateDir = defaultSystemStateDir
	}

	out, err := statusBase(filepath.Join(stateDir, "layat"), manifest.RootKindHome)
	if err != nil {
		return nil, err
	}
	system := filepath.Join(systemStateDir, "layat")
	sys, err := statusBase(system, manifest.RootKindSystem)
	if err != nil {
		warnf("layat: skipping the system record base (%s): %v", system, err)
		return out, nil
	}
	return append(out, sys...), nil
}

// statusBase lists the records under one record base; mode is the base's home / system. An
// absent base is empty; only a base that exists but cannot be read is an error.
func statusBase(base, mode string) ([]StatusSeries, error) {
	if _, err := os.Stat(base); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("layat: cannot read the record base (%s): %w", base, err)
	}
	series, err := paths.ListRootHashSeries(base)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("layat: cannot list the record base (%s): %w", base, err)
	}
	sort.Slice(series, func(i, j int) bool { return series[i].RootHash < series[j].RootHash })

	isSeries := map[string]bool{}
	for _, s := range series {
		isSeries[s.RootHash] = true
	}
	var direct []string
	for _, e := range entries {
		if e.IsDir() && !isSeries[e.Name()] {
			direct = append(direct, e.Name())
		}
	}
	sort.Strings(direct)

	var out []StatusSeries
	for _, name := range direct {
		if s, ok := statusRecord(filepath.Join(base, name), mode, "", ""); ok {
			out = append(out, s)
		}
	}
	for _, sr := range series {
		out = append(out, statusRootHashSeries(base, mode, sr)...)
	}
	return out, nil
}

// statusRootHashSeries lists the records of one <roothash> series. A series whose <name>
// entries cannot be listed, or whose backref cannot be read and has no record, is one entry.
func statusRootHashSeries(base, mode string, sr paths.RootHashSeries) []StatusSeries {
	whole := StatusSeries{Mode: mode, RootHash: sr.RootHash}
	if sr.NamesErr != nil {
		whole.Err = errors.Join(sr.BackrefErr, sr.NamesErr)
		return []StatusSeries{whole}
	}
	names := append([]string(nil), sr.Names...)
	sort.Strings(names)

	var out []StatusSeries
	for _, name := range names {
		s, ok := statusRecord(filepath.Join(base, sr.RootHash, name), mode, sr.RootHash, sr.Root)
		if !ok {
			continue
		}
		if sr.BackrefErr != nil {
			s.Mode = mode
			s.Err = errors.Join(sr.BackrefErr, s.Err)
		}
		out = append(out, s)
	}
	if len(out) == 0 && sr.BackrefErr != nil {
		whole.Err = sr.BackrefErr
		return []StatusSeries{whole}
	}
	return out
}

// statusRecord reads the record in dir. ok is false when dir holds no record. A <roothash>
// series record takes its Mode from the manifest; an Err keeps the base's mode.
func statusRecord(dir, mode, rootHash, root string) (StatusSeries, bool) {
	file := filepath.Join(dir, "last-applied.json")
	s := StatusSeries{Mode: mode, Name: filepath.Base(dir), Root: root, RootHash: rootHash, RecordFile: file}
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		return s, false
	} else if err != nil {
		s.Err = fmt.Errorf("layat: cannot stat the last-applied record (%s): %w", file, err)
		return s, true
	}
	rec, m, err := readRecord(file)
	if err != nil {
		s.Err = err
		return s, true
	}
	if rec == nil {
		return s, false
	}
	s.AppliedAt, s.LinkFarm, s.Manifest = rec.AppliedAt, rec.LinkFarm, m
	if rootHash != "" {
		s.Mode = m.Root.RootKind
	}
	return s, true
}
