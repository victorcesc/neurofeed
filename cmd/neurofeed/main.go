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
	"github.com/victorcesc/neurofeed/internal/domain"
	"github.com/victorcesc/neurofeed/internal/ingest"
	"github.com/victorcesc/neurofeed/internal/notify"
	"github.com/victorcesc/neurofeed/internal/pipeline"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}
	if err := config.ValidatePhase1(cfg); err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Single-run deadline; RSS and Telegram use the same HTTP timeout baseline.
	runCtx, cancel := context.WithTimeout(ctx, cfg.HTTPClientTimeout+time.Minute)
	defer cancel()

	httpClient := ingest.HTTPClient(cfg.HTTPClientTimeout)

	fetcher := &ingest.RSSFetcher{
		URL:         cfg.RSSFeedURL,
		Client:      httpClient,
		UserAgent:   "",
		DefaultTier: domain.SourceTierNews,
	}
	notifier := &notify.TelegramNotifier{
		Token:  cfg.TelegramBotToken,
		ChatID: cfg.TelegramChatID,
		Client: httpClient,
	}

	contentPipeline := pipeline.New(cfg, logger, fetcher, ai.HeadlineSummarizer{}, notifier)
	if err := contentPipeline.Run(runCtx); err != nil {
		logger.Error("run failed", "err", err)
		os.Exit(1)
	}
	logger.Info("neurofeed phase 1 run OK")
}
