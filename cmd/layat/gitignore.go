package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/manifest"
)

// gitignoreInfo is gitignore's result.info: the anchor-form target list. It is a pointer so
// failures before the listing omit info.
type gitignoreInfo struct {
	Paths []string `json:"paths"`
}

// gitignoreRun is gitignore's concrete run instantiation, threaded from RunE.
type gitignoreRun = outturnRun[*gitignoreInfo, *struct{}]

// beginGitignoreRun starts gitignore's run.
func beginGitignoreRun(command string) *gitignoreRun {
	return beginOutturnRun[*gitignoreInfo, *struct{}](command)
}

func newGitignoreCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "gitignore [name]",
		Short: "Print placement targets for .gitignore to stdout (project mode only; no writes)",
		Long: "List the placement targets of layat.<name> for .gitignore on stdout (writes no file). " +
			"Output is the root-relative target with a leading / in anchor form (e.g. /.claude/skills/nix), one per line, " +
			"covering every target regardless of method (symlink / copy). project mode only; " +
			"--all sorts and de-duplicates the targets of all projectRoot configs into this one listing " +
			"(under --json it keeps them per config instead; see the --all flag).",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// beginGitignoreRun publishes the run to outturnReport, so main emits the envelope after Execute.
			run := beginGitignoreRun(cmd.Name())
			if all {
				if len(args) > 0 {
					return fmt.Errorf("layat: gitignore cannot combine <name> with --all")
				}
				return runGitignoreAll(run)
			}
			if len(args) != 1 {
				return fmt.Errorf("layat: gitignore requires <name> or --all")
			}
			return runGitignore(run, args[0])
		},
	}
	cmd.Flags().BoolVar(&all, "all", false,
		"Sort and de-duplicate the targets of all projectRoot configs (under --json each config keeps its own targets instead, un-deduplicated)")
	return cmd
}

// runGitignore lists a single project-mode config's placement targets. Other modes are rejected,
// since the anchor form presupposes the git toplevel.
func runGitignore(run *gitignoreRun, name string) error {
	// The config name is the run's single outturn subject.
	subject := run.beginSubject(name)
	gen, err := newGenerator()
	if err != nil {
		return err
	}
	if err := gen.Discover(flagFile); err != nil {
		return err
	}

	// Confirm project mode via rootKind pre-resolution eval (rejecting cheaply before build).
	root, err := gen.Roots(name)
	if err != nil {
		return err
	}
	if rootKind := root.RootKind; rootKind != manifest.RootKindProject {
		return fmt.Errorf("layat: gitignore is project mode only (layat.%s has rootKind=%q; the .gitignore anchor is meaningless for home / fixed)", name, rootKind)
	}

	targets, err := configTargets(gen, name)
	if err != nil {
		return err
	}
	// The listing rides result.info as anchor-form paths; items stays [].
	subject.setPayload(gitignorePayload(targets))
	printGitignore(targets)
	return nil
}

// gitignorePayload wraps one config's targets as its SubjectResult payload in result.info.
func gitignorePayload(targets []string) *outturnPayload[*gitignoreInfo] {
	return &outturnPayload[*gitignoreInfo]{info: &gitignoreInfo{Paths: gitignoreAnchors(targets)}}
}

// gitignoreAnchors maps targets into their /-anchor form (non-nil even when empty, so a
// config without entries still lists "paths": []).
func gitignoreAnchors(targets []string) []string {
	anchors := make([]string, 0, len(targets))
	for _, t := range targets {
		anchors = append(anchors, gitignoreAnchor(t))
	}
	return anchors
}

// runGitignoreAll lists the targets of all projectRoot configs. stdout gets the sorted,
// de-duplicated union; --json gives each config its own undeduplicated paths.
func runGitignoreAll(run *gitignoreRun) error {
	gen, err := newGenerator()
	if err != nil {
		return err
	}
	if err := gen.Discover(flagFile); err != nil {
		return err
	}

	roots, err := gen.AllRoots()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(roots))
	for name := range roots {
		names = append(names, name)
	}
	sort.Strings(names)

	var selected []string
	for _, name := range names {
		if roots[name].RootKind == manifest.RootKindProject {
			selected = append(selected, name)
		}
	}
	return enumerateGitignoreAll(run, selected, func(name string) ([]string, error) {
		return configTargets(gen, name)
	})
}

// enumerateGitignoreAll lists each selected config's targets as its own outturn subject and prints
// the de-duplicated union to stdout. targetsFor reads one config's targets.
func enumerateGitignoreAll(run *gitignoreRun, selected []string, targetsFor func(name string) ([]string, error)) error {
	var all []string
	for _, name := range selected {
		subject := run.beginSubject(name)
		targets, err := targetsFor(name)
		if err != nil {
			// The listing stops at the first failure, which this config's subject carries.
			subject.finish(err)
			return err
		}
		subject.setPayload(gitignorePayload(targets))
		subject.finish(nil)
		all = append(all, targets...)
	}
	// The text output is the cross-config sorted, de-duplicated union.
	printGitignore(dedupeSorted(all))
	return nil
}

// configTargets lists every placement target of the config. It gets the link-farm via DryBuild,
// without a gcroot.
func configTargets(gen generator.Generator, name string) ([]string, error) {
	store, err := gen.DryBuild(name)
	if err != nil {
		return nil, err
	}
	m, err := manifest.Load(store)
	if err != nil {
		return nil, err
	}
	targets := make([]string, 0, len(m.Entries))
	for _, e := range m.Entries {
		targets = append(targets, e.Target)
	}
	return targets, nil
}

// dedupeSorted sorts and de-duplicates (for --all's combined listing).
func dedupeSorted(in []string) []string {
	sort.Strings(in)
	out := in[:0]
	var prev string
	for i, s := range in {
		if i == 0 || s != prev {
			out = append(out, s)
		}
		prev = s
	}
	return out
}

// printGitignore prints targets to stdout in /-anchor form, one per line, so the output can be
// appended to .gitignore. It prints nothing under --json.
func printGitignore(targets []string) {
	if flagJSON {
		return
	}
	for _, t := range targets {
		fmt.Println(gitignoreAnchor(t))
	}
}

// gitignoreAnchor normalizes a root-relative target into /-anchor form: a leading /, no trailing /.
func gitignoreAnchor(target string) string {
	t := strings.TrimSuffix(target, "/")
	t = strings.TrimPrefix(t, "/")
	return "/" + t
}
