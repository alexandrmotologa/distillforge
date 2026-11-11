package benchmark

import (
	"sort"
	"time"
)

// LatencyResult records latency measurements for inference benchmarking.
type LatencyResult struct {
	TotalRuns     int     `json:"total_runs"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	MinLatencyMs  float64 `json:"min_latency_ms"`
	MaxLatencyMs  float64 `json:"max_latency_ms"`
	P50LatencyMs  float64 `json:"p50_latency_ms"`
	P90LatencyMs  float64 `json:"p90_latency_ms"`
	P95LatencyMs  float64 `json:"p95_latency_ms"`
	P99LatencyMs  float64 `json:"p99_latency_ms"`
	AvgTTFTMs     float64 `json:"avg_ttft_ms"`
	TokensPerSec  float64 `json:"tokens_per_sec"`
}

// LatencyProfiler aggregates duration metrics over repeated inference executions.
type LatencyProfiler struct {
	durations []time.Duration
	ttfts     []time.Duration
	tokens    []int
}

// NewLatencyProfiler creates an empty profiler.
func NewLatencyProfiler() *LatencyProfiler {
	return &LatencyProfiler{
		durations: make([]time.Duration, 0),
		ttfts:     make([]time.Duration, 0),
		tokens:    make([]int, 0),
	}
}

// Record adds a single run's duration, TTFT, and generated token count.
func (p *LatencyProfiler) Record(duration, ttft time.Duration, tokenCount int) {
	p.durations = append(p.durations, duration)
	p.ttfts = append(p.ttfts, ttft)
	p.tokens = append(p.tokens, tokenCount)
}

// Calculate compiles the latency statistics and percentiles.
func (p *LatencyProfiler) Calculate() LatencyResult {
	n := len(p.durations)
	if n == 0 {
		return LatencyResult{}
	}

	sorted := make([]float64, n)
	var totalDurMs float64
	var totalTTFTMs float64
	var totalTokens int

	for i, d := range p.durations {
		ms := float64(d.Microseconds()) / 1000.0
		sorted[i] = ms
		totalDurMs += ms
		totalTTFTMs += float64(p.ttfts[i].Microseconds()) / 1000.0
		totalTokens += p.tokens[i]
	}

	sort.Float64s(sorted)

	avgLatency := totalDurMs / float64(n)
	avgTTFT := totalTTFTMs / float64(n)

	var tokPerSec float64
	if totalDurMs > 0 {
		tokPerSec = float64(totalTokens) / (totalDurMs / 1000.0)
	}

	return LatencyResult{
		TotalRuns:    n,
		AvgLatencyMs: avgLatency,
		MinLatencyMs: sorted[0],
		MaxLatencyMs: sorted[n-1],
		P50LatencyMs: percentile(sorted, 0.50),
		P90LatencyMs: percentile(sorted, 0.90),
		P95LatencyMs: percentile(sorted, 0.95),
		P99LatencyMs: percentile(sorted, 0.99),
		AvgTTFTMs:    avgTTFT,
		TokensPerSec: tokPerSec,
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}
