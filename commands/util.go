package commands

import (
	"fmt"
	"os"
)

// Print error and exit
func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
