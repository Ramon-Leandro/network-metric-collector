package monitor

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "network_collector_requests_total",
			Help: "The total number of network probes executed by the collector.",
		},
		[]string{"target", "status"},
	)

	// Upgraded from GaugeVec to HistogramVec to capture latency percentiles (P50, P99)
	HttpLatencyHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "network_collector_latency_seconds",
			Help:    "The network response latency distribution in fractional seconds.",
			Buckets: prometheus.DefBuckets, // Standard system buckets ranging from 5ms to 10s
		},
		[]string{"target"},
	)
)

func CheckTarget(ctx context.Context, client *http.Client, url string) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("failed to create network request", "target", url, "error", err.Error())
		return
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() == context.Canceled {
			slog.Info("probe aborted due to shutdown signal", "target", url)
			return
		}
		slog.Error("target is unreachable", "target", url, "error", err.Error())
		HttpRequestsTotal.WithLabelValues(url, "error").Inc()
		return
	}
	
	// Crucial structural change: Drain the network stream before closing it.
	// This signals Go's runtime that the TCP socket connection can be safely returned 
	// to the Keep-Alive pool instead of destroying it.
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	statusCodeStr := strconv.Itoa(resp.StatusCode)

	HttpRequestsTotal.WithLabelValues(url, statusCodeStr).Inc()
	HttpLatencyHistogram.WithLabelValues(url).Observe(duration.Seconds()) // Observe into distribution buckets

	slog.Info("network metric collected", 
		"target", url, 
		"status", resp.StatusCode, 
		"latency_ms", duration.Milliseconds(),
	)
}