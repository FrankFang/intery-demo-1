package cmd

import (
	"intery/cmd/generate"
	"intery/db"
	"intery/server"
	"os"

	"github.com/urfave/cli/v2"
)

var rootCmd = &cli.App{
	Commands: []*cli.Command{
		{
			Name:    "task",
			Aliases: []string{"r"},
			Usage:   "run a task",
			Subcommands: []*cli.Command{
				{
					Name: "db:create",
					Action: func(c *cli.Context) error {
						return db.Create()
					},
				},
				{
					Name: "db:migrate",
					Action: func(c *cli.Context) error {
						return db.Migrate()
					},
				},
				{
					Name: "db:rollback",
					Action: func(c *cli.Context) error {
						return db.Rollback()
					},
				},
				{
					Name: "db:drop",
					Action: func(context *cli.Context) error {
						return db.Drop()
					},
				},
				{
					Name: "db:reset",
					Action: func(context *cli.Context) error {
						return db.Reset()
					},
				},
				{
					Name:    "generate",
					Aliases: []string{"g"},
					Action: func(context *cli.Context) error {
						generate.Run()
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
