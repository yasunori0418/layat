package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/generator"
)

// testGeneratorError is a nix generator failure carrying a guidance and captured raw diagnostics.
func testGeneratorError() *generator.Error {
	return generator.NewError(generator.NameNix, generator.StageRoots, generator.KindPrerequisiteMissing,
		"nix eval failed: exit status 1", "enable experimental-features",
		"error: experimental Nix feature nix-command is disabled", nil)
}

// withDebug sets --debug for the test and restores it afterwards.
func withDebug(t *testing.T, debug bool) {
	t.Helper()
	orig := flagDebug
	t.Cleanup(func() { flagDebug = orig })
	flagDebug = debug
}

// TestPrintGeneratorErrorHuman pins the human form: the summary line headed by "layat: " and the
// guidance, without the raw diagnostics.
func TestPrintGeneratorErrorHuman(t *testing.T) {
	withDebug(t, false)
	e := testGeneratorError()
	var buf bytes.Buffer
	printGeneratorError(&buf, "", e)

	if want := "layat: " + e.Message + "\n" + e.Guidance + "\n"; buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
	if strings.Contains(buf.String(), e.Stderr) {
		t.Errorf("output re-shows the raw diagnostics: %q", buf.String())
	}
}

// TestPrintGeneratorErrorHumanNoGuidance pins the human form without a guidance: the summary only.
func TestPrintGeneratorErrorHumanNoGuidance(t *testing.T) {
	withDebug(t, false)
	e := generator.NewError(generator.NameNix, generator.StageBuild, generator.KindFailed,
		"nix build failed: exit status 1", "", "error: boom", nil)
	var buf bytes.Buffer
	printGeneratorError(&buf, "", e)

	if want := "layat: " + e.Message + "\n"; buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// TestPrintGeneratorErrorDebug pins the --debug human form: the generator's name after the summary
// line, then the guidance.
func TestPrintGeneratorErrorDebug(t *testing.T) {
	withDebug(t, true)
	e := testGeneratorError()
	var buf bytes.Buffer
	printGeneratorError(&buf, "", e)

	if want := "layat: " + e.Message + " (generator: nix)\n" + e.Guidance + "\n"; buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// TestPrintGeneratorErrorHead pins a failure line with a head, as apply --all writes it: the head
// before the summary on the first line (with the generator's name under --debug), then the guidance.
func TestPrintGeneratorErrorHead(t *testing.T) {
	e := testGeneratorError()
	for _, debug := range []bool{false, true} {
		withDebug(t, debug)
		var buf bytes.Buffer
		printFailure(&buf, "apply a failed", fmt.Errorf("ctx: %w", e))
		summary := "layat: apply a failed: " + e.Message
		if debug {
			summary += " (generator: nix)"
		}
		if want := summary + "\n" + e.Guidance + "\n"; buf.String() != want {
			t.Errorf("debug=%v: output = %q, want %q", debug, buf.String(), want)
		}
	}
}

// TestGeneratorErrorMessageJSON pins the --json errors[].message: the summary followed by the raw
// diagnostics, with no generator name, with or without --debug.
func TestGeneratorErrorMessageJSON(t *testing.T) {
	e := testGeneratorError()
	want := e.Message + "\n" + e.Stderr
	for _, debug := range []bool{false, true} {
		withDebug(t, debug)
		if got := generatorErrorMessage(e); got != want {
			t.Errorf("debug=%v: message = %q, want %q", debug, got, want)
		}
	}

	noStderr := generator.NewError(generator.NameNix, generator.StageBuild, generator.KindFailed,
		"nix build failed: exit status 1", "", "", nil)
	if got := generatorErrorMessage(noStderr); got != noStderr.Message {
		t.Errorf("message without stderr = %q, want %q", got, noStderr.Message)
	}
}
