package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// disclosureCmdColumn is the column where a disclosure line's nix command starts (after the label).
const disclosureCmdColumn = 20

// noNixLine is the --help sentence that discloses the subcommands running no nix at all.
const noNixLine = "prune and status run no nix commands."

// disclosureEntry is one label of the --help disclosure block with the nix commands it lists.
type disclosureEntry struct {
	label string
	cmds  []string
}

// parseDisclosure reads the block from the "Internal nix commands" line up to the next blank line.
// A label ending in "/" joins the next line's label; a line without a label continues the commands.
func parseDisclosure(t *testing.T, long string) []disclosureEntry {
	t.Helper()
	lines := strings.Split(long, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "Internal nix commands") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatal("rootCmdLong has no \"Internal nix commands\" block")
	}
	var entries []disclosureEntry
	joining := false
	for _, l := range lines[start:] {
		if strings.TrimSpace(l) == "" {
			break
		}
		label, cmd := l, ""
		if len(l) > disclosureCmdColumn {
			label, cmd = l[:disclosureCmdColumn], strings.TrimSpace(l[disclosureCmdColumn:])
		}
		label = strings.TrimSpace(label)
		switch {
		case label != "" && joining:
			entries[len(entries)-1].label += " " + label
		case label != "":
			entries = append(entries, disclosureEntry{label: label})
		case len(entries) == 0:
			t.Fatalf("disclosure line without a label: %q", l)
		}
		if label != "" {
			joining = strings.HasSuffix(label, "/")
		}
		if cmd != "" {
			e := &entries[len(entries)-1]
			e.cmds = append(e.cmds, perConfigNote.ReplaceAllString(cmd, ""))
		}
	}
	return entries
}

// perConfigNote is the trailing remark of a line that runs once per config.
var perConfigNote = regexp.MustCompile(`\s+\(per config\)$`)

// stubNixRecorder puts a fake nix first on PATH that records each call (args separated by \x1f) to
// the returned log. A batch eval prints one config "a" of kind, any other eval prints kind, a
// --no-link build prints a store path, and an --out-link build fails.
func stubNixRecorder(t *testing.T, kind string) string {
	t.Helper()
	bin := t.TempDir()
	log := filepath.Join(bin, "nix.log")
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do printf '%s\\037' \"$a\" >> '" + log + "'; done\n" +
		"printf '\\n' >> '" + log + "'\n" +
		"case \"$1\" in\n" +
		"eval) case \" $* \" in *' --apply '*) printf '%s' '{\"a\":{\"rootKind\":\"" + kind + "\",\"targets\":[\"t\"]}}';; *) printf '" + kind + "';; esac;;\n" +
		"build) case \" $* \" in *' --no-link '*) echo /nix/store/stub-link-farm;; *) echo 'error: stub build failed' >&2; exit 1;; esac;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(bin, "nix"), []byte(script), 0o755); err != nil {
		t.Fatalf("write nix stub: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return log
}

// flakeAttr splits "<ep>#layat.<system>[.<name>]<rest>" for the stub config "a".
var flakeAttr = regexp.MustCompile(`^#layat\.([^.]+)(\.a)?(.*)$`)

// recordedCalls reads the stub's log and renders each call in the --help placeholders.
func recordedCalls(t *testing.T, log, ep, ref, template string) []string {
	t.Helper()
	b, err := os.ReadFile(log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read nix log: %v", err)
	}
	var calls []string
	for _, line := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		args := strings.Split(strings.TrimSuffix(line, "\x1f"), "\x1f")
		for i, a := range args {
			prev := ""
			if i > 0 {
				prev = args[i-1]
			}
			switch {
			case prev == "--apply":
				args[i] = "'<rootKind map>'"
			case prev == "--out-link" && filepath.Base(a) == ".pending":
				args[i] = "<profileDir>/.pending"
			case prev == "-t" && a == ref+"#"+template:
				args[i] = "<ref>#<template>"
			case strings.HasPrefix(a, ep+"#"):
				m := flakeAttr.FindStringSubmatch(strings.TrimPrefix(a, ep))
				if m == nil {
					t.Fatalf("unexpected installable %q", a)
				}
				name := ""
				if m[2] != "" {
					name = ".<name>"
				}
				args[i] = "<ep>#layat.<system>" + name + m[3]
			}
		}
		calls = append(calls, "nix "+strings.Join(args, " "))
	}
	return calls
}

