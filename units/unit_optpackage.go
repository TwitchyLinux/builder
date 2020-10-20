package units

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type optPackageMeta struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Version     string   `json:"version"`
	Packages    []string `json:"top-level-packages"`
}

// OptPackage is a unit which installs packages from apt.
type OptPackage struct {
	OptName, Version string
	DisplayName      string
	Packages         []string
}

// Name implements Unit.
func (i *OptPackage) Name() string {
	return i.OptName
}

// Run implements Unit.
func (i *OptPackage) Run(ctx context.Context, opts Opts) error {
	if err := os.Mkdir(filepath.Join(opts.Dir, "deb-pkgs", i.OptName), 0755); err != nil && !os.IsExist(err) {
		return err
	}

	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := chroot.Shell(ctx, &opts, "apt-get", "clean"); err != nil {
		return err
	}
	args := append([]string{"--download-only", "install", "-y"}, i.Packages...)
	if err := chroot.Shell(ctx, &opts, "apt-get", args...); err != nil {
		return err
	}

	if err := chroot.Shell(ctx, &opts, "bash", "-c",
		"/bin/mv -v /var/cache/apt/archives/*.deb /deb-pkgs/"+i.OptName+"/"); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(opts.Dir, "deb-pkgs", i.OptName, "meta.json"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(optPackageMeta{
		Name:        i.OptName,
		DisplayName: i.DisplayName,
		Version:     i.Version,
		Packages:    i.Packages,
	})
}
