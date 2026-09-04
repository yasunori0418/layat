package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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

// TestVersionFlagOutput drives `nput --version` end-to-end and observes the actual stdout, not just
// the wired field. ADR-0042 requires cobra's default template ("nput version X.Y.Z\n") unchanged, so
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
	want := "nput version " + version + "\n"
	if out != want {
		t.Errorf("`nput --version` output = %q, want %q (cobra default template)", out, want)
	}
}

// TestVersionSubcommandAbsent locks in that `nput version` is NOT a command: cobra's Version field
// adds a --version flag only, never a `version` subcommand. This pins the actual UX so a comment or
// doc claiming otherwise can't drift back in (→ diff-review must finding).
func TestVersionSubcommandAbsent(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"version"})
	if err == nil && cmd != nil && cmd != root {
		t.Errorf("`version` resolved to subcommand %q, want no such subcommand (only --version flag)", cmd.Name())
	}
}

// TestRenameNoticeOnSubcommand drives a real subcommand through Execute and asserts the rename
// notice reaches stderr exactly once (→ ADR-0054 §6, issue #387). `gitignore` with no argument
// passes cobra's Args check (MaximumNArgs(1)) and then returns from RunE's arity branch, so the
// run reaches PersistentPreRun and stops before any entrypoint discovery or filesystem access.
// SilenceErrors keeps cobra from printing that error, so stderr holds the notice and nothing
// else — which lets this assert the exact bytes rather than mere containment.
func TestRenameNoticeOnSubcommand(t *testing.T) {
	// gitignore's RunE calls beginGitignoreRun before its arity check, so the run reaches
	// the package-global nifaceReport. Save and restore it as niface_test.go does, so this
	// test leaves no state behind for whatever runs next.
	origReport := nifaceReport
	defer func() { nifaceReport = origReport }()

	root := newRootCmd()
	root.SetArgs([]string{"gitignore"})
	errOut := captureStderr(t, func() {
		// The command itself is expected to fail on arity; only the notice matters here.
		_ = root.Execute()
	})
	if want := renameNotice + "\n"; errOut != want {
		t.Errorf("stderr = %q, want exactly the notice %q", errOut, want)
	}
}

// TestRenameNoticeAbsentFromVersionFlags pins that `--version` and `--help` never reach
// PersistentPreRun: cobra returns inside execute() before it runs. The notice is a *stderr*
// line, so stderr is what has to be observed — asserting on stdout could not fail even if the
// hook did fire, and TestVersionFlagOutput already pins stdout exactly. Regressing this would
// break flake.nix's installCheckPhase, which matches `nput --version` output exactly
// (→ ADR-0042, ADR-0054 §6).
func TestRenameNoticeAbsentFromVersionFlags(t *testing.T) {
	for _, flag := range []string{"--version", "--help"} {
		t.Run(flag, func(t *testing.T) {
			root := newRootCmd()
			root.SetArgs([]string{flag})
			// --help writes to stdout; discard it so only stderr is under test.
			errOut := captureStderr(t, func() {
				_ = captureStdout(t, func() {
					if err := root.Execute(); err != nil {
						t.Fatalf("Execute(%s): %v", flag, err)
					}
				})
			})
			if errOut != "" {
				t.Errorf("`nput %s` wrote %q to stderr, want nothing (PersistentPreRun must not run)", flag, errOut)
			}
		})
	}
}

// TestRenameNoticeMatchesNixSource pins the byte-equality of the two copies of the notice: the
// Go const above and modules/common.nix's `renameNotice`. Nothing else holds them together — a
// date bumped on one side only would ship a CLI that contradicts the module warning. Rather
// than shelling out to `nix eval` (this suite is stdlib-only and must run without nix), it
// reconstructs the string from the Nix source's literal concatenation. Both copies go away
// with the rename PR (→ issue #388), and so does this test.
func TestRenameNoticeMatchesNixSource(t *testing.T) {
	// The nix build's goSrc is go.mod / go.sum / internal / cmd only, so modules/ is absent
	// when `go test` runs inside the sandbox. Skip there rather than widening the package's
	// source closure for a temporary check; checks.notice-parity (flake.nix) enforces the same
	// pairing in an environment where both files exist, so the contract is never unguarded.
	// Skip only where the whole modules/ tree is absent — that is the nix sandbox. A missing
	// file inside an existing modules/ means it moved or was deleted, which must fail loudly
	// rather than pass as a silent skip.
	src, err := os.ReadFile(filepath.Join("..", "..", "modules", "common.nix"))
	if errors.Is(err, fs.ErrNotExist) {
		if _, dirErr := os.Stat(filepath.Join("..", "..", "modules")); errors.Is(dirErr, fs.ErrNotExist) {
			t.Skip("modules/ is out of tree (nix build sandbox); checks.notice-parity covers this")
		}
	}
	if err != nil {
		t.Fatalf("read modules/common.nix: %v", err)
	}

	// Take everything from `renameNotice =` to the terminating `;`, then concatenate the
	// string literals it is built from.
	body := regexp.MustCompile(`(?s)renameNotice =(.*?);\n`).FindStringSubmatch(string(src))
	if body == nil {
		t.Fatalf("no renameNotice binding found in modules/common.nix")
	}
	var got string
	for _, part := range regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(body[1], -1) {
		got += strings.ReplaceAll(part[1], `\"`, `"`)
	}

	if got != renameNotice {
		t.Errorf("Nix and Go copies of the rename notice differ:\n nix = %q\n  go = %q", got, renameNotice)
	}
}
