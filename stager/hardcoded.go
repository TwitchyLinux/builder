package stager

import (
	"github.com/twitchylinux/builder/units"
)

// Contains hardcoded stages.
var (
	systemBuildUnits = []units.Unit{
		&units.Systemd{},
	}

	defaultFeatures = map[string]bool{
		"graphical":            true,
		"features.rootfs-only": true,
	}

	afterGUIUnits = []units.Unit{}

	finalUnits = []units.Unit{
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
