package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/gitutil"
	"github.com/yasunori0418/layat/internal/manifest"
)

// msgPrefix heads every line layat itself writes to stderr. Error and warning strings do not
// carry it; the functions in this file add it when they write.
const msgPrefix = "layat: "

// printError writes the command's failure to w, one "layat: <message>" per message of err. A
// generator failure is written by printGeneratorError.
func printError(w io.Writer, err error) {
	var ge *generator.Error
	if errors.As(err, &ge) {
		printGeneratorError(w, "", ge)
		return
	}
	for _, m := range messages(err) {
		_, _ = fmt.Fprintln(w, msgPrefix+m)
	}
}

// printFailure writes a failure the command reports and goes on from: "layat: <head>: <message>"
// for the first message of err, then "layat: <message>" for each other one. A generator failure is
// written by printGeneratorError under the head.
func printFailure(w io.Writer, head string, err error) {
	var ge *generator.Error
	if errors.As(err, &ge) {
		printGeneratorError(w, head, ge)
		return
	}
	msgs := messages(err)
	_, _ = fmt.Fprintf(w, msgPrefix+"%s: %s\n", head, msgs[0])
	for _, m := range msgs[1:] {
		_, _ = fmt.Fprintln(w, msgPrefix+m)
	}
}

// messages returns the messages of err that each start a prefixed line: the elements of an
// errors.Join, otherwise err's whole text. Newlines inside a message get no prefix.
func messages(err error) []string {
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return []string{err.Error()}
	}
	var msgs []string
	for _, e := range joined.Unwrap() {
		msgs = append(msgs, e.Error())
	}
	// fmt.Errorf with several %w also unwraps to a list but formats its own text.
	if len(msgs) == 0 || strings.Join(msgs, "\n") != err.Error() {
		return []string{err.Error()}
	}
	return msgs
}

// warnf writes one warning line, "layat: warning: <text>".
func warnf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, msgPrefix+"warning: "+format+"\n", args...)
}

// noticef writes one notice line, "layat: <text>".
func noticef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, msgPrefix+format+"\n", args...)
}

// engineWarnf is the engine's Warnf: warnf writing to os.Stderr, looked up at call time.
func engineWarnf(format string, args ...any) {
	warnf(os.Stderr, format, args...)
}

// engineReportf is the engine's Reportf for the details of a failure: noticef writing to
// os.Stderr, looked up at call time.
func engineReportf(format string, args ...any) {
	noticef(os.Stderr, format, args...)
}

// skipNotice writes the notice that apply name was skipped because another apply holds the lock.
func skipNotice(w io.Writer, name string) {
	noticef(w, "skipped apply %s because another apply is in progress (run layat apply %s after it ends)", name, name)
}

// configWarnf returns apply --all's warning function for config name: warnf writing to w with the
// line headed by "[<name>] ".
func configWarnf(w io.Writer, name string) func(format string, args ...any) {
	return func(format string, args ...any) {
		_, _ = fmt.Fprintf(w, "[%s] "+msgPrefix+"warning: %s\n", name, fmt.Sprintf(format, args...))
	}
}

// configReportf returns apply --all's failure-detail function for config name: noticef writing to
// w with the line headed by "[<name>] ".
func configReportf(w io.Writer, name string) func(format string, args ...any) {
	return func(format string, args ...any) {
		_, _ = fmt.Fprintf(w, "[%s] "+msgPrefix+"%s\n", name, fmt.Sprintf(format, args...))
	}
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

// printNotInRepoHint writes the fix for a project root resolved outside a git repository.
func printNotInRepoHint(w io.Writer) {
	_, _ = fmt.Fprintln(w, msgPrefix+"the project root is the git toplevel of the CWD; "+
		"run inside a git repository, or pass --root <dir> to set the root explicitly")
}

// printFailureHints writes, after the failure lines, each hint that one of errs calls for, once:
// the schemaVersion skew hint and the hint for running outside a git repository.
func printFailureHints(w io.Writer, errs ...error) {
	var skew, notInRepo bool
	for _, err := range errs {
		skew = skew || errors.Is(err, manifest.ErrSchemaVersionUnsupported)
		notInRepo = notInRepo || errors.Is(err, gitutil.ErrNotInRepo)
	}
	if skew {
		printSchemaSkewHint(w)
	}
	if notInRepo {
		printNotInRepoHint(w)
	}
}
