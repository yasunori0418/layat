package generator

import (
	"fmt"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/manifest"
)

// Prebuilt is the generator behind apply --manifest (→ ADR-0055 §5, ADR-0026): the link-farm is
// already built, so Discover takes its path, Roots reads its manifest.json, and Build / DryBuild
// return the path as-is. It runs no external command.
type Prebuilt struct {
	linkFarm string
}

// Discover takes the pre-built link-farm's path (the --manifest value) and makes it absolute.
func (p *Prebuilt) Discover(file string) error {
	abs, err := filepath.Abs(file)
	if err != nil {
		msg := fmt.Sprintf("layat: cannot resolve the --manifest path (%s): %v", file, err)
		return NewError(NamePrebuilt, StageDiscover, KindFailed, msg, "", "", err)
	}
	p.linkFarm = abs
	return nil
}

// Roots reads the root from the link-farm's manifest.json, deriving Targets from its entries.
// The link-farm holds a single config, so name is not consulted.
func (p *Prebuilt) Roots(string) (manifest.Root, error) {
	m, err := manifest.Load(p.linkFarm)
	if err != nil {
		return manifest.Root{}, NewError(NamePrebuilt, StageRoots, KindFailed, err.Error(), "", "", err)
	}
	root := m.Root
	root.Targets = make([]string, 0, len(m.Entries))
	for _, e := range m.Entries {
		root.Targets = append(root.Targets, e.Target)
	}
	return root, nil
}

// AllRoots returns the link-farm's single config keyed by its directory name (a link-farm carries
// no config name of its own).
func (p *Prebuilt) AllRoots() (map[string]manifest.Root, error) {
	root, err := p.Roots("")
	if err != nil {
		return nil, err
	}
	return map[string]manifest.Root{filepath.Base(p.linkFarm): root}, nil
}

// Build returns the pre-built link-farm; there is nothing to lay down at pending.
func (p *Prebuilt) Build(string, string) (string, error) { return p.linkFarm, nil }

// DryBuild returns the pre-built link-farm.
func (p *Prebuilt) DryBuild(string) (string, error) { return p.linkFarm, nil }
