package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Gnome installs the GNOME graphical environment
type Gnome struct {
}

// Name implements Unit.
func (d *Gnome) Name() string {
	return "Gnome"
}

// Run implements Unit.
func (d *Gnome) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := chroot.AptInstall(ctx, &opts, "gnome"); err != nil {
		return err
	}
	if err := d.setupDConf(ctx, &opts); err != nil {
		return err
	}
	if err := Shell(ctx, &opts, "cp", filepath.Join(opts.Resources, "twitchy_background.png"), filepath.Join(opts.Dir, "usr/share/backgrounds/twitchy_background.png")); err != nil {
		return err
	}

	return chroot.Shell(ctx, &opts, "dconf", "update")
}

func (d *Gnome) writeTextConfig(base, resourceDir string) error {
	files, err := ioutil.ReadDir(resourceDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		d, err := ioutil.ReadFile(filepath.Join(resourceDir, f.Name()))
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(base, f.Name()), d, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (d *Gnome) setupDConf(ctx context.Context, opts *Opts) error {
	for _, subdir := range []string{"profile", "db/gdm.d", "db/local.d", "db/local.d/locks"} {
		if err := os.MkdirAll(filepath.Join(opts.Dir, "etc", "dconf", subdir), 0755); err != nil {
			return fmt.Errorf("creating directory %q: %v", subdir, err)
		}
	}

	// Configure user profile defaults.
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "dconf", "profile", "user"), []byte("user-db:user\nsystem-db:local\n"), 0644); err != nil {
		return err
	}
	// Configure GDM profile defaults.
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "dconf", "profile", "gdm"), []byte("user-db:user\nsystem-db:gdm\nfile-db:/usr/share/gdm/greeter-dconf-defaults\n"), 0644); err != nil {
		return err
	}

	// Write out textual configuration from resources.
	if err := d.writeTextConfig(filepath.Join(opts.Dir, "etc", "dconf", "db", "gdm.d"), filepath.Join(opts.Resources, "gdm-dconf")); err != nil {
		return err
	}
	if err := d.writeTextConfig(filepath.Join(opts.Dir, "etc", "dconf", "db", "local.d"), filepath.Join(opts.Resources, "local-dconf")); err != nil {
		return err
	}

	// Configure ?locks?
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "dconf", "db", "local.d", "locks", "screensaver"), []byte("/org/gnome/desktop/screensaver/idle-activation-enabled\n"), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "dconf", "db", "local.d", "locks", "session"), []byte("/org/gnome/desktop/session\n"), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "dconf", "db", "local.d", "locks", "lockdown"), []byte("/org/gnome/desktop/lockdown\n"), 0644); err != nil {
		return err
	}

	return nil
}
