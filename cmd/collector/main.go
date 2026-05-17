package main

import (
	"log"
	"time"

	"github.com/youruser/network-metric-collector/internal/config"
	"github.com/youruser/network-metric-collector/internal/monitor"
)

func main() {
	// Loading configuration at startup to ensure environment consistency
	cfg, err := config.LoadConfig("configs/settings.yaml")
	if err != nil {
		log.Fatalf("Critical failure: could not load configuration: %v", err)
	}

	log.Printf("Starting collector with %d targets...", len(cfg.Settings.Targets))

	// Ticker ensures precise execution intervals regardless of task duration
	ticker := time.NewTicker(time.Duration(cfg.Settings.Interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			runCheck(cfg.Settings.Targets)
		}
	}
}

// runCheck orchestrates the concurrent execution of network probes.
func runCheck(targets []string) {
	for _, target := range targets {
		// Spawning a goroutine per target allows concurrent latency checks,
		// preventing a slow endpoint from bottlenecking the entire pipeline.
		go monitor.CheckTarget(target)
	}
}