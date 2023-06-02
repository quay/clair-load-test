package attacker

import (
	// "bytes"
	"context"
	"encoding/base64"
	// "encoding/json"
	"fmt"
	"net/http"
	"os"
	// "os/exec"
	// "strings"
	"strconv"
	"time"

	// "github.com/quay/zlog"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// generateVegetaRequests generates requests string which can be fed as input to vegeta for HTTP benchmarking.
// It return a consolidated string which has list of request jsons fed all at once to vegeta.
func generateVegetaRequests(requestDicts []map[string]interface{}) ([]vegeta.Target, error) {
	// Convert requestDicts to a slice of Vegeta requests
	var targets []vegeta.Target
	for _, reqDict := range requestDicts {
		// Encode the body as base64-encoded JSON if the request method is not GET
		req_body := ""
		var req_headers http.Header
		if reqDict["body"] != nil && reqDict["method"] != http.MethodGet {
			if body, ok := reqDict["body"].([]byte); ok {
				req_body = base64.StdEncoding.EncodeToString(body)
			}
		}
		// Set the request headers
		if headers, ok := reqDict["header"]; ok {
			req_headers = headers.(http.Header)
		} else {
			req_headers = http.Header{
				"Authorization": []string{"Bearer " + os.Getenv("AUTH_TOKEN")},
				"Content-Type":  []string{"application/json"},
			}
		}
		target := vegeta.Target{
			Method: reqDict["method"].(string),
			URL:    reqDict["url"].(string),
			Header: req_headers,
			Body:   []byte(req_body),
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// // writeVegetaResults writes all the vegeta output to the specified log directory.
// // It returns the filename where results are logged and errors if any.
// func writeVegetaResults(ctx context.Context, UUID, testName string, vegetaOutput []byte) (string, error) {
// 	// Ensure a directory exists for writing vegeta results
// 	log_directory := "./logs"
// 	if _, err := os.Stat(log_directory); os.IsNotExist(err) {
// 		if err := os.MkdirAll(log_directory, 0755); err != nil {
// 			return "", err
// 		}
// 	}
// 	// Write Vegeta Stats to a file
// 	result_filename := fmt.Sprintf("%s/%s_%s_result.json", log_directory, UUID, testName)
// 	cmd := exec.Command("vegeta", "report", "--every=1s", "--type=json", fmt.Sprintf("--output=%s", result_filename))
// 	cmd.Stdin = strings.NewReader(string(vegetaOutput))
// 	err := cmd.Run()
// 	if err != nil {
// 		return "", err
// 	}
// 	zlog.Debug(ctx).Msg(fmt.Sprintf("Results for test %s written to file: %s\n", testName, result_filename))
// 	return result_filename, nil
// }

// // indexVegetaResults uses snafu to process vegeta output and index it to elastic search.
// // It returns an error if any during the execution.
// func indexVegetaResults(ctx context.Context, resultFileName, testName string, attackMap map[string]string) error {
// 	// Use Snafu to push results to Elasticsearch
// 	fmt.Printf("Recording test results into ElasticSearch: %s\n", attackMap["ESHost"])
// 	cmd := exec.Command("run_snafu",
// 		"-t", "vegeta",
// 		"-u", attackMap["UUID"],
// 		"-w", attackMap["Concurrency"],
// 		"-r", resultFileName,
// 		"--target_name", testName,
// 	)
// 	cmd.Env = []string{
// 		fmt.Sprintf("es=%s", attackMap["ESHost"]),
// 		fmt.Sprintf("es_port=%s", attackMap["ESPort"]),
// 		fmt.Sprintf("es_index=%s", attackMap["ESIndex"]),
// 		fmt.Sprintf("clustername=%s", attackMap["Host"]),
// 	}
// 	var stdout bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return err
// 	}
// 	zlog.Debug(ctx).Msg(fmt.Sprintf("stdout: %s", stdout.String()))
// 	zlog.Debug(ctx).Msg(fmt.Sprintf("stderr: %s", stderr.String()))
// 	return nil
// }

// RunVegeta runs vegeta, records their results and indexes to elastic search if provided with connection details.
// It returns an error if any during the execution.
func RunVegeta(ctx context.Context, requestDicts []map[string]interface{}, testName string, attackMap map[string]string) error {
	// var err error
	requests, _ := generateVegetaRequests(requestDicts)
	concurrency, _ := strconv.Atoi(attackMap["Concurrency"])

	rate := vegeta.Rate{Freq: concurrency, Per: time.Second} // Set your desired rate
	duration := 0 * time.Second                              // Set your desired duration
	targeter := vegeta.NewStaticTargeter(requests...)
	attacker := vegeta.NewAttacker()
	var metrics vegeta.Metrics

	totalRequests := len(requests)
	completedRequests := 0
	stop := make(chan struct{}) // Channel to stop the attack

	// Run the attack in a separate goroutine
	go func() {
		for res := range attacker.Attack(targeter, rate, duration, "Vegeta") {
			metrics.Add(res)

			completedRequests++
			if completedRequests == totalRequests {
				stop <- struct{}{} // Send the stop signal if all requests are completed
			}
		}
	}()

	// Wait for the stop signal
	<-stop

	report := vegeta.NewTextReporter(&metrics)
	err := report.Report(os.Stdout)
	if err != nil {
		fmt.Println("Error while generating vegeta report:", err)
	}

	fmt.Println("Vegeta attack completed successfully")

	// for res := range attacker.Attack(targeter, rate, duration, "Vegeta") {
	// 	metrics.Add(res)
	// }
	// report := vegeta.NewTextReporter(&metrics)
	// err := report.Report(os.Stdout)
	// if err != nil {
	// 	fmt.Println("Error while generating vegeta report", err)
	// }
	// fmt.Println("Vegeta attack completed successfully")
	// // Run `vegeta attack` to execute the HTTP Requests
	// cmd := exec.Command("vegeta", "attack", "-lazy", "-format=json", "-rate", attackMap["Concurrency"], "-insecure")
	// cmd.Stdin = strings.NewReader(requests)
	// vegetaOutput, err := cmd.Output()
	// if err != nil {
	// 	zlog.Debug(ctx).Msg("vegeta attack command failure")
	// 	return err
	// }
	// // Show Vegeta Stats
	// cmd = exec.Command("vegeta", "report")
	// cmd.Stdin = strings.NewReader(string(vegetaOutput))
	// cmd.Stdout = os.Stdout
	// err = cmd.Run()
	// if err != nil {
	// 	zlog.Debug(ctx).Msg("vegeta report command failure")
	// 	return err
	// }
	// zlog.Debug(ctx).Msg("Vegeta attack completed successfully")
	// resultFileName, err := writeVegetaResults(ctx, attackMap["UUID"], testName, vegetaOutput)
	// if err != nil {
	// 	zlog.Debug(ctx).Msg("Failed writing results to log dir")
	// 	return err
	// }
	// if attackMap["ESHost"] != "" && attackMap["ESPort"] != "" && attackMap["ESIndex"] != "" {
	// 	err = indexVegetaResults(ctx, resultFileName, testName, attackMap)
	// 	if err != nil {
	// 		zlog.Debug(ctx).Msg("Failed to indexing results to elastic search")
	// 		return err
	// 	}
	// }
	// return nil
	return nil
}
