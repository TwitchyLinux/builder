package units

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/twitchylinux/builder/conf/dconf"
)

// Dconf sets up dconf configuration.
type Dconf struct {
	Profiles   map[string]dconf.Profile
	LocalLocks map[string]dconf.Lock
}

// Name implements Unit.
func (d *Dconf) Name() string {
	return "Dconf"
}

// Run implements Unit.
func (d *Dconf) Run(ctx context.Context, opts Opts) error {
	for _, subdir := range []string{"profile", "db/gdm.d", "db/local.d", "db/local.d/locks"} {
		if err := os.MkdirAll(filepath.Join(opts.Dir, "etc", "dconf", subdir), 0755); err != nil {
			return fmt.Errorf("creating directory %q: %v", subdir, err)
		}
	}

	// Install profile configuration.
	for name, profile := range d.Profiles {
		if err := ioutil.WriteFile(filepath.Join(opts.Dir, dconf.ProfileDir, name), profile.Generate(), 0644); err != nil {
			return err
		}
	}

	// Write out textual configuration from resources.
	if err := InstallConfigResources(ctx, &opts, filepath.Join(dconf.DBDir, "gdm.d"), "gdm-dconf"); err != nil {
		return err
	}
	if err := InstallConfigResources(ctx, &opts, filepath.Join(dconf.DBDir, "local.d"), "local-dconf"); err != nil {
		return err
	}

	// Configure locks on certain paths so users cannot change them.
	for name, lock := range d.LocalLocks {
		if err := ioutil.WriteFile(filepath.Join(opts.Dir, dconf.DBDir, "local.d", dconf.LocksDir, name), lock.Generate(), 0644); err != nil {
			return err
		}
	}

	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := d.setupDefaultTerminal(ctx, chroot, &opts); err != nil {
		return err
	}
	return chroot.Shell(ctx, &opts, "dconf", "update")
}

func (d *Dconf) setupDefaultTerminal(ctx context.Context, chroot *Chroot, opts *Opts) error {
	cmd, err := chroot.CmdContext(ctx, opts, "gsettings", "get", "org.gnome.Terminal.ProfilesList", "default")
	if err != nil {
		return err
	}
	cmd.Stderr = opts.L.Stderr()
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	profile := strings.Replace(strings.Replace(strings.Replace(string(out), "\n", "", -1), "\"", "", -1), "'", "", -1)
	var buff bytes.Buffer
	buff.WriteString("[org/gnome/terminal/legacy/profiles:/:" + profile + "]\n")
	buff.WriteString("use-theme-colors=false\n")
	buff.WriteString("audible-bell=false\n")
	buff.WriteString("allow-bold=true\n")
	buff.WriteString("scrollback-unlimited=false\n")
	buff.WriteString("font='Monospace 12'\n")
	buff.WriteString("foreground-color='rgb(247,247,247)'\n")
	buff.WriteString("palette=['rgb(46,52,54)', 'rgb(204,0,0)', 'rgb(78,154,6)', 'rgb(196,160,0)', 'rgb(52,101,164)', 'rgb(117,80,123)', 'rgb(6,152,154)', 'rgb(211,215,207)', 'rgb(85,87,83)', 'rgb(239,41,41)', 'rgb(138,226,52)', 'rgb(252,233,79)', 'rgb(114,159,207)', 'rgb(173,127,168)', 'rgb(54,226,226)', 'rgb(238,238,236)']\n")
	buff.WriteString("cursor-foreground-color='rgb(255,255,255)'\n")
	buff.WriteString("background-color='rgb(47,47,47)'\n")
	buff.WriteString("highlight-foreground-color='rgb(255,255,255)'\n")
	buff.WriteString("cursor-background-color='rgb(0,0,0)'\n")
	buff.WriteString("highlight-background-color='rgb(255,255,255)'\n")
	buff.WriteString("bold-color='rgb(0,0,0)'\n")
	return ioutil.WriteFile(filepath.Join(opts.Dir, dconf.DBDir, "local.d", "09-default-terminal"), buff.Bytes(), 0644)
}
