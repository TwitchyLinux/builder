package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Cmd represents the execution of a command within a chroot.
type Cmd struct {
	Bin  string
	Args []string
}

// Name implements Unit.
func (c *Cmd) Name() string {
	return "run " + filepath.Base(c.Bin)
}

// Run implements Unit.
func (c *Cmd) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()
	return chroot.Shell(ctx, &opts, c.Bin, c.Args...)
}

// Append appends a line to a file
type Append struct {
	To   string
	Data string
}

// Name implements Unit.
func (c *Append) Name() string {
	return "append " + filepath.Base(c.To)
}

// Run implements Unit.
func (c *Append) Run(ctx context.Context, opts Opts) error {
	var (
		wantPerm os.FileMode
		wantData []byte
		p        string = filepath.Join(opts.Dir, c.To)
	)

	s, err := os.Stat(p)
	switch {
	case err != nil && os.IsNotExist(err):
		wantPerm = 0644
	case err != nil:
		return err
	default:
		wantPerm = s.Mode()
		if wantData, err = ioutil.ReadFile(p); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(p, append(wantData, []byte(c.Data)...), wantPerm)
}

// Download represents the download of a file into the system.
type Download struct {
	URL string
	To  string
}

// Name implements Unit.
func (d *Download) Name() string {
	return "download " + filepath.Base(d.URL)
}

// Run implements Unit.
func (d *Download) Run(ctx context.Context, opts Opts) error {
	return DownloadFile(ctx, &opts, d.URL, filepath.Join(opts.Dir, d.To))
}

// Composite is a unit made up of a sequence of simpler operations.
type Composite struct {
	UnitName string
	Order    int
	Ops      []Unit
}

// Name implements Unit.
func (c *Composite) Name() string {
	return c.UnitName
}

// Run implements Unit.
func (c *Composite) Run(ctx context.Context, opts Opts) error {
	for _, o := range c.Ops {
		opts.L.SetSubstage(o.Name())
		if err := o.Run(ctx, opts); err != nil {
			return fmt.Errorf("%s: %v", o.Name(), err)
		}
	}

	return nil
}
