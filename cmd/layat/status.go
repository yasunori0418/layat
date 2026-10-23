package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/planner"
)

// statusInfo is status' result.info: one config's last-applied record. It is a pointer so a
// subject whose record cannot be read omits info.
type statusInfo struct {
	Mode      string        `json:"mode"`
	Root      string        `json:"root"`
	AppliedAt string        `json:"appliedAt"`
	LinkFarm  string        `json:"linkFarm"`
	Entries   []statusEntry `json:"entries"`
}

// statusEntry is one entry of the recorded manifest.
type statusEntry struct {
	Target  string `json:"target"`
	Method  string `json:"method"`
	SrcKind string `json:"srcKind"`
	Src     string `json:"src"`
	Subpath string `json:"subpath"`
}

// statusRun is status' concrete run instantiation, threaded from RunE.
type statusRun = outturnRun[*statusInfo, *struct{}]

// beginStatusRun starts status' run.
func beginStatusRun(command string) *statusRun {
	return beginOutturnRun[*statusInfo, *struct{}](command)
}

// statusFn lists the last-applied records; tests replace it.
var statusFn = engine.Status

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [name]",
		Short: "Show what is currently applied, from the last-applied records",
		Long: "A read-only command that lists the last-applied records of every config, user base first, " +
			"then the system base. Pass <name> to show only the configs of that name. " +
			"It runs no nix and ignores the entrypoint and the generator.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			run := beginStatusRun(cmd.Name())
			return runStatus(run, args)
		},
	}
}

// runStatus lists the records engine.Status finds, filtered by args[0] when given. A record that
// cannot be read becomes its subject's error and the listing goes on; the command then exits 1.
func runStatus(run *statusRun, args []string) error {
	series, err := statusFn(engine.StatusOptions{SystemStateDir: systemStateDirOverride()})
	if err != nil {
		return err
	}

	listed, failures, printed := 0, 0, false
	for _, s := range series {
		// A series reported as a whole has no <name> and may hold the one asked for, so it stays.
		if len(args) == 1 && s.Name != "" && s.Name != args[0] {
			continue
		}
		listed++
		name := statusSubjectName(s)
		subject := run.beginSubject(name)
		if s.Err != nil {
			subject.finish(s.Err)
			failures++
			fmt.Fprintf(os.Stderr, "layat: status %s: %v\n", name, s.Err)
			continue
		}
		subject.setPayload(&outturnPayload[*statusInfo]{info: statusPayloadInfo(s)})
		subject.finish(nil)
		// Under --json stdout belongs to the envelope.
		if flagJSON {
			continue
		}
		if printed {
			fmt.Println()
		}
		printed = true
		printStatus(name, s)
	}

	if listed == 0 {
		if len(args) == 1 {
			fmt.Fprintf(os.Stderr, "layat: status: no applied config named %s\n", args[0])
		} else {
			fmt.Fprintln(os.Stderr, "layat: status: no applied config")
		}
		return nil
	}
	if failures > 0 {
		return &exitCodeError{code: 1, msg: fmt.Sprintf("layat: status: %d config(s) could not be read", failures)}
	}
	return nil
}

// statusSubjectName names a series: <mode>:<name> when <name>-keyed, <mode>:@<roothash> for a
// series reported as a whole, <mode>:<name>@<roothash> when the backref cannot be read, and
// <mode>:<name>@<root> otherwise.
func statusSubjectName(s engine.StatusSeries) string {
	switch {
	case s.RootHash == "":
		return s.Mode + ":" + s.Name
	case s.Name == "":
		return s.Mode + ":@" + s.RootHash
	case s.Root == "":
		return s.Mode + ":" + s.Name + "@" + s.RootHash
	default:
		return s.Mode + ":" + s.Name + "@" + s.Root
	}
}

// statusPayloadInfo converts a readable record into result.info (entries non-nil even when empty).
func statusPayloadInfo(s engine.StatusSeries) *statusInfo {
	entries := make([]statusEntry, 0, len(s.Manifest.Entries))
	for _, e := range s.Manifest.Entries {
		entries = append(entries, statusEntry{Target: e.Target, Method: e.Method, SrcKind: e.SrcKind, Src: e.Src, Subpath: e.Subpath})
	}
	return &statusInfo{Mode: s.Mode, Root: s.Root, AppliedAt: s.AppliedAt, LinkFarm: s.LinkFarm, Entries: entries}
}

// printStatus prints one config's heading and its <target>\t<method>\t<dest> lines to stdout.
func printStatus(name string, s engine.StatusSeries) {
	fmt.Printf("# %s (applied %s)\n", name, s.AppliedAt)
	for _, e := range s.Manifest.Entries {
		fmt.Printf("%s\t%s\t%s\n", e.Target, e.Method, planner.LinkDest(e))
	}
}
