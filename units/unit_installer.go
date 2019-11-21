package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Installer is a unit which installs the graphical installer.
type Installer struct {
}

// Name implements Unit.
func (i *Installer) Name() string {
	return "graphical-installer"
}

// Run implements Unit.
func (i *Installer) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return fmt.Errorf("failed to initialize chroot: %v", err)
	}
	defer chroot.Close()

	if err := os.MkdirAll(filepath.Join(opts.Dir, "usr", "share", "twlinst"), 0755); err != nil {
		return fmt.Errorf("mkdir %q failed: %v", "/usr/share/twlinst", err)
	}
	os.RemoveAll(filepath.Join(opts.Dir, "tmp-twlinst-build"))
	defer os.RemoveAll(filepath.Join(opts.Dir, "tmp-twlinst-build"))

	opts.L.SetSubstage("Download")
	if err := chroot.Shell(ctx, &opts, "git", "clone", "https://github.com/TwitchyLinux/graphical-installer", "/tmp-twlinst-build"); err != nil {
		return fmt.Errorf("cloning installer: %v", err)
	}

	if err := i.build(ctx, &opts, chroot); err != nil {
		return err
	}
	if err := i.copyResources(ctx, &opts); err != nil {
		return err
	}
	return i.installVersion(ctx, &opts)
}

func (i *Installer) build(ctx context.Context, opts *Opts, chroot *Chroot) error {
	opts.L.SetSubstage("Compile")
	if err := os.MkdirAll(filepath.Join(opts.Dir, "tmp-gocache"), 0755); err != nil {
		return fmt.Errorf("mkdir %q failed: %v", "/tmp-gocache", err)
	}
	defer os.RemoveAll(filepath.Join(opts.Dir, "tmp-gocache"))

	cmd, err := chroot.CmdContext(ctx, opts, "bash", "-c", "cd /tmp-twlinst-build && go build -o /usr/share/twlinst/twlinst -v *.go")
	if err != nil {
		return fmt.Errorf("building installer: %v", err)
	}
	cmd.Env = []string{"GOPATH=/tmp-twlinst-build", "GOCACHE=/tmp-gocache", "PATH=/bin:/usr/bin:/usr/local/go/bin:/sbin:/usr/sbin"}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	return cmd.Run()
}

func (i *Installer) copyResources(ctx context.Context, opts *Opts) error {
	opts.L.SetSubstage("Copy resources")
	for _, sysdFile := range []string{"installer.target", "twl-installer.service"} {
		if err := CopyResource(ctx, opts, filepath.Join("installer", sysdFile), filepath.Join("lib/systemd/system", sysdFile)); err != nil {
			return err
		}
	}
	if err := CopyResource(ctx, opts, filepath.Join("installer", "gnome-installer-setup.json"), "usr/share/gnome-shell/modes/initial-setup.json"); err != nil {
		return err
	}
	if err := CopyResource(ctx, opts, filepath.Join("installer", "twlinst-start"), "usr/sbin/twlinst-start"); err != nil {
		return err
	}
	if err := CopyResource(ctx, opts, filepath.Join("installer", "twl-plain-background.png"), "usr/share/backgrounds/twl-plain-background.png"); err != nil {
		return err
	}
	if err := Shell(ctx, opts, "cp", filepath.Join(opts.Dir, "tmp-twlinst-build", "layout.glade"), filepath.Join(opts.Dir, "usr", "share", "twlinst", "layout.glade")); err != nil {
		return err
	}
	return nil
}

func (i *Installer) installVersion(ctx context.Context, opts *Opts) error {
	scriptPath := filepath.Join(opts.Dir, "usr", "sbin", "twlinst-start")
	script, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return err
	}
	newScript := strings.Replace(string(script), "VERSION_MARKER", opts.Version, -1)
	return ioutil.WriteFile(scriptPath, []byte(newScript), 0755)
}
