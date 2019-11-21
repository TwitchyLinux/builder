package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Linux is a unit that builds the Linux kernel.
type Linux struct {
	Version      string
	URL          string
	SHA256       string
	BuildDepPkgs []string
}

// Name implements Unit.
func (l *Linux) Name() string {
	return "Linux"
}

func (l *Linux) dirFilename() string {
	return "linux-" + l.Version
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

	opts.L.SetSubstage("Installing build dependencies")
	if err := l.installDeps(ctx, chroot, opts); err != nil {
		return err
	}

	opts.L.SetSubstage("Downloading Linux " + l.Version)
	if err := DownloadFile(&opts, l.URL, l.tarPath(&opts, false)); err != nil {
		return fmt.Errorf("Linux source download failed: %v", err)
	}
	if err := CheckSHA256(l.tarPath(&opts, false), l.SHA256); err != nil {
		return err
	}

	opts.L.SetSubstage("Extracting")
	if err := chroot.Shell(ctx, &opts, "tar", "xf", l.tarPath(&opts, true)); err != nil {
		return err
	}

	opts.L.SetSubstage("Preparing")
	if err := chroot.Shell(ctx, &opts, "make", "-C", l.dirFilename(), opts.makeNumThreadsArg(), "mrproper"); err != nil {
		return err
	}

	d, err := ioutil.ReadFile(filepath.Join(opts.Resources, "linux", ".config"))
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, l.dirFilename(), ".config"), d, 0644); err != nil {
		return err
	}

	if err := chroot.Shell(ctx, &opts, "make", "-C", l.dirFilename(), opts.makeNumThreadsArg(), "clean"); err != nil {
		return err
	}
	opts.L.SetSubstage("Building")
	if err := chroot.Shell(ctx, &opts, "make", "-C", l.dirFilename(), opts.makeNumThreadsArg(), "deb-pkg"); err != nil {
		return err
	}

	return l.runInstallLinux(ctx, chroot, opts)
}

func (l *Linux) installDeps(ctx context.Context, chroot *Chroot, opts Opts) error {
	cmd, err := chroot.CmdContext(ctx, &opts, "apt-get", append([]string{"install", "-y"}, l.BuildDepPkgs...)...)
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = chroot.CmdContext(ctx, &opts, "apt-get", "-y", "build-dep", "linux")
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	return cmd.Run()
}

func (l *Linux) runInstallLinux(ctx context.Context, chroot *Chroot, opts Opts) error {
	if err := os.Mkdir(filepath.Join(opts.Dir, "var", "tmp"), 0777); err != nil && !os.IsExist(err) {
		return err
	}
	files, err := ioutil.ReadDir(opts.Dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		for _, wantPkg := range []string{"linux-headers-", "linux-image-"} {
			if strings.Contains(f.Name(), wantPkg) && strings.HasSuffix(f.Name(), ".deb") {
				opts.L.SetSubstage("Install " + wantPkg + l.Version)
				if err := chroot.Shell(ctx, &opts, "dpkg", "--install", f.Name()); err != nil {
					return err
				}
			}
		}
	}

	opts.L.SetSubstage("Install initramfs")
	if err := chroot.AptInstall(ctx, &opts, "initramfs-tools"); err != nil {
		return err
	}
	return chroot.Shell(ctx, &opts, "update-initramfs", "-c", "-k", l.Version)
}
