package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	return chroot.Shell(ctx, &opts, "dconf", "update")
}
