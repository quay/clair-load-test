package reports

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
	"github.com/vishnuchalla/clair-load-test/pkg/token"
	"github.com/vishnuchalla/clair-load-test/pkg/manifests"
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
			EnvVars: []string{"CLAIR_TEST_HOST"},
		},
		&cli.StringFlag{
			Name:    "containers",
			Usage:   "--containers ubuntu:latest,mysql:latest",
			Value:   "",
			EnvVars: []string{"CLAIR_TEST_CONTAINERS"},
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
			EnvVars: []string{"INDEX_REPORT_DELETE"},
		},
		&cli.Float64Flag{
			Name:    "hitsize",
			Usage:   "--hitsize 100",
			Value:   25,
			EnvVars: []string{"CLAIR_TEST_HIT_SIZE"},
		},
		&cli.Float64Flag{
			Name:    "concurrency",
			Usage:   "--concurrency 50",
			Value:   10,
			EnvVars: []string{"CLAIR_TEST_CONCURRENCY"},
		},
	},
}

// Method to create a test configuration from CLI options.
func NewConfig(c *cli.Context) *TestConfig {
	containersArg := c.String("containers")
	return &TestConfig{
		Containers: strings.Split(containersArg, ","),
		Psk:        c.String("psk"),
		Host:       c.String("host"),
		IndexDelete:     c.Bool("delete"),
		HitSize:	c.Float64("hitsize"),
		Concurrency:  c.Float64("concurrency"),
		ESHost: c.String("eshost"),
		ESPort: c.String("esport"),
		ESIndex: c.String("esindex"),
	}
}

func reportAction(c *cli.Context) error {
	ctx := c.Context
	conf := NewConfig(c)
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

func orchestrateWorkload(ctx context.Context, manifests [][]byte, manifestHashes []string, jwt_token string, conf *TestConfig) {
	zlog.Debug(ctx).Msg("Orchestrating reports workload")
	var requests []map[string]interface{}
	var testName string
	requests, testName = createIndexReport(ctx, manifests, conf.Host, jwt_token)
	run_vegeta(requests, testName)

	requests, testName = getIndexReport(ctx, manifestHashes, conf.Host, jwt_token)
	run_vegeta(requests, testName)

	requests, testName = getVulnerabilityReport(ctx, manifestHashes, conf.Host, jwt_token)
	run_vegeta(requests, testName)

	requests, testName = getIndexerState(ctx, conf.HitSize, conf.Host, jwt_token)
	run_vegeta(requests, testName)

	if (conf.IndexDelete) {
		requests, testName = deleteIndexReports(ctx, manifestHashes, conf.Host, jwt_token)
		run_vegeta(requests, testName)
	}
}

func getRequestCommons(endpoint, host, token string) (string, map[string][]string) {
	url := host + endpoint
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	return url, headers
}

func createIndexReport(ctx context.Context, manifests [][]byte, host, token string) ([]map[string]interface{}, string) {
    url, headers := getRequestCommons("/indexer/api/v1/index_report", host, token)
	var requests []map[string]interface{}
	for _, manifest := range manifests {
		requests = append(requests, map[string]interface{}{
			"method":  "POST",
			"url":     url,
			"header": headers,
			"body":    manifest,
		})
	}
	return requests, "post_index_report"
}

func getIndexReport(ctx context.Context, manifestHashes []string, host, token string) ([]map[string]interface{}, string) {
	url, headers := getRequestCommons("/indexer/api/v1/index_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method":  "GET",
			"url":     url + manifestHash,
			"header": headers,
		})
	}
	return requests, "get_index_report"
}

func deleteIndexReports(ctx context.Context, manifestHashes []string, host, token string) ([]map[string]interface{}, string) {
	url, headers := getRequestCommons("/indexer/api/v1/index_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method":  "DELETE",
			"url":     url + manifestHash,
			"header": headers,
		})
	}
	return requests, "delete_index_report"
}

