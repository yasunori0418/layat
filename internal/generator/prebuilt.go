package generator

import (
	"fmt"
	"path/filepath"

	"github.com/yasunori0418/layat/internal/manifest"
)

// Prebuilt is the generator behind apply --manifest: Discover takes the pre-built link-farm's
// path, Roots reads its manifest.json, and Build / DryBuild return the path as-is. With LinkFarms
// set it holds one link-farm per config name instead (apply --all --manifest <name>=<link-farm>).
type Prebuilt struct {
	// LinkFarms maps each config name to its link-farm; nil for the single link-farm.
	LinkFarms map[string]string
	linkFarm  string
}

// Discover takes the pre-built link-farm's path (the --manifest value) and makes it absolute.
// With LinkFarms set, file is ignored and every link-farm in the set is made absolute instead.
func (p *Prebuilt) Discover(file string) error {
	if p.LinkFarms != nil {
		farms := make(map[string]string, len(p.LinkFarms))
		for name, farm := range p.LinkFarms {
			abs, err := absLinkFarm(farm)
			if err != nil {
				return err
			}
			farms[name] = abs
		}
		p.LinkFarms = farms
		return nil
	}
	abs, err := absLinkFarm(file)
	if err != nil {
		return err
	}
	p.linkFarm = abs
	return nil
}

// absLinkFarm makes a --manifest link-farm path absolute.
func absLinkFarm(file string) (string, error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		msg := fmt.Sprintf("layat: cannot resolve the --manifest path (%s): %v", file, err)
		return "", NewError(NamePrebuilt, StageDiscover, KindFailed, msg, "", "", err)
	}
	return abs, nil
}

// farm returns the link-farm of the config name: the single link-farm regardless of name, or
// name's entry in LinkFarms.
func (p *Prebuilt) farm(name string, stage Stage) (string, error) {
	if p.LinkFarms == nil {
		return p.linkFarm, nil
	}
	farm, ok := p.LinkFarms[name]
	if !ok {
		return "", NewError(NamePrebuilt, stage, KindNotFound,
			fmt.Sprintf("layat: no --manifest link-farm is given for config %q", name), "", "", nil)
	}
	return farm, nil
}

// Roots reads the config's root from its link-farm's manifest.json, deriving Targets from its
// entries.
func (p *Prebuilt) Roots(name string) (manifest.Root, error) {
	farm, err := p.farm(name, StageRoots)
	if err != nil {
		return manifest.Root{}, err
	}
	return rootOf(farm)
}

// rootOf reads the root from the link-farm's manifest.json, deriving Targets from its entries.
func rootOf(linkFarm string) (manifest.Root, error) {
	m, err := manifest.Load(linkFarm)
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

// AllRoots returns the root of every config in LinkFarms. A single link-farm has no config list
// to enumerate, so it fails.
func (p *Prebuilt) AllRoots() (map[string]manifest.Root, error) {
	if p.LinkFarms == nil {
		return nil, NewError(NamePrebuilt, StageRoots, KindFailed,
			"layat: a pre-built manifest (--manifest) holds a single config and cannot list all configs", "", "", nil)
	}
	roots := make(map[string]manifest.Root, len(p.LinkFarms))
	for name, farm := range p.LinkFarms {
		root, err := rootOf(farm)
		if err != nil {
			return nil, err
		}
		roots[name] = root
	}
	return roots, nil
}

// Build returns the config's pre-built link-farm; there is nothing to lay down at pending.
func (p *Prebuilt) Build(name, _ string) (string, error) { return p.farm(name, StageBuild) }

// DryBuild returns the config's pre-built link-farm.
func (p *Prebuilt) DryBuild(name string) (string, error) { return p.farm(name, StageBuild) }

// Entrypoint is "": a pre-built link-farm has no entrypoint.
func (p *Prebuilt) Entrypoint() string { return "" }
