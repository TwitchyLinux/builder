package units

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type grubColorPair struct {
	FG, BG string
}

// Grub2 is a unit which installs Grub2.
type Grub2 struct {
	DistroName string
	Quiet      bool

	ColorNormal    grubColorPair
	ColorHighlight grubColorPair
}

// Name implements Unit.
func (i *Grub2) Name() string {
	return "Grub2"
}

// Run implements Unit.
func (i *Grub2) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := chroot.AptInstall(ctx, &opts, "grub2"); err != nil {
		return err
	}
	os.Remove(filepath.Join(opts.Dir, "etc", "grub.d", "05_debian_theme"))

	conf, err := ioutil.ReadFile(filepath.Join(opts.Dir, "etc", "default", "grub"))
	if err != nil {
		return err
	}
	conf = append(conf, []byte("GRUB_COLOR_NORMAL=\""+i.ColorNormal.FG+"/"+i.ColorNormal.BG+"\"\n")...)
	conf = append(conf, []byte("GRUB_COLOR_HIGHLIGHT=\""+i.ColorHighlight.FG+"/"+i.ColorHighlight.BG+"\"\n")...)
	if i.Quiet {
		conf = []byte(strings.Replace(string(conf), "\nGRUB_CMDLINE_LINUX_DEFAULT=\"quiet\"", "\nGRUB_CMDLINE_LINUX_DEFAULT=\"\"", -1))
	}
	conf = []byte(strings.Replace(string(conf), "echo Debian`", "echo Debian/"+i.DistroName+"`", -1))

	return ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "default", "grub"), conf, 0644)
}
