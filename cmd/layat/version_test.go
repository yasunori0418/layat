package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/paths"
)

// TestVersionDefault: without ldflags (as in go test), version stays "dev".
func TestVersionDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("version = %q, want %q (ldflags-unset default)", version, "dev")
	}
}

// TestRootCmdVersionWired asserts cobra's Version field is the package version variable.
func TestRootCmdVersionWired(t *testing.T) {
	root := newRootCmd()
	if root.Version != version {
		t.Errorf("root.Version = %q, want %q (must track main.version)", root.Version, version)
	}
}

// TestVersionFlagOutput drives `layat --version` and checks stdout against cobra's default
// template ("layat version X.Y.Z\n").
func TestVersionFlagOutput(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"--version"})
	out := captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatalf("Execute(--version): %v", err)
		}
	})
	want := "layat version " + version + "\n"
	if out != want {
		t.Errorf("`layat --version` output = %q, want %q (cobra default template)", out, want)
	}
}

// TestVersionSubcommandAbsent: `layat version` is not a command; cobra adds only --version.
func TestVersionSubcommandAbsent(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"version"})
	if err == nil && cmd != nil && cmd != root {
		t.Errorf("`version` resolved to subcommand %q, want no such subcommand (only --version flag)", cmd.Name())
	}
}

// legacyStateDirFixture points $XDG_STATE_HOME at a temp dir holding the pre-rename profile base
// and returns the hint line expected for it, built from the CLI's format string.
func legacyStateDirFixture(t *testing.T) string {
	t.Helper()
	state := t.TempDir()
	if err := os.MkdirAll(legacyStateDir(state), 0o755); err != nil {
		t.Fatalf("MkdirAll(legacy state dir): %v", err)
	}
	t.Setenv("XDG_STATE_HOME", state)
	return fmt.Sprintf(legacyStateDirHintFmt+"\n", legacyStateDir(state), paths.Base(state))
}

// TestLegacyStateDirHintOnSubcommand: with <state>/nix/profiles/nput/ present, a subcommand writes
// the hint to stderr exactly once. `gitignore` with no argument fails in RunE after PersistentPreRun,
// before any entrypoint discovery.
func TestLegacyStateDirHintOnSubcommand(t *testing.T) {
	// gitignore's RunE publishes its run to outturnReport, so save and restore it.
	origReport := outturnReport
	defer func() { outturnReport = origReport }()

	want := legacyStateDirFixture(t)

	root := newRootCmd()
	root.SetArgs([]string{"gitignore"})
	errOut := captureStderr(t, func() {
		// The command itself is expected to fail on arity; only the hint matters here.
		_ = root.Execute()
	})
	if errOut != want {
		t.Errorf("stderr = %q, want exactly the hint %q", errOut, want)
	}
}

// TestLegacyStateDirHintContent checks the facts the hint carries: the old directory exists,
// generations are not carried over, the README section decides, and GC is blocked until removal.
func TestLegacyStateDirHintContent(t *testing.T) {
	hint := fmt.Sprintf(legacyStateDirHintFmt, legacyStateDir("/state"), paths.Base("/state"))
	for _, want := range []string{
		// the old state directory, named as the thing that was found
		"/state/nix/profiles/nput",
		// generations are not carried over
		"generations are not carried over",
		// where the decision (migrate or delete) is written down
		"Migrating from nput",
		// leaving it in place keeps the old generations off the GC's reach
		"garbage collect",
	} {
		if !strings.Contains(hint, want) {
			t.Errorf("hint %q does not mention %q", hint, want)
		}
	}
}

// TestLegacyStateDirHintAbsentWithoutLegacyDir: without the nput profile directory, the run says
// nothing.
func TestLegacyStateDirHintAbsentWithoutLegacyDir(t *testing.T) {
	origReport := outturnReport
	defer func() { outturnReport = origReport }()

	t.Setenv("XDG_STATE_HOME", t.TempDir())

	root := newRootCmd()
	root.SetArgs([]string{"gitignore"})
	errOut := captureStderr(t, func() {
		_ = root.Execute()
	})
	if errOut != "" {
		t.Errorf("stderr = %q, want nothing (no legacy state directory)", errOut)
	}
}

