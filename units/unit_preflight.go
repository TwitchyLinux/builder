package units

import (
	"context"
	"fmt"
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
)

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

	return nil
}
