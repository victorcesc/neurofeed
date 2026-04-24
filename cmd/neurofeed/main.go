package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/victorcesc/neurofeed/internal/ai"
	"github.com/victorcesc/neurofeed/internal/config"
	"github.com/victorcesc/neurofeed/internal/ingest"
	"github.com/victorcesc/neurofeed/internal/notify"
	"github.com/victorcesc/neurofeed/internal/pipeline"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Single-run deadline; real I/O in later phases should use narrower per-request timeouts.
	runCtx, cancel := context.WithTimeout(ctx, cfg.HTTPClientTimeout+time.Minute)
	defer cancel()

	p := pipeline.New(cfg, log, ingest.StubFetcher{}, ai.StubSummarizer{}, notify.StubNotifier{})
	if err := p.Run(runCtx); err != nil {
		log.Error("run failed", "err", err)
		os.Exit(1)
	}
	log.Info("neurofeed phase 0 baseline OK")
}
