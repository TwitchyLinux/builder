package units

import (
	"context"
	"fmt"
	"io"

	"github.com/twitchylinux/builder/conf/dconf"
)

// DebianOpts configures the debian URL and track the system will
// be based on.
type DebianOpts struct {
	URL   string
	Track string
}

// Opts describes options provided to the units.
type Opts struct {
	// Dir represents the path the system is being built at.
	Dir string
	// Resources is the path to the builder resources directory.
	Resources string

	// Num indicates which unit (in execution order) the unit is.
	Num int
	// L is a logger which units can use to communicate state.
	L Logger

	// NumThreads is the number of concurrent threads to be used while building.
	NumThreads int

	Debian DebianOpts
}

func (o *Opts) makeNumThreadsArg() string {
	return fmt.Sprintf("-j%d", o.NumThreads)
}

// Logger implements status reporting and logging for executing units.
type Logger interface {
	io.Writer
}

// Unit describes an execution unit for building the system.
type Unit interface {
	Name() string
	Run(ctx context.Context, opts Opts) error
}

// Units contains the ordered set of all units needed to build the
// target system.
// TODO: Switch to method?
var Units = []Unit{
	&Preflight{},
	&Debootstrap{},
	&FinalizeApt{},
	&Locale{
		Area:     "America",
		Zone:     "Los_Angeles",
		Generate: []string{"en_US.UTF-8 UTF-8", "en_US ISO-8859-1"},
		Default:  "en_US.UTF-8",
	},
	&BaseBuildtools{},
	&Linux{},
	&Systemd{},
	&ShellCustomization{
		AdditionalSkel:          additionalSkel,
		AddtionalProfileScripts: profiledScripts,
		Users:                   defaultUsers,
	},
	fsToolsInstall,
	netToolsInstall,
	compressionToolsInstall,
	cliToolsInstall,
	usbInstall,
	cToolchainInstall,
	wifiInstall,
	&InstallTools{name: "sudo", pkgs: []string{"sudo"}},
	&InstallTools{name: "bash-completion", pkgs: []string{"bash-completion", "bash-doc", "bash-builtins"}},
	&InstallTools{name: "cryptsetup", pkgs: []string{"cryptsetup", "kbd", "console-setup", "keyutils"}},
	&InstallTools{name: "firmware", pkgs: []string{"firmware-iwlwifi", "firmware-amd-graphics", "firmware-atheros", "firmware-brcm80211",
		"firmware-cavium", "firmware-intel-sound", "intel-microcode", "firmware-misc-nonfree", "firmware-realtek", "firmware-zd1211"}},
	&Gnome{
		NeedPkgs: []string{"gnome"},
	},
	&Dconf{
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
	nmapToolsInstall,
	&InstallTools{name: "gui-dev-tools", pkgs: []string{"gpick", "glade", "mesa-utils", "libgtk-3-dev", "libcairo2-dev", "libglib2.0-dev"}},
	&Clean{},
	&Grub2{
		DistroName: "TwitchyLinux",
		Quiet:      true,
		ColorNormal: grubColorPair{
			FG: "white",
			BG: "black",
		},
		ColorHighlight: grubColorPair{
			FG: "black",
			BG: "light-gray",
		},
	},
}
