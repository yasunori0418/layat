package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/manifest"
)

// TestGitignoreAnchor: a root-relative target becomes /-anchor form (leading /, no trailing /).
func TestGitignoreAnchor(t *testing.T) {
	cases := map[string]string{
		".claude/skills/nix": "/.claude/skills/nix",
		".config/nvim":       "/.config/nvim",
		"dir/":               "/dir", // no trailing / is added
		"/already/anchored":  "/already/anchored",
		"file.txt":           "/file.txt",
	}
	for in, want := range cases {
		if got := gitignoreAnchor(in); got != want {
			t.Errorf("gitignoreAnchor(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDedupeSorted(t *testing.T) {
	in := []string{".b", ".a", ".b", ".c", ".a"}
	want := []string{".a", ".b", ".c"}
	if got := dedupeSorted(in); !reflect.DeepEqual(got, want) {
		t.Errorf("dedupeSorted = %v, want %v", got, want)
	}
	if got := dedupeSorted(nil); len(got) != 0 {
		t.Errorf("dedupeSorted(nil) = %v, want empty", got)
	}
}

// TestGitignoreJSONInfoPaths pins gitignore's --json shape: anchor-form result.info.paths, empty
// items, no generation slot, and a conformant envelope.
func TestGitignoreJSONInfoPaths(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}

	r, buf := newGitignoreTestRun()
	r.beginSubject("docs").setPayload(&outturnPayload[*gitignoreInfo]{info: &gitignoreInfo{
		Paths: gitignoreAnchors([]string{".claude/skills/nix", ".layat-out/docs"}),
	}})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings: %v\ndocument: %s", findings, buf.String())
	}

	doc := decodeEnvelope(t, buf)
	if doc["dryRun"] != false {
		t.Errorf("dryRun = %v, want false (a read-only command reflects only the flag)", doc["dryRun"])
	}
	sr := doc["results"].([]any)[0].(map[string]any)
	if _, ok := sr["generation"]; ok {
		t.Errorf("generation slot present, want absent for gitignore: %v", sr["generation"])
	}
	res := sr["result"].(map[string]any)
	if items := res["items"].([]any); len(items) != 0 {
		t.Errorf("items = %v, want empty (the enumeration is info, not items)", items)
	}
	paths := res["info"].(map[string]any)["paths"].([]any)
	if len(paths) != 2 || paths[0] != "/.claude/skills/nix" || paths[1] != "/.layat-out/docs" {
		t.Errorf("info.paths = %v, want the anchor-form targets", paths)
	}
}

// TestGitignoreJSONInfoAbsentWithoutEnumeration: when gitignore fails after registering the subject
// but before the listing exists, result.info stays absent.
func TestGitignoreJSONInfoAbsentWithoutEnumeration(t *testing.T) {
	r, buf := newGitignoreTestRun()
	r.beginSubject("docs")
	if err := r.emit(errors.New("gitignore is project mode only")); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertNoInfoKeys(t, decodeEnvelope(t, buf))
}

// TestGitignoreJSONEmptyPathsStaysArray: a zero-entry run emits "paths": [] rather than dropping
// the key.
func TestGitignoreJSONEmptyPathsStaysArray(t *testing.T) {
	r, buf := newGitignoreTestRun()
	r.beginSubject("empty").setPayload(&outturnPayload[*gitignoreInfo]{info: &gitignoreInfo{Paths: gitignoreAnchors(nil)}})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	info := decodeEnvelope(t, buf)["results"].([]any)[0].(map[string]any)["result"].(map[string]any)["info"].(map[string]any)
	paths, ok := info["paths"].([]any)
	if !ok {
		t.Fatalf("info = %v, want the paths key present as an array", info)
	}
	if len(paths) != 0 {
		t.Errorf("info.paths = %v, want []", paths)
	}
}

// TestGitignoreImplicitInputRecord: gitignore records the discovered entrypoint and the adopted
// generator in the top-level info, with or without --all, and never records a root. Each given
// flag drops its own key, and with both given the top-level info is absent.
func TestGitignoreImplicitInputRecord(t *testing.T) {
	gitignore := func(t *testing.T, all bool) map[string]any {
		t.Helper()
		run, buf := newGitignoreTestRun()
		var err error
		captureOutErr(t, func() {
			if all {
				err = runGitignoreAll(run)
			} else {
				err = runGitignore(run, "p")
			}
		})
		if err != nil {
			t.Fatalf("gitignore: %v", err)
		}
		if err := run.emit(nil); err != nil {
			t.Fatalf("emit: %v", err)
		}
		return decodeEnvelope(t, buf)
	}
	const evalAll = `{"p":{"rootKind":"project","targets":[]}}`
	for _, all := range []bool{false, true} {
		name := map[bool]string{false: "single", true: "--all"}[all]
		t.Run(name+" with --generator given", func(t *testing.T) {
			repo := implicitInputProject(t, manifest.RootKindProject, evalAll)
			flagGenerator = "nix"
			if top, _ := envelopeInfo(t, gitignore(t, all), 0); len(top) != 1 || top["entrypoint"] != repo {
				t.Errorf("info = %v, want {entrypoint: %s}", top, repo)
			}
		})
		t.Run(name+" with -f and --generator given", func(t *testing.T) {
			repo := implicitInputProject(t, manifest.RootKindProject, evalAll)
			flagFile, flagGenerator = repo, "nix"
			if v, ok := gitignore(t, all)["info"]; ok {
				t.Errorf("info = %v, want absent", v)
			}
		})
	}
	for _, all := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "--all"}[all], func(t *testing.T) {
			repo := implicitInputProject(t, manifest.RootKindProject, evalAll)
			top, result := envelopeInfo(t, gitignore(t, all), 0)
			if top["entrypoint"] != repo || top["generator"] != "nix" || len(top) != 2 {
				t.Errorf("info = %v, want {entrypoint: %s, generator: nix}", top, repo)
			}
			if _, ok := result["root"]; ok {
				t.Errorf("result.info = %v, want no root", result)
			}
		})
	}
}
