// outturn.go is the --json machine-readable output foundation (→ ADR-0043, issue #130): the
// layat instantiation of the outturn envelope (specVersion 1), the item-id derivation seam, and
// the emit helper that writes exactly one envelope document to stdout at command completion.
//
// The CLI output contract is closed inside the cmd layer: engine results are never marshaled
// directly — #131 / #132 convert them through the DTO types below, so engine-internal structure
// changes cannot break the output contract (→ ADR-0043 §8).
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

// outturnSpecVersion is the outturn output-spec version layat produces. Independent of both the
// manifest.json schemaVersion (engine input contract) and tool.version (layat release)
// (→ ADR-0043 §2).
const outturnSpecVersion = 1

// The layat instantiation of the outturn generic envelope. The four type parameters are the
// tool-specific info slots (item / change / per-subject result / envelope-wide). The item /
// change slots are pinned to the concrete mutation DTOs shared by every command (pointers, so
// an absent info is omitted · #131); the per-subject result (TInfo) and envelope-wide
// (TEnvInfo) slots vary per command, so they stay type parameters here and each command
// instantiates them with its own info type (→ issue #196, outturn ADR-0018).
//
// These stay aliases (parameterized aliases, Go 1.24+), so the emitted values remain outturn's
// own types and keep outturn's MarshalJSON — the required-array normalization must not be lost.
type (
	layatEnvelope[TInfo, TEnvInfo any] = outturn.Envelope[*outturnEntryInfo, *outturnChangeInfo, TInfo, TEnvInfo]
	layatSubjectResult[TInfo any]      = outturn.SubjectResult[*outturnEntryInfo, *outturnChangeInfo, TInfo]
	layatResult[TInfo any]             = outturn.Result[*outturnEntryInfo, *outturnChangeInfo, TInfo]
	layatItem                          = outturn.Item[*outturnEntryInfo]
	layatChange                        = outturn.Change[*outturnChangeInfo]
)

// entryItemID derives the outturn item id for a placement entry: identity kind="entry",
// key={target} (the root-relative target only — the config name stays out of the key; consumers
// resolve references by the (tool.name, subject, id) triple · → ADR-0043 §3, outturn ADR-0024).
// Shared by the #131 / #132 payload builders; the identity shape is pinned against outturn's
// id-vectors in TestEntryItemIDMatchesVectors.
func entryItemID(target string) (string, error) {
	return outturn.DeriveID(outturn.Identity{Kind: "entry", Key: map[string]any{"target": target}})
}

// outturnTimestamp renders t for the envelope's startedAt/finishedAt: RFC 3339, "T" separator,
// explicit UTC offset (the local offset; UTC itself renders as "Z" — both satisfy outturn
// ADR-0025's format assertion).
func outturnTimestamp(t time.Time) string { return t.Format(time.RFC3339) }

// outturnRun accumulates what the --json envelope needs across one command invocation. Each layat
// subcommand's RunE builds it first thing and begins it (command name + start time — flag parsing
// and cobra's argument validation have succeeded by then, and cobra's utility commands help /
// completion / __complete never reach a RunE of ours, so they keep stdout for their own text),
// commands register their subject as it becomes known, and main emits once after Execute
// returns: the single stdout write point (→ issue #130, ADR-0043 §2).
//
// TInfo / TEnvInfo are the command's own result.info / envelope-info types (→ issue #196): each
// command instantiates the run with its concrete pair in RunE and threads that concrete value
// through its run functions, because the info-typed setters cannot cross the emitter interface
// (Go forbids generic methods).
type outturnRun[TInfo, TEnvInfo any] struct {
	now     func() time.Time // injectable clock; tests pin it to a fixed time
	out     io.Writer        // envelope sink (os.Stdout; tests substitute a buffer)
	command string           // executed subcommand name ("" until begin · → began)
	dryRun  bool             // envelope dryRun field, captured from flagDryrun at begin (tests set it directly)
	started time.Time
	// subjects are the run's registered subjects in registration order, one SubjectResult each
	// (→ issue #164). A single-config command registers exactly one (N=1, the shape #130 wired);
	// --all registers one per selected config; init and a pre-enumeration failure register none.
	subjects []*outturnSubject[TInfo]
	info     TEnvInfo // envelope-wide tool info (init's run info · outturn ADR-0018, issue #132)
}

