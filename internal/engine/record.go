package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// The last-applied record is what layat last placed under a root: the baseline for stale
// removal, drift repair, reset and rollback, read ahead of the profile's current generation.

// recordVersion is the only last-applied record format the engine reads and writes.
const recordVersion = 1

// lastApplied is the on-disk last-applied record.
type lastApplied struct {
	RecordVersion int             `json:"recordVersion"`
	LinkFarm      string          `json:"linkFarm"`
	AppliedAt     string          `json:"appliedAt"`
	Manifest      json.RawMessage `json:"manifest"`
}

// recordFor locates the last-applied record of a config, resolving the state base like ProfileFor.
func recordFor(stateDir, name, rootKind, root string, rootOverride bool) (paths.Record, error) {
	if stateDir == "" {
		var err error
		if stateDir, err = paths.StateDir(); err != nil {
			return paths.Record{}, err
		}
	}
	return paths.ResolveRecord(stateDir, name, rootKind, root, rootOverride), nil
}

// readRecord reads the record at file and its embedded manifest. An absent record is nil with
// no error; a record that is not valid JSON, lacks a field, or has another version is an error.
func readRecord(file string) (*lastApplied, *manifest.Manifest, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("layat: cannot read the last-applied record (%s): %w", file, err)
	}
	rec, m, err := parseRecord(data)
	if err != nil {
		return nil, nil, fmt.Errorf("layat: invalid last-applied record (%s): %w", file, err)
	}
	return rec, m, nil
}

func parseRecord(data []byte) (*lastApplied, *manifest.Manifest, error) {
	var raw struct {
		RecordVersion *int            `json:"recordVersion"`
		LinkFarm      *string         `json:"linkFarm"`
		AppliedAt     *string         `json:"appliedAt"`
		Manifest      json.RawMessage `json:"manifest"`
	}
	if err := strictDecode(data, &raw); err != nil {
		return nil, nil, err
	}
	switch {
	case raw.RecordVersion == nil:
		return nil, nil, errors.New("recordVersion is missing")
	case *raw.RecordVersion != recordVersion:
		return nil, nil, fmt.Errorf("recordVersion %d is unsupported (this engine reads v%d)", *raw.RecordVersion, recordVersion)
	case raw.LinkFarm == nil || *raw.LinkFarm == "":
		return nil, nil, errors.New("linkFarm is missing")
	case raw.AppliedAt == nil:
		return nil, nil, errors.New("appliedAt is missing")
	case len(raw.Manifest) == 0 || bytes.Equal(raw.Manifest, []byte("null")):
		return nil, nil, errors.New("manifest is missing")
	}
	if _, err := time.Parse(time.RFC3339, *raw.AppliedAt); err != nil {
		return nil, nil, fmt.Errorf("appliedAt is not RFC3339: %w", err)
	}
	var m manifest.Manifest
	if err := strictDecode(raw.Manifest, &m); err != nil {
		return nil, nil, fmt.Errorf("manifest: %w", err)
	}
	if m.SchemaVersion < 1 || m.SchemaVersion > manifest.SchemaVersion || m.Root.RootKind == "" {
		return nil, nil, fmt.Errorf("manifest: schemaVersion %d / rootKind %q is not a readable manifest", m.SchemaVersion, m.Root.RootKind)
	}
	rec := &lastApplied{
		RecordVersion: *raw.RecordVersion,
		LinkFarm:      *raw.LinkFarm,
		AppliedAt:     *raw.AppliedAt,
		Manifest:      raw.Manifest,
	}
	return rec, &m, nil
}

func strictDecode(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// writeRecord atomically replaces the record with manifestJSON placed from linkFarm at time at,
// creating the record directory and, for a roothash series, its backref.
func writeRecord(rec paths.Record, root, linkFarm string, manifestJSON []byte, at time.Time) error {
	data, err := json.Marshal(lastApplied{
		RecordVersion: recordVersion,
		LinkFarm:      linkFarm,
		AppliedAt:     at.Format(time.RFC3339),
		Manifest:      manifestJSON,
	})
	if err != nil {
		return fmt.Errorf("layat: cannot encode the last-applied record: %w", err)
	}
	if err := os.MkdirAll(rec.Dir, 0o755); err != nil {
		return fmt.Errorf("layat: cannot create the last-applied record directory (%s): %w", rec.Dir, err)
	}
	if rec.Backref != "" {
		if err := os.WriteFile(rec.Backref, []byte(root+"\n"), 0o644); err != nil {
			return fmt.Errorf("layat: cannot write backref (%s): %w", rec.Backref, err)
		}
	}
	tmp, err := os.CreateTemp(rec.Dir, ".last-applied-*.tmp")
	if err != nil {
		return fmt.Errorf("layat: cannot write the last-applied record (%s): %w", rec.File, err)
	}
	_, werr := tmp.Write(data)
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Rename(tmp.Name(), rec.File)
	}
	if werr != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("layat: cannot write the last-applied record (%s): %w", rec.File, werr)
	}
	return nil
}

// removeRecord deletes the record; an absent record is not an error.
func removeRecord(file string) error {
	if err := os.Remove(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("layat: cannot remove the last-applied record (%s): %w", file, err)
	}
	return nil
}

// baseline is what layat last placed under the root, with the sources it was chosen from.
type baseline struct {
	// manifest is the baseline manifest; nil is the empty baseline.
	manifest *manifest.Manifest
	// record is the last-applied record, nil when there is none.
	record *lastApplied
	// profileLinkFarm is the link-farm the profile link resolves to, "" without a profile.
	profileLinkFarm string
	// profileErr is why the profile could not serve as the baseline when no record did.
	profileErr error
}

// linkFarm is the link-farm the baseline was placed from: the record's, else the profile's.
func (b baseline) linkFarm() string {
	if b.record != nil {
		return b.record.LinkFarm
	}
	return b.profileLinkFarm
}

// loadBaseline reads the baseline in the order record → profile's current generation → empty.
// It warns once when the record and the profile disagree and when both are absent; only a
// record that cannot be read is an error.
func loadBaseline(recordFile, profileLink string, warnf func(string, ...any)) (baseline, error) {
	var b baseline
	rec, m, err := readRecord(recordFile)
	if err != nil {
		return b, err
	}
	b.record = rec

	hasProfile := false
	switch _, err := os.Stat(profileLink); {
	case err == nil:
		hasProfile = true
		if b.profileLinkFarm, err = filepath.EvalSymlinks(profileLink); err != nil {
			b.profileErr = err
		}
	case !errors.Is(err, fs.ErrNotExist):
		b.profileErr = err
	}

	if rec != nil {
		b.manifest = m
		if b.profileLinkFarm != "" && b.profileLinkFarm != rec.LinkFarm {
			warnf("layat: the last-applied record (%s) disagrees with the profile (%s); using the record as the baseline", rec.LinkFarm, b.profileLinkFarm)
		}
		return b, nil
	}
	if b.profileErr != nil {
		return b, nil
	}
	if hasProfile {
		b.manifest, b.profileErr = manifest.Load(profileLink)
		return b, nil
	}
	warnf("layat: no last-applied record; starting from an empty baseline")
	return b, nil
}
