package attacker

import (
	"os"
	"fmt"
	"bytes"
	"strings"
	"strconv"
	"os/exec"
	"encoding/base64"
	"encoding/json"
	"github.com/vishnuchalla/clair-load-test/pkg/utils"
)

// Method to generate requests string to feed it as input to vegeta.
func generateVegetaRequests(requestDicts []map[string]interface{}) (string) {
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
	return requests
}

// Method to write vegeta results to log file.
func writeVegetaResults(Uuid, testName string, vegetaOutput []byte) (string) {
	// Ensure a directory exists for writing vegeta results
	log_directory := "./logs"
	if _, err := os.Stat(log_directory); os.IsNotExist(err) {
		if err := os.MkdirAll(log_directory, 0755); err != nil {
			// Handle the error here
			panic(err)
		}
	}
	// Write Vegeta Stats to a file
	result_filename := fmt.Sprintf("%s/%s_%s_result.json", log_directory, Uuid, testName)
	cmd := exec.Command("vegeta", "report", "--every=1s", "--type=json", fmt.Sprintf("--output=%s", result_filename))
	cmd.Stdin = strings.NewReader(string(vegetaOutput))
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Results for test %s written to file: %s\n", testName, result_filename)
	return result_filename
}

// Method to index vegeta results using snafu wrapper.
func indexVegetaResults(resultFileName, testName string, conf *utils.TestConfig){
	// Use Snafu to push results to Elasticsearch
	fmt.Printf("Recording test results into ElasticSearch: %s\n", conf.ESHost)
	cmd := exec.Command("run_snafu",
		"-t", "vegeta",
		"-u", conf.Uuid,
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
		panic(err)
	}
	fmt.Printf("stdout: %s", stdout.String())
	fmt.Printf("stderr: %s", stderr.String())
}

// Method to trigger vegeta workflow.
func RunVegeta(requestDicts []map[string]interface{}, testName string, conf *utils.TestConfig) {
	requests := generateVegetaRequests(requestDicts)
	// Run `vegeta attack` to execute the HTTP Requests
	cmd := exec.Command("vegeta", "attack", "-lazy", "-format=json", "-rate", strconv.Itoa(conf.Concurrency), "-insecure")
	cmd.Stdin = strings.NewReader(requests)
	vegetaOutput, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	// Show Vegeta Stats
	cmd = exec.Command("vegeta", "report")
	cmd.Stdin = strings.NewReader(string(vegetaOutput))
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("Vegeta attack completed successfully")
	resultFileName := writeVegetaResults(conf.Uuid, testName, vegetaOutput)
	indexVegetaResults(resultFileName, testName, conf)
}