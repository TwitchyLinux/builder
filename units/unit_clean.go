package units

import (
	"context"
	"os"
	"path/filepath"
)

// Clean is a unit which cleans up unneeded files from the system.
type Clean struct {
}

// Name implements Unit.
func (i *Clean) Name() string {
	return "Clean"
}

// Run implements Unit.
func (i *Clean) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := os.Mkdir(filepath.Join(opts.Dir, "deb-pkgs"), 0755); err != nil && !os.IsNotExist(err) && !os.IsExist(err) {
		return err
	}

	if err := chroot.Shell(ctx, &opts, "bash", "-c", "mv -v /*.deb /deb-pkgs"); err != nil {
		return err
	}
	if err := chroot.Shell(ctx, &opts, "bash", "-c", "rm -rf /linux-*"); err != nil {
		return err
	}
	return chroot.Shell(ctx, &opts, "apt-get", "clean")
}
