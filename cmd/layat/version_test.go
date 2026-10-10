package main

import "testing"

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