// outturnSubject is one subject's (config's) accumulation for its SubjectResult. Commands hold
// it as the handle beginSubject returns and attach everything through it, so a multi-subject run
// cannot mis-attribute one config's result to another — the binding is the handle, not the
// registration order (the precondition for #149's parallel apply --all · → issue #164).
type outturnSubject[TInfo any] struct {
	name    string
	started time.Time
	// payload is the command-built items/changes/generation/warnings mapping (→ #131's
	// outturn_payload.go builders). nil keeps the minimal #130 shape (empty items) — the
	// read-only commands until #132, and failures before any engine result exists.
	payload *outturnPayload[TInfo]
	// finished marks that this subject's outcome is settled: err below is final and nothing may
	// re-attribute the command-level error to it. A multi-subject run (--all) finishes every
	// subject as its config completes — one config's failure must not colour the ones that
	// succeeded — while the single-config commands leave it unset so emit can settle their one
	// subject with the command error, which for them IS that subject's (→ issue #164).
	finished bool
	err      error
}

// setPayload attaches this subject's result payload (→ #131 / #132 payload builders).
func (s *outturnSubject[TInfo]) setPayload(p *outturnPayload[TInfo]) { s.payload = p }

// finish settles this subject's own outcome: err is that config's error (nil = success), which
// decides its SubjectResult status and errors[] independently of the other subjects and of the
// command's aggregate error (→ issue #164). Calling it twice keeps the first outcome: the config's
// own result is the truth, and emit's later sweep must not overwrite it with the aggregate error.
func (s *outturnSubject[TInfo]) finish(err error) {
	if s.finished {
		return
	}
	s.finished, s.err = true, err
}

// itemBorne reports whether this subject's payload already represents the failure as a failed item
// (an entry failure / conflict). It is the single source for that question: the failed item both
// makes this subject error and keeps the error out of its errors[] (outturn §2: item 起因のエラーを
// errors[] に置いてはならない), so nothing needs to say "this failed" a second time.
func (s *outturnSubject[TInfo]) itemBorne() bool {
	return s.payload != nil && s.payload.itemBorne
}

// failed reports whether this subject settled on an error — by its own error or by a failed item in
// its payload. It is the aggregate's error source (any one subject in error makes the envelope
// error · → outturn §2, ADR-0043 §6, outturn ADR-0002 "a failed item makes the result error").
func (s *outturnSubject[TInfo]) failed() bool {
	return s.finished && (s.err != nil || s.itemBorne())
}

// setEnvelopeInfo registers the envelope-wide tool info (top-level info — run-scoped facts not
// tied to any subject; init's template expansion · outturn ADR-0018, issue #132).
func (r *outturnRun[TInfo, TEnvInfo]) setEnvelopeInfo(info TEnvInfo) { r.info = info }

// emitter is the type-erased face of a run: the two methods main needs after Execute returns,
// neither of which mentions the info types. Which subcommand runs is unknown before cobra
// dispatches, so the process-wide report cannot be a single static instantiation; the
// info-typed operations (begin / beginSubject / setPayload / setEnvelopeInfo) stay on the
// concrete run each command holds (→ issue #196).
type emitter interface {
	began() bool
	emit(error) error
}

// noopEmitter is outturnReport's initial value: no RunE has run, so no envelope exists. It keeps
// main's emit gate total without a nil check on the paths that never reach a RunE (--help /
// --version / help / completion / flag- and argument-validation failures).
type noopEmitter struct{}