// TestHelpDisclosureMatchesNixCalls runs each disclosed subcommand against a recording nix and
// requires the --help block to list exactly the nix calls it made, label for label.
func TestHelpDisclosureMatchesNixCalls(t *testing.T) {
	const ref, template = "github:example/templates", "standalone"
	run := func(t *testing.T, kind string, f func() error) []string {
		t.Helper()
		withFlakeEntrypoint(t)
		installFakeBackend(t, "")
		t.Setenv("LAYAT_TEMPLATE_REF", ref)
		origDebug, origJSON := flagDebug, flagJSON
		t.Cleanup(func() { flagDebug, flagJSON = origDebug, origJSON })
		flagDebug, flagJSON = false, false
		log := stubNixRecorder(t, kind)
		_ = f()
		return recordedCalls(t, log, flagFile, ref, template)
	}
	observed := map[string]func(t *testing.T) [][]string{
		"init <template>": func(t *testing.T) [][]string {
			r, _ := newInitTestRun()
			return [][]string{run(t, "home", func() error { return runInit(r, template) })}
		},
		"apply <name>": func(t *testing.T) [][]string {
			r, _ := newApplyTestRun()
			return [][]string{run(t, "home", func() error { return runApply(r, "a") })}
		},
		"apply --all": func(t *testing.T) [][]string {
			r, _ := newApplyTestRun()
			return [][]string{run(t, "home", func() error { return runApplyAll(r) })}
		},
		"gitignore <name>": func(t *testing.T) [][]string {
			r, _ := newGitignoreTestRun()
			return [][]string{run(t, "project", func() error { return runGitignore(r, "a") })}
		},
		"reset / rollback / list-generations": func(t *testing.T) [][]string {
			rs, _ := newResetTestRun()
			rb, _ := newRollbackTestRun()
			lg, _ := newListGenerationsTestRun()
			return [][]string{
				run(t, "home", func() error { return runReset(rs, "a", nil, true) }),
				run(t, "home", func() error { return runRollback(rb, "a") }),
				run(t, "home", func() error { return runListGenerations(lg, "a") }),
			}
		},
	}

	entries := parseDisclosure(t, rootCmdLong)
	var labels []string
	for _, e := range entries {
		labels = append(labels, e.label)
		probe, ok := observed[e.label]
		if !ok {
			t.Errorf("--help discloses %q, which this test does not run", e.label)
			continue
		}
		t.Run(e.label, func(t *testing.T) {
			for _, got := range probe(t) {
				if !reflect.DeepEqual(got, e.cmds) {
					t.Errorf("nix calls differ from --help\n got: %q\nhelp: %q", got, e.cmds)
				}
			}
		})
	}
	for label := range observed {
		if !slices.Contains(labels, label) {
			t.Errorf("--help no longer discloses %q", label)
		}
	}
}

// TestHelpDisclosesNoNixSubcommands requires the --help sentence on prune / status, and that they
// succeed over a record and an orphan series without running nix; prune --yes deletes the series.
func TestHelpDisclosesNoNixSubcommands(t *testing.T) {
	if !strings.Contains(rootCmdLong, "\n"+noNixLine+"\n") {
		t.Errorf("rootCmdLong lacks the line %q", noNixLine)
	}
	cases := []struct {
		name        string
		run         func(t *testing.T) error
		wantDeleted bool
	}{
		{"prune --dryrun", func(*testing.T) error { r, _ := newPruneTestRun(); return runPrune(r, true, false) }, false},
		{"prune --yes", func(*testing.T) error { r, _ := newPruneTestRun(); return runPrune(r, false, false) }, true},
		{"status", func(t *testing.T) error {
			r, buf := newStatusTestRun()
			flagJSON = true // restored by the subtest's cleanup
			if err := runStatus(r, nil); err != nil {
				return err
			}
			if err := r.emit(nil); err != nil {
				return err
			}
			// The fixture's record must be the one listed, or the run proves nothing.
			if results, _ := decodeEnvelope(t, buf)["results"].([]any); len(results) != 1 {
				return fmt.Errorf("status listed %d records, want the fixture's 1", len(results))
			}
			return nil
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withFlakeEntrypoint(t)
			origDebug, origJSON, origYes := flagDebug, flagJSON, flagYes
			t.Cleanup(func() { flagDebug, flagJSON, flagYes = origDebug, origJSON, origYes })
			flagDebug, flagJSON, flagYes = false, false, true
			t.Setenv(systemBaseEnv, t.TempDir())
			t.Setenv(systemStateEnv, t.TempDir())
			state := t.TempDir()
			t.Setenv("XDG_STATE_HOME", state)
			series := writeNoNixFixture(t, installFakeBackend(t, "").UserBase(state), state)
			log := stubNixRecorder(t, "home")
			if err := tc.run(t); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if calls := recordedCalls(t, log, flagFile, "", ""); len(calls) != 0 {
				t.Errorf("%s ran nix: %q", tc.name, calls)
			}
			if _, err := os.Lstat(series); os.IsNotExist(err) != tc.wantDeleted {
				t.Errorf("%s: orphan series deleted = %v, want %v", tc.name, os.IsNotExist(err), tc.wantDeleted)
			}
		})
	}
}

// writeNoNixFixture puts a home last-applied record for "a" under state and an orphan <roothash>
// series (its root is missing) under base, and returns the series directory.
func writeNoNixFixture(t *testing.T, base, state string) string {
	t.Helper()
	write := func(path, data string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	write(filepath.Join(state, "layat", "a", "last-applied.json"),
		`{"recordVersion":2,"linkFarm":"/nix/store/stub-link-farm","appliedAt":"2026-11-01T00:00:00Z",`+
			`"origin":{"generator":"nix"},"manifest":{"schemaVersion":1,"root":{"rootKind":"home"},"entries":[]}}`)
	series := filepath.Join(base, "orphanhash")
	write(filepath.Join(series, ".root"), filepath.Join(t.TempDir(), "gone")+"\n")
	if err := os.MkdirAll(filepath.Join(series, "a"), 0o755); err != nil {
		t.Fatalf("mkdir series profileDir: %v", err)
	}
	return series
}
