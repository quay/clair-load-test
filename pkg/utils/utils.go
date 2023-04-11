package utils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/quay/zlog"
)

// Method to return some of the HTTP request commons.
func GetRequestCommons(ctx context.Context, endpoint, host, token string) (string, map[string][]string) {
	url := host + endpoint
	zlog.Debug(ctx).Str("endpoint", url).Msg("preparing headers")
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	return url, headers
}

// Method to return list of containers using repo prefix.
func GetContainersList(ctx context.Context, testRepoPrefix string, hitSize int) []string {
	var containers []string
	for i := 1; i <= hitSize; i++ {
		containers = append(containers, testRepoPrefix+"_tag_"+strconv.Itoa(i))
	}
	return containers
}
