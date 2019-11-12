package main

import "os"

type rawOutput struct{}

func (o *rawOutput) registerUnit(unit *unitState) {}

func (o *rawOutput) updated(unit *unitState) {}

func (o *rawOutput) unitWrite(unit *unitState, in []byte, stderr bool) (int, error) {
	if stderr {
		return os.Stderr.Write(in)
	}
	return os.Stdout.Write(in)
}
