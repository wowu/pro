package commands

import (
	"fmt"
	"os"
)

// Print error and exit if error is present
func handleError(err error, reason string) {
	if err != nil {
		if reason != "" {
			fmt.Println(reason+":", err)
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}
