package cmd

import (
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
						return db.Create("intery_development")
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
					Name: "db:destroy",
					Action: func(context *cli.Context) error {
						return db.Destroy("intery_development")
					},
				},
				{
					Name: "db:reset",
					Action: func(context *cli.Context) error {
						_ = db.Destroy("intery_development")
						err := db.Create("intery_development")
						if err != nil {
							return err
						}
						return db.Migrate()
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
