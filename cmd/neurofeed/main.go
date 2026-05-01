package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
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
	llmSmoke := flag.Bool("llm-smoke", false, "send one minimal OpenAI chat completion and exit (requires LLM_API_KEY; optional LLM_BASE_URL, LLM_MODEL, LLM_PROVIDER)")
	flag.Parse()

	// Structured logs on stderr; each major step uses the same "step" attribute for easy grepping.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	logger.Info("neurofeed", "step", "startup", "detail", "logger ready")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("neurofeed", "step", "config_load", "err", err)
		os.Exit(1)
	}
	logger.Info("neurofeed", "step", "config_loaded", "feeds", len(cfg.RSSFeeds), "http_timeout", cfg.HTTPClientTimeout.String(), "llm_timeout", cfg.LLMRequestTimeout.String())

	if *llmSmoke {
		if err := config.ValidateLLMSmoke(cfg); err != nil {
			logger.Error("neurofeed", "step", "llm_smoke_validate", "err", err)
			os.Exit(1)
		}
		llmClient := &http.Client{Timeout: cfg.LLMRequestTimeout}
		chatClient, err := ai.NewOpenAIChatClientFromConfig(cfg, llmClient)
		if err != nil {
			logger.Error("neurofeed", "step", "llm_client", "err", err)
			os.Exit(1)
		}
		smokeCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		callCtx, callCancel := context.WithTimeout(smokeCtx, cfg.LLMRequestTimeout+30*time.Second)
		defer callCancel()
		logger.Info("neurofeed", "step", "llm_smoke_start", "detail", "POST chat/completions")
		reply, err := chatClient.ChatCompletion(callCtx, []ai.ChatMessage{
			{Role: "user", Content: "Reply with exactly the two characters OK and nothing else."},
		})
		if err != nil {
			logger.Error("neurofeed", "step", "llm_smoke_failed", "err", err)
			os.Exit(1)
		}
		logger.Info("neurofeed", "step", "llm_smoke_ok", "reply_chars", len(reply), "reply", reply)
		os.Exit(0)
	}

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
