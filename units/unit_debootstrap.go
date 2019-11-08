package units

import (
	"context"
	"os"
	"os/exec"
)

type Debootstrap struct {
}

func (d *Debootstrap) Name() string {
	return "Debootstrap"
}

func (d *Debootstrap) Run(ctx context.Context, opts Opts) error {
	dbstrp := exec.CommandContext(ctx, "debootstrap")
	dbstrp.Args = []string{"debootstrap", opts.Debian.Track, opts.Dir, opts.Debian.URL}
	dbstrp.Stdout = opts.L
	dbstrp.Stderr = os.Stderr
	return dbstrp.Run()
}
