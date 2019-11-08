package units

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

const tzData = `
tzdata tzdata/Areas select America
tzdata tzdata/Zones/Europe select Los_Angeles
`

type Locale struct {
}

func (d *Locale) Name() string {
	return "Locale"
}

func (d *Locale) Run(ctx context.Context, opts Opts) error {
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "tz-data"), []byte(tzData), 0644); err != nil {
		return err
	}
	defer os.Remove(filepath.Join(opts.Dir, "tz-data"))

	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	cmd, err := chroot.CmdContext(ctx, "debconf-set-selections", "/tz-data")
	if err != nil {
		return err
	}
	cmd.Env = []string{
		"DEBIAN_FRONTEND=noninteractive",
		"DEBCONF_NONINTERACTIVE_SEEN=true",
	}
	cmd.Stdout = opts.L
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
