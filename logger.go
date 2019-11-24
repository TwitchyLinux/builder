package main

import (
	"fmt"
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
	subStage string

	showProgress bool
	progress     float64
	progressMsg  string

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

// SetSubstage tells the logger the unit is entering a new substage.
func (u *unitState) SetSubstage(ss string) {
	u.subStage = ss
	u.output.updated(u)
}

// SetProgress shows a progress bar.
func (u *unitState) SetProgress(msg string, fraction float64) {
	if _, supportsProgress := u.output.(*interactiveOutput); supportsProgress {
		if fraction == 0 {
			u.showProgress = false
		} else {
			u.showProgress = true
			u.progress = fraction
			u.progressMsg = msg
		}
	} else {
		u.output.unitWrite(u, []byte(msg+": "+fmt.Sprint(int(fraction*100))+"%\n"), false)
	}
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
