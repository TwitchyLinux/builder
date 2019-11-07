package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/twitchylinux/builder/units"
)

var (
	debianURL   = flag.String("debian-url", "http://deb.debian.org/debian/", "Mirror to download debian packages from.")
	debianTrack = flag.String("debian-track", "stable", "Which debian track to use.")
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "USAGE: %s [options] <build-directory>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	ctx := context.Background()
	flag.Usage = printUsage
	flag.Parse()

	config := units.Opts{
		Dir: buildDir(),
		Debian: units.DebianOpts{
			URL:   *debianURL,
			Track: *debianTrack,
		}}

	err := run(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, config units.Opts) error {
	for i, unit := range units.Units {
		opts := config
		opts.Num = i
		shouldSkip, err := skipUnit(opts, unit)
		if err != nil {
			return err
		}
		if shouldSkip {
			continue
		}

		if err := unit.Run(ctx, opts); err != nil {
			return err
		}

		if err := recordUnitStatus(opts, unit, StatusDone); err != nil {
			return err
		}
	}
	return nil
}

// buildDir returns the absolute path to the build directory, setting it up
// if it does not exist. The program exists if it is not specified or
// the path it references is not a r/x directory.
func buildDir() string {
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}
	dir := flag.Arg(0)
	if !filepath.IsAbs(dir) {
		wd, _ := os.Getwd()
		dir = filepath.Join(wd, dir)
	}

	s, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
			return dir
		}
		fmt.Fprintf(os.Stderr, "Error: Could not stat build directory: %v\n", err)
		os.Exit(1)
	}

	if !s.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dir)
		os.Exit(1)
	}
	if s.Mode()&0111 == 0 {
		fmt.Fprintf(os.Stderr, "Error: %s is not executable\n", dir)
		os.Exit(1)
	}
	return dir
}
