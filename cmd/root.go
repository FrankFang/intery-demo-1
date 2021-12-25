package cmd

import (
	"context"
	"intery/cmd/docker"
	"intery/cmd/generate"
	"intery/db"
	"intery/server"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var rootCmd = &cli.App{
	Commands: []*cli.Command{
		{
			Name:    "task",
			Aliases: []string{"t"},
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
				{
					Name: "nginx:start",
					Action: func(c *cli.Context) error {
						ctx := context.Background()
						containerId, err := docker.StartNginxContainer(ctx)
						if err != nil {
							return err
						}
						log.Println(containerId)
						return nil
					},
				},
				{
					Name: "clear",
					Action: func(c *cli.Context) error {
						// remove all files in /tmp/socket
						cwd, err := os.Getwd()
						if err != nil {
							return err
						}
						socketDir := filepath.Join(cwd, "userspace/socket")
						err = os.RemoveAll(socketDir)
						if err != nil {
							return err
						}
						err = os.MkdirAll(socketDir, 0777)
						if err != nil {
							return err
						}
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
