package stager

import (
	"fmt"
	"sort"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

// InstallAction describes a step during installation.
type InstallAction struct {
	Action string `toml:"action"`

	URL  string `toml:"url"`
	From string `toml:"from"`
	To   string `toml:"to"`
	Data string `toml:"data"`
	Dir  string `toml:"dir"`

	Expected string `toml:"expected"`

	Bin  string   `toml:"bin"`
	Args []string `toml:"args"`
}

// InstallConf desribes a set of packages to be installed.
type InstallConf struct {
	Order    int             `toml:"order_priority"`
	Packages []string        `toml:"packages"`
	Actions  []InstallAction `toml:"do"`
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
			ut, err := makeInstallUnit(k, c)
			if err != nil {
				return nil, err
			}
			out = append(out, ut)
		}

		sort.Slice(out, func(i int, j int) bool {
			var lhs, rhs int
			switch l := out[i].(type) {
			case *units.InstallTools:
				lhs = l.Order
			case *units.Composite:
				lhs = l.Order
			}
			switch r := out[j].(type) {
			case *units.InstallTools:
				rhs = r.Order
			case *units.Composite:
				rhs = r.Order
			}
			return lhs > rhs
		})

		return out, nil
	}

	return nil, nil
}

func makeInstallUnit(k string, c InstallConf) (units.Unit, error) {
	// Simple case - only packages to install.
	if len(c.Actions) == 0 {
		return &units.InstallTools{
			UnitName: k,
			Pkgs:     c.Packages,
			Order:    c.Order,
		}, nil
	}

	out := units.Composite{
		UnitName: k,
		Order:    c.Order,
	}
	// Add the packages.
	out.Ops = []units.Unit{&units.InstallTools{
		UnitName: k,
		Pkgs:     c.Packages,
	}}
	// Add the actions.
	for _, a := range c.Actions {
		u, err := actionToUnit(a)
		if err != nil {
			return nil, err
		}
		out.Ops = append(out.Ops, u)
	}
	return &out, nil
}

func actionToUnit(a InstallAction) (units.Unit, error) {
	switch a.Action {
	case "download":
		return &units.Download{URL: a.URL, To: a.To}, nil
	case "run":
		return &units.Cmd{Bin: a.Bin, Args: a.Args}, nil
	case "sha256sum":
		return &units.CheckHash{File: a.From, ExpectedHash: a.Expected}, nil
	case "append":
		return &units.Append{To: a.To, Data: a.Data}, nil
	case "mkdir":
		return &units.Mkdir{Dir: a.Dir}, nil
	}
	return nil, fmt.Errorf("unknown action: %q", a.Action)
}
