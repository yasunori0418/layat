// outturn.go is the --json output foundation: layat's outturn envelope instantiation, item-id
// derivation, and the emit helper writing one envelope to stdout. Engine results reach the
// envelope only through the DTO types, never marshaled directly.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/yasunori0418/outturn/go"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/lock"
)

// outturnSpecVersion is the outturn output-spec version layat produces, independent of the
// manifest.json schemaVersion and tool.version.
const outturnSpecVersion = 1

// layat's instantiation of the outturn envelope: item / change info are the shared mutation DTOs,
// while result.info (TInfo) and envelope info (TEnvInfo) are chosen per command. Aliases keep
// outturn's own types and their MarshalJSON.
type (
	layatEnvelope[TInfo, TEnvInfo any] = outturn.Envelope[*outturnEntryInfo, *outturnChangeInfo, TInfo, TEnvInfo]
	layatSubjectResult[TInfo any]      = outturn.SubjectResult[*outturnEntryInfo, *outturnChangeInfo, TInfo]
	layatResult[TInfo any]             = outturn.Result[*outturnEntryInfo, *outturnChangeInfo, TInfo]
	layatItem                          = outturn.Item[*outturnEntryInfo]
	layatChange                        = outturn.Change[*outturnChangeInfo]
)

// entryItemID derives the outturn item id for a placement entry: kind="entry", key={target}
// (root-relative target only; the config name stays out of the key).
func entryItemID(target string) (string, error) {
	return outturn.DeriveID(outturn.Identity{Kind: "entry", Key: map[string]any{"target": target}})
}

// outturnTimestamp renders t as RFC 3339 with an explicit offset for startedAt / finishedAt.
func outturnTimestamp(t time.Time) string { return t.Format(time.RFC3339) }

// outturnRun accumulates the --json envelope for one command invocation. RunE begins it, commands
// register their subjects, and main emits it once after Execute returns. TInfo / TEnvInfo are the
// command's result.info / envelope-info types.
type outturnRun[TInfo, TEnvInfo any] struct {
	now     func() time.Time // injectable clock; tests pin it to a fixed time
	out     io.Writer        // envelope sink (os.Stdout; tests substitute a buffer)
	command string           // executed subcommand name ("" until begin · → began)
	dryRun  bool             // envelope dryRun field, captured from flagDryrun at begin (tests set it directly)
	started time.Time
	// subjects are the registered subjects in registration order, one SubjectResult each.
	subjects []*outturnSubject[TInfo]
	info     TEnvInfo // envelope-wide tool info
}

// outturnSubject accumulates one subject's (config's) SubjectResult. Commands attach through the
// handle beginSubject returns, so results bind to the handle rather than registration order.
type outturnSubject[TInfo any] struct {
	name    string
	started time.Time
	// payload is the command-built items / changes / generation / warnings; nil leaves items empty.
	payload *outturnPayload[TInfo]
	// finished marks that err is final. --all finishes each subject as its config completes; a
	// single-config command leaves it to emit, which settles it with the command error.
	finished bool
	err      error
}

// setPayload attaches this subject's result payload.
func (s *outturnSubject[TInfo]) setPayload(p *outturnPayload[TInfo]) { s.payload = p }

// finish settles this subject's outcome (nil = success) independently of other subjects.
// A second call keeps the first outcome.
func (s *outturnSubject[TInfo]) finish(err error) {
	if s.finished {
		return
	}
	s.finished, s.err = true, err
}

// itemBorne reports whether the payload already represents the failure as a failed item.
// Such a failure makes the subject error and stays out of its errors[].
func (s *outturnSubject[TInfo]) itemBorne() bool {
	return s.payload != nil && s.payload.itemBorne
}

// failed reports whether this subject settled on an error, by its own error or by a failed item.
func (s *outturnSubject[TInfo]) failed() bool {
	return s.finished && (s.err != nil || s.itemBorne())
}

// setEnvelopeInfo registers the envelope-wide tool info (run-scoped facts tied to no subject).
func (r *outturnRun[TInfo, TEnvInfo]) setEnvelopeInfo(info TEnvInfo) { r.info = info }

// emitter is the type-erased face of a run that main needs after Execute returns. The info-typed
// operations stay on the concrete run each command holds.
type emitter interface {
	began() bool
	emit(error) error
}

// noopEmitter is outturnReport's initial value for paths that never reach a RunE (--help /
// --version / completion / flag and argument validation failures).
type noopEmitter struct{}

func (noopEmitter) began() bool      { return false }
func (noopEmitter) emit(error) error { return nil }

// outturnReport is the process-wide run that main emits after Execute returns.
var outturnReport emitter = noopEmitter{}

// beginOutturnRun builds a run on the real clock and stdout, begins it, and publishes it to
// outturnReport. Each command wraps it in a begin<Command>Run helper next to its run alias.
func beginOutturnRun[TInfo, TEnvInfo any](command string) *outturnRun[TInfo, TEnvInfo] {
	r := &outturnRun[TInfo, TEnvInfo]{now: time.Now, out: os.Stdout}
	r.begin(command)
	outturnReport = r
	return r
}

// begin records the subcommand, the start time and the parsed --dryrun value.
func (r *outturnRun[TInfo, TEnvInfo]) begin(command string) {
	r.command = command
	r.dryRun = flagDryrun
	r.started = r.now()
}

