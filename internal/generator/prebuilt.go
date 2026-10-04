package generator

import (
	"fmt"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/manifest"
)

// Prebuilt is the generator behind apply --manifest: Discover takes the pre-built link-farm's
// path, Roots reads its manifest.json, and Build / DryBuild return the path as-is.
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

// AllRoots fails: a pre-built link-farm is a single config with no config list to enumerate
// (the CLI never selects prebuilt for apply --all).
func (p *Prebuilt) AllRoots() (map[string]manifest.Root, error) {
	return nil, NewError(NamePrebuilt, StageRoots, KindFailed,
		"layat: a pre-built manifest (--manifest) holds a single config and cannot list all configs", "", "", nil)
}

// Build returns the pre-built link-farm; there is nothing to lay down at pending.
func (p *Prebuilt) Build(string, string) (string, error) { return p.linkFarm, nil }

// DryBuild returns the pre-built link-farm.
func (p *Prebuilt) DryBuild(string) (string, error) { return p.linkFarm, nil }
