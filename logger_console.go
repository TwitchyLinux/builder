package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/pkg/term"
)

type interactiveOutput struct {
	lock  sync.Mutex
	units []*unitState

	idx                int
	currentUnit        *unitState
	stdoutLinesWritten int

	consoleBuff  [12]string
	lineIsStderr [12]bool
	newest       int
}

func truncateStr(in string, maxLen uint16) string {
	if len(in) > int(maxLen) {
		return in[:maxLen-3] + "..."
	}
	return in
}

func (o *interactiveOutput) resetCursor() {
	for ; o.stdoutLinesWritten > 0; o.stdoutLinesWritten-- {
		os.Stdout.Write([]byte("\033[A"))  // Move cursor up a line.
		os.Stdout.Write([]byte("\033[2K")) // Erase line.
	}
	os.Stdout.Write([]byte("\033[0G")) // Move to beginning of line.
}

func (o *interactiveOutput) writeConsoleBuffer(ws *term.Winsize) {
	for i := 0; i < len(o.consoleBuff); i++ {
		idx := (1 + i + o.newest) % len(o.consoleBuff)
		if o.lineIsStderr[idx] {
			os.Stdout.Write([]byte("\033[1;31m")) // Set text red.
		}
		os.Stdout.Write([]byte(truncateStr(o.consoleBuff[idx], ws.Width) + "\n"))
		os.Stdout.Write([]byte("\033[0m")) // Reset text colors.
		o.stdoutLinesWritten++
	}
}

func (o *interactiveOutput) writeHeader(ws *term.Winsize) {
	fmt.Fprintf(os.Stdout, "Building TwitchyLinux \033[1;32m(%d\033[1;0m/\033[1;32m%d)\033[1;0m --- \033[1;34m%s\033[1;0m\n", o.idx+1, len(o.units), o.currentUnit.unit.Name())
	o.stdoutLinesWritten++
}

func (o *interactiveOutput) flush() {
	ws, err := term.GetWinsize(os.Stdin.Fd())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine terminal size: %v\n", err)
		return
	}
	o.resetCursor()
	o.writeHeader(ws)
	o.writeConsoleBuffer(ws)
}

func (o *interactiveOutput) findIndex(unit *unitState) (int, bool) {
	for i := range o.units {
		if o.units[i] == unit {
			return i, true
		}
	}
	return -1, false
}

func (o *interactiveOutput) registerUnit(unit *unitState) {
	o.units = append(o.units, unit)
}

func (o *interactiveOutput) updated(unit *unitState) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if unit != o.currentUnit {
		idx, found := o.findIndex(unit)
		if !found {
			panic("could not find unit!")
		}
		o.idx = idx
		o.currentUnit = unit
	}
	o.flush()
}

func (o *interactiveOutput) unitWrite(unit *unitState, in []byte, stderr bool) (int, error) {
	o.lock.Lock()
	defer o.lock.Unlock()
	for _, line := range strings.Split(string(in), "\n") {
		if line == "" {
			continue
		}
		o.newest = (o.newest + 1) % len(o.consoleBuff)
		o.consoleBuff[o.newest] = line
		o.lineIsStderr[o.newest] = stderr
	}

	o.flush()
	return len(in), nil
}
