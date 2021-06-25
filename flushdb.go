package main

import (
	"bufio"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
)

var FlushDBCmd = &cli.Command{
	Name:        "flushdb",
	Description: "truncate relevant tables in the DB",
	Usage:       "clair-load-test flushdb",
	Action:      flushDBAction,
	Flags:       []cli.Flag{},
}

func flushDBAction(c *cli.Context) error {
	ctx := c.Context
	//connect to DB
	DBURL := "postgres://clair:clair@localhost:5432/clair"
	conn, err := pgx.Connect(ctx, DBURL)

	zlog.Warn(ctx).Msg("About to delete data, continue? [y/n]")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if string([]byte(input)[0]) != "y" {
		return nil
	}

	zlog.Info(ctx).Msg("Going to TRUNCATE table scanned_layer")
	_, err = conn.Exec(ctx, "TRUNCATE scanned_layer")
	if err != nil {
		return err
	}

	zlog.Info(ctx).Msg("Going to TRUNCATE table scanned_manifest")
	_, err = conn.Exec(ctx, "TRUNCATE scanned_manifest")
	if err != nil {
		return err
	}
	return nil
}
