package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/victorcesc/neurofeed/internal/domain"
)

func rssXML(title, link string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
  <title>Ch</title>
  <item>
    <title>%s</title>
    <link>%s</link>
    <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
  </item>
</channel>
</rss>`, title, link)
}

func TestMultiRSSFetcher_Fetch_twoFeeds(t *testing.T) {
	first := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(rssXML("A", "https://one/item")))
	}))
	second := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(rssXML("B", "https://two/item")))
	}))
	t.Cleanup(first.Close)
	t.Cleanup(second.Close)

	multiFetcher := &MultiRSSFetcher{
		Feeds: []RSSFeedSpec{
			{URL: first.URL, Tier: domain.SourceTierPrimary, Subject: "A"},
			{URL: second.URL, Tier: domain.SourceTierCommunity, Subject: "B"},
		},
		Client: &http.Client{},
		Log:    slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1})),
	}

	articles, err := multiFetcher.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("len=%d", len(articles))
	}
	if articles[0].Title != "A" || articles[0].SourceTier != domain.SourceTierPrimary || articles[0].Subject != "A" {
		t.Fatalf("first: %+v", articles[0])
	}
	if articles[1].Title != "B" || articles[1].SourceTier != domain.SourceTierCommunity || articles[1].Subject != "B" {
		t.Fatalf("second: %+v", articles[1])
	}
}

func TestMultiRSSFetcher_Fetch_partialFailure(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	badServer.Close()

	okServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(rssXML("OK", "https://ok/item")))
	}))
	t.Cleanup(okServer.Close)

	multiFetcher := &MultiRSSFetcher{
		Feeds: []RSSFeedSpec{
			{URL: badServer.URL, Tier: domain.SourceTierNews},
			{URL: okServer.URL, Tier: domain.SourceTierExpert},
		},
		Client: &http.Client{},
		Log:    slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1})),
	}

	articles, err := multiFetcher.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 || articles[0].Title != "OK" {
		t.Fatalf("got %+v", articles)
	}
}

func TestMultiRSSFetcher_Fetch_allFailed(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	badServer.Close()

	multiFetcher := &MultiRSSFetcher{
		Feeds: []RSSFeedSpec{
			{URL: badServer.URL, Tier: domain.SourceTierNews},
		},
		Client: &http.Client{},
		Log:    slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1})),
	}

	_, err := multiFetcher.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
