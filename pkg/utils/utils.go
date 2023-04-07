package utils

import (
	"fmt"
	"context"

	"github.com/quay/zlog"
)

// Method to return some of the HTTP request commons.
func GetRequestCommons(ctx context.Context, endpoint, host, token string) (string, map[string][]string) {
	url := host + endpoint
	zlog.Debug(ctx).Str("endpoint", url).Msg("preparing headers")
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	return url, headers
}