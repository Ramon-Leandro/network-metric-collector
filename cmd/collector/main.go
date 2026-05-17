package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ramon-Leandro/network-metric-collector/internal/config"
	"github.com/Ramon-Leandro/network-metric-collector/internal/monitor"
)

func main() {
	cfg, err := config.LoadConfig("configs/settings.yaml")
	if err != nil {
		log.Fatalf("Critical failure: could not load configuration: %v", err)
	}

	log.Printf("Starting professional collector with %d targets...", len(cfg.Settings.Targets))

	// Base context that will be canceled when the OS sends a termination signal.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listening to OS signals asynchronously.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(cfg.Settings.Interval) * time.Second)
	defer ticker.Stop()

	// Goroutine responsible for handling the graceful shutdown orchestration.
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v. Initiating graceful shutdown...", sig)
		cancel() // Triggers cancellation across all context-aware routines
	}()

	// Execution loop
	for {
		select {
		case <-ticker.C:
			// Spawning the checks. If the context is already canceled, 
			// they will return instantly.
			runCheck(ctx, cfg.Settings.Targets)
		case <-ctx.Done():
			// Ensuring the loop breaks cleanly when the shutdown process finishes.
			log.Println("Collector stopped successfully. Exiting clean.")
			return
		}
	}
}

func runCheck(ctx context.Context, targets []string) {
	for _, target := range targets {
		go monitor.CheckTarget(ctx, target)
	}
}