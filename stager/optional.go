package stager

import (
	"fmt"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

// optPackage describes an optional package installation.
type optPackage struct {
	If          *StepCondition `toml:"if"`
	DisplayName string         `toml:"display_name"`
	Version     string         `toml:"version"`
	Packages    []string       `toml:"packages"`
}

func optPackagesConfig(opts Options, tree *toml.Tree) (*units.Composite, error) {
	conf := map[string]optPackage{}
	t := tree.Get(keyOptPackages)
	if t == nil {
		return nil, nil
	}
	ge, ok := t.(*toml.Tree)
	if !ok {
		return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyOptPackages, t)
	}
	if err := ge.Unmarshal(&conf); err != nil {
		return nil, err
	}
	if len(conf) == 0 {
		return nil, nil
	}

	var out []units.Unit
	for name, pkg := range conf {
		skip, err := pkg.If.ShouldSkip(tree, opts)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}

		out = append(out, &units.OptPackage{
			OptName:     name,
			DisplayName: pkg.DisplayName,
			Version:     pkg.Version,
			Packages:    pkg.Packages,
		})
	}

	if len(out) == 0 {
		return nil, nil
	}
	return &units.Composite{
		UnitName: "opt-packages",
		Ops:      out,
	}, nil
}
