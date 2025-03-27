package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/ryotarai/hashiru/cli/client"
	"github.com/ryotarai/hashiru/cli/server"
	"github.com/urfave/cli/v3"
)

const defaultSocketPath = "/tmp/hashiru.sock"

func main() {
	app := &cli.Command{
		Name:  "hashiru",
		Usage: "A command execution server and client",
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Start the command execution server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "socket",
						Value:   defaultSocketPath,
						Usage:   "Unix domain socket path",
						Aliases: []string{"s"},
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					return server.Run(ctx, c.String("socket"))
				},
			},
			{
				Name:  "client",
				Usage: "Execute a command through the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "socket",
						Value:   defaultSocketPath,
						Usage:   "Unix domain socket path",
						Aliases: []string{"s"},
					},
					&cli.StringFlag{
						Name:    "workdir",
						Usage:   "Working directory (default: current directory)",
						Aliases: []string{"w"},
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() == 0 {
						return errors.New("no command provided")
					}
					exitCode, err := client.Run(ctx, c.String("socket"), c.Args().Slice(), c.String("workdir"))
					if err != nil {
						return err
					}
					os.Exit(exitCode)
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
