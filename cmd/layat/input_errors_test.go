package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// executeInputCmd runs cmd with args and returns RunE's error, restoring outturnReport afterwards.
func executeInputCmd(t *testing.T, cmd *cobra.Command, args ...string) error {
	t.Helper()
	origReport := outturnReport
	t.Cleanup(func() { outturnReport = origReport })
	cmd.SetArgs(args)
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	return cmd.Execute()
}

// withRootFilters sets --project-root / --home-root for one test and restores them afterwards.
func withRootFilters(t *testing.T, project, home bool) {
	t.Helper()
	origProject, origHome := flagProjectRoot, flagHomeRoot
	t.Cleanup(func() { flagProjectRoot, flagHomeRoot = origProject, origHome })
	flagProjectRoot, flagHomeRoot = project, home
}

// TestInputErrorsClassifiedEInput: each input known to be wrong before the run starts is E_INPUT
// under --json and stays a single line on stderr.
func TestInputErrorsClassifiedEInput(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T) error
	}{
		{"apply <name> with --all", func(t *testing.T) error {
			return executeApply(t, "--all", "foo")
		}},
		{"--jobs without --all", func(t *testing.T) error {
			return executeApply(t, "foo", "--jobs", "2")
		}},
		{"more than one root filter", func(t *testing.T) error {
			withRootFilters(t, true, true)
			return executeApply(t, "--all")
		}},
		{"root filter without --all", func(t *testing.T) error {
			withRootFilters(t, true, false)
			return executeApply(t, "foo")
		}},
		{"negative --jobs", func(t *testing.T) error {
			return executeApply(t, "--all", "--jobs", "-1")
		}},
		{"gitignore <name> with --all", func(t *testing.T) error {
			return executeInputCmd(t, newGitignoreCmd(), "--all", "foo")
		}},
		{"list-generations <name> with --all", func(t *testing.T) error {
			return executeInputCmd(t, newListGenerationsCmd(), "--all", "foo")
		}},
		{"gitignore without <name> or --all", func(t *testing.T) error {
			return executeInputCmd(t, newGitignoreCmd())
		}},
		{"list-generations without <name> or --all", func(t *testing.T) error {
			return executeInputCmd(t, newListGenerationsCmd())
		}},
		{"init with an unknown template", func(t *testing.T) error {
			return executeInputCmd(t, newInitCmd(), "bogus")
		}},
		{"destructive operation without --yes in a non-interactive run", func(t *testing.T) error {
			_, err := confirmPolicy(false, false, "reset")
			return err
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.run(t)
			if err == nil {
				t.Fatal("err = nil, want an input error")
			}
			if got := classifyError(err).Code; got != "E_INPUT" {
				t.Errorf("classifyError(%q).Code = %s, want E_INPUT", err, got)
			}
			if strings.Contains(err.Error(), "\n") {
				t.Errorf("err = %q, want a single line", err)
			}
		})
	}
}
