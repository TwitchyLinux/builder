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

// Mkdir represents the execution of mkdir on the target system.
type Mkdir struct {
	Dir string
}

// Name implements Unit.
func (c *Mkdir) Name() string {
	return "mkdir " + filepath.Base(c.Dir)
}

// Run implements Unit.
func (c *Mkdir) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()
	return chroot.Shell(ctx, &opts, "mkdir", "-pv", c.Dir)
}

// CheckHash represents the validation of the sha256 checksum of a file.
type CheckHash struct {
	File         string
	ExpectedHash string
}

// Name implements Unit.
func (c *CheckHash) Name() string {
	return "sha256sum " + filepath.Base(c.File)
}

// Run implements Unit.
func (c *CheckHash) Run(ctx context.Context, opts Opts) error {
	return CheckSHA256(filepath.Join(opts.Dir, c.File), c.ExpectedHash)
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

// EnableUnit enables a systemd unit.
type EnableUnit struct {
	Unit, Target string
}

// Name implements Unit.
func (c *EnableUnit) Name() string {
	return "enable " + c.Unit
}

func (c *EnableUnit) isEnabled(opts Opts, unit, target string) (bool, error) {
	s, err := os.Lstat(filepath.Join(opts.Dir, "etc/systemd/system", target+".wants", unit))
	switch {
	case err != nil && !os.IsNotExist(err):
		return false, err
	case err == nil && s.Mode()&os.ModeSymlink == 0:
		return false, fmt.Errorf("expected symlink on %s", filepath.Join("/etc/systemd/system", target+".wants"))
	case err == nil && s.Mode()&os.ModeSymlink != 0:
		return true, nil
	}

	s, err = os.Lstat(filepath.Join(opts.Dir, "lib/systemd/system", target+".wants", unit))
	switch {
	case err != nil && !os.IsNotExist(err):
		return false, err
	case err == nil && s.Mode()&os.ModeSymlink == 0:
		return false, fmt.Errorf("expected symlink on %s", filepath.Join("/lib/systemd/system", target+".wants"))
	case err == nil && s.Mode()&os.ModeSymlink != 0:
		return true, nil
	}

	return false, nil
}

// Run implements Unit.
func (c *EnableUnit) Run(ctx context.Context, opts Opts) error {
	enabled, err := c.isEnabled(opts, c.Unit, c.Target)
	if err != nil {
		return err
	}
	if enabled {
		fmt.Fprintf(opts.L.Stdout(), "Unit %q was already enabled for target %q.", c.Unit, c.Target)
		return nil
	}

	if _, err := os.Stat(filepath.Join(opts.Dir, "lib/systemd/system", c.Target+".wants")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.Mkdir(filepath.Join(opts.Dir, "lib/systemd/system", c.Target+".wants"), 755); err != nil {
			return err
		}
	}

	return os.Symlink("../"+c.Unit, filepath.Join(opts.Dir, "lib/systemd/system", c.Target+".wants", c.Unit))
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
