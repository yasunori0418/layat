package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yasunori0418/layat/internal/generator/nixgen"
)

// defaultTemplateRef is the fixed flake ref of init's templates. LAYAT_TEMPLATE_REF overrides it.
const defaultTemplateRef = "github:yasunori0418/layat"

// initTemplates is the template names that init accepts (matching the output names in flake.templates).
var initTemplates = []string{"standalone", "project"}

// initInfo is init's envelope-wide info: the template expansion's run facts. It is a pointer so
// failures before setEnvelopeInfo omit info.
type initInfo struct {
	Template string `json:"template"`
	Ref      string `json:"ref"`
}

// initRun is init's concrete run instantiation, threaded from RunE into runInit.
type initRun = outturnRun[*struct{}, *initInfo]

// beginInitRun starts init's run.
func beginInitRun(command string) *initRun {
	return beginOutturnRun[*struct{}, *initInfo](command)
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <template>",
		Short: "Expand a starter template into the CWD via nix flake init (standalone / project)",
		Long: "layat init <template> is a transparent wrapper around `nix flake init -t <ref>#<template>`. " +
			"It expands a starter flake into the CWD (layat generates no files; the nix templates mechanism handles expansion).\n\n" +
			"template:\n" +
			"  standalone  homeRoot example (places under $HOME)\n" +
			"  project     projectRoot example + devShell + shellHook + .gitignore\n\n" +
			"The template reference is a fixed ref (" + defaultTemplateRef + "). Override it with LAYAT_TEMPLATE_REF.\n" +
			"Existing files are not overwritten (inherits nix flake init's behavior).",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			run := beginInitRun(cmd.Name())
			return runInit(run, args[0])
		},
	}
}

// runInit validates the template name and runs `nix flake init -t <ref>#<template>` in the CWD,
// without entrypoint discovery.
func runInit(run *initRun, template string) error {
	if !isValidTemplate(template) {
		return &inputError{err: fmt.Errorf("unknown template: %q (valid values: %s)", template, strings.Join(initTemplates, " / "))}
	}

	ref := defaultTemplateRef
	if env := os.Getenv("LAYAT_TEMPLATE_REF"); env != "" {
		ref = env
	}

	// init has no subject, so the run facts ride the envelope-wide info. They are set before the
	// expansion so a failed init still reports them.
	run.setEnvelopeInfo(&initInfo{Template: template, Ref: ref})

	args := flakeInitArgs(template, ref)
	if flagDebug {
		discloseNix(os.Stderr, args)
	}

	// Capture stderr: forward nix's created-file list on success, classify it on failure.
	cmd := exec.Command("nix", args...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	// Under --json stdout belongs to the envelope, so nix's stdout goes to stderr.
	if flagJSON {
		cmd.Stdout = os.Stderr
	}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return initNixError(args, stderr.String(), err)
	}

	// On success, forward the captured list of created files (nix's output) to stderr.
	if s := stderr.String(); s != "" {
		fmt.Fprint(os.Stderr, s)
	}
	return nil
}

// isValidTemplate reports whether the template name is accepted.
func isValidTemplate(template string) bool {
	return slices.Contains(initTemplates, template)
}

// flakeInitArgs builds the argv for `nix flake init` with the ref#template installable.
func flakeInitArgs(template, ref string) []string {
	return []string{"flake", "init", "-t", ref + "#" + template}
}

// initNixError classifies a failed `nix flake init`: guidance for disabled experimental features,
// otherwise the summary naming the installable; nix's stderr follows on the next lines and its
// exit stays the cause.
func initNixError(args []string, stderr string, runErr error) error {
	trimmed := strings.TrimSpace(stderr)
	if nixgen.IsExperimentalDisabled(stderr) {
		return fmt.Errorf("nix's experimental-features are not enabled (nix flake init): %w\n%s\n\nOriginal nix error:\n%s", runErr, nixgen.ExperimentalGuidance, trimmed)
	}
	installable := args[len(args)-1] // flakeInitArgs puts <ref>#<template> last
	err := fmt.Errorf("cannot run nix flake init (%s): %w", installable, runErr)
	if trimmed == "" {
		return err
	}
	return fmt.Errorf("%w\n%s", err, trimmed)
}
