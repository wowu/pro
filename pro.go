package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var openCommandFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "print",
		Aliases: []string{"p"},
		Usage:   "print the PR URL instead of opening in browser",
	},
}

func main() {
	// cli library API example:
	// https://github.com/urfave/cli/blob/main/docs/v2/manual.md#full-api-example
	app := &cli.App{
		Name:  "pro",
		Usage: "Pull Request opener",
		Flags: openCommandFlags,
		Commands: []*cli.Command{
			{
				Name:      "auth",
				ArgsUsage: "[gitlab|github]",
				Usage:     "Authorize GitLab or GitHub",
				UsageText: fmt.Sprintf("pro auth gitlab\npro login github"),
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						fmt.Println("Please specify provider (github or gitlab)")
						os.Exit(1)
					}

					provider := c.Args().Get(0)

					if provider != "gitlab" && provider != "github" {
						fmt.Println("Please specify provider (github or gitlab)")
						os.Exit(1)
					}

					if provider == "github" {
						fmt.Println("Github provider is not supported yet")
						os.Exit(0)
					}

					auth(provider)

					return nil
				},
			},
			{
				Name:  "open",
				Usage: "Open PR page in browser (default action)",
				Flags: openCommandFlags,
				Action: func(c *cli.Context) error {
					open(".", c.Bool("print"))
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			open(".", c.Bool("print"))

			return nil
		},
	}

	app.Run(os.Args)
}
