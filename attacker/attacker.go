package attacker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/quay/zlog"
	"github.com/vishnuchalla/clair-load-test/pkg/commons"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// generateVegetaRequests generates requests string which can be fed as input to vegeta for HTTP benchmarking.
// It return a consolidated string which has list of request jsons fed all at once to vegeta.
func generateVegetaRequests(requestDicts []map[string]interface{}) string {
	// Convert requestDicts to a slice of Vegeta requests
	var requests string = ""
	for _, reqDict := range requestDicts {
		req := make(map[string]interface{})
		req["url"] = reqDict["url"]
		req["method"] = reqDict["method"]
		// Encode the body as base64-encoded JSON if the request method is not GET
		if reqDict["body"] != nil && reqDict["method"] != http.MethodGet {
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
	return requests
}

// writeVegetaResults writes all the vegeta output to the specified log directory.
// It returns the filename where results are logged and errors if any.
func writeVegetaResults(ctx context.Context, UUID, testName string, vegetaOutput []byte) (string, error) {
	// Ensure a directory exists for writing vegeta results
	log_directory := "./logs"
	if _, err := os.Stat(log_directory); os.IsNotExist(err) {
		if err := os.MkdirAll(log_directory, 0755); err != nil {
			return "", err
		}
	}
	// Write Vegeta Stats to a file
	result_filename := fmt.Sprintf("%s/%s_%s_result.json", log_directory, UUID, testName)
	cmd := exec.Command("vegeta", "report", "--every=1s", "--type=json", fmt.Sprintf("--output=%s", result_filename))
	cmd.Stdin = strings.NewReader(string(vegetaOutput))
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	zlog.Debug(ctx).Msg(fmt.Sprintf("Results for test %s written to file: %s\n", testName, result_filename))
	return result_filename, nil
}

// indexVegetaResults uses snafu to process vegeta output and index it to elastic search.
// It returns an error if any during the execution.
func indexVegetaResults(ctx context.Context, resultFileName, testName string, conf *commons.TestConfig) error {
	// Use Snafu to push results to Elasticsearch
	fmt.Printf("Recording test results into ElasticSearch: %s\n", conf.ESHost)
	cmd := exec.Command("run_snafu",
		"-t", "vegeta",
		"-u", conf.UUID,
		"-w", strconv.Itoa(conf.Concurrency),
		"-r", resultFileName,
		"--target_name", testName,
	)
	cmd.Env = []string{
		fmt.Sprintf("es=%s", conf.ESHost),
		fmt.Sprintf("es_port=%s", conf.ESPort),
		fmt.Sprintf("es_index=%s", conf.ESIndex),
		fmt.Sprintf("clustername=%s", conf.Host),
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	zlog.Debug(ctx).Msg(fmt.Sprintf("stdout: %s", stdout.String()))
	zlog.Debug(ctx).Msg(fmt.Sprintf("stderr: %s", stderr.String()))
	return nil
}

// RunVegeta runs vegeta, records their results and indexes to elastic search if provided with connection details.
// It returns an error if any during the execution.
func RunVegeta(ctx context.Context, requestDicts []map[string]interface{}, testName string, conf *commons.TestConfig) error {
	var err error
	requests := generateVegetaRequests(requestDicts)
	// Run `vegeta attack` to execute the HTTP Requests
	cmd := exec.Command("vegeta", "attack", "-lazy", "-format=json", "-rate", strconv.Itoa(conf.Concurrency), "-insecure")
	cmd.Stdin = strings.NewReader(requests)
	vegetaOutput, err := cmd.Output()
	if err != nil {
		zlog.Debug(ctx).Msg("vegeta attack command failure")
		return err
	}
	// Show Vegeta Stats
	cmd = exec.Command("vegeta", "report")
	cmd.Stdin = strings.NewReader(string(vegetaOutput))
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		zlog.Debug(ctx).Msg("vegeta report command failure")
		return err
	}
	zlog.Debug(ctx).Msg("Vegeta attack completed successfully")
	resultFileName, err := writeVegetaResults(ctx, conf.UUID, testName, vegetaOutput)
	if err != nil {
		zlog.Debug(ctx).Msg("Failed writing results to log dir")
		return err
	}
	if conf.ESHost != "" && conf.ESPort != "" && conf.ESIndex != "" {
		err = indexVegetaResults(ctx, resultFileName, testName, conf)
		if err != nil {
			zlog.Debug(ctx).Msg("Failed to indexing results to elastic search")
			return err
		}
	}
	return nil
}
