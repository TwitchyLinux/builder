package units

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Golang is a unit that install the Go toolchain.
type Golang struct {
	Version string
	URL     string
	SHA256  string
}

// Name implements Unit.
func (l *Golang) Name() string {
	return "Go"
}

func (l *Golang) dirFilename() string {
	return "go-" + l.Version
}

func (l *Golang) tarFilename() string {
	return l.dirFilename() + ".tar.gz"
}

func (l *Golang) tarPath(opts *Opts, inChroot bool) string {
	if inChroot {
		return "/" + l.tarFilename()
	}
	return filepath.Join(opts.Dir, l.tarFilename())
}

// Run implements Unit.
func (l *Golang) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	opts.L.SetSubstage("Downloading Go " + l.Version)
	if err := DownloadFile(&opts, l.URL, l.tarPath(&opts, false)); err != nil {
		return fmt.Errorf("Go source download failed: %v", err)
	}
	if err := CheckSHA256(l.tarPath(&opts, false), l.SHA256); err != nil {
		return err
	}

	opts.L.SetSubstage("Extracting")
	if err := chroot.Shell(ctx, &opts, "tar", "-C", "/usr/local", "-xzf", l.tarPath(&opts, true)); err != nil {
		return err
	}

	opts.L.SetSubstage("Installing to PATH")
	goProfPath := filepath.Join(opts.Dir, "etc", "profile.d", "golang.sh")
	if err := ioutil.WriteFile(goProfPath, []byte("# Make Go tools available via path\nexport PATH=$PATH:/usr/local/go/bin\n"), 0644); err != nil {
		return err
	}
	if err := os.Remove(l.tarPath(&opts, false)); err != nil {
		return err
	}
	return os.Chmod(goProfPath, 0755)
}
