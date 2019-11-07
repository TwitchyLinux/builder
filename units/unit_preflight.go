package units

import (
	"context"
	"fmt"

	ss "github.com/twitchylinux/builder/shellstr"
)

var (
	needBinaries = []string{
		"bash",
		"debdep",
		"ld",
		"gcc",
		"g++",
		"yacc",
		"bzip2",
		"chown",
		"diff",
		"find",
		"gawk",
		"grep",
		"gzip",
		"m4",
		"make",
		"patch",
		"perl",
		"sed",
		"tar",
		"makeinfo",
		"xz",
	}

	neededVersions = []versionCheck{
		{
			bin:        "bash",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}, &ss.Cut{Delim: " ", From: 2, To: 4}},
			minVersion: "3.2",
		},
		{
			bin:        "ld",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}, &ss.Cut{Delim: " ", From: 3}},
			minVersion: "2.25",
		},
		{
			bin:        "bzip2",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}, &ss.Cut{Delim: " ", From: 6}},
			minVersion: "1.0.4",
		},
		{
			bin:        "chown",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}, &ss.Cut{Delim: ")", To: 2, From: 2}},
			minVersion: "6.9",
		},
		{
			bin:        "gcc",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "4.9",
		},
		{
			bin:        "g++",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "4.9",
		},
		{
			bin:        "ldd",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}, &ss.Cut{Delim: " ", From: 2}},
			minVersion: "2.11",
		},
		{
			bin:        "grep",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "2.5.1",
		},
		{
			bin:        "gzip",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "1.3.12",
		},
		{
			bin:        "uname",
			args:       []string{"-a"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "4",
		},
		{
			bin:        "m4",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}, &ss.Cut{Delim: " ", From: 2}},
			minVersion: "1.4.10",
		},
		{
			bin:        "make",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "4.0",
		},
		{
			bin:        "patch",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "2.5.4",
		},
		{
			bin:        "sed",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "4.1.5",
		},
		{
			bin:        "tar",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "1.22",
		},
		{
			bin:        "makeinfo",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "5.0",
		},
		{
			bin:        "xz",
			args:       []string{"--version"},
			trims:      []ss.Op{&ss.Head{N: 1}},
			minVersion: "5.0.0",
		},
	}
)

type versionCheck struct {
	bin   string
	args  []string
	trims []ss.Op

	minVersion string
}

// Preflight checks the building system has the requisite packages and programs
// installed.
type Preflight struct{}

// Name implements Unit.
func (p *Preflight) Name() string {
	return "Preflight"
}

// Run implements Unit.
func (p *Preflight) Run(ctx context.Context, opts Opts) error {
	for _, bin := range needBinaries {
		if _, err := FindBinary(bin); err != nil {
			return fmt.Errorf("could not find %s on host", bin)
		}
	}

	for _, chk := range neededVersions {
		versStr, err := CmdCombined(ctx, chk.bin, chk.args...)
		if err != nil {
			return fmt.Errorf("could not get %s version: %v", chk.bin, err)
		}
		v, err := CompareExtractSemver(ss.Trim(versStr, chk.trims...), chk.minVersion)
		if err != nil {
			return fmt.Errorf("could not extract %s's version: %v", chk.bin, err)
		}
		if v >= 0 {
			return fmt.Errorf("%s must be >= version %s", chk.bin, chk.minVersion)
		}
	}

	return nil
}
