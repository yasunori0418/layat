package main

import (
	"fmt"
	"io"

	"github.com/yasunori0418/layat/internal/generator"
)

// printGeneratorError writes a generator failure for humans to w (main passes stderr). The CLI
// decides how a generator.Error is shown (→ ADR-0055 §6, §8): the summary line (its Message) and
// the guidance, without the raw diagnostics, which already went to the writer the generator was
// given. Under --debug the summary line ends with the generator's name; the failed internal
// command is already in Message (→ ADR-0055 §8).
//
// It and generatorErrorMessage render the generator.Error itself, not the chain around it: every
// command returns a generator's error as-is, so nothing wraps it with context today. A caller that
// starts wrapping one with %w must render the outer error instead, or that context is dropped.
func printGeneratorError(w io.Writer, e *generator.Error) {
	summary := e.Message
	if flagDebug {
		summary += " (generator: " + e.Generator + ")"
	}
	_, _ = fmt.Fprintln(w, summary)
	if e.Guidance != "" {
		_, _ = fmt.Fprintln(w, e.Guidance)
	}
}

// generatorErrorMessage renders a generator failure as the --json errors[].message: the summary
// followed by the captured raw diagnostics, since stdout carries only the envelope (→ ADR-0055 §6,
// §7: no generator name is put in front of it, --debug or not).
func generatorErrorMessage(e *generator.Error) string {
	if e.Stderr == "" {
		return e.Message
	}
	return e.Message + "\n" + e.Stderr
}
