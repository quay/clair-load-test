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
	Usage:       "clair-load-test report",
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

type clairCtlArgs struct {
	host           string
	configFilePath string
}

func reportAction(c *cli.Context) error {
	ctx := c.Context
	containers := c.String("containers")
	concurrency := c.Int64("concurrency")

	conf := &clairCtlArgs{
		host:           c.String("host"),
		configFilePath: confFilePath,
	}

	sem := semaphore.NewWeighted(concurrency)
	g, ctx := errgroup.WithContext(ctx)

	for _, c := range strings.Split(containers, ",") {
		cc := c
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			res, err := reportForContainer(ctx, cc, conf)
			if err != nil {
				zlog.Error(ctx).Str("err", res).Str("container", cc).Msg(err.Error())
				return nil
			}
			zlog.Debug(ctx).Str("container", cc).Msg(res)
			return nil
		})
	}
	return g.Wait()
}

func reportForContainer(ctx context.Context, container string, conf *clairCtlArgs) (string, error) {
	cmd := exec.Command(
		"clairctl", "--config", conf.configFilePath, "report",
		"--host", conf.host, container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("Starting to pull report")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		return errOut.String(), err
	}
	return out.String(), nil
}
