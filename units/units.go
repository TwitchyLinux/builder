package units

import "context"

// Opts describes options provided to the units.
type Opts struct {
	// OutputDir represents the path the system is being built at.
	OutputDir string
	// Num indicates which unit (in execution order) the unit is.
	Num int
}

// Unit describes an execution unit for building the system.
type Unit interface {
	Name() string
	Run(ctx context.Context, opts Opts) error
}

// Units contains the ordered set of all units needed to build the
// target system.
// TODO: Switch to method?
var Units = []Unit{
	&Preflight{},
}
