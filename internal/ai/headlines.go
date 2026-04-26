package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/victorcesc/neurofeed/internal/domain"
)

const headlineSummarizerMaxItems = 40

// HeadlineSummarizer builds a plain-text digest of article titles and links (Phase 1 MVP; no LLM).
type HeadlineSummarizer struct{}

// Summarize implements Summarizer.
func (HeadlineSummarizer) Summarize(ctx context.Context, articles []domain.Article) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(articles) == 0 {
		return "", nil
	}

	// limit := len(articles)
	limit := min(len(articles), headlineSummarizerMaxItems)


	var builder strings.Builder
	for index := range limit {
		article := articles[index]
		title := strings.TrimSpace(article.Title)
		link := strings.TrimSpace(article.Link)
		if title == "" {
			title = "(no title)"
		}
		if link == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(title)
		builder.WriteString("\n")
		builder.WriteString(link)
	}
	if builder.Len() == 0 {
		return "", nil
	}
	if len(articles) > headlineSummarizerMaxItems {
		_, _ = fmt.Fprintf(&builder, "\n\n… and %d more (cap %d for Telegram size).", len(articles)-headlineSummarizerMaxItems, headlineSummarizerMaxItems)
	}
	return builder.String(), nil
}
