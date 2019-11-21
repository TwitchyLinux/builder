// Package units implements logic to build TwitchyLinux as a series of units.
package units

import (
	"context"
	"fmt"
	"io"
)

// Opts describes options provided to the units.
type Opts struct {
	// Dir represents the path the system is being built at.
	Dir string
	// Resources is the path to the builder resources directory.
	Resources string
	// Version is the current version of TwitchyLinux.
	Version string

	// Num indicates which unit (in execution order) the unit is.
	Num int
	// L is a logger which units can use to communicate state.
	L Logger

	// NumThreads is the number of concurrent threads to be used while building.
	NumThreads int
}

func (o *Opts) makeNumThreadsArg() string {
	return fmt.Sprintf("-j%d", o.NumThreads)
}

// Logger implements status reporting and logging for executing units.
type Logger interface {
	Stderr() io.Writer
	Stdout() io.Writer
	SetSubstage(string)
}

// Unit describes an execution unit for building the system.
type Unit interface {
	Name() string
	Run(ctx context.Context, opts Opts) error
}
