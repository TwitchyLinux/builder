package units

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileInfo describes a file to be installed in the target system.
type FileInfo struct {
	Path  string
	Data  []byte
	Perms os.FileMode
}

// InstallFiles installs files into the target system.
type InstallFiles struct {
	UnitName string
	Files    []FileInfo
}

// Name implements Unit.
func (i *InstallFiles) Name() string {
	return i.UnitName
}

// Run implements Unit.
func (i *InstallFiles) Run(ctx context.Context, opts Opts) error {
	for _, f := range i.Files {
		opts.L.SetSubstage("Install " + filepath.Base(f.Path))
		var perms os.FileMode = 0644
		if f.Perms != 0 {
			perms = f.Perms
		}
		if err := ioutil.WriteFile(filepath.Join(opts.Dir, f.Path), f.Data, perms); err != nil {
			return err
		}
	}

	return nil
}
