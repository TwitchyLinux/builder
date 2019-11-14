package stager

import (
	"fmt"
	"sort"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

var (
	graphicalEnvDefault = GraphicsConf{Packages: []string{"gnome"}}

	localeDefault = LocaleConf{
		Area:     "America",
		Zone:     "Los_Angeles",
		Generate: []string{"en_US.UTF-8 UTF-8", "en_US ISO-8859-1"},
		Default:  "en_US.UTF-8",
	}
)

// GraphicsConf describes how a graphical environment should be installed.
type GraphicsConf struct {
	Packages []string `toml:"packages"`
}

func graphicsConf(tree *toml.Tree) (*units.Gnome, error) {
	conf := graphicalEnvDefault
	if t := tree.Get(rootKeyGraphicalEnv); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyGraphicalEnv, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Gnome{
		NeedPkgs: conf.Packages,
	}, nil
}

// LocaleConf describes the locale of the system.
type LocaleConf struct {
	Area     string   `toml:"area"`
	Zone     string   `toml:"zone"`
	Generate []string `toml:"generate_locales"`
	Default  string   `toml:"default"`
}

func localeConf(tree *toml.Tree) (*units.Locale, error) {
	conf := localeDefault
	if t := tree.Get(rootKeyLocale); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyGraphicalEnv, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Locale{
		Area:     conf.Area,
		Zone:     conf.Zone,
		Generate: conf.Generate,
		Default:  conf.Default,
	}, nil
}

// InstallConf desribes a set of packages to be installed.
type InstallConf struct {
	Order    int      `toml:"order_priority"`
	Packages []string `toml:"packages"`
}

func installsUnderKey(tree *toml.Tree, key string) ([]units.Unit, error) {
	if t := tree.Get(key); t != nil {
		installs, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyGraphicalEnv, t)
		}
		var conf map[string]InstallConf
		if err := installs.Unmarshal(&conf); err != nil {
			return nil, err
		}

		out := make([]units.Unit, 0, len(conf))
		for k, c := range conf {
			out = append(out, &units.InstallTools{
				UnitName: k,
				Pkgs:     c.Packages,
				Order:    c.Order,
			})
		}
		sort.Slice(out, func(i int, j int) bool {
			return out[i].(*units.InstallTools).Order > out[j].(*units.InstallTools).Order
		})
		return out, nil
	}

	return nil, nil
}
