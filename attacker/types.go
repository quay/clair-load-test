package attacker

import (
	"time"
)

// Type used to index results to elastic search.
type Document struct {
	Workload       string         `json:"workload"`
	Endpoint       string         `json:"endpoint"`
	RequestTimeout int            `json:"request_timeout"`
	Targets        string         `json:"targets"`
	Hostname       string         `json:"hostname"`
	RPS            int            `json:"rps"`
	Throughput     float64        `json:"throughput"`
	StatusCodes    map[string]int `json:"status_codes"`
	Requests       uint64         `json:"requests"`
	P99Latency     time.Duration  `json:"p99_latency"`
	P95Latency     time.Duration  `json:"p95_latency"`
	MaxLatency     time.Duration  `json:"max_latency"`
	MinLatency     time.Duration  `json:"min_latency"`
	ReqLatency     time.Duration  `json:"req_latency"`
	Timestamp      string         `json:"timestamp"`
	BytesIn        float64        `json:"bytes_in"`
	BytesOut       float64        `json:"bytes_out"`
	RunID          string         `json:"run_id"`
}
