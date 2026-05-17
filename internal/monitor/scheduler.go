package monitor

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CheckTarget performs a concurrent network probe against a specific URL.
// It accepts a context to allow premature cancellation from the main lifecycle.
func CheckTarget(ctx context.Context, url string) {
	start := time.Now()

	// Using a custom client to avoid the global http.DefaultClient bottleneck.
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// Creating a request bound to the application lifetime context.
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create request for %s: %v\n", url, err)
		return
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		// Checking if the error was due to an intentional context cancellation.
		if ctx.Err() == context.Canceled {
			fmt.Printf("[INFO] Probe for %s aborted due to shutdown signal\n", url)
			return
		}
		fmt.Printf("[ERROR] Target %s is unreachable: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[METRIC] Target: %s | Status: %d | Latency: %v\n", url, resp.StatusCode, duration)
}