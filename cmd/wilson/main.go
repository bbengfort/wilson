package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"

	confire "go.rtnl.ai/confire/usage"

	"go.bengfort.dev/wilson/pkg"
	"go.bengfort.dev/wilson/pkg/config"
	"go.bengfort.dev/wilson/pkg/server"
)

func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

	app := &cli.Command{
		Name:    "wilson",
		Usage:   "start and manage the wilson server",
		Version: pkg.Version(false),
		Commands: []*cli.Command{
			{
				Name:     "serve",
				Usage:    "start the wilson server",
				Action:   serve,
				Category: "service",
			},
			{
				Name:     "config",
				Usage:    "print wilson configuration guide",
				Category: "service",
				Action:   usage,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "print in list mode instead of table mode",
					},
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		if ec, ok := err.(cli.ExitCoder); ok {
			os.Exit(ec.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

//===========================================================================
// Server commands
//===========================================================================

// serve starts the endeavor HTTP server.
func serve(ctx context.Context, cmd *cli.Command) (err error) {
	var srv *server.Server
	if srv, err = server.New(); err != nil {
		return cli.Exit(err, 1)
	}
	if err = srv.Serve(); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

// usage prints the endeavor configuration guide.
func usage(ctx context.Context, cmd *cli.Command) error {
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)
	format := confire.DefaultTableFormat
	if cmd.Bool("list") {
		format = confire.DefaultListFormat
	}

	var conf config.Config
	if err := confire.Usagef(config.Prefix, &conf, tabs, format); err != nil {
		return cli.Exit(err, 1)
	}
	tabs.Flush()
	return nil
}
