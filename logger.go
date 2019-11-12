package main

import (
	"io"
	"time"

	"github.com/twitchylinux/builder/units"
)

type logger interface {
	registerUnit(*unitState)
	updated(*unitState)
	unitWrite(unit *unitState, in []byte, stderr bool) (int, error)
}

type unitState struct {
	started  time.Time
	finished time.Time
	done     bool
	skipped  bool
	err      error

	output logger
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

// Stderr returns a writer for writing to stderr.
func (u *unitState) Stderr() io.Writer {
	return &unitWriter{unit: u, stderr: true}
}

// Stderr returns a writer for writing to stdout.
func (u *unitState) Stdout() io.Writer {
	return &unitWriter{unit: u, stderr: false}
}

type unitWriter struct {
	unit   *unitState
	stderr bool
}

func (w *unitWriter) Write(in []byte) (int, error) {
	return w.unit.output.unitWrite(w.unit, in, w.stderr)
}
