// Package config loads and validates runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
)

// Config holds validated settings for the neurofeed pipeline.
// Secrets are read from the environment only; never commit real values.
type Config struct {
	HTTPClientTimeout time.Duration

	// Optional until later phases (Telegram, LLM, RSS).
	TelegramBotToken string
	TelegramChatID   string
	LLMProvider      string
	LLMModel         string
	LLMBaseURL       string
	LLMAPIKey        string
	RSSFeedURL       string
}

// Load reads configuration from the process environment.
func Load() (Config, error) {
	cfg := Config{
		HTTPClientTimeout: defaultHTTPTimeout,
	}

	if httpTimeoutValue := os.Getenv("NEUROFEED_HTTP_TIMEOUT"); httpTimeoutValue != "" {
		timeoutDuration, err := time.ParseDuration(httpTimeoutValue)
		if err != nil {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT: %w", err)
		}
		if timeoutDuration <= 0 {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT must be positive")
		}
		cfg.HTTPClientTimeout = timeoutDuration
	}

	if httpTimeoutSecondsValue := os.Getenv("NEUROFEED_HTTP_TIMEOUT_SECONDS"); httpTimeoutSecondsValue != "" {
		timeoutSeconds, err := strconv.Atoi(httpTimeoutSecondsValue)
		if err != nil {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT_SECONDS: %w", err)
		}
		if timeoutSeconds <= 0 {
			return Config{}, fmt.Errorf("NEUROFEED_HTTP_TIMEOUT_SECONDS must be positive")
		}
		cfg.HTTPClientTimeout = time.Duration(timeoutSeconds) * time.Second
	}

	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.TelegramChatID = os.Getenv("TELEGRAM_CHAT_ID")
	cfg.LLMProvider = os.Getenv("LLM_PROVIDER")
	cfg.LLMModel = os.Getenv("LLM_MODEL")
	cfg.LLMBaseURL = os.Getenv("LLM_BASE_URL")
	cfg.LLMAPIKey = os.Getenv("LLM_API_KEY")
	cfg.RSSFeedURL = os.Getenv("RSS_FEED_URL")

	return cfg, nil
}

// ValidatePhase1 returns an error if required settings for the RSS → Telegram MVP are missing.
func ValidatePhase1(cfg Config) error {
	var missing []string
	if strings.TrimSpace(cfg.RSSFeedURL) == "" {
		missing = append(missing, "RSS_FEED_URL")
	}
	if strings.TrimSpace(cfg.TelegramBotToken) == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if strings.TrimSpace(cfg.TelegramChatID) == "" {
		missing = append(missing, "TELEGRAM_CHAT_ID")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("phase 1 requires %s", strings.Join(missing, ", "))
}
