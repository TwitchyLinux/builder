// Package stager reads config to pick units to run during install.
package stager

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

const (
	rootKeyBase    = "base"
	keyDebian      = rootKeyBase + ".debian"
	keyLocale      = rootKeyBase + ".locale"
	keyLinux       = rootKeyBase + ".linux"
	keyReleaseInfo = rootKeyBase + ".release_info"
	keyShellCust   = rootKeyBase + ".shell_customization"
	keyMainUser    = rootKeyBase + ".main_user"

	rootKeyGraphicalEnv = "graphical_environment"
	installKeyPostBase  = "post_base.install"
	installKeyPostGUI   = rootKeyGraphicalEnv + ".post.install"
	rootKeyUdev         = "udev"
	keyUdevRules        = rootKeyUdev + ".rules"
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

// Options describes settings which change how stages are selected
// and generated.
type Options struct {
	// Overrides specifies the value for a given config key. If the key is
	// already set in the config file, this value will take precedence.
	Overrides map[string]interface{}
}

// UnitsFromConfig returns a set of units that represent the configuration
// in the directory provided.
func UnitsFromConfig(dir string, opts Options) ([]units.Unit, error) {
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

	// Apply any configuration overrides.
	for key, val := range opts.Overrides {
		conf.Set(key, val)
	}

	// Build base system.
	out, err = baseUnitsFromConf(out, conf)
	if err != nil {
		return nil, err
	}

	// Install specified packages.
	installs, err := installsUnderKey(opts, conf, installKeyPostBase, dir)
	if err != nil {
		return nil, err
	}
	out = append(out, installs...)

	ge, err := graphicsConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, ge)

	// Install post-GUI packages.
	if ge != nil {
		if installs, err = installsUnderKey(opts, conf, installKeyPostGUI, dir); err != nil {
			return nil, err
		}
		out = append(out, installs...)
		out = append(out, afterGUIUnits...)
	}

	udev, err := udevConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, udev)

	out = append(out, finalUnits...)
	return out, nil
}

func baseUnitsFromConf(out []units.Unit, conf *toml.Tree) ([]units.Unit, error) {
	out = append(out, &units.Preflight{})
	dbstrp, err := debootstrapConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, dbstrp)
	out = append(out, &units.FinalizeApt{})

	locale, err := localeConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, locale)

	linux, err := linuxConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, linux)

	shellUnits, err := shellUserConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, shellUnits)

	out = append(out, systemBuildUnits...)

	releaseConfUnits, err := releaseConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, releaseConfUnits...)
	return out, nil
}
