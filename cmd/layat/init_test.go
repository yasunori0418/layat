package main

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/generator/nixgen"
)

// TestFlakeInitArgs: the argv is `nix flake init -t <ref>#<template>`.
func TestFlakeInitArgs(t *testing.T) {
	cases := []struct {
		name     string
		template string
		ref      string
		want     []string
	}{
		{
			name:     "default ref",
			template: "project",
			ref:      defaultTemplateRef,
			want:     []string{"flake", "init", "-t", "github:yasunori0418/layat#project"},
		},
		{
			name:     "standalone",
			template: "standalone",
			ref:      defaultTemplateRef,
			want:     []string{"flake", "init", "-t", "github:yasunori0418/layat#standalone"},
		},
		{
			name:     "env override ref (path: local reference)",
			template: "project",
			ref:      "path:/tmp/layat",
			want:     []string{"flake", "init", "-t", "path:/tmp/layat#project"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := flakeInitArgs(tc.template, tc.ref); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("flakeInitArgs(%q, %q) = %v, want %v", tc.template, tc.ref, got, tc.want)
			}
		})
	}
}

// isValidTemplate is true only for accepted template names (rejects invalid values, sending them to the exit 1 path).
func TestIsValidTemplate(t *testing.T) {
	valid := []string{"standalone", "project"}
	for _, v := range valid {
		if !isValidTemplate(v) {
			t.Errorf("isValidTemplate(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "Project", "home", "standalone ", "default"}
	for _, v := range invalid {
		if isValidTemplate(v) {
			t.Errorf("isValidTemplate(%q) = true, want false", v)
		}
	}
}

// TestInitJSONEnvelopeInfoAbsentOnRejectedTemplate: an unknown template fails before
// setEnvelopeInfo, so the envelope's info stays absent.
func TestInitJSONEnvelopeInfoAbsentOnRejectedTemplate(t *testing.T) {
	r, buf := newInitTestRun()
	if err := r.emit(errors.New(`unknown template: "nosuch"`)); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertNoInfoKeys(t, decodeEnvelope(t, buf))
}

// TestInitJSONEnvelopeInfo pins init's --json shape: results stays [] and the template / ref ride
// in the envelope-wide info. The envelope is conformant.
func TestInitJSONEnvelopeInfo(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}

	r, buf := newInitTestRun()
	r.setEnvelopeInfo(&initInfo{Template: "standalone", Ref: defaultTemplateRef})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings: %v\ndocument: %s", findings, buf.String())
	}

	doc := decodeEnvelope(t, buf)
	if results := doc["results"].([]any); len(results) != 0 {
		t.Errorf("results = %v, want [] (init has no subject)", results)
	}
	info := doc["info"].(map[string]any)
	if info["template"] != "standalone" || info["ref"] != defaultTemplateRef {
		t.Errorf("info = %v, want template=standalone ref=%s", info, defaultTemplateRef)
	}
	if doc["status"] != "success" || doc["dryRun"] != false {
		t.Errorf("status/dryRun = %v/%v, want success/false", doc["status"], doc["dryRun"])
	}
}

// TestInitNixError pins a failed nix flake init: the summary naming the installable with nix's
// stderr on the following lines, or the experimental-features guidance, all keeping nix's exit as
// the cause without taking over its exit code.
func TestInitNixError(t *testing.T) {
	runErr := exec.Command("sh", "-c", "exit 3").Run()
	args := flakeInitArgs("project", "path:/tmpl")
	const disabled = "error: experimental Nix feature 'flakes' is disabled"
	cases := []struct{ name, stderr, want string }{
		{"failure", "error: boom\n", "cannot run nix flake init (path:/tmpl#project): exit status 3\nerror: boom"},
		{"no stderr", "", "cannot run nix flake init (path:/tmpl#project): exit status 3"},
		{"experimental", disabled + "\n", "nix's experimental-features are not enabled (nix flake init): exit status 3\n" +
			nixgen.ExperimentalGuidance + "\n\nOriginal nix error:\n" + disabled},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := initNixError(args, c.stderr, runErr)
			if err.Error() != c.want {
				t.Errorf("error = %q, want %q", err, c.want)
			}
			var ee *exec.ExitError
			if !errors.As(err, &ee) {
				t.Errorf("errors.As(*exec.ExitError) = false: %v", err)
			}
			if got := exitCodeFor(err); got != 1 {
				t.Errorf("exitCodeFor = %d, want 1", got)
			}
		})
	}
}
