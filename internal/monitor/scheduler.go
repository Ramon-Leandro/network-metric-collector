package monitor

import (
	"fmt"
	"net/http"
	"time"
)

// CheckTarget performs a HEAD request to minimize bandwidth consumption 
// while validating endpoint reachability and latency.
func CheckTarget(url string) {
	start := time.Now()

	// Using a custom client with a strict timeout to prevent routine leakage
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head(url)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("[ERROR] Target %s is unreachable: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	// Metrics should ideally be exported to a TSDB like Prometheus, 
	// but here we log for stdout observability.
	fmt.Printf("[METRIC] Target: %s | Status: %d | Latency: %v\n", url, resp.StatusCode, duration)
}