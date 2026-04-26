package config

import (
	"testing"
	"time"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("LLM_MODEL", "")
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("RSS_FEED_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPClientTimeout != defaultHTTPTimeout {
		t.Fatalf("timeout: got %v want %v", cfg.HTTPClientTimeout, defaultHTTPTimeout)
	}
}

func TestLoad_HTTPTimeout(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "5s")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "")

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

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_HTTPTimeoutSeconds_priority(t *testing.T) {
	t.Setenv("NEUROFEED_HTTP_TIMEOUT", "1s")
	t.Setenv("NEUROFEED_HTTP_TIMEOUT_SECONDS", "99")

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
}
