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
		projectDir  bool   // layat.toml is a directory: present but unreadable
		noUserDir   bool   // no user config directory could be resolved
		want        string
		wantErr     bool
		errContains string
		wantCode    string // set: the error is not an inputError and classifies as this code
	}{
		{name: "default when nothing is set", want: "nix"},
		{name: "flag wins over everything", flag: "nix", env: "bogus", project: bogus, user: bogus, want: "nix"},
		{name: "env wins over both settings files", env: "nix", project: bogus, user: bogus, want: "nix"},
		{name: "project setting wins over the user setting", project: nixKey, user: bogus, want: "nix"},
		{name: "user setting applies alone", user: nixKey, want: "nix"},
		{name: "empty env passes on", env: "", project: bogus, wantErr: true, errContains: "layat.toml"},
		{name: "project file without the key passes on", project: noKey, user: bogus, wantErr: true, errContains: "config.toml"},
		{name: "user file without the key falls back to nix", user: noKey, want: "nix"},
		{name: "unknown flag value", flag: "bogus", wantErr: true, errContains: "--generator"},
		{name: "unknown env value", env: "bogus", wantErr: true, errContains: "LAYAT_GENERATOR"},
		{name: "unknown project value", project: bogus, wantErr: true, errContains: "layat.toml"},
		{name: "unknown user value", user: bogus, wantErr: true, errContains: "config.toml"},
		{name: "unknown key is rejected (strict)", project: unknown, wantErr: true, errContains: "version"},
		{name: "unknown key in the user file is rejected", user: unknown, wantErr: true, errContains: "version"},
		{name: "malformed TOML is rejected", project: malformed, wantErr: true, errContains: "layat.toml"},
		{name: "a decided step does not read the steps below", project: nixKey, user: malformed, want: "nix"},
		{name: "a set env does not read a broken project file", env: "nix", project: malformed, want: "nix"},
		{name: "an unreadable settings file keeps its own classification", projectDir: true, wantErr: true,
			errContains: "layat.toml", wantCode: "E_IO"},
		{name: "no user config directory skips the user setting", noUserDir: true, want: "nix"},
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
			if c.projectDir {
				if err := os.Mkdir(filepath.Join(projectDir, "layat.toml"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if c.noUserDir {
				userConfigDir = ""
			}
			got, err := resolveGenerator(c.flag, c.env, projectDir, userConfigDir)
			if c.wantErr {
				var ie *inputError
				if c.wantCode != "" {
					if err == nil || errors.As(err, &ie) {
						t.Fatalf("resolveGenerator() error = %v, want a read failure that is not an inputError", err)
					}
					if got := classifyError(err).Code; got != c.wantCode {
						t.Errorf("classifyError().Code = %s, want %s", got, c.wantCode)
					}
				} else if !errors.As(err, &ie) {
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

// withSelectionFlags isolates the flags and environment the generator selection reads.
func withSelectionFlags(t *testing.T) {
	t.Helper()
	origGen, origManifest, origFile, origAll := flagGenerator, flagManifest, flagFile, flagApplyAll
	t.Cleanup(func() { flagGenerator, flagManifest, flagFile, flagApplyAll = origGen, origManifest, origFile, origAll })
	flagGenerator, flagManifest, flagFile, flagApplyAll = "", "", "", false
	t.Setenv(generatorEnv, "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Chdir(t.TempDir())
}

// TestSelectGeneratorManifestIgnoresEnvironment pins that apply --manifest reads neither
// LAYAT_GENERATOR nor any settings file: an invalid value in each still selects prebuilt, so
// module activation does not depend on the environment it runs in (→ ADR-0056 §2).
func TestSelectGeneratorManifestIgnoresEnvironment(t *testing.T) {
	withSelectionFlags(t)
	t.Setenv(generatorEnv, "bogus")
	writeSetting(t, "layat.toml", "bogus = true\n")
	writeSetting(t, filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "layat", "config.toml"), "bogus = true\n")
	flagManifest = "/nonexistent/link-farm"
	got, err := selectGenerator()
	if err != nil || got != "prebuilt" {
		t.Fatalf("selectGenerator() = %q, %v; want prebuilt, nil", got, err)
	}
}

// TestSelectGeneratorRejectsGeneratorWithManifest pins that --generator and --manifest together
// are an input error (→ ADR-0056 §2, §5).
func TestSelectGeneratorRejectsGeneratorWithManifest(t *testing.T) {
	withSelectionFlags(t)
	flagGenerator, flagManifest = "nix", "/nonexistent/link-farm"
	_, err := selectGenerator()
	var ie *inputError
	if !errors.As(err, &ie) {
		t.Fatalf("selectGenerator() error = %v, want an inputError", err)
	}
	if got := classifyError(err).Code; got != "E_INPUT" {
		t.Errorf("classifyError().Code = %s, want E_INPUT", got)
	}
}

// TestSelectGeneratorWiring pins where the command reads each step from: the flag, the
// environment, layat.toml in the -f directory (its parent when -f names a file) or the CWD, and
// config.toml under XDG_CONFIG_HOME. An -f path that cannot be stat'ed skips the settings files,
// leaving the failure to the generator's discovery (→ ADR-0056 §3).
func TestSelectGeneratorWiring(t *testing.T) {
	const bad = "bogus = true\n"
	cases := []struct {
		name    string
		setup   func(t *testing.T)
		wantErr string // "" = selects nix; otherwise the error must name it
	}{
		{"explicit --generator nix beats a bad env", func(t *testing.T) {
			flagGenerator = "nix"
			t.Setenv(generatorEnv, "bogus")
		}, ""},
		{"unset flag reads the env", func(t *testing.T) { t.Setenv(generatorEnv, "bogus") }, generatorEnv},
		{"layat.toml in the CWD", func(t *testing.T) { writeSetting(t, "layat.toml", bad) }, "layat.toml"},
		{"layat.toml in the -f directory, not the CWD", func(t *testing.T) {
			dir := t.TempDir()
			writeSetting(t, filepath.Join(dir, "layat.toml"), bad)
			writeSetting(t, "layat.toml", "generator = \"nix\"\n")
			flagFile = dir
		}, "layat.toml"},
		{"layat.toml beside an -f file", func(t *testing.T) {
			dir := t.TempDir()
			writeSetting(t, filepath.Join(dir, "layat.toml"), bad)
			writeSetting(t, filepath.Join(dir, "flake.nix"), "{}\n")
			flagFile = filepath.Join(dir, "flake.nix")
		}, "layat.toml"},
		{"the -f directory's missing layat.toml does not fall back to the CWD", func(t *testing.T) {
			writeSetting(t, "layat.toml", bad)
			flagFile = t.TempDir()
		}, ""},
		{"an -f path that cannot be stat'ed skips the settings files", func(t *testing.T) {
			writeSetting(t, "layat.toml", bad)
			writeSetting(t, filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "layat", "config.toml"), bad)
			flagFile = filepath.Join(t.TempDir(), "missing")
		}, ""},
		{"config.toml under XDG_CONFIG_HOME", func(t *testing.T) {
			writeSetting(t, filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "layat", "config.toml"), bad)
		}, "config.toml"},
		{"config.toml under ~/.config when XDG_CONFIG_HOME is unset", func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", "")
			writeSetting(t, filepath.Join(home, ".config", "layat", "config.toml"), bad)
		}, filepath.Join(".config", "layat", "config.toml")},
		{"neither XDG_CONFIG_HOME nor HOME skips the user setting", func(t *testing.T) {
			t.Setenv("HOME", "")
			t.Setenv("XDG_CONFIG_HOME", "")
		}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withSelectionFlags(t)
			c.setup(t)
			got, err := selectGenerator()
			if c.wantErr == "" {
				if err != nil || got != "nix" {
					t.Fatalf("selectGenerator() = %q, %v; want nix, nil", got, err)
				}
				return
			}
			var ie *inputError
			if !errors.As(err, &ie) || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("selectGenerator() error = %v, want an inputError naming %q", err, c.wantErr)
			}
		})
	}
}
