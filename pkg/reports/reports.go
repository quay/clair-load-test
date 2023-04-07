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
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/vishnuchalla/clair-load-test/pkg/token"
	"github.com/vishnuchalla/clair-load-test/pkg/manifests"
)

var logout zerolog.Logger

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
	fmt.Println(listOfManifestHashes)
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
	return
}

// func (r *Reporter) reportForContainer(ctx context.Context, container string, delete bool) error {
// 	// Get a token
// 	logout.Debug().Str("container", container).Msg("got manifest")
// 	token, err := token.CreateToken(r.Psk)
// 	if err != nil {
// 		zlog.Debug(ctx).Str("PSK", r.Psk).Msg("creating token")
// 		return fmt.Errorf("could not create token: %w", err)
// 	}
// 	// Send manifest as body to index_report
// 	hash, err := r.createIndexReport(ctx, manifest, token)
// 	if err != nil {
// 		return fmt.Errorf("could not create index report: %w", err)
// 	}

// 	// Request index report
// 	err = r.getIndexReport(ctx, hash, token)
// 	if err != nil {
// 		return fmt.Errorf("could not get index report: %w", err)
// 	}

// 	// Get a token
// 	// Request vuln report
// 	err = r.getVulnerabilityReport(ctx, hash, token)
// 	if err != nil {
// 		return fmt.Errorf("could not get vulnerability report: %w", err)
// 	}
// 	// Delete index_report
// 	if delete {
// 		err = r.deleteIndexReports(ctx, hash, token)
// 		if err != nil {
// 			return fmt.Errorf("could not delete index report: %w", err)
// 		}
// 	}

// 	// Request indexer state
// 	err = r.getIndexerState(ctx, token)
// 	if err != nil {
// 		return fmt.Errorf("could not get index report: %w", err)
// 	}
// 	return nil
// }

// func (r *Reporter) createIndexReport(ctx context.Context, body []byte, token string) (string, error) {
//     url := r.Host + "/indexer/api/v1/index_report"
// 	headers := map[string][]string{
// 		"Content-Type": {"application/json"},
// 		"Authorization": {fmt.Sprintf("Bearer %s", token)},
// 	}
// 	err := json.Unmarshal(body, &blob)
// 	if err != nil {
// 		fmt.Println("Handling error here")
// 	}
// 	var vegetaData []map[string]interface{}
// 	vegetaData = append(vegetaData, map[string]interface{}{
// 		"method":  "POST",
// 		"url":     url,
// 		"header": headers,
// 		"body":    body,
// 	})

// 	run_vegeta(vegetaData,"post_index_report")
// 	return "nil", nil
// }

// func (r *Reporter) getIndexReport(ctx context.Context, hash string, token string) error {
// 	url := r.Host+"/indexer/api/v1/index_report/"+hash
// 	headers := map[string][]string{
// 		"Content-Type": {"application/json"},
// 		"Authorization": {fmt.Sprintf("Bearer %s", token)},
// 	}
// 	zlog.Debug(ctx).Str("hash", hash).Msg("getting index report")
// 	var vegetaData []map[string]interface{}
// 	vegetaData = append(vegetaData, map[string]interface{}{
// 		"method":  "GET",
// 		"url":     url,
// 		"header": headers,
// 	})

// 	run_vegeta(vegetaData, "get_index_report")
// 	return nil
// }

// func (r *Reporter) getIndexerState(ctx context.Context, token string) error {
// 	url := r.Host+"/indexer/api/v1/index_state"
// 	headers := map[string][]string{
// 		"Content-Type": {"application/json"},
// 		"Authorization": {fmt.Sprintf("Bearer %s", token)},
// 	}
// 	zlog.Debug(ctx).Msg("getting indexer state")
// 	var vegetaData []map[string]interface{}
// 	vegetaData = append(vegetaData, map[string]interface{}{
// 		"method":  "GET",
// 		"url":     url,
// 		"header": headers,
// 	})

// 	run_vegeta(vegetaData, "get_indexer_state")
// 	return nil
// }

// func (r *Reporter) getVulnerabilityReport(ctx context.Context, hash string, token string) error {
// 	url := r.Host+"/matcher/api/v1/vulnerability_report/"+hash
// 	headers := map[string][]string{
// 		"Content-Type": {"application/json"},
// 		"Authorization": {fmt.Sprintf("Bearer %s", token)},
// 	}
// 	zlog.Debug(ctx).Str("hash", hash).Msg("getting vulnerability report")
// 	var vegetaData []map[string]interface{}
// 	vegetaData = append(vegetaData, map[string]interface{}{
// 		"method":  "GET",
// 		"url":     url,
// 		"header": headers,
// 	})

// 	run_vegeta(vegetaData, "get_vulnerability_report")
// 	return nil
// }

// func (r *Reporter) deleteIndexReports(ctx context.Context, hash string, token string) error {
// 	url := r.Host+"/indexer/api/v1/index_report/"+hash
// 	headers := map[string][]string{
// 		"Content-Type": {"application/json"},
// 		"Authorization": {fmt.Sprintf("Bearer %s", token)},
// 	}
// 	zlog.Debug(ctx).Str("hash", hash).Msg("deleting index report")
// 	var vegetaData []map[string]interface{}
// 	vegetaData = append(vegetaData, map[string]interface{}{
// 		"method":  "DELETE",
// 		"url":     url,
// 		"header": headers,
// 	})

// 	run_vegeta(vegetaData, "delete_vulnerability_report")
// 	return nil
// }

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
