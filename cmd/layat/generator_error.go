package main

import (
	"fmt"
	"io"

	"github.com/yasunori0418/layat/internal/generator"
)

// printGeneratorError writes a generator failure for humans to w (main passes stderr). The CLI
// decides how a generator.Error is shown (→ ADR-0055 §6, §8); the summary is its Message, the
// generator's raw diagnostics having already gone to the writer it was given.
//
// It and generatorErrorMessage render the generator.Error itself, not the chain around it: every
// command returns a generator's error as-is, so nothing wraps it with context today. A caller that
// starts wrapping one with %w must render the outer error instead, or that context is dropped.
func printGeneratorError(w io.Writer, e *generator.Error) {
	_, _ = fmt.Fprintln(w, e.Message)
}

// generatorErrorMessage renders a generator failure as the --json errors[].message (→ ADR-0055 §6,
// §7: no generator name is put in front of it).
func generatorErrorMessage(e *generator.Error) string {
	return e.Message
}
