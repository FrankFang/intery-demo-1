package cmd

import (
	"intery/server"
	"os"

	"github.com/urfave/cli/v2"
)

var rootCmd = &cli.App{
	Commands: []*cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run a task",
			Subcommands: []*cli.Command{
				{
					Name: "db:create",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		},
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "start a server",
			Action: func(c *cli.Context) error {
				return server.Run()
			},
		},
	},
}

func Execute() error {
	return rootCmd.Run(os.Args)
}
