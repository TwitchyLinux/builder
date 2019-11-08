package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ss "github.com/twitchylinux/builder/shellstr"
)

const (
	linuxVersion = "5.1.18"
	linuxURL     = "https://mirrors.edge.kernel.org/pub/linux/kernel/v5.x/linux-" + linuxVersion + ".tar.xz"
	linuxMD5     = "599391aef003a22abc1d3b7ba3758183"
)

// Linux is a unit that builds the Linux kernel.
type Linux struct {
}

// Name implements Unit.
func (l *Linux) Name() string {
	return "Linux"
}

// Run implements Unit.
func (l *Linux) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	// TODO: Make util function for simple commands requiring no output.
	download, err := chroot.CmdContext(ctx, "wget", "-q", "-O", "/linux-"+linuxVersion+".tar.xz", linuxURL)
	if err != nil {
		return err
	}
	download.Stdout = opts.L
	download.Stderr = os.Stderr
	if err := download.Run(); err != nil {
		return err
	}

	// TODO: Make pure-go util for computing/comparing the hash.
	hsh, err := chroot.CmdContext(ctx, "md5sum", "/linux-"+linuxVersion+".tar.xz")
	if err != nil {
		return err
	}
	out, err := hsh.Output()
	if err != nil {
		return err
	}
	s := strings.TrimSpace(ss.Trim(string(out), &ss.Cut{Delim: " ", From: 1, To: 1}))
	if s != linuxMD5 {
		return fmt.Errorf("MD5 mismatch: %q != %q", s, linuxMD5)
	}

	xtract, err := chroot.CmdContext(ctx, "tar", "xf", "/linux-"+linuxVersion+".tar.xz")
	if err != nil {
		return err
	}
	xtract.Stdout = opts.L
	xtract.Stderr = os.Stderr
	if err := xtract.Run(); err != nil {
		return err
	}

	s1, err := chroot.CmdContext(ctx, "make", "-C", "linux-"+linuxVersion, "-j6", "mrproper")
	if err != nil {
		return err
	}
	s1.Stdout = opts.L
	s1.Stderr = os.Stderr
	if err := s1.Run(); err != nil {
		return err
	}

	d, err := ioutil.ReadFile(filepath.Join(opts.Resources, "linux", ".config"))
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "linux-"+linuxVersion, ".config"), d, 0644); err != nil {
		return err
	}

	s2, err := chroot.CmdContext(ctx, "make", "-C", "linux-"+linuxVersion, "-j6", "clean")
	if err != nil {
		return err
	}
	s2.Stdout = opts.L
	s2.Stderr = os.Stderr
	if err := s2.Run(); err != nil {
		return err
	}

	s3, err := chroot.CmdContext(ctx, "make", "-C", "linux-"+linuxVersion, "-j6", "deb-pkg")
	if err != nil {
		return err
	}
	s3.Stdout = opts.L
	s3.Stderr = os.Stderr
	if err := s3.Run(); err != nil {
		return err
	}
	return nil
}
