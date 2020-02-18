package stager

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/units"
)

var (
	graphicalEnvDefault = GraphicsConf{Packages: []string{"gnome"}}

	debootstrapDefault = DebootstrapConf{
		Track: "stable",
		URL:   "http://deb.debian.org/debian/",
	}

	localeDefault = LocaleConf{
		Area:     "America",
		Zone:     "Los_Angeles",
		Generate: []string{"en_US.UTF-8 UTF-8", "en_US ISO-8859-1"},
		Default:  "en_US.UTF-8",
	}

	linuxDefault = LinuxConf{
		Version: "5.1.18",
		URL:     "https://mirrors.edge.kernel.org/pub/linux/kernel/v5.x/linux-5.1.18.tar.xz",
		SHA256:  "6013e7dcf59d7c1b168d8edce3dbd61ce340ff289541f920dbd0958bef98f36a",
	}
)

// DebootstrapConf describes what to tell debootstrap.
type DebootstrapConf struct {
	Track string `toml:"track"`
	URL   string `toml:"url"`
}

func debootstrapConf(tree *toml.Tree) (*units.Debootstrap, error) {
	conf := debootstrapDefault
	if t := tree.Get(keyDebian); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyDebian, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Debootstrap{
		Track: conf.Track,
		URL:   conf.URL,
	}, nil
}

// LinuxConf describes what Linux kernel to install
type LinuxConf struct {
	Version      string   `toml:"version"`
	URL          string   `toml:"url"`
	SHA256       string   `toml:"sha256"`
	BuildDepPkgs []string `toml:"build_packages"`
}

func linuxConf(tree *toml.Tree) (*units.Linux, error) {
	conf := linuxDefault
	if t := tree.Get(keyLinux); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyLinux, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Linux{
		Version:      conf.Version,
		URL:          conf.URL,
		SHA256:       conf.SHA256,
		BuildDepPkgs: conf.BuildDepPkgs,
	}, nil
}

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

// LocaleConf describes the locale of the system.
type LocaleConf struct {
	Area     string   `toml:"area"`
	Zone     string   `toml:"zone"`
	Generate []string `toml:"generate_locales"`
	Default  string   `toml:"default"`
}

func localeConf(tree *toml.Tree) (*units.Locale, error) {
	conf := localeDefault
	if t := tree.Get(keyLocale); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyLocale, t)
		}
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}
	}

	return &units.Locale{
		Area:     conf.Area,
		Zone:     conf.Zone,
		Generate: conf.Generate,
		Default:  conf.Default,
	}, nil
}

// ReleaseConf describes how the system should self-describe.
type ReleaseConf struct {
	Name       string `toml:"name"`
	PrettyName string `toml:"pretty_name"`
	ID         string `toml:"id"`
	URL        string `toml:"url"`
	Issue      string `toml:"issue"`
}

var osReleaseTmpl = `PRETTY_NAME="{{.PrettyName}}"
NAME="{{.Name}}"
ID={{.ID}}
HOME_URL="{{.URL}}"
`

func releaseConf(tree *toml.Tree) ([]units.Unit, error) {
	var out []units.Unit

	if t := tree.Get(keyReleaseInfo); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyReleaseInfo, t)
		}
		var conf ReleaseConf
		if err := ge.Unmarshal(&conf); err != nil {
			return nil, err
		}

		tmpl, err := template.New("").Parse(osReleaseTmpl)
		if err != nil {
			return nil, err
		}
		var osRelease bytes.Buffer
		if err := tmpl.Execute(&osRelease, conf); err != nil {
			return nil, err
		}
		out = append(out, &units.InstallFiles{
			UnitName: "release-info",
			Files: []units.FileInfo{
				{
					Path: "/etc/os-release",
					Data: osRelease.Bytes(),
				},
				{
					Path: "/etc/issue",
					Data: []byte(conf.Issue + "\n"),
				},
			},
		})

		return out, nil
	}

	return nil, nil
}

// ShellProfile describes a profile to be installed in /etc/profile.d.
type ShellProfile struct {
	Name   string `toml:"name"`
	Script string `toml:"script"`
}

// ShellConf describes the customization of the shell.
type ShellConf struct {
	Skel     string         `toml:"skel"`
	Profiles []ShellProfile `toml:"profile"`
}

type MainUserConf struct {
	Name            string   `toml:"name"`
	DefaultPassword string   `toml:"default_password"`
	Groups          []string `toml:"groups"`
}

func shellUserConf(tree *toml.Tree) (*units.ShellCustomization, error) {
	shellConf := ShellConf{}
	if t := tree.Get(keyShellCust); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyShellCust, t)
		}
		if err := ge.Unmarshal(&shellConf); err != nil {
			return nil, err
		}
	}

	userConf := MainUserConf{}
	if t := tree.Get(keyMainUser); t != nil {
		ge, ok := t.(*toml.Tree)
		if !ok {
			if i, isInt := t.(int64); isInt && i == 0 {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyMainUser, t)
		}
		if err := ge.Unmarshal(&userConf); err != nil {
			return nil, err
		}
	}

	out := &units.ShellCustomization{
		AdditionalSkel:           []byte(shellConf.Skel),
		AdditionalProfileScripts: map[string][]byte{},
		Users: []units.UserSpec{
			{
				Username: userConf.Name,
				Password: userConf.DefaultPassword,
				Groups:   userConf.Groups,
			},
		},
	}
	for _, p := range shellConf.Profiles {
		out.AdditionalProfileScripts[p.Name] = []byte(p.Script)
	}

	return out, nil
}
