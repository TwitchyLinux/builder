package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const (
	linuxVersion = "5.1.18"
	linuxURL     = "https://mirrors.edge.kernel.org/pub/linux/kernel/v5.x/linux-" + linuxVersion + ".tar.xz"
	linuxSHA256  = "6013e7dcf59d7c1b168d8edce3dbd61ce340ff289541f920dbd0958bef98f36a"
)

// Linux is a unit that builds the Linux kernel.
type Linux struct {
}

// Name implements Unit.
func (l *Linux) Name() string {
	return "Linux"
}

func (l *Linux) dirFilename() string {
	return "linux-" + linuxVersion
}

func (l *Linux) tarFilename() string {
	return l.dirFilename() + ".tar.xz"
}

func (l *Linux) tarPath(opts *Opts, inChroot bool) string {
	if inChroot {
		return "/" + l.tarFilename()
	}
	return filepath.Join(opts.Dir, l.tarFilename())
}

// Run implements Unit.
func (l *Linux) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	// TODO: Make util function for simple commands requiring no output.
	if err := DownloadFile(&opts, linuxURL, l.tarPath(&opts, false)); err != nil {
		return fmt.Errorf("Linux source download failed: %v", err)
	}
	if err := CheckSHA256(l.tarPath(&opts, false), linuxSHA256); err != nil {
		return err
	}

	if err := chroot.Shell(ctx, &opts, "tar", "xf", l.tarPath(&opts, true)); err != nil {
		return err
	}

	if err := chroot.Shell(ctx, &opts, "make", "-C", l.dirFilename(), "-j6", "mrproper"); err != nil {
		return err
	}

	d, err := ioutil.ReadFile(filepath.Join(opts.Resources, "linux", ".config"))
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, l.dirFilename(), ".config"), d, 0644); err != nil {
		return err
	}

	if err := chroot.Shell(ctx, &opts, "make", "-C", l.dirFilename(), "-j6", "clean"); err != nil {
		return err
	}
	if err := chroot.Shell(ctx, &opts, "make", "-C", l.dirFilename(), "-j6", "deb-pkg"); err != nil {
		return err
	}
	return nil
}
