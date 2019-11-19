// Package stager reads config to pick units to run during install.
package stager

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

const (
	rootKeyGraphicalEnv = "graphical_environment"
	rootKeyLocale       = "locale"
	rootKeyGolang       = "go_toolchain"
	installKeyPostBase  = "post_base.install"
	installKeyPostGUI   = rootKeyGraphicalEnv + ".post.install"
)

// UnitsFromConfig returns a set of units that represent the configuration
// in the directory provided.
func UnitsFromConfig(dir string) ([]units.Unit, error) {
	var (
		conf, _ = toml.Load("")
		out     []units.Unit
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

	// Build base system.
	out = append(out, earlyBuildUnits...)
	locale, err := localeConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, locale)
	out = append(out, systemBuildUnits...)

	// Install specified packages.
	installs, err := installsUnderKey(conf, installKeyPostBase)
	if err != nil {
		return nil, err
	}
	out = append(out, installs...)

	// Pre-graphics packages.
	got, err := golangConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, got)

	ge, err := graphicsConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, ge)

	// Install post-GUI packages.
	if installs, err = installsUnderKey(conf, installKeyPostGUI); err != nil {
		return nil, err
	}
	out = append(out, installs...)
	out = append(out, afterGUIUnits...)

	out = append(out, finalUnits...)
	return out, nil
}
