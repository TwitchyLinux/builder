package stager

import (
	"fmt"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

// GraphicsConf describes how a graphical environment should be installed.
type GraphicsConf struct {
	Packages []string               `toml:"packages"`
	Steps    map[string]InstallConf `toml:"steps"`
}

func graphicsConf(opts Options, tree *toml.Tree, resDir string) ([]units.Unit, error) {
	wantEnv := tree.GetDefault(keyGraphicalEnvName, "gnome")
	env, ok := wantEnv.(string)
	if !ok {
		return nil, fmt.Errorf("%s is %T, not string", keyGraphicalEnvName, wantEnv)
	}

	conf := graphicalEnvDefault
	if t := tree.Get(rootKeyGraphicalEnv); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyGraphicalEnv, t)
		}
		var allConfs map[string]GraphicsConf
		if err := ge.Unmarshal(&allConfs); err != nil {
			return nil, err
		}
		conf = allConfs[env]
	}

	out := make([]units.Unit, 0, len(conf.Steps))
	for k, c := range conf.Steps {
		skip, err := c.ShouldSkip(tree, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", k, err)
		}
		if skip {
			continue
		}
		ut, err := makeInstallUnit(k, c, tree, resDir)
		if err != nil {
			return nil, err
		}
		out = append(out, ut)
	}

	return append([]units.Unit{&units.InstallTools{
		UnitName: env,
		Pkgs:     conf.Packages,
	}}, out...), nil
}
