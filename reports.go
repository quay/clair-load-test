package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/quay/zlog"
	// "github.com/tsenart/vegeta/lib"
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
		&cli.DurationFlag{
			Name:    "timeout",
			Usage:   "--timeout 1m",
			Value:   time.Minute * 1,
			EnvVars: []string{"TIMEOUT"},
		},
		&cli.Float64Flag{
			Name:    "rate",
			Usage:   "--rate 1",
			Value:   1,
			EnvVars: []string{"RATE"},
		},
	},
}

type IndexReportReponse struct {
	Hash string `json:"manifest_hash"`
}

type testConfig struct {
	Containers []string      `json:"containers"`
	PSK        string        `json:"-"`
	Host       string        `json:"host"`
	Delete     bool          `json:"delete"`
	Timeout    time.Duration `json:"timeout"`
	PerSecond  float64       `json:"rate"`
}

func NewConfig(c *cli.Context) *testConfig {
	containersArg := c.String("containers")
	return &testConfig{
		Containers: strings.Split(containersArg, ","),
		PSK:        c.String("psk"),
		Host:       c.String("host"),
		Delete:     c.Bool("delete"),
		Timeout:    c.Duration("timeout"),
		PerSecond:  c.Float64("rate"),
	}
}

type reporter struct {
	host  string
	psk   string
	cl    *http.Client
}

func NewReporter(host, psk string) *reporter {
	return &reporter{
		host:  host,
		psk:   psk,
		cl:    &http.Client{Timeout: time.Minute * 1},
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

	fmt.Println("%v", conf)
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
	token, err := createToken(r.psk)
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

type RequestData struct {
    Method  string 		`json:"method"`
    URL     string		`json:"url"`
    Headers http.Header	`json:"header"`
    Body    []byte		`json:"body"`
}

func (r *reporter) createIndexReport(ctx context.Context, body []byte, token string) (string, error) {
    url := r.host + "/indexer/api/v1/index_report"
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// handle error
	}
	req.Body.Close()

	var vegetaData []map[string]interface{}
	vegetaData = append(vegetaData, map[string]interface{}{
		"method":  req.Method,
		"url":     req.URL.String(),
		"header": req.Header,
		"body":    bodyBytes,
	})

	run_vegeta(vegetaData)
	return "", nil
}

func (r *reporter) getVulnerabilityReport(ctx context.Context, hash string, token string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		r.host+"/matcher/api/v1/vulnerability_report/"+hash,
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	fmt.Printf("In getVulnerabilityReport Request:- %v\n", req)
	// Start clock
	t := time.Now()
	resp, err := r.cl.Do(req)
	// end clock and report
	diff := time.Now().Sub(t)
	fmt.Println(diff)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (r *reporter) deleteIndexReports(ctx context.Context, hash string, token string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete,
		r.host+"/indexer/api/v1/index_report/"+hash,
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	zlog.Debug(ctx).Str("hash", hash).Msg("deleting index report")
	fmt.Printf("In deleteIndexReport Request:- %v\n", req)
	resp, err := r.cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}


func run_vegeta(requestDicts []map[string]interface{}) {

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
		os.MkdirAll(log_directory, 0755)
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
	result_filename := fmt.Sprintf("%s/%s_%s_result.json", log_directory, "TEST_UUID", "test_name")
	cmd = exec.Command("vegeta", "report", "--every=1s", "--type=json", fmt.Sprintf("--output=%s", result_filename))
	cmd.Stdin = strings.NewReader(string(output))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Results for test %s written to file: %s\n", "test_name", result_filename)

	// Use Snafu to push results to Elasticsearch
	fmt.Printf("Recording test results in ElasticSearch: %s\n", "https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com")
	cmd = exec.Command("run_snafu",
		"-t", "vegeta",
		"-u", "TEST_UUID",
		"-w", fmt.Sprintf("%d", 100),
		"-r", result_filename,
		"--target_name", "test_name",
	)
	cmd.Env = []string{
		fmt.Sprintf("es=%s", "https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com"),
		fmt.Sprintf("es_port=%s", "443"),
		fmt.Sprintf("es_index=%s", "cliar-test-index"),
		fmt.Sprintf("clustername=%s", "example-registry-clair-app-quay-enterprise.apps.vchalla-quay-test.perfscale.devcluster.openshift.com"),
	}
	output, err = cmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", output)
	if cmd.ProcessState.ExitCode() != 0 {
		panic("command failed")
	}
}
