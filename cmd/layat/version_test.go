package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/paths"
)

// TestVersionDefault pins the ldflags-unset default. A plain `go build` (no -X main.version=...)
// must leave version at "dev" so the CLI still works out of tree (→ ADR-0042 acceptance criteria).
// The nix build overrides this via ldflags; go test runs without them (the flake's custom checkPhase
// deliberately omits ldflags — see flake.nix), so this test observes the default.
func TestVersionDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("version = %q, want %q (ldflags-unset default)", version, "dev")
	}
}

// TestRootCmdVersionWired asserts cobra's Version field is wired to the package version variable.
// Guards against the field silently drifting from the variable that ldflags targets (main.version,
// a fixed contract for #130's tool.version supply).
func TestRootCmdVersionWired(t *testing.T) {
	root := newRootCmd()
	if root.Version != version {
		t.Errorf("root.Version = %q, want %q (must track main.version)", root.Version, version)
	}
}

// TestVersionFlagOutput drives `layat --version` end-to-end and observes the actual stdout, not just
// the wired field. ADR-0042 requires cobra's default template ("layat version X.Y.Z\n") unchanged, so
// this catches drift the field-equality check can't — e.g. an errant SetVersionTemplate. cobra prints
// the version via OutOrStdout(), which falls back to os.Stdout, so captureStdout observes it.
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

// TestVersionSubcommandAbsent locks in that `layat version` is NOT a command: cobra's Version field
// adds a --version flag only, never a `version` subcommand. This pins the actual UX so a comment or
// doc claiming otherwise can't drift back in (→ diff-review must finding).
func TestVersionSubcommandAbsent(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"version"})
	if err == nil && cmd != nil && cmd != root {
		t.Errorf("`version` resolved to subcommand %q, want no such subcommand (only --version flag)", cmd.Name())
	}
}

// legacyStateDirFixture points $XDG_STATE_HOME at a temp dir holding the pre-rename profile
// base and returns the exact line the hint must produce for it. The wanted line is built from
// the same format string the CLI uses, so the assertions below pin the hint's *contents*
// (which paths it names) without restating the sentence a second time.
func legacyStateDirFixture(t *testing.T) string {
	t.Helper()
	state := t.TempDir()
	if err := os.MkdirAll(legacyStateDir(state), 0o755); err != nil {
		t.Fatalf("MkdirAll(legacy state dir): %v", err)
	}
	t.Setenv("XDG_STATE_HOME", state)
	return fmt.Sprintf(legacyStateDirHintFmt+"\n", legacyStateDir(state), paths.Base(state))
}

