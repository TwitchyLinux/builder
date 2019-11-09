package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var localeEnv = []string{
	"DEBIAN_FRONTEND=noninteractive",
	"DEBCONF_NONINTERACTIVE_SEEN=true",
}

// Chroot represents a directory configured to be used as a chroot.
type Chroot struct {
	Dir        string
	chrootPath string

	// mounts describes additional mounts that need to be torn down.
	mounts struct {
		sys  bool
		proc bool
		dev  bool
	}

	previousResolv []byte
}

// Close releases all resources associated with the chroot.
func (c *Chroot) Close() error {
	if c.previousResolv != nil {
		if err := ioutil.WriteFile(filepath.Join(c.Dir, "etc", "resolv.conf"), c.previousResolv, 0755); err != nil {
			return err
		}
		c.previousResolv = nil
	}

	if c.mounts.dev {
		if err := syscall.Unmount(filepath.Join(c.Dir, "dev"), 0); err != nil {
			return err
		}
		c.mounts.dev = false
	}
	if c.mounts.proc {
		if err := syscall.Unmount(filepath.Join(c.Dir, "proc"), 0); err != nil {
			return err
		}
		c.mounts.proc = false
	}
	if c.mounts.sys {
		if err := syscall.Unmount(filepath.Join(c.Dir, "sys"), 0); err != nil {
			return err
		}
		c.mounts.sys = false
	}

	return nil
}

// CmdContext prepares an execution within the chroot.
func (c *Chroot) CmdContext(ctx context.Context, bin string, args ...string) (*exec.Cmd, error) {
	p, err := FindBinary(bin)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", bin, err)
	}
	cmd := exec.CommandContext(ctx, c.chrootPath)
	cmd.Args = append([]string{c.chrootPath, c.Dir, p}, args...)
	return cmd, nil
}

// Shell runs a simple command within the chroot.
func (c *Chroot) Shell(ctx context.Context, opts *Opts, bin string, args ...string) error {
	cmd, err := c.CmdContext(ctx, bin, args...)
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AptInstall installs the given packages.
func (c *Chroot) AptInstall(ctx context.Context, opts *Opts, packages ...string) error {
	cmd, err := c.CmdContext(ctx, "apt-get", append([]string{"install", "-y"}, packages...)...)
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L
	cmd.Stderr = os.Stderr
	cmd.Env = localeEnv
	return cmd.Run()
}

func prepareChroot(root string) (out *Chroot, err error) {
	p, err := FindBinary("chroot")
	if err != nil {
		return nil, fmt.Errorf("could not find chroot: %v", err)
	}
	out = &Chroot{Dir: root, chrootPath: p}

	defer func() {
		if err != nil {
			out.Close()
		}
	}()

	if err = syscall.Mount("sysfs", filepath.Join(root, "sys"), "sysfs", 0, ""); err != nil {
		return nil, err
	}
	out.mounts.sys = true
	if err = syscall.Mount("proc", filepath.Join(root, "proc"), "proc", 0, ""); err != nil {
		return nil, err
	}
	out.mounts.proc = true
	if err = syscall.Mount("/dev", filepath.Join(root, "dev"), "bind", syscall.MS_BIND, ""); err != nil {
		return nil, err
	}
	out.mounts.dev = true

	prev, err := ioutil.ReadFile(filepath.Join(root, "etc", "resolv.conf"))
	if err != nil {
		return nil, err
	}
	d, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(filepath.Join(root, "etc", "resolv.conf"), d, 0755); err != nil {
		return nil, err
	}
	out.previousResolv = prev
	return out, nil
}