// TestLegacyStateDirHintUnresolvableStateBase: with neither $XDG_STATE_HOME nor $HOME set, the hint
// says nothing, even when a relative nix/profiles/nput exists under the working directory.
func TestLegacyStateDirHintUnresolvableStateBase(t *testing.T) {
	origReport := outturnReport
	defer func() { outturnReport = origReport }()

	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "")
	// Put a decoy at the relative path an unguarded run would stat, derived from legacyStateDir("").
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(legacyStateDir(""), 0o755); err != nil {
		t.Fatalf("MkdirAll(relative decoy): %v", err)
	}

	root := newRootCmd()
	root.SetArgs([]string{"gitignore"})
	errOut := captureStderr(t, func() {
		_ = root.Execute()
	})
	if errOut != "" {
		t.Errorf("stderr = %q, want nothing (no state base can be resolved)", errOut)
	}
}

// TestLegacyStateDirHintNonDirectoryIgnored: a plain file at <state>/nix/profiles/nput gets no
// hint.
func TestLegacyStateDirHintNonDirectoryIgnored(t *testing.T) {
	origReport := outturnReport
	defer func() { outturnReport = origReport }()

	state := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(legacyStateDir(state)), 0o755); err != nil {
		t.Fatalf("MkdirAll(profiles): %v", err)
	}
	if err := os.WriteFile(legacyStateDir(state), nil, 0o644); err != nil {
		t.Fatalf("WriteFile(legacy state path): %v", err)
	}
	t.Setenv("XDG_STATE_HOME", state)

	root := newRootCmd()
	root.SetArgs([]string{"gitignore"})
	errOut := captureStderr(t, func() {
		_ = root.Execute()
	})
	if errOut != "" {
		t.Errorf("stderr = %q, want nothing (the path is a file, not the state directory)", errOut)
	}
}

// TestLegacyStateDirHintNotInJSONEnvelope: under --json, stdout holds only the envelope and the
// hint goes to stderr. The emit runs inside the capture, since main (not Execute) emits it; flagJSON
// is restored afterwards.
func TestLegacyStateDirHintNotInJSONEnvelope(t *testing.T) {
	origReport := outturnReport
	origJSON := flagJSON
	defer func() {
		outturnReport = origReport
		flagJSON = origJSON
	}()
	outturnReport = noopEmitter{}

	wantHint := legacyStateDirFixture(t)

	var execErr, emitErr error
	var errOut string
	out := captureStdout(t, func() {
		errOut = captureStderr(t, func() {
			root := newRootCmd()
			root.SetArgs([]string{"--json", "gitignore"})
			execErr = root.Execute()
			if !outturnReport.began() {
				return
			}
			emitErr = outturnReport.emit(execErr)
		})
	})
	if emitErr != nil {
		t.Fatalf("emit: %v", emitErr)
	}
	if execErr == nil {
		t.Fatal("gitignore without <name> must fail")
	}
	if !outturnReport.began() {
		t.Fatal("gitignore's RunE did not publish a begun outturn run")
	}
	// The envelope must be the whole of stdout.
	if !strings.HasPrefix(out, "{") {
		t.Fatalf("stdout must hold the envelope alone (the --json contract), got %q", out)
	}
	if strings.Contains(out, "Migrating from nput") || strings.Contains(out, "profiles/nput") {
		t.Errorf("the --json envelope carries the hint: %q", out)
	}
	if !strings.Contains(errOut, wantHint) {
		t.Errorf("stderr = %q, want it to contain the hint %q", errOut, wantHint)
	}
}

// TestLegacyStateDirHintAbsentFromVersionFlags: `--version` and `--help` return before
// PersistentPreRun, so stderr carries no hint.
func TestLegacyStateDirHintAbsentFromVersionFlags(t *testing.T) {
	legacyStateDirFixture(t)

	for _, flag := range []string{"--version", "--help"} {
		t.Run(flag, func(t *testing.T) {
			root := newRootCmd()
			root.SetArgs([]string{flag})
			// captureStdout / captureStderr restore the streams after f returns, not in a defer, so the error
			// is carried out and judged here instead of calling t.Fatal inside.
			var err error
			// --help writes to stdout; discard it so only stderr is under test.
			errOut := captureStderr(t, func() {
				_ = captureStdout(t, func() {
					err = root.Execute()
				})
			})
			if err != nil {
				t.Fatalf("Execute(%s): %v", flag, err)
			}
			if errOut != "" {
				t.Errorf("`layat %s` wrote %q to stderr, want nothing (PersistentPreRun must not run)", flag, errOut)
			}
		})
	}
}
