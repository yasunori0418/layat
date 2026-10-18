package nixprofile

import (
	"path/filepath"
	"testing"
)

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
