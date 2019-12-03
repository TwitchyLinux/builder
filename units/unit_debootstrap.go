package units

import (
	"context"
	"os/exec"
	"path/filepath"
)

// Debootstrap bootstraps the base debian system.
type Debootstrap struct {
	Track string
	URL   string
}

// Name implements Unit.
func (d *Debootstrap) Name() string {
	return "Debootstrap"
}

// Run implements Unit.
func (d *Debootstrap) Run(ctx context.Context, opts Opts) error {
	dbstrp := exec.CommandContext(ctx, "debootstrap")
	if opts.DebProxy != "" {
		dbstrp.Env = append(dbstrp.Env, "http_proxy=http://"+opts.DebProxy)
	}
	dbstrp.Args = []string{"debootstrap", d.Track, opts.Dir, d.URL}
	dbstrp.Stdout = opts.L.Stdout()
	dbstrp.Stderr = opts.L.Stderr()
	if err := dbstrp.Run(); err != nil {
		return err
	}

	return Shell(ctx, &opts, "cp", filepath.Join(opts.Resources, "fstab"), filepath.Join(opts.Dir, "etc", "fstab"))
}
