package ingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/victorcesc/neurofeed/internal/domain"
)

// RSSFeedSpec is one feed URL with its configured source tier for multi-feed ingestion.
type RSSFeedSpec struct {
	URL  string
	Tier domain.SourceTier
}

// MultiRSSFetcher fetches several RSS URLs and concatenates articles. Partial failures are logged;
// if every feed fails, it returns a joined error. Successful feeds still contribute articles.
type MultiRSSFetcher struct {
	Feeds     []RSSFeedSpec
	Client    *http.Client
	UserAgent string
	Log       *slog.Logger
}

// Fetch implements FeedFetcher. Feeds are fetched sequentially so logs stay ordered; each feed uses its own RSSFetcher.
func (multiFetcher *MultiRSSFetcher) Fetch(ctx context.Context) ([]domain.Article, error) {
	if len(multiFetcher.Feeds) == 0 {
		return nil, fmt.Errorf("multi rss fetcher: no feeds configured")
	}
	if multiFetcher.Client == nil {
		return nil, fmt.Errorf("multi rss fetcher: nil HTTP client")
	}

	if multiFetcher.Log != nil {
		multiFetcher.Log.Info("ingest", "step", "multi_fetch_start", "feed_count", len(multiFetcher.Feeds))
	}

	var combined []domain.Article
	var failures []error
	for feedIndex := range multiFetcher.Feeds {
		spec := multiFetcher.Feeds[feedIndex]
		if multiFetcher.Log != nil {
			multiFetcher.Log.Info("ingest", "step", "feed_fetch_start", "feed_index", feedIndex, "url", spec.URL, "tier", spec.Tier.String())
		}

		singleFetcher := &RSSFetcher{
			URL:         spec.URL,
			Client:      multiFetcher.Client,
			UserAgent:   multiFetcher.UserAgent,
			DefaultTier: spec.Tier,
		}
		batch, err := singleFetcher.Fetch(ctx)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", spec.URL, err))
			if multiFetcher.Log != nil {
				multiFetcher.Log.Warn("ingest", "step", "feed_fetch_failed", "feed_index", feedIndex, "url", spec.URL, "err", err)
			}
			continue
		}
		if multiFetcher.Log != nil {
			multiFetcher.Log.Info("ingest", "step", "feed_fetch_ok", "feed_index", feedIndex, "url", spec.URL, "items", len(batch))
		}
		combined = append(combined, batch...)
	}
	if len(combined) == 0 && len(failures) > 0 {
		return nil, fmt.Errorf("all rss feeds failed: %w", errors.Join(failures...))
	}
	if multiFetcher.Log != nil {
		multiFetcher.Log.Info("ingest", "step", "multi_fetch_done", "total_items", len(combined), "failed_feeds", len(failures))
	}
	return combined, nil
}
