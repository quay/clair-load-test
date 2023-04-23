package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/quay/clair-load-test/attacker"
	"github.com/quay/clair-load-test/manifests"
	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
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

// Type to store the test config.
type TestConfig struct {
	Containers     []string `json:"containers"`
	Concurrency    int      `json:"concurrency"`
	TestRepoPrefix string   `json:"testrepoprefix"`
	ESHost         string   `json:"eshost"`
	ESPort         string   `json:"esport"`
	ESIndex        string   `json:"esindex"`
	Host           string   `json:"host"`
	HitSize        int      `json:"hitsize"`
	IndexDelete    bool     `json:"delete"`
	PSK            string   `json:"-"`
	UUID           string   `json:"uuid"`
}

// NewConfig creates and returns a test configuration from CLI options.
func NewConfig(c *cli.Context) *TestConfig {
	containersArg := c.String("containers")
	return &TestConfig{
		Containers:     strings.Split(containersArg, ","),
		TestRepoPrefix: c.String("testrepoprefix"),
		PSK:            c.String("psk"),
		UUID:           c.String("uuid"),
		Host:           c.String("host"),
		IndexDelete:    c.Bool("delete"),
		HitSize:        c.Int("hitsize"),
		Concurrency:    c.Int("concurrency"),
		ESHost:         c.String("eshost"),
		ESPort:         c.String("esport"),
		ESIndex:        c.String("esindex"),
	}
}

// getContainersList returns list of containers from test repo used in load phase.
// It returns a list of strings which is a list of container names.
func getContainersList(ctx context.Context, testRepoPrefix string, hitSize int) []string {
	var containers []string
	for i := 1; i <= hitSize; i++ {
		containers = append(containers, testRepoPrefix+"_tag_"+strconv.Itoa(i))
	}
	return containers
}

// reportAction drives the report action logic.
// It returns an error if any during the execution.
func reportAction(c *cli.Context) error {
	ctx := c.Context
	conf := NewConfig(c)
	if (c.String("containers") == "" && conf.TestRepoPrefix == "") || ((c.String("containers") != "") && conf.TestRepoPrefix != "") {
		return fmt.Errorf("Please specify either of --containers or --testrepoprefix options. Both are mutually exclusive")
	}
	if conf.TestRepoPrefix != "" {
		conf.Containers = getContainersList(ctx, conf.TestRepoPrefix, conf.HitSize)
	}
	if len(conf.Containers) > conf.HitSize {
		conf.Containers = conf.Containers[:conf.HitSize]
	}
	listOfManifests, listOfManifestHashes := manifests.GetManifest(ctx, conf.Containers, conf.Concurrency)
	jwt_token, err := CreateToken(conf.PSK)
	if err != nil {
		zlog.Debug(ctx).Str("PSK", conf.PSK).Msg("creating token")
		return fmt.Errorf("could not create token: %w", err)
	}
	err = orchestrateWorkload(ctx, listOfManifests, listOfManifestHashes, jwt_token, conf)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}
	return nil
}

// orchestrateWorkload triggers the api endpoint hits and writes results to the desired location.
// It returns an error if any during the execution.
func orchestrateWorkload(ctx context.Context, manifests [][]byte, manifestHashes []string, jwt_token string, conf *TestConfig) error {
	zlog.Debug(ctx).Msg("Orchestrating reports workload")
	zlog.Info(ctx).Str("UUID", conf.UUID).Msg("Run details")
	var requests []map[string]interface{}
	var err error
	attackMap := map[string]string{
		"UUID":        conf.UUID,
		"Concurrency": strconv.Itoa(conf.Concurrency),
		"ESHost":      conf.ESHost,
		"ESPort":      conf.ESPort,
		"ESIndex":     conf.ESIndex,
		"Host":        conf.Host,
	}
	requests = attacker.CreateIndexReportRequests(ctx, manifests, conf.Host, jwt_token)
	err = attacker.RunVegeta(ctx, requests, "post_index_report", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running POST operation on index_report: %w", err)
	}
	requests = attacker.GetIndexReportRequests(ctx, manifestHashes, conf.Host, jwt_token)
	err = attacker.RunVegeta(ctx, requests, "get_index_report", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running GET operation on index_report: %w", err)
	}
	requests = attacker.GetVulnerabilityReportRequests(ctx, manifestHashes, conf.Host, jwt_token)
	err = attacker.RunVegeta(ctx, requests, "get_vulnerability_report", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running GET operation on vulnerability_report: %w", err)
	}
	requests = attacker.GetIndexerStateRequests(ctx, len(manifests), conf.Host, jwt_token)
	err = attacker.RunVegeta(ctx, requests, "get_indexer_state", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running GET operation on indexer_state: %w", err)
	}
	if conf.IndexDelete {
		requests = attacker.DeleteIndexReportsRequests(ctx, manifestHashes, conf.Host, jwt_token)
		err = attacker.RunVegeta(ctx, requests, "delete_index_report", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running DELETE operation on index_report: %w", err)
		}
	}
	return nil
}
