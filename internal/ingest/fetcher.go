// Package ingest fetches raw items from RSS and other sources and maps them to domain articles.
package ingest

import (
	"context"
	"net/http"
	"time"

	"github.com/victorcesc/neurofeed/internal/domain"
)

// FeedFetcher retrieves articles from external feeds. All implementations must respect ctx cancellation and deadlines.
type FeedFetcher interface {
	Fetch(ctx context.Context) ([]domain.Article, error)
}

// StubFetcher is a phase-0 placeholder that performs no network I/O.
type StubFetcher struct{}

// Fetch implements FeedFetcher.
func (StubFetcher) Fetch(ctx context.Context) ([]domain.Article, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

// HTTPClient returns an http.Client with the given timeout for use by real fetchers in later phases.
func HTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
