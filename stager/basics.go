package stager

import (
	"fmt"
	"sort"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

var (
	graphicalEnvDefault = GraphicsConf{Packages: []string{"gdm3", "adwaita-icon-theme", "at-spi2-core", "baobab", "caribou", "dconf-cli", "dconf-gsettings-backend", "eog", "evince",
		"fonts-cantarell", "gedit", "glib-networking", "gnome-backgrounds", "gnome-bluetooth", "gnome-calculator", "gnome-characters", "gnome-control-center",
		"gnome-disk-utility", "gnome-font-viewer", "gnome-keyring", "gnome-logs", "gnome-menus", "gnome-session", "gnome-shell", "gnome-settings-daemon",
		"gnome-shell-extensions", "gnome-system-monitor", "gnome-terminal", "gsettings-desktop-schemas", "gstreamer1.0-packagekit", "gstreamer1.0-plugins-base",
		"gstreamer1.0-plugins-good", "gstreamer1.0-pulseaudio", "gvfs-backends", "gvfs-fuse", "libatk-adaptor", "libcanberra-pulse", "libglib2.0-bin",
		"libpam-gnome-keyring", "libproxy1-plugin-gsettings", "libproxy1-plugin-webkit", "nautilus", "pulseaudio", "pulseaudio-module-bluetooth",
		"sound-theme-freedesktop", "system-config-printer-common", "system-config-printer-udev", "totem", "zenity", "libproxy1-plugin-networkmanager",
		"network-manager-gnome"}}
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

//.InstallConf desribes a set of packages to be installed.
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
