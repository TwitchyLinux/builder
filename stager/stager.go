// Stager reads config to select and configure units to run during install.
package stager

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

const (
	rootKeyGraphicalEnv = "graphical_environment"
	installKeyPostBase  = "post_base.install"
	installKeyPostGUI   = "post_gui.install"
)

// UnitsFromConfig returns a set of units that represent the configuration
// in the directory provided.
func UnitsFromConfig(dir string) ([]units.Unit, error) {
	var (
		conf, _ = toml.Load("")
		out     = append([]units.Unit{}, baseSystemUnits...)
	)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.IsDir() {
			t, err := toml.LoadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				return nil, err
			}
			for _, k := range t.Keys() {
				conf.Set(k, t.Get(k))
			}
		}
	}

	installs, err := installsUnderKey(conf, installKeyPostBase)
	if err != nil {
		return nil, err
	}
	out = append(out, installs...)

	ge, err := graphicsConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, ge)

	if installs, err = installsUnderKey(conf, installKeyPostGUI); err != nil {
		return nil, err
	}
	out = append(out, installs...)

	out = append(out, finalUnits...)
	return out, nil
}
