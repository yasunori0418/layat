package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/engine"
	"github.com/yasunori0418/layat/internal/generation"
	"github.com/yasunori0418/layat/internal/manifest"
	"github.com/yasunori0418/layat/internal/paths"
)

// generationsInfo is list-generations' result.info: the profile's mode ("home" / "system") and the
// generation inventory. It is a pointer so failures before the listing omit info.
type generationsInfo struct {
	Mode        string          `json:"mode"`
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
		Short: "List generations (home / system mode only)",
		Long: "A read-only command that lists the generations of the home / system mode profile. " +
			"Pass <name> for that config, or --all to list every home mode config, then every system mode config.",
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
	cmd.Flags().BoolVar(&all, "all", false, "List generations for every home mode config, then every system mode config")
	return cmd
}

// runListGenerations confirms rootKind via eval pre-resolution (home / system mode only), resolves profileDir, and lists generations.
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
	if rootKind != manifest.RootKindHome && rootKind != manifest.RootKindSystem {
		return fmt.Errorf("layat: list-generations is home / system mode only (layat.%s has rootKind=%q)", name, rootKind)
	}

	backend := newGenerationBackend()
	prof, _, err := engine.ProfileFor(engine.ProfileOptions{
		Name:         name,
		RootKind:     rootKind,
		FixedRoot:    fixedRoot,
		RootOverride: flagRoot,
		Backend:      backend,
		SystemDir:    systemDirOverride(),
	})
	if err != nil {
		return err
	}
	gens, err := backend.List(prof.Profile)
	if err != nil {
		return err
	}
	// The listing rides result.info; items stays [] and the generation slot stays absent.
	subject.setPayload(generationsPayload(rootKind, gens))
	printGenerations(gens)
	return nil
}

// generationsPayload wraps a config's listing and its mode as its SubjectResult payload in result.info.
func generationsPayload(mode string, gens []generation.Generation) *outturnPayload[*generationsInfo] {
	return &outturnPayload[*generationsInfo]{info: &generationsInfo{Mode: mode, Generations: generationRows(gens)}}
}

// generationRow is one generation of the --json inventory. date is the backend's display
// timestamp verbatim.
type generationRow struct {
	Number  int    `json:"number"`
	Date    string `json:"date"`
	Current bool   `json:"current"`
}

// generationRows converts the backend listing into the info inventory (non-nil even when empty,
// so an empty profile still lists "generations": []).
func generationRows(gens []generation.Generation) []generationRow {
	rows := make([]generationRow, 0, len(gens))
	for _, g := range gens {
		rows = append(rows, generationRow{Number: g.Number, Date: g.Date, Current: g.Current})
	}
	return rows
}

// runListAllGenerations lists the generations of every profile directly under the backend's user
// base (home), then under its system base (system), one subject per config. It reads the disk only;
// an unreadable system base is warned about and skipped.
func runListAllGenerations(run *listGenerationsRun) error {
	stateDir, err := paths.StateDir()
	if err != nil {
		return err
	}
	backend := newGenerationBackend()
	userBase := backend.UserBase(stateDir)
	homeNames, err := profileNamesUnder(userBase, manifest.RootKindHome)
	if err != nil {
		return fmt.Errorf("layat: cannot read the profile base (%s): %w", userBase, err)
	}
	systemBase := systemDirOverride()
	if systemBase == "" {
		systemBase = backend.SystemBase()
	}
	systemNames, err := profileNamesUnder(systemBase, manifest.RootKindSystem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "layat: warning: skipping the system profile base (%s): %v\n", systemBase, err)
	}

	printed := false
	for _, b := range []struct {
		base, mode string
		names      []string
	}{
		{userBase, manifest.RootKindHome, homeNames},
		{systemBase, manifest.RootKindSystem, systemNames},
	} {
		for _, name := range b.names {
			subject := run.beginSubject(name)
			prof := paths.Resolve(b.base, name, b.mode, "", false)
			gens, err := backend.List(prof.Profile)
			if err != nil {
				// The listing stops at the first failure, which this config's subject carries.
				subject.finish(err)
				return err
			}
			subject.setPayload(generationsPayload(b.mode, gens))
			subject.finish(nil)
			// Under --json stdout belongs to the envelope; the listing still runs for the exit behavior.
			if flagJSON {
				continue
			}
			if printed {
				fmt.Println()
			}
			printed = true
			if b.mode == manifest.RootKindSystem {
				fmt.Printf("# %s (system)\n", name)
			} else {
				fmt.Printf("# %s\n", name)
			}
			printGenerations(gens)
		}
	}
	return nil
}

// profileNamesUnder returns, sorted, the names under base holding a profile directly (roothash-family
// directories hold none and are skipped). A base that does not exist yet lists nothing.
func profileNamesUnder(base, mode string) ([]string, error) {
	dents, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, d := range dents {
		if !d.IsDir() {
			continue
		}
		prof := paths.Resolve(base, d.Name(), mode, "", false)
		if _, err := os.Lstat(prof.Profile); err != nil {
			continue
		}
		names = append(names, d.Name())
	}
	sort.Strings(names)
	return names, nil
}

// printGenerations prints the generation list to stdout. It prints nothing under --json; the --all
// path gates its own header lines.
func printGenerations(gens []generation.Generation) {
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
