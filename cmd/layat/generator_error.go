package main

import (
	"fmt"
	"io"

	"github.com/yasunori0418/layat/internal/generator"
)

// printGeneratorError writes a generator failure for humans to w: the summary line and the
// guidance, without the raw diagnostics already written. --debug appends the generator's name.
// It renders the generator.Error itself, not any wrapping context.
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
// followed by the captured raw diagnostics.
func generatorErrorMessage(e *generator.Error) string {
	if e.Stderr == "" {
		return e.Message
	}
	return e.Message + "\n" + e.Stderr
}
