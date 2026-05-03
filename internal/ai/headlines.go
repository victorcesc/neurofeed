package ai

import (
	"context"

	"github.com/victorcesc/neurofeed/internal/domain"
)

// HeadlineSummarizer builds Telegram HTML digests without the LLM (Phase 4 layout: up to two links per subject).
type HeadlineSummarizer struct{}

// Summarize implements Summarizer. Pure string building (no HTTP); caller logs pipeline steps around this call.
func (HeadlineSummarizer) Summarize(ctx context.Context, articles []domain.Article) (string, error) {
	return HeadlineSummarizer{}.SummarizeWithSubjectOrder(ctx, articles, nil)
}

// SummarizeWithSubjectOrder implements SubjectOrderedSummarizer.
func (HeadlineSummarizer) SummarizeWithSubjectOrder(ctx context.Context, articles []domain.Article, subjectOrder []string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(articles) == 0 {
		return "", nil
	}
	if !domain.HasAnyConfiguredSubject(articles) {
		return FormatStaticHeadlineHTML(articles, nil, false), nil
	}
	order := effectiveSectionOrder(subjectOrder, articles)
	return FormatStaticHeadlineHTML(articles, order, true), nil
}
