package units

import (
	"context"
)

// BaseBuildtools is a unit which installs build dependencies for
// the Linux kernel.
type BaseBuildtools struct {
}

// Name implements Unit.
func (d *BaseBuildtools) Name() string {
	return "Base-buildtools"
}

// Run implements Unit.
func (d *BaseBuildtools) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	cmd, err := chroot.CmdContext(ctx, "apt-get", "install", "-y", "build-essential", "fakeroot", "devscripts", "wget", "libncurses-dev")
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = chroot.CmdContext(ctx, "apt-get", "-y", "build-dep", "linux")
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	return cmd.Run()
}
