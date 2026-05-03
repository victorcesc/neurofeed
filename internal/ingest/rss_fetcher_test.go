package ingest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/victorcesc/neurofeed/internal/domain"
)

func TestRSSFetcher_Fetch(t *testing.T) {
	const sampleRSS = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
  <title>Test Channel</title>
  <item>
    <title>Hello</title>
    <link>https://example.com/a</link>
    <description>Desc</description>
    <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
  </item>
</channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		_, _ = writer.Write([]byte(sampleRSS))
	}))
	t.Cleanup(server.Close)

	fetcher := &RSSFetcher{
		URL:         server.URL,
		Client:      server.Client(),
		UserAgent:   "neurofeed/test",
		DefaultTier: domain.SourceTierExpert,
		Subject:     "Tech",
	}

	articles, err := fetcher.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 {
		t.Fatalf("len=%d", len(articles))
	}
	if articles[0].Title != "Hello" || articles[0].Link != "https://example.com/a" {
		t.Fatalf("unexpected article %+v", articles[0])
	}
	if articles[0].Source != "Test Channel" {
		t.Fatalf("source: got %q", articles[0].Source)
	}
	if articles[0].Description != "Desc" {
		t.Fatalf("description: got %q", articles[0].Description)
	}
	if articles[0].SourceTier != domain.SourceTierExpert {
		t.Fatalf("tier: got %v", articles[0].SourceTier)
	}
	if articles[0].Subject != "Tech" {
		t.Fatalf("subject: got %q", articles[0].Subject)
	}
	if articles[0].Published.IsZero() {
		t.Fatal("expected published time from pubDate")
	}
}

func TestRSSFetcher_Fetch_maxItemsPerFeed_keepsNewestTwo(t *testing.T) {
	t.Parallel()
	const sampleRSS = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
  <title>Ch</title>
  <item>
    <title>Oldest</title>
    <link>https://example.com/old</link>
    <pubDate>Mon, 01 Jan 2007 12:00:00 GMT</pubDate>
  </item>
  <item>
    <title>Newest</title>
    <link>https://example.com/new</link>
    <pubDate>Wed, 03 Jan 2007 12:00:00 GMT</pubDate>
  </item>
  <item>
    <title>Middle</title>
    <link>https://example.com/mid</link>
    <pubDate>Tue, 02 Jan 2007 12:00:00 GMT</pubDate>
  </item>
</channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		_, _ = writer.Write([]byte(sampleRSS))
	}))
	t.Cleanup(server.Close)

	fetcher := &RSSFetcher{
		URL:             server.URL,
		Client:          server.Client(),
		UserAgent:       "neurofeed/test",
		DefaultTier:     domain.SourceTierNews,
		MaxItemsPerFeed: 2,
	}

	articles, err := fetcher.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("len=%d", len(articles))
	}
	if articles[0].Title != "Newest" || articles[1].Title != "Middle" {
		t.Fatalf("order/newest: %+v, %+v", articles[0], articles[1])
	}
}
