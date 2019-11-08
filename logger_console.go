package main

import (
	"os"
	"time"

	"github.com/twitchylinux/builder/units"
)

type interactiveOutput struct {
	units []*unitState
}

func (o *interactiveOutput) updated(unit *unitState) {
}

func (o *interactiveOutput) writeOccured(unit *unitState) {
}

type unitState struct {
	started  time.Time
	finished time.Time
	done     bool
	skipped  bool
	err      error

	output *interactiveOutput
	unit   units.Unit
	opts   *units.Opts
}

func (u *unitState) setSkipped() {
	u.started = time.Now()
	u.done = false
	u.skipped = true
	u.output.updated(u)
}

func (u *unitState) setStarting() {
	u.started = time.Now()
	u.done = false
	u.skipped = false
	u.output.updated(u)
}

func (u *unitState) setFinalState(err error) {
	u.finished = time.Now()
	u.done = true
	u.err = err
	u.output.updated(u)
}

// Write implements units.Logger.
func (u *unitState) Write(in []byte) (int, error) {
	defer u.output.writeOccured(u)
	return os.Stdout.Write(in)
}
