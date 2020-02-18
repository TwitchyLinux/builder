package stager

import (
	"github.com/twitchylinux/builder/conf/dconf"
	"github.com/twitchylinux/builder/units"
)

// Contains hardcoded stages.
var (
	systemBuildUnits = []units.Unit{
		&units.Systemd{},
	}

	afterGUIUnits = []units.Unit{
		&units.Dconf{
			Profiles: map[string]dconf.Profile{
				"user": dconf.Profile{
					RW: dconf.Directive{
						Type: dconf.User,
						Name: "user",
					},
					ROs: []dconf.Directive{
						dconf.Directive{
							Type: dconf.System,
							Name: "local",
						},
					},
				},
				"gdm": dconf.Profile{
					RW: dconf.Directive{
						Type: dconf.User,
						Name: "user",
					},
					ROs: []dconf.Directive{
						dconf.Directive{
							Type: dconf.System,
							Name: "gdm",
						},
						dconf.Directive{
							Type: dconf.File,
							Name: "/usr/share/gdm/greeter-dconf-defaults",
						},
					},
				},
			},
			LocalLocks: map[string]dconf.Lock{
				"screensaver": dconf.Lock("/org/gnome/desktop/screensaver/idle-activation-enabled"),
				"session":     dconf.Lock("/org/gnome/desktop/session"),
				"lockdown":    dconf.Lock("/org/gnome/desktop/lockdown"),
			},
		},
	}

	finalUnits = []units.Unit{
		&units.Installer{},
		&units.Clean{},
		&units.Grub2{
			DistroName: "TwitchyLinux",
			Quiet:      true,
			ColorNormal: units.GrubColorPair{
				FG: "white",
				BG: "black",
			},
			ColorHighlight: units.GrubColorPair{
				FG: "black",
				BG: "light-gray",
			},
		},
	}
)
