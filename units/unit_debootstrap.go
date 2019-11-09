package units

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// Debootstrap bootstraps the base debian system.
type Debootstrap struct {
}

// Name implements Unit.
func (d *Debootstrap) Name() string {
	return "Debootstrap"
}

// Run implements Unit.
func (d *Debootstrap) Run(ctx context.Context, opts Opts) error {
	dbstrp := exec.CommandContext(ctx, "debootstrap")
	dbstrp.Args = []string{"debootstrap", opts.Debian.Track, opts.Dir, opts.Debian.URL}
	dbstrp.Stdout = opts.L
	dbstrp.Stderr = os.Stderr
	if err := dbstrp.Run(); err != nil {
		return err
	}

	return Shell(ctx, &opts, "cp", filepath.Join(opts.Resources, "fstab"), filepath.Join(opts.Dir, "etc", "fstab"))
}
