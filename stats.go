package main

import (
	"sync/atomic"
)

type Stats struct {
	TotalIndexReportRequests                           int64   `json:"total_index_report_requests"`
	TotalVulnerabilityReportRequests                   int64   `json:"total_vulnerability_report_requests"`
	TotalIndexReportRequestLatencyMilliseconds         int64   `json:"total_index_report_latency_milliseconds"`
	TotalVulnerabilityReportRequestLatencyMilliseconds int64   `json:"total_vulnerability_report_latency_milliseconds"`
	LatencyPerIndexReportRequest                       float64 `json:"latency_per_index_report_request"`
	LatencyPerVulnerabilityReportRequest               float64 `json:"latency_per_vulnerability_report_request"`
	Non2XXIndexReportResponses                         int64   `json:"non_2XX_index_report_responses"`
	Non2XXVulnerabilityReportResponses                 int64   `json:"non_2XX_vulnerability_report_responses"`
	MaxIndexReportRequestLatencyMilliseconds           int64   `json:"max_index_report_request_latency_milliseconds"`
	MaxVulnerabilityReportRequestLatencyMilliseconds   int64   `json:"max_vulnerability_report_request_latency_milliseconds"`
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) IncrTotalIndexReportRequests(by int64) {
	atomic.AddInt64((*int64)(&s.TotalIndexReportRequests), by)
}

func (s *Stats) IncrTotalVulnerabilityReportRequests(by int64) {
	atomic.AddInt64((*int64)(&s.TotalVulnerabilityReportRequests), by)
}

func (s *Stats) IncrTotalIndexReportRequestLatencyMilliseconds(by int64) {
	if by > s.MaxIndexReportRequestLatencyMilliseconds {
		atomic.SwapInt64(&s.MaxIndexReportRequestLatencyMilliseconds, by)
	}
	atomic.AddInt64((*int64)(&s.TotalIndexReportRequestLatencyMilliseconds), by)
}

func (s *Stats) IncrTotalVulnerabilityReportRequestLatencyMilliseconds(by int64) {
	if by > s.MaxVulnerabilityReportRequestLatencyMilliseconds {
		atomic.SwapInt64(&s.MaxVulnerabilityReportRequestLatencyMilliseconds, by)
	}
	atomic.AddInt64((*int64)(&s.TotalVulnerabilityReportRequestLatencyMilliseconds), by)
}

func (s *Stats) IncrNon2XXIndexReportResponses(by int64) {
	atomic.AddInt64((*int64)(&s.Non2XXIndexReportResponses), by)
}

func (s *Stats) IncrNon2XXVulnerabilityReportResponses(by int64) {
	atomic.AddInt64((*int64)(&s.Non2XXVulnerabilityReportResponses), by)
}

func (s *Stats) GetStats() *Stats {
	s.LatencyPerIndexReportRequest = float64(s.TotalIndexReportRequestLatencyMilliseconds) / float64(s.TotalIndexReportRequests)
	s.LatencyPerVulnerabilityReportRequest = float64(s.TotalVulnerabilityReportRequestLatencyMilliseconds) / float64(s.TotalVulnerabilityReportRequests)
	return s
}
