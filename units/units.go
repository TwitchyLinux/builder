package units

import "context"

// DebianOpts configures the debian URL and track the system will
// be based on.
type DebianOpts struct {
	URL   string
	Track string
}

// Opts describes options provided to the units.
type Opts struct {
	// Dir represents the path the system is being built at.
	Dir string
	// Num indicates which unit (in execution order) the unit is.
	Num int

	Debian DebianOpts
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
	&Debootstrap{},
	&FinalizeApt{},
	&BaseBuildtools{},
}
