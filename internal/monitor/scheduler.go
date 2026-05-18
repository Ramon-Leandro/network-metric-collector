package monitor

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Defining professional Prometheus metrics vectors. 
// We use vector types to safely inject dynamic labels like 'target' and 'status'.
var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "network_collector_requests_total",
			Help: "The total number of network probes executed by the collector.",
		},
		[]string{"target", "status"},
	)

	HttpLatencyGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "network_collector_latency_seconds",
			Help: "The network response latency measured in fractional seconds.",
		},
		[]string{"target"},
	)
)

// CheckTarget performs a concurrent network probe using a scoped HTTP GET method.
// It records structural execution telemetry directly into the Prometheus registry.
func CheckTarget(ctx context.Context, client *http.Client, url string) {
	start := time.Now()

	// Switching to GET ensures compatibility with servers rejecting HEAD commands.
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
	// Instantly discarding body streams to prevent kernel memory socket allocation leaks
	defer resp.Body.Close()

	// Parse values into structural string metrics labels
	statusCodeStr := strconv.Itoa(resp.StatusCode)

	// Update Prometheus vectors safely across parallel threads
	HttpRequestsTotal.WithLabelValues(url, statusCodeStr).Inc()
	HttpLatencyGauge.WithLabelValues(url).Set(duration.Seconds())

	slog.Info("network metric collected", 
		"target", url, 
		"status", resp.StatusCode, 
		"latency_ms", duration.Milliseconds(),
	)
}