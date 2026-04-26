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
	if articles[0].Published.IsZero() {
		t.Fatal("expected published time from pubDate")
	}
}
