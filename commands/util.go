package commands

import (
	"fmt"
	"os"
)

// Print error and exit if error is present
func handleError(err error, reason string) {
	if err != nil {
		if reason != "" {
			fmt.Fprintln(os.Stderr, reason+":", err)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}

		os.Exit(1)
	}
}
