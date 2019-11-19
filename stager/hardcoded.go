package stager

import (
	"github.com/twitchylinux/builder/conf/dconf"
	"github.com/twitchylinux/builder/units"
)

// Contains hardcoded stages.
var (
	// baseSystemUnits are run first.
	earlyBuildUnits = []units.Unit{
		&units.Preflight{},
		&units.Debootstrap{},
		&units.FinalizeApt{},
	}
	systemBuildUnits = []units.Unit{
		&units.BaseBuildtools{},
		&units.Linux{},
		&units.Systemd{},
		&units.ShellCustomization{
			AdditionalSkel:          additionalSkel,
			AddtionalProfileScripts: profiledScripts,
			Users:                   defaultUsers,
		},
		&units.Golang{},
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

// Hardcoded shell customization.
var (
	additionalSkel = []byte(`

  # Start TwitchyLinux section
  alias ls='ls --color=auto'
  alias grep='grep --color=auto'
  export GCC_COLORS='error=01;31:warning=01;35:note=01;36:caret=01;32:locus=01:quote=01'
  export PS1="\[\033[38;5;2m\][\u\[$(tput sgr0)\]\[\033[38;5;1m\]@\[$(tput sgr0)\]\[\033[38;5;2m\]\h]\[$(tput sgr0)\]\[\033[38;5;15m\]:\[$(tput bold)\]\[\033[38;5;6m\]\w\[$(tput sgr0)\]\[\033[38;5;15m\]> \[$(tput sgr0)\]"
  if [ "$UID" -eq "0" ]; then
    export PS1="\[\033[38;5;2m\][\[$(tput sgr0)\]\[\033[38;5;11m\]\u\[$(tput sgr0)\]\[\033[38;5;1m\]@\[$(tput sgr0)\]\[\033[38;5;2m\]\h]\[$(tput sgr0)\]\[\033[38;5;15m\]:\[$(tput bold)\]\[\033[38;5;6m\]\w\[$(tput sgr0)\]\[\033[38;5;15m\]> \[$(tput sgr0)\]"
  fi
  alias edit='nano'
  # End TwitchyLinux section
  `)

	profiledScripts = map[string][]byte{
		"twl.sh": []byte(`
  export LANG=en_US.UTF-8
  # Setup for /bin/ls and /bin/grep to support color.
  if [ -f "/etc/dircolors" ] ; then
          eval $(dircolors -b /etc/dircolors)
  fi
  if [ -f "$HOME/.dircolors" ] ; then
          eval $(dircolors -b $HOME/.dircolors)
  fi
  alias ls='ls --color=auto'
  alias grep='grep --color=auto'
  #colored GCC stuff
  export GCC_COLORS='error=01;31:warning=01;35:note=01;36:caret=01;32:locus=01:quote=01'
  export PS1="\[\033[38;5;2m\][\u\[$(tput sgr0)\]\[\033[38;5;1m\]@\[$(tput sgr0)\]\[\033[38;5;2m\]\h]\[$(tput sgr0)\]\[\033[38;5;15m\]:\[$(tput bold)\]\[\033[38;5;6m\]\w\[$(tput sgr0)\]\[\033[38;5;15m\]> \[$(tput sgr0)\]"
  if [ "$UID" -eq "0" ]; then
    export PS1="\[\033[38;5;2m\][\[$(tput sgr0)\]\[\033[38;5;11m\]\u\[$(tput sgr0)\]\[\033[38;5;1m\]@\[$(tput sgr0)\]\[\033[38;5;2m\]\h]\[$(tput sgr0)\]\[\033[38;5;15m\]:\[$(tput bold)\]\[\033[38;5;6m\]\w\[$(tput sgr0)\]\[\033[38;5;15m\]> \[$(tput sgr0)\]"
  fi
  alias edit='nano'
  alias reload='. ~/.bashrc'
  `),
	}

	defaultUsers = []units.UserSpec{
		{
			Username: "twl",
			Password: "twl",
			Groups:   []string{"sudo", "systemd-journal", "netdev"},
		},
	}
)
