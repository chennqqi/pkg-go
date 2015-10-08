package pkgcobra //import "go.pedge.io/pkg/cobra"

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Bounds represent min and max bounds.
type Bounds struct {
	Min int
	Max int
}

// CobraRunFixedArgs makes a new cobra run function that checks that the number of args is equal to numArgs.
func CobraRunFixedArgs(numArgs int, run func(args []string) error) func(*cobra.Command, []string) {
	return CobraRunBoundedArgs(Bounds{Min: numArgs, Max: numArgs}, run)
}

// CobraRunBoundedArgs makes a new cobra run function that checks that the number of args is within argBounds.
func CobraRunBoundedArgs(argBounds Bounds, run func(args []string) error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		if argBounds.Max == 0 && len(args) < argBounds.Min {
			errorAndExit("Expected at least %d args, got %d", argBounds.Min, len(args))
		} else if argBounds.Min == 0 && len(args) > argBounds.Max {
			errorAndExit("Expected at most %d args, got %d", argBounds.Max, len(args))
		} else if len(args) < argBounds.Min || len(args) > argBounds.Max {
			if argBounds.Min == argBounds.Max {
				errorAndExit("Expected %d args, got %d", argBounds.Min, len(args))
			}
			errorAndExit("Expected between %d and %d args, got %d", argBounds.Min, argBounds.Max, len(args))
		}
		check(run(args))
	}
}

func check(err error) {
	if err != nil {
		errorAndExit(err.Error())
	}
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
