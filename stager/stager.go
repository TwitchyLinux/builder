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
	rootKeyLinux        = "linux"
	rootKeyGolang       = "go_toolchain"
	installKeyPostBase  = "post_base.install"
	installKeyPostGUI   = rootKeyGraphicalEnv + ".post.install"
)

func unionTree(target, in *toml.Tree, inPrefix []string) error {
	for _, k := range in.Keys() {
		v := in.Get(k)
		t, isTree := v.(*toml.Tree)

		if !isTree {
			target.SetPath(append(inPrefix, k), v)
			continue
		}

		if err := unionTree(target, t, append(inPrefix, k)); err != nil {
			return err
		}
	}
	return nil
}

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
			if err := unionTree(conf, t, nil); err != nil {
				return nil, err
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

	out = append(out, &units.BaseBuildtools{})
	linux, err := linuxConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, linux)

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
