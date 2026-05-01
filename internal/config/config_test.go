package config

import (
	"testing"
	"time"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "")
	t.Setenv("NEUROFEED_LLM_TIMEOUT", "")
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
	if len(cfg.RSSFeeds) != 0 {
		t.Fatalf("RSSFeeds: got %d want 0", len(cfg.RSSFeeds))
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