// began satisfies emitter; it reports whether begin has run.
func (r *outturnRun[TInfo, TEnvInfo]) began() bool { return r.command != "" }

// beginSubject registers a subject (config name) and returns its handle. Once a subject exists,
// a command failure goes to its errors[] rather than the top-level errors[].
func (r *outturnRun[TInfo, TEnvInfo]) beginSubject(name string) *outturnSubject[TInfo] {
	s := &outturnSubject[TInfo]{name: name, started: r.now()}
	r.subjects = append(r.subjects, s)
	return s
}

// subjectResult renders this subject's SubjectResult from its settled state. finishedAt is the
// run's single finish timestamp.
func (s *outturnSubject[TInfo]) subjectResult(finishedAt string) layatSubjectResult[TInfo] {
	status := outturn.StatusSuccess
	if s.failed() {
		status = outturn.StatusError
	}
	sr := layatSubjectResult[TInfo]{
		Subject:    outturn.Subject{Name: s.name},
		Status:     status,
		StartedAt:  outturnTimestamp(s.started),
		FinishedAt: finishedAt,
	}
	if p := s.payload; p != nil {
		sr.Generation = p.generation
		sr.Warnings = p.warnings
		sr.Result = layatResult[TInfo]{Items: p.items, Changes: p.changes, Info: p.info}
	}
	// A failure the items do not carry goes to the subject's errors[].
	if s.err != nil && !s.itemBorne() {
		sr.Errors = append(sr.Errors, classifyError(s.err))
	}
	return sr
}

// emit writes the outturn envelope (one JSON document and a newline) to r.out. Each subject decides
// its own status; one subject in error makes the envelope error. cmdErr nil ⇔ exit 0 ⇔ "success".
func (r *outturnRun[TInfo, TEnvInfo]) emit(cmdErr error) error {
	finished := outturnTimestamp(r.now())

	env := layatEnvelope[TInfo, TEnvInfo]{
		SpecVersion: outturnSpecVersion,
		Tool:        outturn.Tool{Name: "layat", Version: version},
		Command:     r.command,
		Status:      outturn.StatusSuccess,
		DryRun:      r.dryRun,
		StartedAt:   outturnTimestamp(r.started),
		FinishedAt:  finished,
		Info:        r.info,
	}
	if cmdErr != nil {
		env.Status = outturn.StatusError
	}
	// Settle unsettled subjects with cmdErr. Subjects --all already finished keep their own outcome.
	for _, s := range r.subjects {
		s.finish(cmdErr)
		if s.failed() {
			env.Status = outturn.StatusError
		}
		env.Results = append(env.Results, s.subjectResult(finished))
	}
	if cmdErr != nil && len(r.subjects) == 0 {
		// With no subject registered, the failure is from before subject enumeration and goes to the
		// top-level errors[].
		env.Errors = append(env.Errors, classifyError(cmdErr))
	}

	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.out, "%s\n", data)
	return err
}

// classifyError maps a command failure onto an outturn error code: E_INPUT, E_LAYAT_COLLISION,
// E_LOCK, E_LAYAT_BUILD, E_NOTFOUND, E_PERMISSION, E_IO, falling back to E_LAYAT_FAILED.
// Specific sentinels win over the generic E_IO shape check.
func classifyError(err error) outturn.Error {
	code := "E_LAYAT_FAILED"
	message := err.Error()
	var ee *exitError
	var ie *inputError
	bf := buildFailure(err)
	switch {
	case errors.As(err, &ie):
		code = "E_INPUT"
	case errors.As(err, &ee) && ee.code == 2:
		// exit 2 is apply --dryrun's conflict; its exitError carries no message, so supply one.
		code = "E_LAYAT_COLLISION"
		if message == "" {
			message = "conflict(s) detected in dryrun"
		}
	case errors.Is(err, lock.ErrLocked):
		code = "E_LOCK"
	case bf != nil:
		code = "E_LAYAT_BUILD"
		message = generatorErrorMessage(bf)
	case errors.Is(err, fs.ErrNotExist):
		code = "E_NOTFOUND"
	case errors.Is(err, fs.ErrPermission):
		code = "E_PERMISSION"
	case isIOError(err):
		code = "E_IO"
	}
	return outturn.Error{Code: code, Message: message}
}

// inputError marks a rejected input so --json classifies it as E_INPUT. It wraps transparently.
type inputError struct{ err error }

func (e *inputError) Error() string { return e.err.Error() }
func (e *inputError) Unwrap() error { return e.err }

// buildFailure returns err's generator.Error when a non-prebuilt generator failed at the roots or
// build stage, and nil otherwise.
func buildFailure(err error) *generator.Error {
	var ge *generator.Error
	if errors.As(err, &ge) && ge.Generator != generator.NamePrebuilt &&
		(ge.Stage == generator.StageRoots || ge.Stage == generator.StageBuild) {
		return ge
	}
	return nil
}

// isIOError reports whether err has a filesystem / external-I/O failure shape (E_IO). It runs after
// the fs.ErrNotExist / fs.ErrPermission checks.
func isIOError(err error) bool {
	var pe *fs.PathError
	var le *os.LinkError
	var se *os.SyscallError
	return errors.As(err, &pe) || errors.As(err, &le) || errors.As(err, &se)
}
