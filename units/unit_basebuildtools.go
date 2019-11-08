package units

import (
	"context"
	"os"
)

type BaseBuildtools struct {
}

func (d *BaseBuildtools) Name() string {
	return "Base-buildtools"
}

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
	cmd.Stdout = opts.L
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = chroot.CmdContext(ctx, "apt-get", "-y", "build-dep", "linux")
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
