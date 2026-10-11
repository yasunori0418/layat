package main

import (
	"github.com/yasunori0418/layat/internal/generation"
	"github.com/yasunori0418/layat/internal/generation/nixprofile"
)

// newGenerationBackend is the single injection point of the generation backend every subcommand
// hands to the engine. There is no selection: the nix profile backend is the only one. Tests
// replace it with the Fake.
var newGenerationBackend = func() generation.Backend {
	return nixprofile.New()
}
