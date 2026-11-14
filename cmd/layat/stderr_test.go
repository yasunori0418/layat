package main

import (
	"bytes"
	"errors"
	"testing"
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
