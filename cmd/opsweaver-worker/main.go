package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
	"github.com/Gooooodman/opsweaver/internal/platform/health"
	"github.com/Gooooodman/opsweaver/internal/platform/logging"
	"github.com/Gooooodman/opsweaver/internal/platform/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceName     = "opsweaver-worker"
	shutdownTimeout = 10 * time.Second
	probeTimeout    = 2 * time.Second
)

func main() {
	configPath := flag.String("config", "./config.yaml", "path to config file")
	flag.Parse()

	logger := logging.New(logging.Options{
		Level:   slog.LevelInfo,
		Service: serviceName,
		Writer:  os.Stdout,
	})

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("load config", "error", err.Error())
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	m, err := metrics.New(metrics.Options{
		Namespace: "opsweaver",
		Service:   serviceName,
	}, prometheus.NewRegistry())
	if err != nil {
		logger.Error("init metrics", "error", err.Error())
		os.Exit(1)
	}

	checker := health.New(health.Options{Timeout: probeTimeout})

	mux := http.NewServeMux()
	mux.Handle("/healthz", checker.LiveHandler())
	mux.Handle("/readyz", checker.ReadyHandler())
	mux.Handle("/metrics", m.Handler())

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Worker.HealthPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("health server listening", "port", cfg.Worker.HealthPort)
		if lerr := srv.ListenAndServe(); lerr != nil && !errors.Is(lerr, http.ErrServerClosed) {
			serverErr <- lerr
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case lerr, ok := <-serverErr:
		if ok && lerr != nil {
			logger.Error("health server failed", "error", lerr.Error())
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if serr := srv.Shutdown(shutdownCtx); serr != nil {
		logger.Error("health server shutdown", "error", serr.Error())
	}
}
