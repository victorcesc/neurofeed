// Package ai wraps LLM providers (e.g. OpenAI) for summarization. All calls accept context for cancellation and timeouts.
package ai

import (
	"context"

	"github.com/victorcesc/neurofeed/internal/domain"
)

// Summarizer produces a digest string from a batch of articles.
type Summarizer interface {
	Summarize(ctx context.Context, articles []domain.Article) (string, error)
}

// StubSummarizer is a phase-0 placeholder that performs no external I/O.
type StubSummarizer struct{}

// Summarize implements Summarizer.
func (StubSummarizer) Summarize(ctx context.Context, _ []domain.Article) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", nil
}
