package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/yasunori0418/layat/internal/generator"
)

// msgPrefix heads every line layat itself writes to stderr.
const msgPrefix = "layat: "

// printError writes the command's failure to w. A generator failure is written by
// printGeneratorError; anything else is written as is.
func printError(w io.Writer, err error) {
	var ge *generator.Error
	if errors.As(err, &ge) {
		printGeneratorError(w, ge)
		return
	}
	_, _ = fmt.Fprintln(w, err)
}

// printFailure writes one failure line, "layat: <head>: <err>", for a failure the command reports
// and goes on from.
func printFailure(w io.Writer, head string, err error) {
	_, _ = fmt.Fprintf(w, msgPrefix+"%s: %v\n", head, err)
}

// warnf writes one warning line, "layat: warning: <text>".
func warnf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, msgPrefix+"warning: "+format+"\n", args...)
}

// noticef writes one notice line, "layat: <text>".
func noticef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, msgPrefix+format+"\n", args...)
}

// discloseNix writes the --debug disclosure of a nix command, "layat: + nix <args>".
func discloseNix(w io.Writer, args []string) {
	_, _ = fmt.Fprintf(w, msgPrefix+"+ nix %s\n", strings.Join(args, " "))
}

// printSchemaSkewHint writes the cause and the fix of a manifest schemaVersion skew between the CLI
// and the flake pin.
func printSchemaSkewHint(w io.Writer) {
	_, _ = fmt.Fprintln(w, "\n"+msgPrefix+"the layat version pinned by the CLI (engine) and by the flake may be out of sync.\n"+
		"  The flake's layat input is generating a manifest newer than the CLI.\n"+
		"  Update the CLI, or lower the flake's layat input to match the CLI so both versions align.")
}
