package units

import (
	"context"
	"os"
	"path/filepath"
)

// Gnome installs the GNOME graphical environment
type Gnome struct {
	NeedPkgs []string
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

	if err := os.MkdirAll(filepath.Join(opts.Dir, "usr/share/backgrounds"), 0755); err != nil {
		return err
	}
	if err := CopyResource(ctx, &opts, "twitchy_background.png", "usr/share/backgrounds/twitchy_background.png"); err != nil {
		return err
	}

	return chroot.AptInstall(ctx, &opts, d.NeedPkgs...)
}
