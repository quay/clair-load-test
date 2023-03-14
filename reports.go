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

	fmt.Println("Request in body bytes")
	fmt.Println(req.Body)
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// handle error
	}
	req.Body.Close()
	fmt.Println("After read io")
	fmt.Println(bodyBytes)

	requestData := RequestData{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers:  req.Header,
		Body:    bodyBytes,
	}
	
	file, err := os.Create("requests.json")
	if err != nil {
		// handle error
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	err = encoder.Encode(requestData)
	if err != nil {
		// handle error
	}
	
	cmd := exec.Command("vegeta", "attack", "-targets", "requests.json")
	err = cmd.Run()
	if err != nil {
		// handle error
	}
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
	var requests []string
	for _, reqDict := range requestDicts {
		req := make(map[string]interface{})
		req["url"] = reqDict["url"]
		req["method"] = reqDict["method"]

		// Encode the body as base64-encoded JSON if the request method is not GET
		if body, ok := reqDict["body"].(interface{}).(*bytes.Buffer); ok && reqDict["method"] != "GET" {
			// jsonData, err := json.Marshal(body)
			// if err != nil {
			// 	fmt.Fprintf(os.Stderr, "Error converting body into base64 json: %v\n", err)
			// 	os.Exit(1)
			// }
			req["body"] = base64.StdEncoding.EncodeToString(body.Bytes())
			// req["body"] = body
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
		fmt.Println(req)
		reqString, err := json.Marshal(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("After json marshal")
		fmt.Println(reqString)
		requests = append(requests, string(reqString))
	}

	// Write the Vegeta requests to a temporary file
	reqFile, err := ioutil.TempFile("", "vegeta-requests-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary file for requests: %v\n", err)
		os.Exit(1)
	}
	//defer os.Remove(reqFile.Name())
	_, err = reqFile.WriteString(strings.Join(requests, "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing requests to temporary file: %v\n", err)
		os.Exit(1)
	}

	// Run the Vegeta attack and capture the output
	vegetaCmd := exec.Command(
		"vegeta",
		"attack",
		"-lazy",
		"-format=json",
		"-rate", "50",
	)
	output, err := vegetaCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running Vegeta attack: %v\n", err)
		os.Exit(1)
	}

	// Show Vegeta stats
	reportCmd := exec.Command("vegeta", "report")
	reportCmd.Stdin = strings.NewReader(string(output))
	reportOutput, err := reportCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running Vegeta report: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(reportOutput))

	// Write Vegeta stats to a file
	resultFile, err := ioutil.TempFile("", "vegeta-results-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary file for results: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(resultFile)
}
