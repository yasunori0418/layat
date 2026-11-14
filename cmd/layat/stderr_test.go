package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/yasunori0418/layat/internal/generator"
)

// TestStderrLineForms pins the line form of each kind of stderr output.
func TestStderrLineForms(t *testing.T) {
	cases := []struct {
		name  string
		write func(*bytes.Buffer)
		want  string
	}{
		{"notice", func(b *bytes.Buffer) { noticef(b, "reset %s done (root=%s)", "a", "/r") },
			"layat: reset a done (root=/r)\n"},
		{"warning", func(b *bytes.Buffer) { warnf(b, "skipping the system profile base (%s): %v", "/s", errors.New("boom")) },
			"layat: warning: skipping the system profile base (/s): boom\n"},
		{"failure", func(b *bytes.Buffer) { printFailure(b, "apply a failed", errors.New("boom")) },
			"layat: apply a failed: boom\n"},
		{"nix disclosure", func(b *bytes.Buffer) { discloseNix(b, []string{"build", "--no-link", ".#layat"}) },
			"layat: + nix build --no-link .#layat\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			c.write(&buf)
			if buf.String() != c.want {
				t.Errorf("output = %q, want %q", buf.String(), c.want)
			}
		})
	}
}

// TestPrintSchemaSkewHint pins the schemaVersion skew hint: a blank line, then the hint headed by
// the prefix.
func TestPrintSchemaSkewHint(t *testing.T) {
	var buf bytes.Buffer
	printSchemaSkewHint(&buf)
	want := "\nlayat: the layat version pinned by the CLI (engine) and by the flake may be out of sync.\n" +
		"  The flake's layat input is generating a manifest newer than the CLI.\n" +
		"  Update the CLI, or lower the flake's layat input to match the CLI so both versions align.\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

// TestPrintErrorPrefix pins the command failure's form: "layat: " heads each element of an
// errors.Join, a newline inside a message gets none, and a generator failure is its summary and
// guidance.
func TestPrintErrorPrefix(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"single", errors.New("unknown flag: --foo"), "layat: unknown flag: --foo\n"},
		{"wrapped", fmt.Errorf("cannot read the manifest: %w", errors.New("boom")), "layat: cannot read the manifest: boom\n"},
		{"multi-line message", errors.New("apply --all: same target:\n  t: claimed by both a and b"),
			"layat: apply --all: same target:\n  t: claimed by both a and b\n"},
		{"joined", errors.Join(errors.New("cannot remove a"), errors.New("cannot write the record")),
			"layat: cannot remove a\nlayat: cannot write the record\n"},
		{"several %w", fmt.Errorf("x: %w, %w", errors.New("a"), errors.New("b")), "layat: x: a, b\n"},
		{"exit code", &exitCodeError{code: 1, msg: "apply --all: 1 config(s) failed"}, "layat: apply --all: 1 config(s) failed\n"},
		{"generator", fmt.Errorf("ctx: %w", generator.NewError(generator.NameNix, generator.StageBuild, generator.KindFailed,
			"nix build failed: exit status 1", "see above", "error: boom", nil)), "layat: nix build failed: exit status 1\nsee above\n"},
	}
	withDebug(t, false)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			printError(&buf, c.err)
			if buf.String() != c.want {
				t.Errorf("output = %q, want %q", buf.String(), c.want)
			}
		})
	}
}

// TestPrintFailureJoined pins a failure line over an errors.Join: the head carries the first
// element, and each other element starts its own "layat: " line.
func TestPrintFailureJoined(t *testing.T) {
	var buf bytes.Buffer
	printFailure(&buf, "status home:a", errors.Join(errors.New("cannot read the backref"), errors.New("cannot read the record")))
	want := "layat: status home:a: cannot read the backref\nlayat: cannot read the record\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}
