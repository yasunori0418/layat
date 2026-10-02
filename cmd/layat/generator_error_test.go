package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yasunori0418/layat/internal/generator"
)

// testGeneratorError is a nix generator failure carrying a guidance and captured raw diagnostics.
func testGeneratorError() *generator.Error {
	return generator.NewError(generator.NameNix, generator.StageRoots, generator.KindPrerequisiteMissing,
		"layat: nix eval failed: exit status 1", "enable experimental-features",
		"error: experimental Nix feature nix-command is disabled", nil)
}

// withDebug sets --debug for the test and restores it afterwards.
func withDebug(t *testing.T, debug bool) {
	t.Helper()
	orig := flagDebug
	t.Cleanup(func() { flagDebug = orig })
	flagDebug = debug
}

// TestPrintGeneratorErrorHuman pins the human form: the summary line and the guidance, without
// re-showing the raw diagnostics that already went to the generator's writer (→ ADR-0055 §6).
func TestPrintGeneratorErrorHuman(t *testing.T) {
	withDebug(t, false)
	e := testGeneratorError()
	var buf bytes.Buffer
	printGeneratorError(&buf, e)

	if want := e.Message + "\n" + e.Guidance + "\n"; buf.String() != want {
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
		"layat: nix build failed: exit status 1", "", "error: boom", nil)
	var buf bytes.Buffer
	printGeneratorError(&buf, e)

	if want := e.Message + "\n"; buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// TestPrintGeneratorErrorDebug pins the --debug human form: the generator's name after the summary
// line (the failed internal command is already in Message), then the guidance (→ ADR-0055 §8).
func TestPrintGeneratorErrorDebug(t *testing.T) {
	withDebug(t, true)
	e := testGeneratorError()
	var buf bytes.Buffer
	printGeneratorError(&buf, e)

	if want := e.Message + " (generator: nix)\n" + e.Guidance + "\n"; buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// TestGeneratorErrorMessageJSON pins the --json errors[].message: the summary followed by the
// captured raw diagnostics, with no generator name put in front, whether or not --debug is set
// (→ ADR-0055 §6, §7, §8).
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
		"layat: nix build failed: exit status 1", "", "", nil)
	if got := generatorErrorMessage(noStderr); got != noStderr.Message {
		t.Errorf("message without stderr = %q, want %q", got, noStderr.Message)
	}
}
