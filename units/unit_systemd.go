package units

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Systemd is a unit which installs systemd and sets up basic configuration.
type Systemd struct {
}

// Name implements Unit.
func (s *Systemd) Name() string {
	return "Systemd"
}

// Run implements Unit.
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
