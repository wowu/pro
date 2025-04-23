package main

import (
	"fmt"
	"os"

	"github.com/wowu/pro/command"

	"github.com/urfave/cli/v2"
)

var openCommandFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "print",
		Aliases: []string{"p"},
		Usage:   "print URL instead of opening in browser",
	},
	&cli.BoolFlag{
		Name:    "copy",
		Aliases: []string{"c"},
		Usage:   "copy URL to clipboard instead of opening in browser",
	},
}

func main() {
	// cli library API example:
	// https://github.com/urfave/cli/blob/main/docs/v2/manual.md#full-api-example
	app := &cli.App{
		Name:    "pro",
		Usage:   "Pull Request Opener",
		Version: "v0.4.1",
		Flags:   openCommandFlags,
		Commands: []*cli.Command{
			{
				Name:      "auth",
				ArgsUsage: "[gitlab|github]",
				Usage:     "Authorize GitLab or GitHub",
				UsageText: "pro auth gitlab\npro login github",
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

					command.Auth(provider)

					return nil
				},
			},
			{
				Name:  "open",
				Usage: "Open PR page in browser (default action)",
				Flags: openCommandFlags,
				Action: func(c *cli.Context) error {
					command.Open(".", c.Bool("print"), c.Bool("copy"))
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			command.Open(".", c.Bool("print"), c.Bool("copy"))

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
