package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
)

var ReportsCmd = &cli.Command{
	Name:        "report",
	Description: "request reports for named containers",
	Usage:       "clairctl report",
	Action:      reportAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Usage:   "--host localhost:6060/",
			Value:   "http://localhost:6060/",
			EnvVars: []string{"CLAIR_API"},
		},
		&cli.StringFlag{
			Name:    "containers",
			Usage:   "--containers ubuntu:latest,mysql:latest",
			Value:   "",
			EnvVars: []string{"CONTAINERS"},
		},
		&cli.Int64Flag{
			Name:    "concurrency",
			Usage:   "--concurrency 10",
			Value:   1,
			EnvVars: []string{"CONCURRENCY"},
		},
	},
}

func reportAction(c *cli.Context) error {
	ctx := c.Context
	containers := c.String("containers")
	host := c.String("host")
	concurrency := c.Int64("concurrency")

	sem := semaphore.NewWeighted(concurrency)
	g, ctx := errgroup.WithContext(ctx)

	for _, c := range strings.Split(containers, ",") {
		cc := c
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			res, err := reportForContainer(ctx, cc, host)
			if err != nil {
				return err
			}
			zlog.Debug(ctx).Str("container", cc).Msg(res)
			return nil
		})
	}
	return g.Wait()
}

func reportForContainer(ctx context.Context, container, host string) (string, error) {
	cmd := exec.Command("clairctl", "report", "--host", host, container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("Starting to pull report")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}
