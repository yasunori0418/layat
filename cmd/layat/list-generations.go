package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// generationsInfo is list-generations' result.info: the generation inventory. It is a pointer so
// failures before the listing omit info.
type generationsInfo struct {
	Generations []generationRow `json:"generations"`
}

// listGenerationsRun is list-generations' concrete run instantiation, threaded from RunE.
type listGenerationsRun = outturnRun[*generationsInfo, *struct{}]

// beginListGenerationsRun starts list-generations' run.
func beginListGenerationsRun(command string) *listGenerationsRun {
	return beginOutturnRun[*generationsInfo, *struct{}](command)
}

func newListGenerationsCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "list-generations [name]",
		Short: "List generations (home mode only)",
		Long: "A read-only command that lists the generations of the home mode profile. " +
			"Pass <name> for that config, or --all to list every home mode config.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// beginListGenerationsRun publishes the run to outturnReport, so main emits the envelope after
			// Execute.
			run := beginListGenerationsRun(cmd.Name())
			if all {
				if len(args) > 0 {
					return fmt.Errorf("layat: list-generations cannot combine <name> with --all")
				}
				return runListAllGenerations(run)
			}
			if len(args) != 1 {
				return fmt.Errorf("layat: list-generations requires <name> or --all")
			}
			return runListGenerations(run, args[0])
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "List generations for every home mode config")
	return cmd
}

// runListGenerations confirms rootKind via eval pre-resolution (home mode only), resolves profileDir, and lists generations.
func runListGenerations(run *listGenerationsRun, name string) error {
	// The config name is the run's single outturn subject.
	subject := run.beginSubject(name)
	gen, err := newGenerator()
	if err != nil {
		return err
	}
	if err := gen.Discover(flagFile); err != nil {
		return err
	}

	root, err := gen.Roots(name)
	if err != nil {
		return err
	}
	rootKind, fixedRoot := root.RootKind, root.Root
	if rootKind != manifest.RootKindHome {
		return fmt.Errorf("layat: list-generations is home mode only (layat.%s has rootKind=%q)", name, rootKind)
	}

	prof, _, err := engine.ProfileFor(engine.ProfileOptions{
		Name:         name,
		RootKind:     rootKind,
		FixedRoot:    fixedRoot,
		RootOverride: flagRoot,
	})
	if err != nil {
		return err
	}
	gens, err := engine.ListGenerations(prof.Profile)
	if err != nil {
		return err
	}
	// The listing rides result.info; items stays [] and the generation slot stays absent.
	subject.setPayload(generationsPayload(gens))
	printGenerations(gens)
	return nil
}

// generationsPayload wraps a config's listing as its SubjectResult payload in result.info.
func generationsPayload(gens []engine.Generation) *outturnPayload[*generationsInfo] {
	return &outturnPayload[*generationsInfo]{info: &generationsInfo{Generations: generationRows(gens)}}
}

// generationRow is one generation of the --json inventory. date is nix-env's display timestamp
// verbatim.
type generationRow struct {
	Number  int    `json:"number"`
	Date    string `json:"date"`
	Current bool   `json:"current"`
}

// generationRows converts the engine listing into the info inventory (non-nil even when empty,
// so an empty profile still lists "generations": []).
func generationRows(gens []engine.Generation) []generationRow {
	rows := make([]generationRow, 0, len(gens))
	for _, g := range gens {
		rows = append(rows, generationRow{Number: g.Number, Date: g.Date, Current: g.Current})
	}
	return rows
}

// runListAllGenerations lists the generations of every home profile directly under
// <state>/nix/profiles/layat, one subject per config. It reads the disk only; roothash-family
// directories hold no profile directly and are skipped.
func runListAllGenerations(run *listGenerationsRun) error {
	stateDir, err := paths.StateDir()
	if err != nil {
		return err
	}
	base := paths.Base(stateDir)
	dents, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no profile created yet = empty listing.
		}
		return fmt.Errorf("layat: cannot read the profile base (%s): %w", base, err)
	}

	var names []string
	for _, d := range dents {
		if !d.IsDir() {
			continue
		}
		prof := paths.Resolve(stateDir, d.Name(), manifest.RootKindHome, "", false)
		if _, err := os.Lstat(prof.Profile); err != nil {
			continue // no profile directly under it = roothash family / empty directory.
		}
		names = append(names, d.Name())
	}
	sort.Strings(names)

	for i, name := range names {
		subject := run.beginSubject(name)
		prof := paths.Resolve(stateDir, name, manifest.RootKindHome, "", false)
		gens, err := engine.ListGenerations(prof.Profile)
		if err != nil {
			// The listing stops at the first failure, which this config's subject carries.
			subject.finish(err)
			return err
		}
		subject.setPayload(generationsPayload(gens))
		subject.finish(nil)
		// Under --json stdout belongs to the envelope; the listing still runs for the exit behavior.
		if flagJSON {
			continue
		}
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("# %s\n", name)
		printGenerations(gens)
	}
	return nil
}

// printGenerations prints the generation list to stdout. It prints nothing under --json; the --all
// path gates its own header lines.
func printGenerations(gens []engine.Generation) {
	if flagJSON {
		return
	}
	for _, g := range gens {
		marker := ""
		if g.Current {
			marker = "\t(current)"
		}
		fmt.Printf("%d\t%s%s\n", g.Number, g.Date, marker)
	}
}
