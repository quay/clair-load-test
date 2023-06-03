package attacker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/quay/zlog"
)

// getRequestCommons returns the common inputs for each and every HTTP request.
// It returns a url and headers for the specified input.
func getRequestCommons(ctx context.Context, endpoint, host, token string) (string, http.Header) {
	url := host + endpoint
	zlog.Debug(ctx).Str("endpoint", url).Msg("preparing headers")
	headers := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
	}
	return url, headers
}

// CreateIndexReportRequests returns the list of requests to perform POST operation on index_report.
func CreateIndexReportRequests(ctx context.Context, manifests [][]byte, host, token string) []map[string]interface{} {
	zlog.Debug(ctx).Int("number of requests", len(manifests)).Msg("preparing requests for POST operation in index_report")
	url, headers := getRequestCommons(ctx, "/indexer/api/v1/index_report", host, token)
	var requests []map[string]interface{}
	for _, manifest := range manifests {
		requests = append(requests, map[string]interface{}{
			"method": http.MethodPost,
			"url":    url,
			"header": headers,
			"body":   manifest,
		})
	}
	return requests
}

// GetIndexReportRequests returns the list of requests to perform GET operation on index_report.
func GetIndexReportRequests(ctx context.Context, manifestHashes []string, host, token string) []map[string]interface{} {
	zlog.Debug(ctx).Int("number of requests", len(manifestHashes)).Msg("preparing requests for GET operation in index_report")
	url, headers := getRequestCommons(ctx, "/indexer/api/v1/index_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method": http.MethodGet,
			"url":    url + manifestHash,
			"header": headers,
		})
	}
	return requests
}

// DeleteIndexReportsRequests returns the list of requests to perform DELETE operation on index_report.
func DeleteIndexReportsRequests(ctx context.Context, manifestHashes []string, host, token string) []map[string]interface{} {
	zlog.Debug(ctx).Int("number of requests", len(manifestHashes)).Msg("preparing requests for DELETE operation in index_report")
	url, headers := getRequestCommons(ctx, "/indexer/api/v1/index_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method": http.MethodDelete,
			"url":    url + manifestHash,
			"header": headers,
		})
	}
	return requests
}

// GetVulnerabilityReportRequests returns the list of requests to perform GET operation on vulnerability_report.
func GetVulnerabilityReportRequests(ctx context.Context, manifestHashes []string, host, token string) []map[string]interface{} {
	zlog.Debug(ctx).Int("number of requests", len(manifestHashes)).Msg("preparing requests for GET operation in vulnerability_report")
	url, headers := getRequestCommons(ctx, "/matcher/api/v1/vulnerability_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method": http.MethodGet,
			"url":    url + manifestHash,
			"header": headers,
		})
	}
	return requests
}

// GetIndexerStateRequests returns the list of requests to perform GET operation on index_state.
func GetIndexerStateRequests(ctx context.Context, hitsize int, host, token string) []map[string]interface{} {
	zlog.Debug(ctx).Int("number of requests", hitsize).Msg("preparing requests for GET operation in index_state")
	url, headers := getRequestCommons(ctx, "/indexer/api/v1/index_state", host, token)
	var requests []map[string]interface{}
	for i := 0; i < int(hitsize); i++ {
		requests = append(requests, map[string]interface{}{
			"method": http.MethodGet,
			"url":    url,
			"header": headers,
		})
	}
	return requests
}
