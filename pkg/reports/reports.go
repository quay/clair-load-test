package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
	"github.com/vishnuchalla/clair-load-test/pkg/attacker"
	"github.com/vishnuchalla/clair-load-test/pkg/manifests"
	"github.com/vishnuchalla/clair-load-test/pkg/token"
	"github.com/vishnuchalla/clair-load-test/pkg/utils"
	"os"
	"strings"
)

// Command line to handle reports functionality.
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
			EnvVars: []string{"CLAIR_TEST_HOST"},
		},
		&cli.StringFlag{
			Name:    "uuid",
			Usage:   "--uuid f519d9b2-aa62-44ab-9ce8-4156b712f6d2",
			Value:   uuid.New().String(),
			EnvVars: []string{"CLAIR_TEST_UUID"},
		},
		&cli.StringFlag{
			Name:    "containers",
			Usage:   "--containers ubuntu:latest,mysql:latest",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_CONTAINERS"},
		},
		&cli.StringFlag{
			Name:    "testrepoprefix",
			Usage:   "--testrepoprefix quay.io/vchalla/clair-load-test:mysql_8.0.25",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_REPO_PREFIX"},
		},
		&cli.StringFlag{
			Name:    "psk",
			Usage:   "--psk secretkey",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_PSK"},
		},
		&cli.StringFlag{
			Name:    "eshost",
			Usage:   "--eshost eshosturl",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_ES_HOST"},
		},
		&cli.StringFlag{
			Name:    "esport",
			Usage:   "--esport esport",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_ES_PORT"},
		},
		&cli.StringFlag{
			Name:    "esindex",
			Usage:   "--esindex esindex",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_ES_INDEX"},
		},
		&cli.BoolFlag{
			Name:    "delete",
			Usage:   "--delete",
			Value:   false,
			EnvVars: []string{"CLAIR_TEST_INDEX_REPORT_DELETE"},
		},
		&cli.IntFlag{
			Name:    "hitsize",
			Usage:   "--hitsize 100",
			Value:   25,
			EnvVars: []string{"CLAIR_TEST_HIT_SIZE"},
		},
		&cli.IntFlag{
			Name:    "concurrency",
			Usage:   "--concurrency 50",
			Value:   10,
			EnvVars: []string{"CLAIR_TEST_CONCURRENCY"},
		},
	},
}

// Method to create a test configuration from CLI options.
func NewConfig(c *cli.Context) *utils.TestConfig {
	containersArg := c.String("containers")
	return &utils.TestConfig{
		Containers:     strings.Split(containersArg, ","),
		TestRepoPrefix: c.String("testrepoprefix"),
		Psk:            c.String("psk"),
		Uuid:           c.String("uuid"),
		Host:           c.String("host"),
		IndexDelete:    c.Bool("delete"),
		HitSize:        c.Int("hitsize"),
		Concurrency:    c.Int("concurrency"),
		ESHost:         c.String("eshost"),
		ESPort:         c.String("esport"),
		ESIndex:        c.String("esindex"),
	}
}

// Method to report action based on parameters.
func reportAction(c *cli.Context) error {
	ctx := c.Context
	conf := NewConfig(c)
	if (c.String("containers") == "" && conf.TestRepoPrefix == "") || ((c.String("containers") != "") && conf.TestRepoPrefix != "") {
		return fmt.Errorf("Please specify either of --containers or --testrepoprefix options. Both are mutually exclusive")
	}
	if conf.TestRepoPrefix != "" {
		conf.Containers = utils.GetContainersList(ctx, conf.TestRepoPrefix, conf.HitSize)
	}
	conf.Containers = conf.Containers[:conf.HitSize]
	listOfManifests, listOfManifestHashes := manifests.GetManifest(ctx, conf.Containers)
	var err error
	jwt_token, err := token.CreateToken(conf.Psk)
	if err != nil {
		zlog.Debug(ctx).Str("PSK", conf.Psk).Msg("creating token")
		return fmt.Errorf("could not create token: %w", err)
	}
	orchestrateWorkload(ctx, listOfManifests, listOfManifestHashes, jwt_token, conf)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}
	return nil
}

// Method to orchestrate the workload.
func orchestrateWorkload(ctx context.Context, manifests [][]byte, manifestHashes []string, jwt_token string, conf *utils.TestConfig) {
	zlog.Debug(ctx).Msg("Orchestrating reports workload")
	zlog.Info(ctx).Str("Uuid", conf.Uuid).Msg("Run details")
	var requests []map[string]interface{}
	var testName string
	requests, testName = CreateIndexReport(ctx, manifests, conf.Host, jwt_token)
	attacker.RunVegeta(requests, testName, conf)

	requests, testName = GetIndexReport(ctx, manifestHashes, conf.Host, jwt_token)
	attacker.RunVegeta(requests, testName, conf)

	requests, testName = GetVulnerabilityReport(ctx, manifestHashes, conf.Host, jwt_token)
	attacker.RunVegeta(requests, testName, conf)

	requests, testName = GetIndexerState(ctx, len(manifests), conf.Host, jwt_token)
	attacker.RunVegeta(requests, testName, conf)

	if conf.IndexDelete {
		requests, testName = DeleteIndexReports(ctx, manifestHashes, conf.Host, jwt_token)
		attacker.RunVegeta(requests, testName, conf)
	}
}
