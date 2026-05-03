package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "")
	t.Setenv("NEUROFEED_LLM_TIMEOUT", "")
	t.Setenv("NEUROFEED_LLM_MAX_ARTICLES", "")
	t.Setenv("NEUROFEED_LLM_MAX_OUTPUT_TOKENS", "")
	t.Setenv("NEUROFEED_RSS_ITEMS_PER_FEED", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("LLM_MODEL", "")
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("RSS_FEED_URL", "")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPClientTimeout != defaultHTTPTimeout {
		t.Fatalf("timeout: got %v want %v", cfg.HTTPClientTimeout, defaultHTTPTimeout)
	}
	if cfg.LLMRequestTimeout != defaultLLMRequestTimeout {
		t.Fatalf("LLM timeout: got %v want %v", cfg.LLMRequestTimeout, defaultLLMRequestTimeout)
	}
	if cfg.LLMMaxDigestArticles != defaultLLMMaxDigestArticles {
		t.Fatalf("LLM max articles: got %d want %d", cfg.LLMMaxDigestArticles, defaultLLMMaxDigestArticles)
	}
	if cfg.LLMMaxOutputTokens != defaultLLMMaxOutputTokens {
		t.Fatalf("LLM max output tokens: got %d want %d", cfg.LLMMaxOutputTokens, defaultLLMMaxOutputTokens)
	}
	if len(cfg.RSSFeeds) != 0 {
		t.Fatalf("RSSFeeds: got %d want 0", len(cfg.RSSFeeds))
	}
	if cfg.RSSMaxItemsPerFeed != 2 {
		t.Fatalf("RSSMaxItemsPerFeed: got %d want 2", cfg.RSSMaxItemsPerFeed)
	}
}

func TestLoad_LLMTimeout(t *testing.T) {
	t.Setenv("NEUROFEED_LLM_TIMEOUT", "90s")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMRequestTimeout != 90*time.Second {
		t.Fatalf("got %v", cfg.LLMRequestTimeout)
	}
}

func TestLoad_LLMMaxArticles(t *testing.T) {
	t.Setenv("NEUROFEED_LLM_MAX_ARTICLES", "8")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMMaxDigestArticles != 8 {
		t.Fatalf("got %d", cfg.LLMMaxDigestArticles)
	}
}

func TestLoad_RSSMaxItemsPerFeed_zeroDisables(t *testing.T) {
	t.Setenv("NEUROFEED_RSS_ITEMS_PER_FEED", "0")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RSSMaxItemsPerFeed != 0 {
		t.Fatalf("got %d", cfg.RSSMaxItemsPerFeed)
	}
}

func TestLoad_RSSMaxItemsPerFeed_invalid(t *testing.T) {
	t.Setenv("NEUROFEED_RSS_ITEMS_PER_FEED", "999")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_LLMMaxArticles_invalid(t *testing.T) {
	t.Setenv("NEUROFEED_LLM_MAX_ARTICLES", "99")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_LLMMaxOutputTokens(t *testing.T) {
	t.Setenv("NEUROFEED_LLM_MAX_OUTPUT_TOKENS", "1024")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMMaxOutputTokens != 1024 {
		t.Fatalf("got %d", cfg.LLMMaxOutputTokens)
	}
}

func TestLoad_LLMTimeout_invalid(t *testing.T) {
	t.Setenv("NEUROFEED_LLM_TIMEOUT", "not-a-duration")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_HTTPTimeout(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "5s")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPClientTimeout != 5*time.Second {
		t.Fatalf("got %v", cfg.HTTPClientTimeout)
	}
}

func TestLoad_HTTPTimeoutSeconds(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "12")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPClientTimeout != 12*time.Second {
		t.Fatalf("got %v", cfg.HTTPClientTimeout)
	}
}

func TestLoad_invalidDuration(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "not-a-duration")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_HTTPTimeoutSeconds_priority(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "1s")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "99")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPClientTimeout != 99*time.Second {
		t.Fatalf("expected seconds to win, got %v", cfg.HTTPClientTimeout)
	}
}

