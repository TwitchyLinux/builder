package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/twitchylinux/builder/units"
)

var (
	debianURL    = flag.String("debian-url", "http://deb.debian.org/debian/", "Mirror to download debian packages from.")
	debianTrack  = flag.String("debian-track", "stable", "Which debian track to use.")
	resourcesDir = flag.String("resources-dir", "resources", "Path to the builder resources directory.")
	outputOnly   = flag.Bool("output-only", false, "Only output to stdout in a non-interactive fashion.")

	defaultNumThreads = int(math.Max(1, float64(runtime.NumCPU()-1)))
	numThreads        = flag.Int("j", defaultNumThreads, "Number of concurrent threads to use while building.")
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
		Dir:        buildDir(),
		Resources:  resourceDir(),
		NumThreads: *numThreads,
		Debian: units.DebianOpts{
			URL:   *debianURL,
			Track: *debianTrack,
		}}

	var logger logger
	if *outputOnly {
		logger = &rawOutput{}
	} else {
		logger = &interactiveOutput{}
	}

	err := run(ctx, config, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cancelCtxOnSignal(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		defer signal.Reset(syscall.SIGTERM, os.Interrupt)
		<-sigs
		cancel()
	}()
}

func selectUnits(config units.Opts, logger logger) ([]*unitState, error) {
	candidateUnits := make([]*unitState, 0, len(units.Units))
	for i, unit := range units.Units {
		opts := config
		opts.Num = i
		ul := &unitState{
			opts:   &opts,
			unit:   unit,
			output: logger,
		}
		opts.L = ul
		logger.registerUnit(ul)

		shouldSkip, err := skipUnit(opts, unit)
		if err != nil {
			return nil, err
		}
		if shouldSkip {
			ul.setSkipped()
			continue
		}
		candidateUnits = append(candidateUnits, ul)
	}
	return candidateUnits, nil
}

func run(ctx context.Context, config units.Opts, logger logger) error {
	ctx, cancel := context.WithCancel(ctx)
	cancelCtxOnSignal(cancel)

	candidateUnits, err := selectUnits(config, logger)
	if err != nil {
		return err
	}

	for _, ul := range candidateUnits {
		unit := ul.unit
		ul.setStarting()
		if err := unit.Run(ctx, *ul.opts); err != nil {
			ul.setFinalState(err)
			return fmt.Errorf("%s: %v", unit.Name(), err)
		}

		ul.setFinalState(nil)
		if err := recordUnitStatus(*ul.opts, unit, StatusDone); err != nil {
			return err
		}
	}
	return nil
}

// resourceDir returns the path to the resources directory. The program
// exits if the path it references is not valid.
func resourceDir() string {
	s, err := os.Stat(*resourcesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not stat resources directory: %v\n", err)
		os.Exit(1)
	}
	if !s.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", *resourcesDir)
		os.Exit(1)
	}
	return *resourcesDir
}

// buildDir returns the absolute path to the build directory, setting it up
// if it does not exist. The program exits if it is not specified or
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
