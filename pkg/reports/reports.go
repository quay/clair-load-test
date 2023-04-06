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
	"golang.org/x/sync/errgroup"
	"github.com/quay/zlog"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/vishnuchalla/clair-load-test/pkg/token"
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
			EnvVars: []string{"CLAIR_HOST"},
		},
		&cli.StringFlag{
			Name:    "containers",
			Usage:   "--containers ubuntu:latest,mysql:latest",
			Value:   "",
			EnvVars: []string{"CONTAINERS"},
		},
		&cli.StringFlag{
			Name:    "psk",
			Usage:   "--psk secretkey",
			Value:   "",
			EnvVars: []string{"PSK"},
		},
		&cli.BoolFlag{
			Name:    "delete",
			Usage:   "--delete",
			Value:   false,
			EnvVars: []string{"DELETE"},
		},
		&cli.Float64Flag{
			Name:    "rate",
			Usage:   "--rate 1",
			Value:   1,
			EnvVars: []string{"RATE"},
		},
	},
}

type testConfig struct {
	Containers []string      `json:"containers"`
	PSK        string        `json:"-"`
	Host       string        `json:"host"`
	Delete     bool          `json:"delete"`
	PerSecond  float64       `json:"rate"`
}

type Blob struct {
    Hash string `json:"hash"`
}

func NewConfig(c *cli.Context) *testConfig {
	containersArg := c.String("containers")
	return &testConfig{
		Containers: strings.Split(containersArg, ","),
		PSK:        c.String("psk"),
		Host:       c.String("host"),
		Delete:     c.Bool("delete"),
		PerSecond:  c.Float64("rate"),
	}
}

type reporter struct {
	host  string
	psk   string
}

func NewReporter(host, psk string) *reporter {
	return &reporter{
		host:  host,
		psk:   psk,
	}
}

func reportAction(c *cli.Context) error {
	ctx := c.Context
	conf := NewConfig(c)

	reporter := NewReporter(conf.Host, conf.PSK)

	g, ctx := errgroup.WithContext(ctx)
	for i := 0;  i < len(conf.Containers); i++ {
		cc := conf.Containers[i]
		g.Go(func() error {
			err := reporter.reportForContainer(ctx, cc, conf.Delete)
			if err != nil {
				zlog.Error(ctx).Str("container", cc).Msg(err.Error())
				return nil
			}
			zlog.Debug(ctx).Str("container", cc).Msg("completed")
			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err = enc.Encode(conf)
	if err != nil {
		return err
	}
	return nil
}

func (r *reporter) reportForContainer(ctx context.Context, container string, delete bool) error {
	// Call clairctl for the manifest
	manifest, err := getManifest(ctx, container)
	if err != nil {
		return fmt.Errorf("could not generate manifest: %w", err)
	}
	// Get a token
	logout.Debug().Str("container", container).Msg("got manifest")
	token, err := token.CreateToken(r.psk)
	if err != nil {
		zlog.Debug(ctx).Str("PSK", r.psk).Msg("creating token")
		return fmt.Errorf("could not create token: %w", err)
	}
	// Send manifest as body to index_report
	hash, err := r.createIndexReport(ctx, manifest, token)
	if err != nil {
		return fmt.Errorf("could not create index report: %w", err)
	}
	// Get a token
	// Request vuln report
	err = r.getVulnerabilityReport(ctx, hash, token)
	if err != nil {
		return fmt.Errorf("could not get vulnerability report: %w", err)
	}
	// Delete index_report
	if delete {
		err = r.deleteIndexReports(ctx, hash, token)
		if err != nil {
			return fmt.Errorf("could not delete index report: %w", err)
		}
	}
	return nil
}

func getManifest(ctx context.Context, container string) ([]byte, error) {
	cmd := exec.Command("clairctl", "manifest", container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("getting manifest")
	return cmd.Output()
}

func (r *reporter) createIndexReport(ctx context.Context, body []byte, token string) (string, error) {
    url := r.host + "/indexer/api/v1/index_report"
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	var blob Blob
	err := json.Unmarshal(body, &blob)
	if err != nil {
		fmt.Println("Handling error here")
	}
	hashToReturn := blob.Hash
	var vegetaData []map[string]interface{}
	vegetaData = append(vegetaData, map[string]interface{}{
		"method":  "POST",
		"url":     url,
		"header": headers,
		"body":    body,
	})

	run_vegeta(vegetaData,"post_index_report")
	return hashToReturn, nil
}

func (r *reporter) getVulnerabilityReport(ctx context.Context, hash string, token string) error {
	url := r.host+"/matcher/api/v1/vulnerability_report/"+hash
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	zlog.Debug(ctx).Str("hash", hash).Msg("getting vulnerability report")
	var vegetaData []map[string]interface{}
	vegetaData = append(vegetaData, map[string]interface{}{
		"method":  "GET",
		"url":     url,
		"header": headers,
	})

	run_vegeta(vegetaData, "get_vulnerability_report")
	return nil
}

func (r *reporter) deleteIndexReports(ctx context.Context, hash string, token string) error {
	url := r.host+"/indexer/api/v1/index_report/"+hash
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	zlog.Debug(ctx).Str("hash", hash).Msg("deleting index report")
	var vegetaData []map[string]interface{}
	vegetaData = append(vegetaData, map[string]interface{}{
		"method":  "DELETE",
		"url":     url,
		"header": headers,
	})

	run_vegeta(vegetaData, "delete_vulnerability_report")
	return nil
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
