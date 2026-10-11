package nixprofile

import (
	"path/filepath"
	"testing"
)

func TestParseGenerations(t *testing.T) {
	out := "" +
		"   1   2026-06-01 10:00:00   \n" +
		"   2   2026-06-02 11:30:00   (current)\n"
	gens, err := parseGenerations(out)
	if err != nil {
		t.Fatalf("parseGenerations: %v", err)
	}
	if len(gens) != 2 {
		t.Fatalf("len = %d, want 2", len(gens))
	}
	if gens[0].Number != 1 || gens[0].Current {
		t.Errorf("gen[0] = %+v", gens[0])
	}
	if gens[0].Date != "2026-06-01 10:00:00" {
		t.Errorf("gen[0].Date = %q", gens[0].Date)
	}
	if gens[1].Number != 2 || !gens[1].Current {
		t.Errorf("gen[1] = %+v", gens[1])
	}
	if gens[1].Date != "2026-06-02 11:30:00" {
		t.Errorf("gen[1].Date = %q (should not include (current))", gens[1].Date)
	}
}

func TestParseGenerationsEmpty(t *testing.T) {
	gens, err := parseGenerations("\n  \n")
	if err != nil {
		t.Fatalf("parseGenerations: %v", err)
	}
	if len(gens) != 0 {
		t.Errorf("len = %d, want 0", len(gens))
	}
}

func TestParseGenerationsBadLine(t *testing.T) {
	if _, err := parseGenerations("not-a-number 2026-06-01\n"); err == nil {
		t.Fatal("expected parse error for non-numeric generation, got nil")
	}
}

// TestBases pins the two bases the nix profile backend owns: the user state base under <state>
// and the fixed system base.
func TestBases(t *testing.T) {
	b := New()
	if got, want := b.UserBase("/state"), filepath.Join("/state", "nix", "profiles", "layat"); got != want {
		t.Errorf("UserBase = %q, want %q", got, want)
	}
	if got, want := b.SystemBase(), "/nix/var/nix/profiles/layat"; got != want {
		t.Errorf("SystemBase = %q, want %q", got, want)
	}
}