func TestLoad_envPassthrough(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	t.Setenv("TELEGRAM_CHAT_ID", "12345")
	t.Setenv("LLM_PROVIDER", "openai")
	t.Setenv("LLM_MODEL", "gpt-4o-mini")
	t.Setenv("LLM_BASE_URL", "https://api.openai.com/v1")
	t.Setenv("LLM_API_KEY", "key")
	t.Setenv("RSS_FEED_URL", "https://example.com/feed")
	t.Setenv("RSS_FEED_TIER", "")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TelegramBotToken != "tok" ||
		cfg.TelegramChatID != "12345" ||
		cfg.LLMProvider != "openai" ||
		cfg.LLMModel != "gpt-4o-mini" ||
		cfg.LLMBaseURL != "https://api.openai.com/v1" ||
		cfg.LLMAPIKey != "key" ||
		cfg.RSSFeedURL != "https://example.com/feed" {
		t.Fatalf("unexpected cfg %+v", cfg)
	}
	if len(cfg.RSSFeeds) != 1 || cfg.RSSFeeds[0].URL != "https://example.com/feed" || cfg.RSSFeeds[0].Tier != "news" {
		t.Fatalf("RSSFeeds: %+v", cfg.RSSFeeds)
	}
}

func TestLoad_RSSFeedsJSON(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "")
	t.Setenv("NEUROFEED_RSS_FEEDS", `[{"url":"https://a.example/feed","tier":"primary"},{"url":"https://b.example/atom","tier":"expert"}]`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.RSSFeeds) != 2 {
		t.Fatalf("len %d", len(cfg.RSSFeeds))
	}
	if cfg.RSSFeeds[0].URL != "https://a.example/feed" || cfg.RSSFeeds[0].Tier != "primary" {
		t.Fatalf("feed0 %+v", cfg.RSSFeeds[0])
	}
	if cfg.RSSFeeds[1].URL != "https://b.example/atom" || cfg.RSSFeeds[1].Tier != "expert" {
		t.Fatalf("feed1 %+v", cfg.RSSFeeds[1])
	}
}

func TestLoad_RSSFeedsJSON_subjectTrim(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "")
	t.Setenv("NEUROFEED_RSS_FEEDS", `[{"url":"https://only.json/feed","tier":"news","subject":"  AI  "}]`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.RSSFeeds) != 1 || cfg.RSSFeeds[0].Subject != "AI" {
		t.Fatalf("RSSFeeds %+v", cfg.RSSFeeds)
	}
}

func TestLoad_RSSFeedsJSON_subjectTooLong(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "")
	longSubject := strings.Repeat("x", maxRSSFeedSubjectRunes+1)
	t.Setenv("NEUROFEED_RSS_FEEDS", fmt.Sprintf(`[{"url":"https://only.json/feed","tier":"news","subject":%q}]`, longSubject))

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_RSSFeedsJSON_prefersOverSingleURL(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "https://legacy.example/feed")
	t.Setenv("NEUROFEED_RSS_FEEDS", `[{"url":"https://only.json/feed","tier":"news"}]`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.RSSFeeds) != 1 || cfg.RSSFeeds[0].URL != "https://only.json/feed" {
		t.Fatalf("RSSFeeds %+v", cfg.RSSFeeds)
	}
}

func TestLoad_RSSFeedsJSON_invalidTier(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "")
	t.Setenv("NEUROFEED_RSS_FEEDS", `[{"url":"https://a.example/feed","tier":"nope"}]`)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_RSSFeedTier_invalid(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "https://example.com/feed")
	t.Setenv("RSS_FEED_TIER", "invalid")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_RSSFeedTier_override(t *testing.T) {
	t.Setenv("RSS_FEED_URL", "https://example.com/feed")
	t.Setenv("RSS_FEED_TIER", "community")
	t.Setenv("NEUROFEED_RSS_FEEDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.RSSFeeds) != 1 || cfg.RSSFeeds[0].Tier != "community" {
		t.Fatalf("RSSFeeds %+v", cfg.RSSFeeds)
	}
}

func TestConfig_DigestSubjectSections_orderAndDedupe(t *testing.T) {
	t.Parallel()
	cfg := Config{RSSFeeds: []RSSFeedEntry{
		{URL: "https://nba.example/feed", Tier: "news", Subject: "NBA"},
		{URL: "https://ai.example/feed", Tier: "news", Subject: "AI"},
		{URL: "https://ai2.example/feed", Tier: "news", Subject: "AI"},
	}}
	got := cfg.DigestSubjectSections()
	if len(got) != 2 || got[0] != "NBA" || got[1] != "AI" {
		t.Fatalf("got %#v", got)
	}
}
