package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"
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
		return fmt.Errorf("layat: unknown template: %q (valid values: %s)", template, strings.Join(initTemplates, " / "))
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
		fmt.Fprintf(os.Stderr, "layat: + nix %s\n", strings.Join(args, " "))
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
// otherwise the raw nix stderr attached.
func initNixError(args []string, stderr string, runErr error) error {
	if initExperimentalDisabled(stderr) {
		return initExperimentalGuidance(stderr)
	}
	trimmed := strings.TrimSpace(stderr)
	if trimmed == "" {
		return fmt.Errorf("layat: nix %s failed: %w", args[0], runErr)
	}
	return fmt.Errorf("layat: nix %s failed:\n%s", args[0], trimmed)
}

// initExperimentalDisabled detects the nix-command / flakes not-enabled error.
func initExperimentalDisabled(stderr string) bool {
	return strings.Contains(stderr, "experimental Nix feature") ||
		strings.Contains(stderr, "experimental-features") ||
		(strings.Contains(stderr, "flakes") && strings.Contains(stderr, "disabled"))
}

// initExperimentalGuidance builds an error explaining how to enable experimental features, with
// the raw nix error attached. The CLI never adds --extra-experimental-features itself.
func initExperimentalGuidance(stderr string) error {
	return fmt.Errorf(`layat: nix's experimental-features are not enabled.
This command internally uses `+"`nix eval`"+` / `+"`nix build`"+` (the new CLI) and flakes,
so experimental-features = nix-command flakes is required.

How to enable (either one):
  - Append to ~/.config/nix/nix.conf or /etc/nix/nix.conf:
      experimental-features = nix-command flakes
  - Temporarily via an environment variable:
      export NIX_CONFIG="experimental-features = nix-command flakes"

layat does not add --extra-experimental-features automatically (it will not override your environment settings).

Original nix error:
%s`, strings.TrimSpace(stderr))
}
