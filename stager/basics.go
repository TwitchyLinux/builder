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

	golangDefault = GolangConf{
		Version: "1.13.4",
		URL:     "https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz",
		SHA256:  "692d17071736f74be04a72a06dab9cac1cd759377bd85316e52b2227604c004c",
	}

	linuxDefault = LinuxConf{
		Version: "5.1.18",
		URL:     "https://mirrors.edge.kernel.org/pub/linux/kernel/v5.x/linux-5.1.18.tar.xz",
		SHA256:  "6013e7dcf59d7c1b168d8edce3dbd61ce340ff289541f920dbd0958bef98f36a",
	}
)

// GolangConf describes what Go toolchain to install.
type GolangConf struct {
	Version string `toml:"version"`
	URL     string `toml:"url"`
	SHA256  string `toml:"sha256"`
}

func golangConf(tree *toml.Tree) (*units.Golang, error) {
	conf := golangDefault
	if t := tree.Get(rootKeyGolang); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyGolang, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Golang{
		Version: conf.Version,
		URL:     conf.URL,
		SHA256:  conf.SHA256,
	}, nil
}

// LinuxConf describes what Linux kernel to install
type LinuxConf struct {
	Version string `toml:"version"`
	URL     string `toml:"url"`
	SHA256  string `toml:"sha256"`
}

func linuxConf(tree *toml.Tree) (*units.Linux, error) {
	conf := linuxDefault
	if t := tree.Get(rootKeyLinux); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyLinux, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Linux{
		Version: conf.Version,
		URL:     conf.URL,
		SHA256:  conf.SHA256,
	}, nil
}

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
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", rootKeyLocale, t)
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
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", key, t)
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