func getVulnerabilityReport(ctx context.Context, manifestHashes []string, host, token string) ([]map[string]interface{}, string) {
	url, headers := getRequestCommons("/matcher/api/v1/vulnerability_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method":  "GET",
			"url":     url + manifestHash,
			"header": headers,
		})
	}
	return requests, "get_vulnerability_report"
}

func getIndexerState(ctx context.Context, hitsize float64, host, token string) ([]map[string]interface{}, string) {
	url, headers := getRequestCommons("/indexer/api/v1/index_state", host, token)
	var requests []map[string]interface{}
	for i := 0; i < int(hitsize); i++ {
		requests = append(requests, map[string]interface{}{
			"method":  "GET",
			"url":     url,
			"header": headers,
		})
	}
	return requests, "get_indexer_state"
}

func run_vegeta(requestDicts []map[string]interface{}, testName string) {

	// Convert requestDicts to a slice of Vegeta requests
	var requests string = ""
	for _, reqDict := range requestDicts {
		req := make(map[string]interface{})
		req["url"] = reqDict["url"]
		req["method"] = reqDict["method"]

		// Encode the body as base64-encoded JSON if the request method is not GET
		if reqDict["body"] != nil && reqDict["method"] != "GET" {
			if body, ok := reqDict["body"].([]byte); ok {
				req["body"] = base64.StdEncoding.EncodeToString(body)
			}
		}

		// Set the request headers
		if headers, ok := reqDict["header"]; ok {
			req["header"] = headers
		} else {
			req["header"] = map[string][]string{
				"Authorization": {"Bearer " + os.Getenv("AUTH_TOKEN")},
				"Content-Type":  {"application/json"},
			}
		}
		reqString, err := json.Marshal(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
			os.Exit(1)
		}
		requests += string(reqString) + "\n"
	}

	if strings.TrimSpace(requests) == "" {
		panic("Something is wrong with requests. Requests string cannot be empty")
	}

	// Ensure a directory exists for writing vegeta results
	log_directory := "./logs"
	if _, err := os.Stat(log_directory); os.IsNotExist(err) {
		if err := os.MkdirAll(log_directory, 0755); err != nil {
			// Handle the error here
			panic(err)
		}
	}

	// Run `vegeta attack` to execute the HTTP Requests
	cmd := exec.Command("vegeta", "attack", "-lazy", "-format=json", "-rate", "100", "-insecure")
	cmd.Stdin = strings.NewReader(requests)
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// Show Vegeta Stats
	cmd = exec.Command("vegeta", "report")
	cmd.Stdin = strings.NewReader(string(output))
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("Vegeta attack completed successfully")

	// Write Vegeta Stats to a file
	result_filename := fmt.Sprintf("%s/%s_%s_result.json", log_directory, "TEST_UUID", testName)
	cmd = exec.Command("vegeta", "report", "--every=1s", "--type=json", fmt.Sprintf("--output=%s", result_filename))
	cmd.Stdin = strings.NewReader(string(output))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Results for test %s written to file: %s\n", testName, result_filename)

	// Use Snafu to push results to Elasticsearch
	fmt.Printf("Recording test results in ElasticSearch: %s\n", "https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com")
	cmd = exec.Command("run_snafu",
		"-t", "vegeta",
		"-u", "TEST_UUID",
		"-w", fmt.Sprintf("%d", 100),
		"-r", result_filename,
		"--target_name", testName,
	)
	cmd.Env = []string{
		fmt.Sprintf("es=%s", "https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com"),
		fmt.Sprintf("es_port=%s", "443"),
		fmt.Sprintf("es_index=%s", "cliar-test-index"),
		fmt.Sprintf("clustername=%s", "example-registry-clair-app-quay-enterprise.apps.vchalla-quay-test.perfscale.devcluster.openshift.com"),
	}
	var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        panic(err)
    }
    fmt.Printf("stdout: %s", stdout.String())
    fmt.Printf("stderr: %s", stderr.String())
}
