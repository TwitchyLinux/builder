package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/twitchylinux/builder/units"
)

const statusDir = "build-status"

type unitStatus string

const (
	StatusFailed unitStatus = "failed"
	StatusDone   unitStatus = "complete"
)

func recordUnitStatus(buildOpts units.Opts, unit units.Unit, status unitStatus) error {
	if _, err := os.Stat(filepath.Join(buildOpts.Dir, statusDir)); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(filepath.Join(buildOpts.Dir, statusDir), 0755); err != nil {
				return err
			} else {
				return err
			}
		}
	}

	return ioutil.WriteFile(filepath.Join(buildOpts.Dir, statusDir, unit.Name()), []byte(status), 0644)
}

func skipUnit(buildOpts units.Opts, unit units.Unit) (bool, error) {
	d, err := ioutil.ReadFile(filepath.Join(buildOpts.Dir, statusDir, unit.Name()))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	switch unitStatus(d) {
	case StatusDone:
		return true, nil
	case StatusFailed:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status for unit %q: %q", unit.Name(), string(d))
	}
}
