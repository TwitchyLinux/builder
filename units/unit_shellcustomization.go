package units

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tredoe/osutil/user/crypt"
	_ "github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

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

	defaultUsers = []userSpec{
		{
			Username: "twl",
			Password: "twl",
			Groups:   []string{"sudo", "systemd-journal", "netdev"},
		},
	}
)

type userSpec struct {
	Username string
	Password string
	Groups   []string
}

// ShellCustomization is a unit which customizes the accounts + shell.
type ShellCustomization struct {
	AdditionalSkel          []byte
	AddtionalProfileScripts map[string][]byte
	Users                   []userSpec
}

// Name implements Unit.
func (d *ShellCustomization) Name() string {
	return "Shell-customization"
}

func (d *ShellCustomization) updateShadowPassword(dir, user, pw string) error {
	shadowData, err := ioutil.ReadFile(filepath.Join(dir, "etc", "shadow"))
	if err != nil {
		return err
	}
	var out strings.Builder
	for _, line := range strings.Split(string(shadowData), "\n") {
		if strings.HasPrefix(line, user+":") {
			spl := strings.Split(line, ":")
			out.WriteString(user + ":" + pw + strings.Join(spl[2:], ":"))
		} else {
			out.WriteString(line)
		}
		out.WriteRune('\n')
	}

	return ioutil.WriteFile(filepath.Join(dir, "etc", "shadow"), []byte(out.String()), 0640)
}

func (d *ShellCustomization) makeUser(ctx context.Context, opts *Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := chroot.AptInstall(ctx, opts, "passwd"); err != nil {
		return err
	}

	for _, usr := range d.Users {
		if err := chroot.Shell(ctx, opts, "adduser", "--disabled-password", "--gecos", "", usr.Username); err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				return err
			}
		}

		if usr.Password != "" {
			s, err := crypt.New(crypt.SHA512).Generate([]byte(usr.Password), nil)
			if err != nil {
				return err
			}
			if err := d.updateShadowPassword(opts.Dir, usr.Username, s); err != nil {
				return err
			}
		}

		for _, g := range usr.Groups {
			if err := chroot.Shell(ctx, opts, "usermod", "-a", "-G", g, usr.Username); err != nil {
				return err
			}
		}
	}
	return nil
}

// Run implements Unit.
func (d *ShellCustomization) Run(ctx context.Context, opts Opts) error {
	if err := os.MkdirAll(filepath.Join(opts.Dir, "etc", "profile.d"), 0755); err != nil && !os.IsExist(err) {
		return err
	}
	for fname, contents := range d.AddtionalProfileScripts {
		if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "profile.d", fname), contents, 0644); err != nil {
			return err
		}
	}

	skel, err := ioutil.ReadFile(filepath.Join(opts.Dir, "etc", "skel", ".bashrc"))
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "skel", ".bashrc"), append(skel, d.AdditionalSkel...), 0644); err != nil {
		return err
	}

	return d.makeUser(ctx, &opts)
}
