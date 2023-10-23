package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/quay/clair-load-test/attacker"
	"github.com/quay/clair-load-test/manifests"
	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
)

// Constants
var validLayers = []int{-1, 5, 10, 15, 20, 25, 30, 35, 40}

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
			Name:    "runid",
			Usage:   "--runid f519d9b2-aa62-44ab-9ce8-4156b712f6d2",
			Value:   uuid.New().String(),
			EnvVars: []string{"CLAIR_TEST_RUNID"},
		},
		&cli.StringFlag{
			Name:    "containers",
			Usage:   "--containers ubuntu:latest,mysql:latest",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_CONTAINERS"},
		},
		&cli.StringFlag{
			Name:    "testrepoprefix",
			Usage:   "--testrepoprefix quay.io/vchalla/clair-load-test:mysql_8.0.25,quay.io/quay-qetest/clair-load-test:hadoop_latest",
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
			Name:    "layers",
			Usage:   "--layers [-1, 5, 10, 15, 20, 25, 30, 35, 40]",
			Value:   5,
			EnvVars: []string{"CLAIR_TEST_LAYERS"},
			Action: func(ctx *cli.Context, v int) error {
				for _, layer := range validLayers {
					if layer == v {
						return nil
					}
				}
				return fmt.Errorf("Invalid layer value. Must be one among: %v", validLayers)
			},
		},
		&cli.IntFlag{
			Name:    "concurrency",
			Usage:   "--concurrency 50",
			Value:   10,
			EnvVars: []string{"CLAIR_TEST_CONCURRENCY"},
		},
	},
	Before: func(c *cli.Context) error {
		if (c.String("containers") == "" && c.String("testrepoprefix") == "") || ((c.String("containers") != "") && c.String("testrepoprefix") != "") {
			return fmt.Errorf("Please specify either --containers or --testrepoprefix options. Both are mutually exclusive")
		}
		return nil
	},
}

// Type to store the test config.
type TestConfig struct {
	Containers     []string `json:"containers"`
	Concurrency    int      `json:"concurrency"`
	TestRepoPrefix []string `json:"testrepoprefix"`
	ESHost         string   `json:"eshost"`
	ESPort         string   `json:"esport"`
	ESIndex        string   `json:"esindex"`
	Host           string   `json:"host"`
	HitSize        int      `json:"hitsize"`
	Layers         int      `json:"layers"`
	IndexDelete    bool     `json:"delete"`
	PSK            string   `json:"-"`
	RUNID          string   `json:"runid"`
}

// NewConfig creates and returns a test configuration from CLI options.
func NewConfig(c *cli.Context) *TestConfig {
	containersArg := c.String("containers")
	testRepoPrefixArg := c.String("testrepoprefix")
	return &TestConfig{
		Containers:     strings.Split(strings.TrimSpace(containersArg), ","),
		TestRepoPrefix: strings.Split(strings.TrimSpace(testRepoPrefixArg), ","),
		PSK:            c.String("psk"),
		RUNID:          c.String("runid"),
		Host:           c.String("host"),
		IndexDelete:    c.Bool("delete"),
		HitSize:        c.Int("hitsize"),
		Layers:         c.Int("layers"),
		Concurrency:    c.Int("concurrency"),
		ESHost:         c.String("eshost"),
		ESPort:         c.String("esport"),
		ESIndex:        c.String("esindex"),
	}
}

// calculateLayers calculates the layers number used while fetching images.
// It returns an integer indicating amount of layers.
func calculateLayers(ctx context.Context, layers int, validLayers []int) int {
	if layers == (-1) {
		index := rand.Intn(len(validLayers) - 1)
		return validLayers[1+index]
	} else {
		return layers
	}
}

// containerName generates the container name string.
// It returns a string.
func containerName(prefix string, nLayers, nTag int) string {
	return prefix + "_layers_" + strconv.Itoa(nLayers) + "_tag_" + strconv.Itoa(nTag)
}

// getContainersList returns list of containers from test repo used in load phase.
// It returns a list of strings which is a list of container names.
func getContainersList(ctx context.Context, testRepoPrefix []string, hitSize, layers int, validLayers []int) []string {
	var containers []string
	testRepos := len(testRepoPrefix)
	imagesPerRepo := (hitSize / testRepos)
	leftOverImages := (hitSize % testRepos)
	nLayers := calculateLayers(ctx, layers, validLayers)

	if imagesPerRepo > 0 {
		for _, repoPrefix := range testRepoPrefix {
			for i := 1; i <= imagesPerRepo; i++ {
				containers = append(containers, containerName(repoPrefix, nLayers, i))
			}
		}
	}

	idx := 0
	for idx < leftOverImages {
		containers = append(containers, containerName(testRepoPrefix[idx], nLayers, imagesPerRepo+1))
		idx++
	}
	return containers
}

// reportAction drives the report action logic.
// It returns an error if any during the execution.
func reportAction(c *cli.Context) error {
	startTime := time.Now()
	ctx := c.Context
	conf := NewConfig(c)
	if c.String("testrepoprefix") != "" {
		conf.Containers = getContainersList(ctx, conf.TestRepoPrefix, conf.HitSize, conf.Layers, validLayers)
	}
	if len(conf.Containers) > conf.HitSize {
		conf.Containers = conf.Containers[:conf.HitSize]
	}
	jwt_token, err := CreateToken(conf.PSK)
	if err != nil {
		zlog.Debug(ctx).Str("PSK", conf.PSK).Msg("creating token")
		return fmt.Errorf("could not create token: %w", err)
	}

	zlog.Debug(ctx).Msg("Fetching manifests for an actual workload")
	listOfManifests, listOfManifestHashes := manifests.GetManifest(ctx, conf.Containers, conf.Concurrency)
	zlog.Info(ctx).Msg("🔥 Orchestrating the workload")
	err = orchestrateWorkload(ctx, listOfManifests, listOfManifestHashes, jwt_token, conf)
	if err != nil {
		return err
	}
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	zlog.Info(ctx).Stringer("duration", elapsedTime).Msg("Total time taken for completion")
	return nil
}

// orchestrateWorkload triggers the api endpoint hits and writes results to the desired location.
// It returns an error if any during the execution.
func orchestrateWorkload(ctx context.Context, manifests [][]byte, manifestHashes []string, jwt_token string, conf *TestConfig) error {
	zlog.Info(ctx).Str("RUNID", conf.RUNID).Msg("Run details")
	var requests []map[string]interface{}
	var err error
	attackMap := map[string]string{
		"RUNID":       conf.RUNID,
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

	zlog.Info(ctx).Str("RUNID", conf.RUNID).Msg("👋 Exiting clair-load-test")
	return nil
}
