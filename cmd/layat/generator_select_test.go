package main

// Tests for the manifest generator selection (→ ADR-0056): the five-step precedence on the pure
// resolver, the strict settings file, and the --manifest path that reads neither the environment
// nor any settings file. None of them runs nix.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSetting writes body to path, creating its directory.
func writeSetting(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestResolveGenerator pins the precedence --generator > LAYAT_GENERATOR > layat.toml >
// config.toml > nix, where an empty value and a settings file without the generator key pass on
// to the next step, and an invalid value or settings file is an inputError.
func TestResolveGenerator(t *testing.T) {
	const (
		bogus     = "generator = \"bogus\"\n"
		nixKey    = "generator = \"nix\"\n"
		noKey     = "# no generator here\n"
		unknown   = "generator = \"nix\"\nversion = 1\n"
		malformed = "generator = \n"
	)
	cases := []struct {
		name        string
		flag, env   string
		project     string // layat.toml body; "" = no file
		user        string // config.toml body; "" = no file
		want        string
		wantErr     bool
		errContains string
	}{
		{name: "default when nothing is set", want: "nix"},
		{name: "flag wins over env", flag: "nix", env: "bogus", want: "nix"},
		{name: "env applies alone", env: "nix", want: "nix"},
		{name: "unknown flag value", flag: "bogus", wantErr: true, errContains: "--generator"},
		{name: "unknown env value", env: "bogus", wantErr: true, errContains: "LAYAT_GENERATOR"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			projectDir, userConfigDir := t.TempDir(), t.TempDir()
			if c.project != "" {
				writeSetting(t, filepath.Join(projectDir, "layat.toml"), c.project)
			}
			if c.user != "" {
				writeSetting(t, filepath.Join(userConfigDir, "layat", "config.toml"), c.user)
			}
			got, err := resolveGenerator(c.flag, c.env, projectDir, userConfigDir)
			if c.wantErr {
				var ie *inputError
				if !errors.As(err, &ie) {
					t.Fatalf("resolveGenerator() error = %v, want an inputError", err)
				}
				msg := err.Error()
				if !strings.Contains(msg, c.errContains) {
					t.Errorf("error %q does not name %q", msg, c.errContains)
				}
				if strings.Contains(msg, "\n") {
					t.Errorf("error %q must be a single line", msg)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveGenerator() error = %v", err)
			}
			if got != c.want {
				t.Errorf("resolveGenerator() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestResolveGeneratorUnsetDirs pins that an empty directory skips that settings file (no
// resolvable user config directory, for one).
func TestResolveGeneratorUnsetDirs(t *testing.T) {
	got, err := resolveGenerator("", "", "", "")
	if err != nil || got != "nix" {
		t.Fatalf("resolveGenerator with no dirs = %q, %v; want nix, nil", got, err)
	}
}
