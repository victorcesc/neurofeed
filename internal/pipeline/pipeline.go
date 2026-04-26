// Package pipeline orchestrates ingest → domain logic → AI → notify.
package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/victorcesc/neurofeed/internal/ai"
	"github.com/victorcesc/neurofeed/internal/config"
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

// Run executes one full cycle: fetch → summarize → notify.
func (contentPipeline *Pipeline) Run(ctx context.Context) error {
	articles, err := contentPipeline.fetcher.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	contentPipeline.log.Info("ingest complete", "count", len(articles))

	summary, err := contentPipeline.summarizer.Summarize(ctx, articles)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	if summary == "" {
		contentPipeline.log.Info("nothing to send", "articles", len(articles))
		return nil
	}

	if err := contentPipeline.notifier.Notify(ctx, summary); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	return nil
}
