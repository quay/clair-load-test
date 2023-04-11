package reports

import (
	"context"
	"strconv"

	"github.com/quay/zlog"
	"github.com/vishnuchalla/clair-load-test/pkg/utils"
)

// Method to test POST operation in index report.
func CreateIndexReport(ctx context.Context, manifests [][]byte, host, token string) ([]map[string]interface{}, string) {
	zlog.Debug(ctx).Str("requests", strconv.Itoa(len(manifests))).Msg("preparing requests for POST operation in index_report")
	url, headers := utils.GetRequestCommons(ctx, "/indexer/api/v1/index_report", host, token)
	var requests []map[string]interface{}
	for _, manifest := range manifests {
		requests = append(requests, map[string]interface{}{
			"method": "POST",
			"url":    url,
			"header": headers,
			"body":   manifest,
		})
	}
	return requests, "post_index_report"
}

// Method to test GET operation in index report.
func GetIndexReport(ctx context.Context, manifestHashes []string, host, token string) ([]map[string]interface{}, string) {
	zlog.Debug(ctx).Str("requests", strconv.Itoa(len(manifestHashes))).Msg("preparing requests for GET operation in index_report")
	url, headers := utils.GetRequestCommons(ctx, "/indexer/api/v1/index_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method": "GET",
			"url":    url + manifestHash,
			"header": headers,
		})
	}
	return requests, "get_index_report"
}

// Method to test DELETE operation in index report.
func DeleteIndexReports(ctx context.Context, manifestHashes []string, host, token string) ([]map[string]interface{}, string) {
	zlog.Debug(ctx).Str("requests", strconv.Itoa(len(manifestHashes))).Msg("preparing requests for DELETE operation in index_report")
	url, headers := utils.GetRequestCommons(ctx, "/indexer/api/v1/index_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method": "DELETE",
			"url":    url + manifestHash,
			"header": headers,
		})
	}
	return requests, "delete_index_report"
}

// Method to test GET operation in matcher vulnerability report.
func GetVulnerabilityReport(ctx context.Context, manifestHashes []string, host, token string) ([]map[string]interface{}, string) {
	zlog.Debug(ctx).Str("requests", strconv.Itoa(len(manifestHashes))).Msg("preparing requests for GET operation in vulnerability_report")
	url, headers := utils.GetRequestCommons(ctx, "/matcher/api/v1/vulnerability_report/", host, token)
	var requests []map[string]interface{}
	for _, manifestHash := range manifestHashes {
		requests = append(requests, map[string]interface{}{
			"method": "GET",
			"url":    url + manifestHash,
			"header": headers,
		})
	}
	return requests, "get_vulnerability_report"
}

// Method to test GET operation of indexer state.
func GetIndexerState(ctx context.Context, hitsize int, host, token string) ([]map[string]interface{}, string) {
	zlog.Debug(ctx).Str("requests", strconv.Itoa(hitsize)).Msg("preparing requests for GET operation in index_state")
	url, headers := utils.GetRequestCommons(ctx, "/indexer/api/v1/index_state", host, token)
	var requests []map[string]interface{}
	for i := 0; i < int(hitsize); i++ {
		requests = append(requests, map[string]interface{}{
			"method": "GET",
			"url":    url,
			"header": headers,
		})
	}
	return requests, "get_indexer_state"
}
