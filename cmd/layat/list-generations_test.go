package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/generation"
)

// TestListGenerationsJSONInfoGenerations pins list-generations' --json shape: {number, date,
// current} rows in result.info.generations, empty items, no generation slot, conformant envelope.
func TestListGenerationsJSONInfoGenerations(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}

	gens := []generation.Generation{
		{Number: 1, Date: "2026-07-18 09:00:00"},
		{Number: 2, Date: "2026-07-19 12:00:00", Current: true},
	}
	r, buf := newListGenerationsTestRun()
	r.beginSubject("home").setPayload(&outturnPayload[*generationsInfo]{info: &generationsInfo{Generations: generationRows(gens)}})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings: %v\ndocument: %s", findings, buf.String())
	}

	doc := decodeEnvelope(t, buf)
	if doc["dryRun"] != false {
		t.Errorf("dryRun = %v, want false (a read-only command reflects only the flag)", doc["dryRun"])
	}
	sr := doc["results"].([]any)[0].(map[string]any)
	if _, ok := sr["generation"]; ok {
		t.Errorf("generation slot present, want absent (info.generations is the single encoding): %v", sr["generation"])
	}
	res := sr["result"].(map[string]any)
	if items := res["items"].([]any); len(items) != 0 {
		t.Errorf("items = %v, want empty (the enumeration is info, not items)", items)
	}
	rows := res["info"].(map[string]any)["generations"].([]any)
	if len(rows) != 2 {
		t.Fatalf("info.generations = %v, want 2 rows", rows)
	}
	current := rows[1].(map[string]any)
	if current["number"] != json.Number("2") || current["date"] != "2026-07-19 12:00:00" || current["current"] != true {
		t.Errorf("rows[1] = %v, want number=2 date verbatim current=true", current)
	}
	if first := rows[0].(map[string]any); first["current"] != false {
		t.Errorf("rows[0].current = %v, want false (the key is always present)", first["current"])
	}
}

// TestListGenerationsJSONInfoAbsentWithoutListing: when list-generations fails after registering
// the subject but before any listing exists, result.info stays absent.
func TestListGenerationsJSONInfoAbsentWithoutListing(t *testing.T) {
	r, buf := newListGenerationsTestRun()
	r.beginSubject("home")
	if err := r.emit(errors.New("layat: list-generations is home mode only")); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertNoInfoKeys(t, decodeEnvelope(t, buf))
}

// TestListGenerationsJSONEmptyStaysArray: an empty profile emits "generations": [] rather than
// dropping the key.
func TestListGenerationsJSONEmptyStaysArray(t *testing.T) {
	r, buf := newListGenerationsTestRun()
	r.beginSubject("empty").setPayload(&outturnPayload[*generationsInfo]{info: &generationsInfo{Generations: generationRows(nil)}})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	info := decodeEnvelope(t, buf)["results"].([]any)[0].(map[string]any)["result"].(map[string]any)["info"].(map[string]any)
	rows, ok := info["generations"].([]any)
	if !ok {
		t.Fatalf("info = %v, want the generations key present as an array", info)
	}
	if len(rows) != 0 {
		t.Errorf("info.generations = %v, want []", rows)
	}
}

