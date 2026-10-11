package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// testRecord returns the record location for a home-keyed config under state.
func testRecord(state, name string) paths.Record {
	return paths.ResolveRecord(state, name, manifest.RootKindHome, "/home/me", false)
}

// putRecord writes a last-applied record with the given link-farm and manifest.
func putRecord(t *testing.T, rec paths.Record, linkFarm string, m manifest.Manifest) {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRecord(rec, "/home/me", linkFarm, raw, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

// putProfile points <dir>/profile at linkFarm, as a committed generation would.
func putProfile(t *testing.T, dir, linkFarm string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(dir, "profile")
	if err := os.Symlink(linkFarm, profile); err != nil {
		t.Fatal(err)
	}
	return profile
}

// entryTargets lists the targets of a manifest's entries ("<nil>" for an empty baseline).
func entryTargets(m *manifest.Manifest) string {
	if m == nil {
		return "<nil>"
	}
	var ts []string
	for _, e := range m.Entries {
		ts = append(ts, e.Target)
	}
	return strings.Join(ts, ",")
}

func TestReadRecordAbsentIsNoRecord(t *testing.T) {
	rec, m, err := readRecord(testRecord(realTempDir(t), "c").File)
	if err != nil || rec != nil || m != nil {
		t.Fatalf("readRecord(absent) = %v, %v, %v; want nil, nil, nil", rec, m, err)
	}
}

func TestWriteRecordRoundTrip(t *testing.T) {
	state := realTempDir(t)
	root := realTempDir(t)
	rec := paths.ResolveRecord(state, "c", manifest.RootKindProject, root, false)
	m := projectManifest(storeEntry("/nix/store/x", ".", "a"))
	raw, _ := json.Marshal(m)
	at := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	if err := writeRecord(rec, root, "/nix/store/lf", raw, at); err != nil {
		t.Fatal(err)
	}

	got, gotM, err := readRecord(rec.File)
	if err != nil {
		t.Fatal(err)
	}
	if got.RecordVersion != 1 || got.LinkFarm != "/nix/store/lf" || got.AppliedAt != at.Format(time.RFC3339) {
		t.Errorf("record = %+v", got)
	}
	if entryTargets(gotM) != "a" {
		t.Errorf("manifest entries = %s, want a", entryTargets(gotM))
	}
	if b, err := os.ReadFile(rec.Backref); err != nil || string(b) != root+"\n" {
		t.Errorf("backref = %q, %v; want %q", b, err, root+"\n")
	}
	// Only the record remains in its directory: the temp file was renamed into place.
	ents, _ := os.ReadDir(rec.Dir)
	if len(ents) != 1 || ents[0].Name() != "last-applied.json" {
		t.Errorf("record dir holds %v, want only last-applied.json", ents)
	}
}

func TestRemoveRecord(t *testing.T) {
	rec := testRecord(realTempDir(t), "c")
	putRecord(t, rec, "/nix/store/lf", homeManifest())
	if err := removeRecord(rec.File); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(rec.File); !os.IsNotExist(err) {
		t.Errorf("record still exists: %v", err)
	}
	if err := removeRecord(rec.File); err != nil {
		t.Errorf("removing an absent record = %v, want nil", err)
	}
}

// TestReadRecordRejectsInvalid verifies an unreadable record is an error, never "no record".
func TestReadRecordRejectsInvalid(t *testing.T) {
	const m = `{"schemaVersion":1,"root":{"rootKind":"home"},"entries":[]}`
	for _, tc := range []struct{ desc, body string }{
		{"broken JSON", `{"recordVersion":1,`},
		{"recordVersion missing", `{"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"recordVersion 2", `{"recordVersion":2,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"linkFarm missing", `{"recordVersion":1,"appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"linkFarm wrong type", `{"recordVersion":1,"linkFarm":3,"appliedAt":"2026-10-11T00:00:00Z","manifest":` + m + `}`},
		{"appliedAt missing", `{"recordVersion":1,"linkFarm":"/lf","manifest":` + m + `}`},
		{"appliedAt not RFC3339", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"yesterday","manifest":` + m + `}`},
		{"manifest missing", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z"}`},
		{"manifest null", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":null}`},
		{"manifest wrong type", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":"x"}`},
		{"manifest without rootKind", `{"recordVersion":1,"linkFarm":"/lf","appliedAt":"2026-10-11T00:00:00Z","manifest":{"schemaVersion":1,"root":{},"entries":[]}}`},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			file := filepath.Join(realTempDir(t), "last-applied.json")
			if err := os.WriteFile(file, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if rec, _, err := readRecord(file); err == nil {
				t.Errorf("readRecord = %+v, nil; want an error", rec)
			}
		})
	}
}

// TestLoadBaselinePrecedence verifies the baseline order record → profile → empty, with a
// warning on a record/profile mismatch and on the empty baseline.
func TestLoadBaselinePrecedence(t *testing.T) {
	recLF := writeLinkFarm(t, homeManifest(storeEntry("/nix/store/x", ".", "from-record")))
	profLF := writeLinkFarm(t, homeManifest(storeEntry("/nix/store/x", ".", "from-profile")))

	for _, tc := range []struct {
		desc              string
		record, profile   bool
		recordLinkFarm    string
		want, wantWarning string
	}{
		{desc: "record only", record: true, recordLinkFarm: recLF, want: "from-record"},
		{desc: "profile only", profile: true, want: "from-profile"},
		{desc: "neither", want: "<nil>", wantWarning: "no last-applied record"},
		{desc: "both agree", record: true, profile: true, recordLinkFarm: profLF, want: "from-record"},
		{desc: "both disagree", record: true, profile: true, recordLinkFarm: recLF, want: "from-record", wantWarning: "disagrees"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			state := realTempDir(t)
			rec := testRecord(state, "c")
			profile := filepath.Join(state, "profiles", "c", "profile")
			if tc.record {
				putRecord(t, rec, tc.recordLinkFarm, homeManifest(storeEntry("/nix/store/x", ".", "from-record")))
			}
			if tc.profile {
				putProfile(t, filepath.Dir(profile), profLF)
			}
			var warns []string
			b, err := loadBaseline(rec.File, profile, collectFormatted(&warns))
			if err != nil {
				t.Fatal(err)
			}
			if got := entryTargets(b.manifest); got != tc.want {
				t.Errorf("baseline = %s, want %s", got, tc.want)
			}
			switch {
			case tc.wantWarning == "" && len(warns) != 0:
				t.Errorf("warnings = %q, want none", warns)
			case tc.wantWarning != "" && (len(warns) != 1 || !strings.Contains(warns[0], tc.wantWarning)):
				t.Errorf("warnings = %q, want one mentioning %q", warns, tc.wantWarning)
			}
		})
	}
}

// TestLoadBaselineRejectsBrokenRecord verifies a broken record stops the baseline read even
// when a profile is there to fall back on.
func TestLoadBaselineRejectsBrokenRecord(t *testing.T) {
	state := realTempDir(t)
	rec := testRecord(state, "c")
	if err := os.MkdirAll(rec.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rec.File, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := putProfile(t, filepath.Join(state, "profiles", "c"), writeLinkFarm(t, homeManifest()))
	if _, err := loadBaseline(rec.File, profile, func(string, ...any) {}); err == nil {
		t.Error("loadBaseline with a broken record = nil error, want an error")
	}
}
