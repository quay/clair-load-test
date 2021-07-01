package main

import (
	"context"
	"os"
	"time"

	"github.com/quay/zlog"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	logout = zerolog.New(&zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()

	commonClaim  = jwt.Claims{}
	confFilePath = "config.yaml"
)

func main() {
	var exit int
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &cli.App{
		Name:                 "clairstress",
		Version:              "0.0.1",
		Usage:                "",
		Description:          "A command-line tool for stress testing clair v4.",
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			if c.IsSet("q") {
				logout = logout.Level(zerolog.WarnLevel)
			}
			if c.IsSet("D") {
				logout = logout.Level(zerolog.DebugLevel)
			}
			zlog.Set(&logout)
			commonClaim.Issuer = c.String("issuer")
			if c.IsSet("config") {
				confFilePath = c.Path("config")
			}
			return nil
		},
		Commands: []*cli.Command{
			ReportsCmd,
			FlushDBCmd,
			CreateTokenCmd,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "D",
				Usage: "print debugging logs",
			},
			&cli.BoolFlag{
				Name:  "q",
				Usage: "quieter log output",
			},
			&cli.PathFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Usage:     "clair configuration file",
				Value:     "config.yaml",
				TakesFile: true,
				EnvVars:   []string{"CLAIR_CONF"},
			},
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				exit = 1
				if err, ok := err.(cli.ExitCoder); ok {
					exit = err.ExitCode()
				}
				logout.Error().Err(err).Send()
			}
		},
	}
	app.RunContext(ctx, os.Args)

}