// stubNixRootKind puts a fake nix first on PATH whose every eval prints kind, so the eval
// pre-resolution of a named config reports that rootKind (and that root, for fixed).
func stubNixRootKind(t *testing.T, kind string) {
	t.Helper()
	bin := t.TempDir()
	script := "#!/bin/sh\ncase \"$1\" in\neval) printf '" + kind + "';;\n*) echo 'error: stub build failed' >&2; exit 1;;\nesac\n"
	if err := os.WriteFile(filepath.Join(bin, "nix"), []byte(script), 0o755); err != nil {
		t.Fatalf("write nix stub: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// makeProfilesUnder creates <base>/<name>/profile links with one generation each, the layout
// directly keyed on <name> that both bases share.
func makeProfilesUnder(t *testing.T, base string, names ...string) {
	t.Helper()
	for _, name := range names {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.Symlink("/nowhere", filepath.Join(dir, "profile-1-link")); err != nil {
			t.Fatalf("symlink generation: %v", err)
		}
		if err := os.Symlink("profile-1-link", filepath.Join(dir, "profile")); err != nil {
			t.Fatalf("symlink profile: %v", err)
		}
	}
}

// TestGenerationCommandsModeGate: rollback and list-generations <name> accept home and system and
// refuse project / fixed before touching any profile. A system rollback reaching the engine's
// root check (instead of the CLI gate) proves the gate let it through.
func TestGenerationCommandsModeGate(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("the system rollback case relies on the engine refusing a non-root caller")
	}
	origJSON := flagJSON
	t.Cleanup(func() { flagJSON = origJSON })
	flagJSON = true

	for _, kind := range []string{"project", "fixed"} {
		for _, command := range []string{"list-generations", "rollback"} {
			t.Run(command+" refuses "+kind, func(t *testing.T) {
				withFlakeEntrypoint(t)
				stubNixRootKind(t, kind)
				installFakeBackend(t, "")
				var err error
				if command == "rollback" {
					run, _ := newRollbackTestRun()
					err = runRollback(run, "c")
				} else {
					run, _ := newListGenerationsTestRun()
					err = runListGenerations(run, "c")
				}
				want := "layat: " + command + " is home / system mode only (layat.c has rootKind=\"" + kind + "\""
				if err == nil || !strings.HasPrefix(err.Error(), want) {
					t.Fatalf("err = %v, want it to start with %q", err, want)
				}
			})
		}
	}

	t.Run("list-generations lists a system profile", func(t *testing.T) {
		withFlakeEntrypoint(t)
		stubNixRootKind(t, "system")
		fake := installFakeBackend(t, "")
		fake.System = t.TempDir()
		makeProfilesUnder(t, fake.System, "c")
		run, buf := newListGenerationsTestRun()
		if err := runListGenerations(run, "c"); err != nil {
			t.Fatalf("runListGenerations: %v", err)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		byName, _ := subjectResults(t, buf)
		if got := infoArray(t, byName["c"], "generations"); len(got) != 1 {
			t.Errorf("generations = %v, want the system profile's listing", got)
		}
		if mode := byName["c"]["result"].(map[string]any)["info"].(map[string]any)["mode"]; mode != "system" {
			t.Errorf("info.mode = %v, want system", mode)
		}
	})

	t.Run("rollback passes a system config on to the engine", func(t *testing.T) {
		withFlakeEntrypoint(t)
		stubNixRootKind(t, "system")
		installFakeBackend(t, "").System = t.TempDir()
		run, _ := newRollbackTestRun()
		if err := runRollback(run, "c"); !errors.Is(err, fs.ErrPermission) {
			t.Fatalf("err = %v, want the engine's root check (fs.ErrPermission)", err)
		}
	})
}

// TestListGenerationsAllScansSystemBase: --all lists the user base, then the system base (or the
// LAYAT_SYSTEM_PROFILE_BASE override), naming each subject <mode>:<name> beside its info.mode. A missing system base is
// an empty listing; an unreadable one is one warning line and skipped.
func TestListGenerationsAllScansSystemBase(t *testing.T) {
	origJSON := flagJSON
	t.Cleanup(func() { flagJSON = origJSON })

	listAll := func() (*listGenerationsRun, *bytes.Buffer, error) {
		run, buf := newListGenerationsTestRun()
		return run, buf, runListAllGenerations(run)
	}

	t.Run("user base then system base, each subject tagged with its mode", func(t *testing.T) {
		flagJSON = true
		fake := installFakeBackend(t, "")
		makeHomeProfiles(t, fake, "a", "b")
		fake.System = t.TempDir()
		makeProfilesUnder(t, fake.System, "a", "s")
		run, buf := newListGenerationsTestRun()
		stderr := captureStderr(t, func() {
			if err := runListAllGenerations(run); err != nil {
				t.Errorf("runListAllGenerations: %v", err)
			}
		})
		if stderr != "" {
			t.Errorf("stderr = %q, want silence", stderr)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		checkConformance(t, buf)
		byName, order := subjectResults(t, buf)
		if want := []string{"home:a", "home:b", "system:a", "system:s"}; !slices.Equal(order, want) {
			t.Fatalf("subjects = %v, want %v (always <mode>:<name>, user base first)", order, want)
		}
		for _, name := range order {
			mode, _, _ := strings.Cut(name, ":")
			if got := byName[name]["result"].(map[string]any)["info"].(map[string]any)["mode"]; got != mode {
				t.Errorf("subject %s info.mode = %v, want %s", name, got, mode)
			}
		}
	})

	t.Run("plain output heads system configs with (system)", func(t *testing.T) {
		flagJSON = false
		fake := installFakeBackend(t, "")
		makeHomeProfiles(t, fake, "a")
		fake.System = t.TempDir()
		makeProfilesUnder(t, fake.System, "s")
		out := captureStdout(t, func() {
			if _, _, err := listAll(); err != nil {
				t.Errorf("runListAllGenerations: %v", err)
			}
		})
		if !strings.Contains(out, "# a\n") || !strings.Contains(out, "# s (system)\n") {
			t.Errorf("stdout = %q, want the headings \"# a\" and \"# s (system)\"", out)
		}
	})

	t.Run("LAYAT_SYSTEM_PROFILE_BASE replaces the system base", func(t *testing.T) {
		flagJSON = true
		fake := installFakeBackend(t, "")
		makeHomeProfiles(t, fake)
		fake.System = t.TempDir()
		makeProfilesUnder(t, fake.System, "ignored")
		override := t.TempDir()
		makeProfilesUnder(t, override, "s")
		t.Setenv(systemBaseEnv, override)
		run, buf := newListGenerationsTestRun()
		if err := runListAllGenerations(run); err != nil {
			t.Fatalf("runListAllGenerations: %v", err)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		if _, order := subjectResults(t, buf); !slices.Equal(order, []string{"system:s"}) {
			t.Errorf("subjects = %v, want [system:s] (the override, not the backend's base)", order)
		}
	})

	t.Run("a missing system base lists nothing, silently", func(t *testing.T) {
		flagJSON = true
		fake := installFakeBackend(t, "")
		makeHomeProfiles(t, fake, "a")
		fake.System = filepath.Join(t.TempDir(), "absent")
		var err error
		stderr := captureStderr(t, func() { _, _, err = listAll() })
		if err != nil || stderr != "" {
			t.Errorf("(err, stderr) = (%v, %q), want (nil, \"\")", err, stderr)
		}
	})

	t.Run("an unreadable system base warns once and is skipped", func(t *testing.T) {
		flagJSON = true
		fake := installFakeBackend(t, "")
		makeHomeProfiles(t, fake, "a")
		notDir := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(notDir, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		fake.System = notDir
		var run *listGenerationsRun
		var buf *bytes.Buffer
		var err error
		stderr := captureStderr(t, func() { run, buf, err = listAll() })
		if err != nil {
			t.Fatalf("runListAllGenerations = %v, want nil (the system base is skipped)", err)
		}
		if lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n"); len(lines) != 1 || !strings.Contains(stderr, notDir) {
			t.Errorf("stderr = %q, want one warning line naming %s", stderr, notDir)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		if _, order := subjectResults(t, buf); !slices.Equal(order, []string{"home:a"}) {
			t.Errorf("subjects = %v, want [home:a] (the user base still listed)", order)
		}
	})
}
