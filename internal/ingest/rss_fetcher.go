package ingest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/victorcesc/neurofeed/internal/domain"
)

const defaultRSSUserAgent = "neurofeed/1.0 (+https://github.com/victorcesc/neurofeed)"

// RSSFetcher downloads and parses a single RSS or Atom feed URL into domain articles.
type RSSFetcher struct {
	URL         string
	Client      *http.Client
	UserAgent   string
	DefaultTier domain.SourceTier
}

// Fetch implements FeedFetcher: HTTP GET + parse via gofeed, then map each item to domain.Article with this fetcher's tier.
func (fetcher *RSSFetcher) Fetch(ctx context.Context) ([]domain.Article, error) {
	if strings.TrimSpace(fetcher.URL) == "" {
		return nil, fmt.Errorf("rss fetcher: empty feed URL")
	}
	if fetcher.Client == nil {
		return nil, fmt.Errorf("rss fetcher: nil HTTP client")
	}

	// gofeed uses parser.Client for the HTTP request; User-Agent identifies neurofeed to feed operators.
	parser := gofeed.NewParser()
	parser.Client = fetcher.Client
	if strings.TrimSpace(fetcher.UserAgent) != "" {
		parser.UserAgent = strings.TrimSpace(fetcher.UserAgent)
	} else {
		parser.UserAgent = defaultRSSUserAgent
	}

	feed, err := parser.ParseURLWithContext(fetcher.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("rss parse: %w", err)
	}

	// Prefer channel title, then feed link, then URL, so Article.Source is never empty when the feed parsed.
	sourceName := strings.TrimSpace(feed.Title)
	if sourceName == "" {
		sourceName = feed.FeedLink
	}
	if sourceName == "" {
		sourceName = fetcher.URL
	}

	tier := fetcher.DefaultTier
	if tier == domain.SourceTierUnspecified {
		tier = domain.SourceTierNews
	}

	// Skip nil or empty items; description falls back to content when the feed uses content:encoded only.
	articles := make([]domain.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item == nil {
			continue
		}
		title := strings.TrimSpace(item.Title)
		link := strings.TrimSpace(item.Link)
		if title == "" && link == "" {
			continue
		}
		description := strings.TrimSpace(item.Description)
		if description == "" {
			description = strings.TrimSpace(item.Content)
		}

		var published time.Time
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			published = *item.UpdatedParsed
		}

		articles = append(articles, domain.Article{
			Title:       title,
			Link:        link,
			Description: description,
			Source:      sourceName,
			SourceTier:  tier,
			Published:   published,
		})
	}

	return articles, nil
}
