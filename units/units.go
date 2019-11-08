package units

import (
	"context"
	"io"
)

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
	// Resources is the path to the builder resources directory.
	Resources string

	// Num indicates which unit (in execution order) the unit is.
	Num int
	// L is a logger which units can use to communicate state.
	L Logger

	Debian DebianOpts
}

// Logger implements status reporting and logging for executing units.
type Logger interface {
	io.Writer
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
	&Linux{},
}
