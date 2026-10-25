package main

// Tests for apply --all over a --manifest set mixing a project-rootKind and a home-rootKind
// link-farm: each lands under its own root, and --project-root / --home-root narrow the set.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yasunori0418/layat/internal/manifest"
)

// writeRootKindLinkFarm writes a manifest.json-only link-farm whose root is rootKind.
func writeRootKindLinkFarm(t *testing.T, rootKind string, entries ...manifest.Entry) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(manifest.Manifest{
		SchemaVersion: 1,
		Root:          manifest.Root{RootKind: rootKind},
		Entries:       entries,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifest.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestRunApplyAllManifestProjectAndHome drives the real engine from inside a git repository: with
// no filter both configs are placed (project under the git toplevel, home under $HOME), and each
// root filter places only its own config. No run calls nix.
func TestRunApplyAllManifestProjectAndHome(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	cases := []struct {
		name               string
		projectRoot        bool
		homeRoot           bool
		wantProj, wantHome bool
	}{
		{"no filter", false, false, true, true},
		{"--project-root", true, false, true, false},
		{"--home-root", false, true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withFlakeEntrypoint(t)
			installFakeBackend(t, "")
			origProject := flagProjectRoot
			t.Cleanup(func() { flagProjectRoot = origProject })
			repo, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("git", "init", repo).CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			// Run from a subdirectory so the project root must come from the git toplevel, not cwd.
			sub := filepath.Join(repo, "sub")
			if err := os.Mkdir(sub, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Chdir(sub)
			home := os.Getenv("HOME")
			flagFile, flagDryrun = "", false
			flagProjectRoot, flagHomeRoot = tc.projectRoot, tc.homeRoot
			withManifestFlags(t, true,
				"proj="+writeRootKindLinkFarm(t, manifest.RootKindProject, symlinkEntries(t, "from-proj")...),
				"home="+writeRootKindLinkFarm(t, manifest.RootKindHome, symlinkEntries(t, "from-home")...))
			log := stubNixForApply(t, `{}`)
			run, _ := newApplyTestRun()
			if err := runApplyAll(run); err != nil {
				t.Fatalf("runApplyAll: %v", err)
			}
			for _, c := range []struct {
				path string
				want bool
			}{
				{filepath.Join(repo, "from-proj"), tc.wantProj},
				{filepath.Join(home, "from-home"), tc.wantHome},
				{filepath.Join(home, "from-proj"), false},
				{filepath.Join(repo, "from-home"), false},
				{filepath.Join(sub, "from-proj"), false},
			} {
				_, err := os.Lstat(c.path)
				if placed := err == nil; placed != c.want {
					t.Errorf("%s placed = %v, want %v", c.path, placed, c.want)
				}
			}
			if calls := nixCalls(t, log); len(calls) != 0 {
				t.Errorf("nix calls = %q, want none", calls)
			}
		})
	}
}
