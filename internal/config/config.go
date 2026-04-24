// Package config loads and validates runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
)

// Config holds validated settings for the neurofeed pipeline.
// Secrets are read from the environment only; never commit real values.
type Config struct {
	HTTPClientTimeout time.Duration

	// Optional until later phases (Telegram, OpenAI, RSS).
	TelegramBotToken string
	OpenAIAPIKey     string
	RSSFeedURL       string
}

// Load reads configuration from the process environment.
func Load() (Config, error) {
	cfg := Config{
		HTTPClientTimeout: defaultHTTPTimeout,
	}

	if v := os.Getenv("NEUROFEED_HTTP_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT: %w", err)
		}
		if d <= 0 {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT must be positive")
		}
		cfg.HTTPClientTimeout = d
	}

	if v := os.Getenv("NEUROFEED_HTTP_TIMEOUT_SECONDS"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT_SECONDS: %w", err)
		}
		if sec <= 0 {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT_SECONDS must be positive")
		}
		cfg.HTTPClientTimeout = time.Duration(sec) * time.Second
	}

	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	cfg.RSSFeedURL = os.Getenv("RSS_FEED_URL")

	return cfg, nil
}
