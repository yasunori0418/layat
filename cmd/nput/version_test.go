package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

// TestRenameNoticeOnSubcommand drives a subcommand through Execute and asserts the rename
// notice reaches stderr exactly once (→ ADR-0054 §6, issue #387). `gitignore` is used because
// it is a real subcommand whose RunE fails fast without an entrypoint, so the run reaches
// PersistentPreRun and returns without touching the filesystem.
func TestRenameNoticeOnSubcommand(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"gitignore", "--file", filepath.Join(t.TempDir(), "absent.nix")})
	errOut := captureStderr(t, func() {
		// The command itself is expected to fail (no entrypoint); only the notice matters here.
		_ = root.Execute()
	})
	if got := strings.Count(errOut, "will be renamed to layat"); got != 1 {
		t.Errorf("rename notice appeared %d times on stderr, want exactly 1; stderr = %q", got, errOut)
	}
	if !strings.Contains(errOut, renameNotice) {
		t.Errorf("stderr does not carry the full notice; stderr = %q, want it to contain %q", errOut, renameNotice)
	}
}

// TestRenameNoticeSkippedForCompletion pins the one exemption: cobra's completion request must
// not carry the notice, or it would land inside a completion script's output (→ ADR-0054 §6).
func TestRenameNoticeSkippedForCompletion(t *testing.T) {
	for _, name := range []string{cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd} {
		t.Run(name, func(t *testing.T) {
			errOut := captureStderr(t, func() {
				printRenameNotice(&cobra.Command{Use: name})
			})
			if errOut != "" {
				t.Errorf("printRenameNotice(%q) wrote %q to stderr, want nothing", name, errOut)
			}
		})
	}
}

// TestRenameNoticeAbsentFromVersionOutput guards the contract the nix installCheckPhase relies on:
// `nput --version` must print the cobra template to stdout and nothing else, since the phase
// matches the output exactly (→ ADR-0042, flake.nix installCheckPhase).
func TestRenameNoticeAbsentFromVersionOutput(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"--version"})
	out := captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatalf("Execute(--version): %v", err)
		}
	})
	if strings.Contains(out, "will be renamed to layat") {
		t.Errorf("`nput --version` stdout carries the rename notice: %q", out)
	}
}
