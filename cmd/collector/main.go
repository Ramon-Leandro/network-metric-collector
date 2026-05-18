package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Ramon-Leandro/network-metric-collector/internal/config"
	"github.com/Ramon-Leandro/network-metric-collector/internal/monitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig("configs/settings.yaml")
	if err != nil {
		slog.Error("critical failure during configuration bootstrap", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("starting professional collector engine", "configured_targets", len(cfg.Settings.Targets))

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Setting up a dedicated HTTP multiplexer for telemetry scraping
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    ":2112", // Standard industrial port convention for custom exporters
		Handler: mux,
	}

	// Spin up the Prometheus scraping server asynchronously
	go func() {
		slog.Info("telemetry endpoint online", "address", "http://localhost:2112/metrics")
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics engine server failed", "error", err.Error())
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(cfg.Settings.Interval) * time.Second)
	defer ticker.Stop()

	go func() {
		sig := <-sigChan
		slog.Info("system signal received, initiating graceful shutdown", "signal", sig.String())
		
		// Shut down the metrics server cleanly within a 5 second time window
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = metricsServer.Shutdown(shutdownCtx)

		cancel()
	}()

	for {
		select {
		case <-ticker.C:
			runCheck(ctx, httpClient, cfg.Settings.Targets)
		case <-ctx.Done():
			slog.Info("collector engine stopped successfully")
			return
		}
	}
}

func runCheck(ctx context.Context, client *http.Client, targets []string) {
	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			monitor.CheckTarget(ctx, client, t)
		}(target)
	}
	wg.Wait()
}