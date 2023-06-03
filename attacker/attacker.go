package attacker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloud-bulldozer/go-commons/indexers"
	"github.com/quay/zlog"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// generateVegetaRequests generates requests which can be fed as input to vegeta for HTTP benchmarking.
// It return a consolidated targets list which has all the requests fed all at once to vegeta.
func generateVegetaRequests(requestDicts []map[string]interface{}) []vegeta.Target {
	// Convert requestDicts to a slice of Vegeta requests
	var targets []vegeta.Target
	for _, reqDict := range requestDicts {
		var req_body []byte
		var req_headers http.Header
		// Prepare request body
		if reqDict["body"] != nil && reqDict["method"] != http.MethodGet {
			req_body, _ = reqDict["body"].([]byte)
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
		// Vegeta Target
		target := vegeta.Target{
			Method: reqDict["method"].(string),
			URL:    reqDict["url"].(string),
			Header: req_headers,
			Body:   req_body,
		}
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		panic("Something is wrong with requests. Requests list cannot be empty")
	}
	return targets
}

// indexVegetaResults to process vegeta output and index the results to elastic search.
// It returns an error if any during the execution.
func indexVegetaResults(ctx context.Context, metrics vegeta.Metrics, testName string, attackMap map[string]string) error {
	var indexer *indexers.Indexer
	indexerConfig := indexers.IndexerConfig{
		Type:               "opensearch",
		Servers:            []string{attackMap["ESHost"] + ":" + attackMap["ESPort"]},
		Index:              attackMap["ESIndex"],
		InsecureSkipVerify: true,
	}
	zlog.Debug(ctx).Msg(fmt.Sprintf("Creating indexer: %s", indexerConfig.Type))
	indexer, err := indexers.NewIndexer(indexerConfig)
	if err != nil {
		return fmt.Errorf("Failure while connnecting to Elasticsearch: %w", err)
	}
	zlog.Debug(ctx).Msg(fmt.Sprintf("Connected to : %s", indexerConfig.Servers[0]))
	concurrency, _ := strconv.Atoi(attackMap["Concurrency"])
	hostname, _ := os.Hostname()
	zlog.Debug(ctx).Msg(fmt.Sprintf("Indexing documents in %s", attackMap["ESIndex"]))
	resp, err := (*indexer).Index([]interface{}{Document{
		Workload:       "clair-load-test",
		Endpoint:       attackMap["Host"],
		RequestTimeout: 120,
		Targets:        testName,
		Hostname:       hostname,
		RPS:            concurrency,
		Throughput:     metrics.Throughput,
		StatusCodes:    metrics.StatusCodes,
		Requests:       metrics.Requests,
		P99Latency:     metrics.Latencies.P99,
		P95Latency:     metrics.Latencies.P95,
		MaxLatency:     metrics.Latencies.Max,
		MinLatency:     metrics.Latencies.Min,
		ReqLatency:     metrics.Latencies.Mean,
		Timestamp:      time.Now().Format("2006-01-02T15:04:05.999999Z07:00"),
		BytesIn:        metrics.BytesIn.Mean,
		BytesOut:       metrics.BytesOut.Mean,
		RunID:          attackMap["RUNID"],
	}}, indexers.IndexingOpts{})
	if err != nil {
		return fmt.Errorf(err.Error())
	} else {
		zlog.Debug(ctx).Msg(resp)
	}
	return nil
}

// RunVegeta runs vegeta, records their results and indexes to elastic search if provided with connection details.
// It returns an error if any during the execution.
func RunVegeta(ctx context.Context, requestDicts []map[string]interface{}, testName string, attackMap map[string]string) error {
	requests := generateVegetaRequests(requestDicts)
	concurrency, _ := strconv.Atoi(attackMap["Concurrency"])
	rate := vegeta.Rate{Freq: concurrency, Per: time.Second}
	duration := 0 * time.Second
	targeter := vegeta.NewStaticTargeter(requests...)
	attacker := vegeta.NewAttacker(vegeta.Timeout(120 * time.Second))
	totalRequests := len(requests)
	completedRequests := 0
	stop := make(chan struct{}) // Channel to stop the attack

	var metrics vegeta.Metrics
	go func() {
		for res := range attacker.Attack(targeter, rate, duration, "Vegeta Attack") {
			metrics.Add(res)
			completedRequests++
			if completedRequests == totalRequests {
				stop <- struct{}{} // Send the stop signal if all requests are completed
			}
		}
	}()
	// Wait for the stop signal
	<-stop
	metrics.Close()

	// Generate Vegeta text report
	report := vegeta.NewTextReporter(&metrics)
	err := report.Report(os.Stdout)
	if err != nil {
		zlog.Debug(ctx).Msg("vegeta report command failure")
		return err
	}
	zlog.Debug(ctx).Msg("Vegeta attack completed successfully")

	// Indexing results to elastic search
	if attackMap["ESHost"] != "" && attackMap["ESPort"] != "" && attackMap["ESIndex"] != "" {
		err = indexVegetaResults(ctx, metrics, testName, attackMap)
		if err != nil {
			zlog.Debug(ctx).Msg("Failed to indexing results to elastic search")
			return err
		}
	}
	return nil
}
