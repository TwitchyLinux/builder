// Package stager reads config to pick units to run during install.
package stager

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
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
	keyGraphicalEnvName = "features.graphical_environment"
	installKeyPostBase  = "post_base.install"
	installKeyPostGUI   = rootKeyGraphicalEnv + ".post.install"
	rootKeyUdev         = "udev"
	keyUdevRules        = rootKeyUdev + ".rules"
	rootKeySysd         = "systemd"
	keySysdNetworks     = rootKeySysd + ".networks"
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

	doGraphicalInstaller, err := featuresAreSet([]string{"graphical"}, conf)
	if err != nil {
		return nil, err
	}
	if doGraphicalInstaller {
		ge, err := graphicsConf(opts, conf, dir)
		if err != nil {
			return nil, err
		}
		out = append(out, ge...)
		// Install post-GUI packages.
		if ge != nil {
			if installs, err = installsUnderKey(opts, conf, installKeyPostGUI, dir); err != nil {
				return nil, err
			}
			out = append(out, installs...)
			out = append(out, afterGUIUnits...)
		}
	}

	udev, err := udevConf(opts, conf)
	if err != nil {
		return nil, err
	}
	if udev != nil {
		out = append(out, udev)
	}
	sysdNet, err := systemdNetConfig(opts, conf)
	if err != nil {
		return nil, err
	}
	if sysdNet != nil {
		out = append(out, sysdNet)
	}

	if doGraphicalInstaller {
		out = append(out, &units.Installer{})
	}
	out = append(out, finalUnits...)
	return out, nil
}

func featuresAreSet(wantFeatures []string, tree *toml.Tree) (bool, error) {
	env, err := cel.NewEnv(cel.Declarations(decls.NewIdent("features", decls.Dyn, nil)))
	if err != nil {
		return false, err
	}

	var f map[string]interface{}
	if features := tree.Get("features"); features != nil {
		switch ft := features.(type) {
		case *toml.Tree:
			f = ft.ToMap()
		case map[string]interface{}:
			f = ft
		case []string:
			f = make(map[string]interface{}, len(ft))
			for _, v := range ft {
				f[v] = true
			}
		}
	}
	if f == nil {
		f = map[string]interface{}{}
	}
	for feature, v := range defaultFeatures {
		if _, present := f[feature]; !present {
			f[feature] = v
		}
	}

	m := map[string]interface{}{
		"conf":     tree.ToMap(),
		"features": f,
	}

	for _, e := range wantFeatures {
		outcome, err := evalCEL(env, "features."+e, m)
		if err != nil {
			return false, fmt.Errorf("evaluating feature %q: %v", e, err)
		}
		if !outcome {
			return false, nil
		}
	}
	return true, nil
}

func evalCEL(env *cel.Env, expr string, m map[string]interface{}) (bool, error) {
	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	prg, err := env.Program(checked)
	if err != nil {
		return false, err
	}
	out, _, err := prg.Eval(m)
	if err != nil {
		return false, err
	}
	v, err := out.ConvertToNative(reflect.TypeOf(true))
	if err != nil {
		return false, err
	}
	return v.(bool), nil
}

func baseUnitsFromConf(out []units.Unit, conf *toml.Tree) ([]units.Unit, error) {
	out = append(out, &units.Preflight{})
	dbstrp, err := debootstrapConf(conf)
	if err != nil {
		return nil, err
	}
	out = append(out, dbstrp)
	out = append(out, &units.FinalizeApt{Track: dbstrp.Track})

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
