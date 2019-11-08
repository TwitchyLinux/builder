package units

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Systemd struct {
}

func (s *Systemd) Name() string {
	return "Systemd"
}

func (s *Systemd) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := chroot.AptInstall(ctx, &opts, "systemd", "systemd-sysv"); err != nil {
		return err
	}

	return s.disableScreenClearing(ctx, &opts)
}

func (s *Systemd) disableScreenClearing(ctx context.Context, opts *Opts) error {
	gettyConfDir := filepath.Join(opts.Dir, "etc", "systemd", "system", "getty@tty1.service.d")
	if err := os.MkdirAll(gettyConfDir, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(gettyConfDir, "noclear.conf"), []byte(`
[Service]
TTYVTDisallocate=no
`), 0644)
}