func (noopEmitter) began() bool      { return false }
func (noopEmitter) emit(error) error { return nil }

// outturnReport is the process-wide run the CLI wires: each subcommand's RunE assigns its own
// concrete run here so main can emit it after Execute returns (tests build their own runs).
var outturnReport emitter = noopEmitter{}

// beginOutturnRun starts a command's run: it builds the concrete instantiation against the real
// clock and stdout, begins it for command, and publishes it to outturnReport so main can emit it
// after Execute returns — keeping build / begin / publish in one place, because a run that is
// begun but never published emits no envelope at all (main's gate reads outturnReport, not the
// command's local variable · → issue #196).
//
// Commands do not call this directly with their type arguments; each one wraps it in a
// begin<Command>Run helper declared next to its run alias, so the command's (TInfo, TEnvInfo)
// pair is spelled exactly once per command and the alias stays the single source of truth.
func beginOutturnRun[TInfo, TEnvInfo any](command string) *outturnRun[TInfo, TEnvInfo] {
	r := &outturnRun[TInfo, TEnvInfo]{now: time.Now, out: os.Stdout}
	r.begin(command)
	outturnReport = r
	return r
}

// begin records the executed subcommand, the run start time, and the parsed --dryrun value
// (cobra has finished flag parsing by RunE, where begin is called).
func (r *outturnRun[TInfo, TEnvInfo]) begin(command string) {
	r.command = command
	r.dryRun = flagDryrun
	r.started = r.now()
}

// began satisfies emitter. A published run is always a begun one — beginOutturnRun is the only
// production constructor and it begins before publishing — so this is true for anything main
// finds in outturnReport; the not-started state is noopEmitter's, which owns the list of paths
// that never reach a RunE. It still reads the field rather than returning true, because tests
// build runs directly and main's gate must stay honest for them too.
func (r *outturnRun[TInfo, TEnvInfo]) began() bool { return r.command != "" }

// beginSubject registers a subject (config name) once it is known and returns its handle, which
// the command uses to attach that config's payload. From then on a command failure attaches to a
// subject's SubjectResult.errors[] rather than the top-level errors[] (top level = failures
// before/outside subject enumeration only · → ADR-0043 §6). A single-config command calls it once
// (N=1); --all calls it per selected config (→ issue #164); init registers none (results stays []).
func (r *outturnRun[TInfo, TEnvInfo]) beginSubject(name string) *outturnSubject[TInfo] {
	s := &outturnSubject[TInfo]{name: name, started: r.now()}
	r.subjects = append(r.subjects, s)
	return s
}

// subjectResult renders this subject's SubjectResult from its own settled state alone — finish is
// the single place its outcome is decided, so reading this result never requires knowing what the
// caller passed in (→ issue #164). finishedAt is the run's single finish timestamp, shared by every
// result. emit is the only caller and settles every subject immediately before rendering it.
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
	// A failure the items already carry stays out of errors[] (→ itemBorne). Everything else that
	// failed with the subject established is subject-borne (build / lock / commit ...) and lands
	// here rather than at the top level (→ ADR-0043 §6).
	if s.err != nil && !s.itemBorne() {
		sr.Errors = append(sr.Errors, classifyError(s.err))
	}
	return sr
}

