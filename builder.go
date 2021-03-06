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
	"strconv"
	"strings"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/twitchylinux/builder/stager"
	"github.com/twitchylinux/builder/units"
)

var (
	resourcesDir = flag.String("resources-dir", "resources", "Path to the builder resources directory.")
	outputOnly   = flag.Bool("output-only", false, "Only output to stdout in a non-interactive fashion.")
	version      = flag.String("twl-version", "0.8.3", "The current version of TwitchyLinux.")
	debProxyAddr = flag.String("deb-proxy-addr", "", "The address:port of a proxy to use when fetching deb packages.")
	printUnits   = flag.Bool("print-units", false, "Print the computed build units before exiting.")

	defaultNumThreads = int(math.Max(1, float64(runtime.NumCPU()-1)))
	numThreads        = flag.Int("j", defaultNumThreads, "Number of concurrent threads to use while building.")
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "USAGE: %s [options] <build-directory> [<build-options>...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nBuild options:\n")
	fmt.Fprintf(os.Stderr, "  -D <key>=<value>\n    \tOverride or set a configuration value.\n")
	fmt.Fprintf(os.Stderr, "\nRegular options:\n")
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
		Version:    *version,
		DebProxy:   *debProxyAddr,
	}

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
	opts, err := stageConfigOpts(flag.CommandLine)
	if err != nil {
		return nil, err
	}
	uts, err := stager.UnitsFromConfig(filepath.Join(config.Resources, "stage-conf"), opts)
	if err != nil {
		return nil, err
	}

	candidateUnits := make([]*unitState, 0, len(uts))
	for i, unit := range uts {
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

	if *printUnits {
		for i, s := range candidateUnits {
			fmt.Printf("Unit %d/%d: %s\n", i, len(candidateUnits), s.unit.Name())
			spew.Dump(s.unit)
			fmt.Println()
		}
		return nil
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

// stageConfigOpts computes options to be provided to the stager.
func stageConfigOpts(f *flag.FlagSet) (stager.Options, error) {
	out := stager.Options{
		Overrides: map[string]interface{}{},
	}

	for i := 1; i < f.NArg(); i++ {
		switch a := f.Arg(i); a {
		case "-D", "--D":
			s := f.Arg(i + 1)
			eqIdx := strings.Index(s, "=")
			if eqIdx == -1 {
				return stager.Options{}, fmt.Errorf("invalid override string %q: must be form key=value", s)
			}

			var v interface{} = s[eqIdx+1:]
			switch v {
			case "false", "true":
				v, _ = strconv.ParseBool(s[eqIdx+1:])
			}
			out.Overrides[s[:eqIdx]] = v
			i++

		default:
			return stager.Options{}, fmt.Errorf("invalid option: %q", a)
		}
	}

	return out, nil
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