// TestLegacyStateDirHintOnSubcommand drives a real subcommand through Execute with
// $XDG_STATE_HOME pointed at a temp dir holding <state>/nix/profiles/nput/, and asserts the
// migration hint reaches stderr exactly once (→ ADR-0054 §8, issue #389). `gitignore` with no
// argument passes cobra's Args check (MaximumNArgs(1)) and then returns from RunE's arity
// branch, so the run reaches PersistentPreRun and stops before any entrypoint discovery.
// SilenceErrors keeps cobra from printing that error, so stderr holds the hint and nothing
// else — which lets this assert the exact bytes rather than mere containment.
func TestLegacyStateDirHintOnSubcommand(t *testing.T) {
	// gitignore's RunE calls beginGitignoreRun before its arity check, so the run reaches
	// the package-global nifaceReport. Save and restore it as niface_test.go does, so this
	// test leaves no state behind for whatever runs next.
	origReport := nifaceReport
	defer func() { nifaceReport = origReport }()

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

// TestLegacyStateDirHintContent pins the four facts ADR-0054 §8 requires the one line to carry:
// the old directory exists, generations are not carried over, the README's "Migrating from nput"
// section decides between migrating and deleting, and nothing is GC-collected until it is gone.
// Asserting on substrings rather than the whole line keeps the wording free to change while the
// facts stay; the hint is removed wholesale in the next minor (→ issue #392).
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

// TestLegacyStateDirHintAbsentWithoutLegacyDir is the negative half: with $XDG_STATE_HOME at a
// temp dir that has no nput profile directory, the run must say nothing. os.Stat is the only
// filesystem access the hint performs, so an empty state base is the whole condition.
func TestLegacyStateDirHintAbsentWithoutLegacyDir(t *testing.T) {
	origReport := nifaceReport
	defer func() { nifaceReport = origReport }()

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

// TestLegacyStateDirHintUnresolvableStateBase pins the silence when no state base can be
// resolved at all: with neither $XDG_STATE_HOME nor $HOME set, paths.StateDir fails and the
// hint says nothing (→ main.go). Without that guard the empty state dir would make
// legacyStateDir return the relative "nix/profiles/nput", and the stat would then be answered
// by whatever happens to sit under the working directory — a hint naming a relative path in a
// checkout that has one.
func TestLegacyStateDirHintUnresolvableStateBase(t *testing.T) {
	origReport := nifaceReport
	defer func() { nifaceReport = origReport }()

	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "")
	// legacyStateDir("") is the relative path an unguarded run would stat, so the decoy goes
	// exactly there, under the working directory: the test fails if the hint is produced from
	// it. Deriving the decoy's path from the implementation keeps the two from drifting apart
	// — a literal here would quietly stop covering anything if the layout ever moved.
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

// TestLegacyStateDirHintNonDirectoryIgnored pins the other half of the stat verdict: a plain
// file at <state>/nix/profiles/nput is not the state directory the hint describes, so nothing
// is said about it. Without the IsDir test the line would tell a stray file a story about
// generations it does not hold (→ ADR-0054 §8).
func TestLegacyStateDirHintNonDirectoryIgnored(t *testing.T) {
	origReport := nifaceReport
	defer func() { nifaceReport = origReport }()

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

// TestLegacyStateDirHintNotInJSONEnvelope pins that the hint stays off the --json envelope: it
// is a niface-conformant machine contract and must not carry a tool-side announcement, so
// stdout holds the document alone while the hint goes to stderr (→ ADR-0043, ADR-0054 §6).
// nifaceReport.emit is called from main, not from Execute (→ main.go), so the emit has to run
// inside the capture the way TestJSONEndToEndSubjectBorneFailure does it — otherwise no
// envelope is produced and the stdout assertion passes against an empty string. `gitignore`
// with no argument fails after beginGitignoreRun has published the run, which is what makes
// the envelope exist at all. flagJSON is a package global that cobra's flag parsing sets, so
// it is restored here rather than left set for the next test.
func TestLegacyStateDirHintNotInJSONEnvelope(t *testing.T) {
	origReport := nifaceReport
	origJSON := flagJSON
	defer func() {
		nifaceReport = origReport
		flagJSON = origJSON
	}()
	nifaceReport = noopEmitter{}

	wantHint := legacyStateDirFixture(t)

	var execErr, emitErr error
	var errOut string
	out := captureStdout(t, func() {
		errOut = captureStderr(t, func() {
			root := newRootCmd()
			root.SetArgs([]string{"--json", "gitignore"})
			execErr = root.Execute()
			if !nifaceReport.began() {
				return
			}
			emitErr = nifaceReport.emit(execErr)
		})
	})
	if emitErr != nil {
		t.Fatalf("emit: %v", emitErr)
	}
	if execErr == nil {
		t.Fatal("gitignore without <name> must fail")
	}
	if !nifaceReport.began() {
		t.Fatal("gitignore's RunE did not publish a begun niface run")
	}
	// The envelope must be the whole of stdout: a leading "{" is what ADR-0043 §2's
	// stdout ownership amounts to here, and the hint is what must not be inside it.
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

// TestLegacyStateDirHintAbsentFromVersionFlags pins that `--version` and `--help` never reach
// PersistentPreRun: cobra returns inside execute() before it runs. The hint is a *stderr* line,
// so stderr is what has to be observed — asserting on stdout could not fail even if the hook did
// fire, and TestVersionFlagOutput already pins stdout exactly. Nothing else covers this:
// flake.nix's installCheckPhase captures only stdout (`got=$(... --version)`), so a hint leaking
// onto stderr there would go unnoticed (→ ADR-0042).
func TestLegacyStateDirHintAbsentFromVersionFlags(t *testing.T) {
	legacyStateDirFixture(t)

	for _, flag := range []string{"--version", "--help"} {
		t.Run(flag, func(t *testing.T) {
			root := newRootCmd()
			root.SetArgs([]string{flag})
			// captureStdout / captureStderr restore the streams after f returns rather than
			// in a defer, so a t.Fatal inside the closure would leave os.Stdout / os.Stderr
			// pointing at a dead pipe for the rest of the package. The error is carried out
			// and judged here instead.
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
