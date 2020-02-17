package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/tredoe/osutil/user/crypt/sha512_crypt"
	"github.com/twitchylinux/builder/conf/user"
)

// UserSpec configures a user account.
type UserSpec struct {
	Username string
	Password string
	Groups   []string
}

// ShellCustomization is a unit which customizes the accounts + shell.
type ShellCustomization struct {
	AdditionalSkel           []byte
	AdditionalProfileScripts map[string][]byte
	Users                    []UserSpec
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
	lines := strings.Split(string(shadowData), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, user+":") {
			spl := strings.Split(line, ":")
			out.WriteString(user + ":" + pw + ":" + strings.Join(spl[2:], ":"))
		} else {
			out.WriteString(line)
		}
		if i < len(lines)-1 {
			out.WriteRune('\n')
		}
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
	c, err := user.ReadConfig(opts.Dir)
	if err != nil {
		return fmt.Errorf("reading static user configuration: %v", err)
	}

	for _, usr := range d.Users {
		if err := c.UpsertUser(usr.Username, user.OptCreateSkel()); err != nil {
			return fmt.Errorf("could not upsert user %q: %v", usr.Username, err)
		}

		if usr.Password != "" {
			s, err := user.ShadowHash(usr.Password)
			if err != nil {
				return err
			}
			if err := c.SetPassword(usr.Username, s); err != nil {
				return err
			}
		}
		for _, g := range usr.Groups {
			if err := c.UpsertMembership(usr.Username, g); err != nil {
				return err
			}
		}
	}

	if err := c.Flush(); err != nil {
		return fmt.Errorf("writing static user config: %v", err)
	}
	return nil
}

// Run implements Unit.
func (d *ShellCustomization) Run(ctx context.Context, opts Opts) error {
	if err := os.MkdirAll(filepath.Join(opts.Dir, "etc", "profile.d"), 0755); err != nil && !os.IsExist(err) {
		return err
	}
	for fname, contents := range d.AdditionalProfileScripts {
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
