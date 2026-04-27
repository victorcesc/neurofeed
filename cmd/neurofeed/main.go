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
	// Structured logs on stderr; each major step uses the same "step" attribute for easy grepping.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	logger.Info("neurofeed", "step", "startup", "detail", "logger ready")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("neurofeed", "step", "config_load", "err", err)
		os.Exit(1)
	}
	logger.Info("neurofeed", "step", "config_loaded", "feeds", len(cfg.RSSFeeds), "http_timeout", cfg.HTTPClientTimeout.String())

	if err := config.ValidatePhase1(cfg); err != nil {
		logger.Error("neurofeed", "step", "config_validate", "err", err)
		os.Exit(1)
	}
	logger.Info("neurofeed", "step", "config_validated", "detail", "RSS and Telegram settings OK")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// One pipeline run must finish before this deadline (feeds + Telegram under one umbrella).
	runCtx, cancel := context.WithTimeout(ctx, cfg.HTTPClientTimeout+time.Minute)
	defer cancel()
	logger.Info("neurofeed", "step", "run_context", "deadline", cfg.HTTPClientTimeout+time.Minute)

	httpClient := ingest.HTTPClient(cfg.HTTPClientTimeout)

	// Resolve each configured feed to a concrete URL + parsed tier for the multi-feed fetcher.
	feedSpecs := make([]ingest.RSSFeedSpec, 0, len(cfg.RSSFeeds))
	for index := range cfg.RSSFeeds {
		entry := cfg.RSSFeeds[index]
		tier, err := domain.ParseSourceTier(entry.Tier)
		if err != nil {
			logger.Error("neurofeed", "step", "feed_tier_parse", "feed_index", index, "err", err)
			os.Exit(1)
		}
		feedSpecs = append(feedSpecs, ingest.RSSFeedSpec{URL: entry.URL, Tier: tier})
	}
	logger.Info("neurofeed", "step", "feeds_wired", "count", len(feedSpecs))

	fetcher := &ingest.MultiRSSFetcher{
		Feeds:     feedSpecs,
		Client:    httpClient,
		UserAgent: "",
		Log:       logger,
	}
	notifier := &notify.TelegramNotifier{
		Token:  cfg.TelegramBotToken,
		ChatID: cfg.TelegramChatID,
		Client: httpClient,
	}

	contentPipeline := pipeline.New(cfg, logger, fetcher, ai.HeadlineSummarizer{}, notifier)
	logger.Info("neurofeed", "step", "pipeline_run_start", "detail", "fetch → dedup → summarize → notify")

	if err := contentPipeline.Run(runCtx); err != nil {
		logger.Error("neurofeed", "step", "pipeline_run_failed", "err", err)
		os.Exit(1)
	}
	logger.Info("neurofeed", "step", "pipeline_run_ok", "detail", "finished without error")
}
