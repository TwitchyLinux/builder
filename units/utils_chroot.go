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

	env map[string]string

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
		if err := unmount(filepath.Join(c.Dir, "dev")); err != nil {
			return err
		}
		c.mounts.dev = false
	}
	if c.mounts.proc {
		miscPath := filepath.Join(c.Dir, "proc", "sys", "fs", "binfmt_misc")
		if _, err := os.Stat(miscPath); err == nil {
			if err := unmount(miscPath); err != nil {
				return err
			}
		}
		if err := unmount(filepath.Join(c.Dir, "proc")); err != nil {
			return err
		}
		c.mounts.proc = false
	}
	if c.mounts.sys {
		if err := unmount(filepath.Join(c.Dir, "sys")); err != nil {
			return err
		}
		c.mounts.sys = false
	}

	return nil
}

// CmdContext prepares an execution within the chroot.
func (c *Chroot) CmdContext(ctx context.Context, opts *Opts, bin string, args ...string) (*exec.Cmd, error) {
	var p string
	if _, err := os.Stat(filepath.Join(opts.Dir, bin)); err == nil {
		p = bin
	} else {
		if p, err = FindBinary(bin); err != nil {
			return nil, fmt.Errorf("%s: %v", bin, err)
		}
	}

	cmd := exec.CommandContext(ctx, c.chrootPath)
	cmd.Args = append([]string{c.chrootPath, c.Dir, p}, args...)
	return cmd, nil
}

// Shell runs a simple command within the chroot.
func (c *Chroot) Shell(ctx context.Context, opts *Opts, bin string, args ...string) error {
	cmd, err := c.CmdContext(ctx, opts, bin, args...)
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	return cmd.Run()
}

// AptInstall installs the given packages.
func (c *Chroot) AptInstall(ctx context.Context, opts *Opts, packages ...string) error {
	cmd, err := c.CmdContext(ctx, opts, "apt-get", append([]string{"install", "-y"}, packages...)...)
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	cmd.Env = localeEnv
	if opts.DebProxy != "" {
		cmd.Env = append(cmd.Env, "http_proxy=http://"+opts.DebProxy)
	}
	return cmd.Run()
}

func prepareChroot(root string) (out *Chroot, err error) {
	p, err := FindBinary("chroot")
	if err != nil {
		return nil, fmt.Errorf("could not find chroot: %v", err)
	}
	out = &Chroot{Dir: root, chrootPath: p}

	defer func(out *Chroot) {
		if err != nil {
			out.Close()
		}
	}(out)

	if mp, err := mountpointType(filepath.Join(root, "sys")); err != nil || mp != "sysfs" {
		if err = syscall.Mount("sysfs", filepath.Join(root, "sys"), "sysfs", 0, ""); err != nil {
			return nil, fmt.Errorf("mounting sysfs: %v", err)
		}
	}
	out.mounts.sys = true
	if mp, err := mountpointType(filepath.Join(root, "proc")); err != nil || mp != "proc" {

		if err = syscall.Mount("proc", filepath.Join(root, "proc"), "proc", 0, ""); err != nil {
			return nil, fmt.Errorf("mounting proc: %v", err)
		}
	}
	out.mounts.proc = true
	if err = syscall.Mount("/dev", filepath.Join(root, "dev"), "bind", syscall.MS_BIND, ""); err != nil {
		return nil, fmt.Errorf("bind-mounting dev: %v", err)
	}
	out.mounts.dev = true

	prev, err := ioutil.ReadFile(filepath.Join(root, "etc", "resolv.conf"))
	if err != nil {
		return nil, fmt.Errorf("reading initial resolv.conf: %v", err)
	}
	d, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("reading system resolv.conf: %v", err)
	}
	if err = ioutil.WriteFile(filepath.Join(root, "etc", "resolv.conf"), d, 0755); err != nil {
		return nil, fmt.Errorf("writing resolv.conf: %v", err)
	}
	out.previousResolv = prev
	return out, nil
}