// emit writes the outturn envelope — exactly one JSON document, trailing newline, nothing else —
// to r.out. cmdErr is the command's overall error: nil ⇔ exit 0 ⇔ status "success" (the only
// status contract consumers may rely on · → ADR-0043 §6).
//
// emit only aggregates (→ issue #164): every subject decides its own status and errors[] through
// finish, and emit folds them into the envelope's status — one subject in error makes the whole
// document error — while the top-level errors[] takes only a failure that belongs to no subject,
// i.e. one from before/outside subject enumeration. cmdErr reaches a subject through the same
// finish, so a single-config command (which settles nothing itself) keeps #130's attribution
// unchanged, and a --all run's already-settled configs are immune to the aggregate error.
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
	// Settle whatever the command left unsettled, so every subject's outcome is decided in exactly
	// one place (finish) before anything is rendered. A single-config command settles nothing —
	// it registers one subject and stops at its first failure, so cmdErr IS that subject's, and
	// this is where it attaches. A --all path settles each config as it completes, so every
	// subject here is already finished and finish's first-wins rule leaves them untouched: the
	// aggregate cmdErr ("N config(s) failed") never overwrites a config's own outcome.
	for _, s := range r.subjects {
		s.finish(cmdErr)
		if s.failed() {
			env.Status = outturn.StatusError
		}
		env.Results = append(env.Results, s.subjectResult(finished))
	}
	if cmdErr != nil && len(r.subjects) == 0 {
		// The failure belongs to no subject because none was ever registered: it happened before
		// or outside subject enumeration (entrypoint discovery, the batch eval, an argument
		// rejection), which is the only thing the top-level errors[] carries (→ ADR-0043 §6,
		// issue #164). Once subjects exist, the run's failure is theirs — a --all aggregate error
		// merely restates that some of them failed, and repeating it here would report the same
		// failure twice, at a layer that promises a different meaning.
		env.Errors = append(env.Errors, classifyError(cmdErr))
	}

	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.out, "%s\n", data)
	return err
}

// classifyError maps a command-level failure onto an outturn error object (two-layer code naming ·
// outturn §6, ADR-0043 §8): E_INPUT (an invalid input combination, via the inputError marker ·
// → ADR-0056), tool-specific E_LAYAT_COLLISION (dryrun conflict exit) / E_LAYAT_BUILD
// (a generator's failure to evaluate or build the manifest, via the generator.Error marker ·
// → buildFailure), and the common registry
// codes E_LOCK / E_NOTFOUND / E_PERMISSION / E_IO. Specific sentinels win over the generic
// E_IO shape check, so a not-found PathError stays E_NOTFOUND. E_LAYAT_FAILED is the
// tool-generic fallback for a command failure not otherwise classified.
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
		// exit 2 is apply --dryrun's conflict detection (→ docs/spec.md exit code table). Its
		// exitError deliberately carries no message (the plan went to stdout), so supply one.
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

// inputError marks a rejected input (an unknown generator, an invalid setting, a forbidden flag
// combination), so the --json classification maps it to E_INPUT (→ ADR-0056, ADR-0043). It wraps
// transparently: Error/Unwrap keep the message and chain untouched.
type inputError struct{ err error }

func (e *inputError) Error() string { return e.err.Error() }
func (e *inputError) Unwrap() error { return e.err }

// buildFailure returns err's generator.Error when it is a generator's failure to evaluate or build
// the manifest — Stage roots / build of a generator other than prebuilt — and nil otherwise. A Stage discover failure (no entrypoint,
// a missing -f path) and a prebuilt failure (an unreadable --manifest link-farm) are not: they fall
// through so the cause chain keeps the classification it had before generators (→ ADR-0055 §7).
func buildFailure(err error) *generator.Error {
	var ge *generator.Error
	if errors.As(err, &ge) && ge.Generator != generator.NamePrebuilt &&
		(ge.Stage == generator.StageRoots || ge.Stage == generator.StageBuild) {
		return ge
	}
	return nil
}

// isIOError reports whether err carries a filesystem / external-I/O failure shape (outturn §6
// common code E_IO). Checked after the more specific fs.ErrNotExist / fs.ErrPermission
// sentinels, so it only catches the remaining I/O failures (EEXIST, ENOTEMPTY, EIO, ...).
func isIOError(err error) bool {
	var pe *fs.PathError
	var le *os.LinkError
	var se *os.SyscallError
	return errors.As(err, &pe) || errors.As(err, &le) || errors.As(err, &se)
}
