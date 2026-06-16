package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
	"github.com/Gooooodman/opsweaver/internal/platform/database"
	"github.com/Gooooodman/opsweaver/internal/platform/health"
	"github.com/Gooooodman/opsweaver/internal/platform/logging"
	"github.com/Gooooodman/opsweaver/internal/platform/metrics"
	"github.com/Gooooodman/opsweaver/internal/platform/servicehttp"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceName     = "opsweaver-gateway"
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

	db, err := database.Open(ctx, database.Options{
		DSN:    cfg.Gateway.Database.DSN,
		Logger: logger,
	})
	if err != nil {
		logger.Error("open database", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		if cerr := database.Close(db); cerr != nil {
			logger.Error("close database", "error", cerr.Error())
		}
	}()

	m, err := metrics.New(metrics.Options{
		Namespace: "opsweaver",
		Service:   serviceName,
	}, prometheus.NewRegistry())
	if err != nil {
		logger.Error("init metrics", "error", err.Error())
		os.Exit(1)
	}

	checker := health.New(health.Options{
		Timeout:          probeTimeout,
		RecordDependency: m.RecordDependency,
	})
	checker.Register("postgres", func(ctx context.Context) error {
		sqlDB, derr := db.DB()
		if derr != nil {
			return derr
		}
		return sqlDB.PingContext(ctx)
	})

	mux := http.NewServeMux()
	mux.Handle("/healthz", m.Middleware("/healthz", checker.LiveHandler()))
	mux.Handle("/readyz", m.Middleware("/readyz", checker.ReadyHandler()))
	mux.Handle("/metrics", m.Handler())

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Gateway.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	logger.Info("http server listening", "port", cfg.Gateway.Port)
	if err := servicehttp.Serve(ctx, srv, shutdownTimeout); err != nil {
		logger.Error("http server failed", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("shutdown signal received")
}
