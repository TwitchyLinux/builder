package units

import (
	"context"
)

// InstallTools is a unit which installs packages from apt.
type InstallTools struct {
	UnitName string
	Pkgs     []string

	Order int
}

// Name implements Unit.
func (i *InstallTools) Name() string {
	return i.UnitName
}

// Run implements Unit.
func (i *InstallTools) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	return chroot.AptInstall(ctx, &opts, i.Pkgs...)
}
