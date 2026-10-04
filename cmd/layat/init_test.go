package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"
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
	if err := r.emit(errors.New(`layat: unknown template: "nosuch"`)); err != nil {
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
