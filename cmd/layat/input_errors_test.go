package main

import (
	"bytes"
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

// emitApplyEnvelope runs apply with args and decodes the envelope main would emit for its error.
func emitApplyEnvelope(t *testing.T, args ...string) map[string]any {
	t.Helper()
	out := captureStdout(t, func() {
		err := executeApply(t, args...)
		if err == nil {
			t.Error("apply err = nil, want an input error")
		}
		if emitErr := outturnReport.emit(err); emitErr != nil {
			t.Errorf("emit: %v", emitErr)
		}
	})
	return decodeEnvelope(t, bytes.NewBufferString(out))
}

// errorCodes returns the code of each element of an envelope's errors[] array.
func errorCodes(v any) []string {
	list, _ := v.([]any)
	codes := make([]string, 0, len(list))
	for _, e := range list {
		m, _ := e.(map[string]any)
		code, _ := m["code"].(string)
		codes = append(codes, code)
	}
	return codes
}

// TestNamedApplyInputErrorIsSubjectBorne: an input error of apply <name> is found after the subject
// is fixed, so it goes to results[0].errors[] and the top-level errors[] stays empty.
func TestNamedApplyInputErrorIsSubjectBorne(t *testing.T) {
	t.Run("--jobs without --all", func(t *testing.T) {
		assertSubjectBorneInputError(t, emitApplyEnvelope(t, "foo", "--jobs", "2"))
	})
	t.Run("root filter without --all", func(t *testing.T) {
		withRootFilters(t, true, false)
		assertSubjectBorneInputError(t, emitApplyEnvelope(t, "foo"))
	})
}

// assertSubjectBorneInputError checks that doc carries one E_INPUT on subject foo and none at the top.
func assertSubjectBorneInputError(t *testing.T, doc map[string]any) {
	t.Helper()
	if top := errorCodes(doc["errors"]); len(top) != 0 {
		t.Errorf("top-level errors = %v, want none", top)
	}
	results, _ := doc["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results = %v, want exactly one subject", doc["results"])
	}
	sr, _ := results[0].(map[string]any)
	if name, _ := sr["subject"].(map[string]any)["name"].(string); name != "foo" {
		t.Errorf("results[0].subject = %v, want foo", sr["subject"])
	}
	if codes := errorCodes(sr["errors"]); len(codes) != 1 || codes[0] != "E_INPUT" {
		t.Errorf("results[0].errors codes = %v, want [E_INPUT]", codes)
	}
}

// TestApplyAllWithNameInputErrorIsTopLevel: apply --all <name> is rejected before any subject
// exists, so its error goes to the top-level errors[].
func TestApplyAllWithNameInputErrorIsTopLevel(t *testing.T) {
	doc := emitApplyEnvelope(t, "--all", "foo")
	if results, _ := doc["results"].([]any); len(results) != 0 {
		t.Errorf("results = %v, want none", results)
	}
	if codes := errorCodes(doc["errors"]); len(codes) != 1 || codes[0] != "E_INPUT" {
		t.Errorf("top-level errors codes = %v, want [E_INPUT]", codes)
	}
}
