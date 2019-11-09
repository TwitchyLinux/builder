package units

import (
	"context"
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

	return chroot.Shell(ctx, &opts, "apt-get", "clean")
}
