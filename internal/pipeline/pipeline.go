// Package pipeline orchestrates ingest → domain logic → AI → notify.
package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/victorcesc/neurofeed/internal/ai"
	"github.com/victorcesc/neurofeed/internal/config"
	"github.com/victorcesc/neurofeed/internal/domain"
	"github.com/victorcesc/neurofeed/internal/ingest"
	"github.com/victorcesc/neurofeed/internal/notify"
)

// Pipeline wires boundary interfaces. Dependencies are injected for testing.
type Pipeline struct {
	cfg        config.Config
	fetcher    ingest.FeedFetcher
	summarizer ai.Summarizer
	notifier   notify.Notifier
	log        *slog.Logger
}

// New builds a Pipeline with explicit dependencies.
func New(cfg config.Config, log *slog.Logger, fetcher ingest.FeedFetcher, summarizer ai.Summarizer, notifier notify.Notifier) *Pipeline {
	if log == nil {
		log = slog.Default()
	}
	return &Pipeline{
		cfg:        cfg,
		fetcher:    fetcher,
		summarizer: summarizer,
		notifier:   notifier,
		log:        log,
	}
}

// Run executes one full cycle: fetch → deduplicate by title → summarize → notify.
func (contentPipeline *Pipeline) Run(ctx context.Context) error {
	contentPipeline.log.Info("pipeline", "step", "fetch_start", "detail", "downloading and parsing all configured feeds")

	articles, err := contentPipeline.fetcher.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	contentPipeline.log.Info("pipeline", "step", "fetch_done", "article_count", len(articles))

	beforeDedup := len(articles)
	articles = domain.DeduplicateArticlesByTitle(articles)
	contentPipeline.log.Info("pipeline", "step", "dedup_done", "article_count", len(articles), "deduped_from", beforeDedup)

	contentPipeline.log.Info("pipeline", "step", "summarize_start", "article_count", len(articles), "detail", "headline digest (no LLM yet)")

	summary, err := contentPipeline.summarizer.Summarize(ctx, articles)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}
	contentPipeline.log.Info("pipeline", "step", "summarize_done", "summary_bytes", len(summary), "non_empty", summary != "")

	if summary == "" {
		contentPipeline.log.Info("pipeline", "step", "notify_skipped", "reason", "empty summary", "articles_after_dedup", len(articles))
		return nil
	}

	contentPipeline.log.Info("pipeline", "step", "notify_start", "detail", "POST Telegram sendMessage")

	if err := contentPipeline.notifier.Notify(ctx, summary); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	contentPipeline.log.Info("pipeline", "step", "notify_done", "detail", "Telegram accepted the message")
	return nil
}
