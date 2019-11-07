package main

import (
	"context"
	"fmt"
	"os"

	"github.com/twitchylinux/builder/units"
)

func main() {
	ctx := context.Background()

	for i, unit := range units.Units {
		if err := unit.Run(ctx, units.Opts{Num: i}); err != nil {
			fmt.Fprintf(os.Stderr, "Unit %s errored: %v\n", unit.Name(), err)
			return
		}
	}
}
