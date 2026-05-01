// Package config loads and validates runtime configuration from the environment.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/victorcesc/neurofeed/internal/domain"
)

const (
	defaultHTTPTimeout        = 30 * time.Second
	defaultLLMRequestTimeout  = 60 * time.Second
	envLLMRequestTimeout      = "NEUROFEED_LLM_TIMEOUT"
	envRSSFeedsJSON           = "NEUROFEED_RSS_FEEDS"
	envRSSFeedURL             = "RSS_FEED_URL"
	envRSSFeedTier            = "RSS_FEED_TIER"
	defaultSingleFeedTierName = "news"
)

// RSSFeedEntry is one RSS source with an optional tier string (primary, expert, news, community).
// Empty tier defaults to news when resolved via domain.ParseSourceTier.
type RSSFeedEntry struct {
	URL  string
	Tier string
}

// Config holds validated settings for the neurofeed pipeline.
// Secrets are read from the environment only; never commit real values.
type Config struct {
	HTTPClientTimeout time.Duration
	// LLMRequestTimeout bounds each LLM HTTP call (chat completions). Separate from RSS/Telegram client timeout.
	LLMRequestTimeout time.Duration

	// Optional until later phases (Telegram, LLM, RSS).
	TelegramBotToken string
	TelegramChatID   string
	LLMProvider      string
	LLMModel         string
	LLMBaseURL       string
	LLMAPIKey        string

	// RSSFeedURL is the raw single-feed env value (legacy); canonical list is RSSFeeds after Load.
	RSSFeedURL string
	RSSFeeds   []RSSFeedEntry
}

type rssFeedJSON struct {
	URL  string `json:"url"`
	Tier string `json:"tier"`
}

// Load reads configuration from the process environment.
func Load() (Config, error) {
	cfg := Config{
		HTTPClientTimeout: defaultHTTPTimeout,
		LLMRequestTimeout: defaultLLMRequestTimeout,
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

	if llmTimeoutValue := strings.TrimSpace(os.Getenv(envLLMRequestTimeout)); llmTimeoutValue != "" {
		llmTimeoutDuration, err := time.ParseDuration(llmTimeoutValue)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", envLLMRequestTimeout, err)
		}
		if llmTimeoutDuration <= 0 {
			return Config{}, fmt.Errorf("%s must be positive", envLLMRequestTimeout)
		}
		cfg.LLMRequestTimeout = llmTimeoutDuration
	}

	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.TelegramChatID = os.Getenv("TELEGRAM_CHAT_ID")
	cfg.LLMProvider = os.Getenv("LLM_PROVIDER")
	cfg.LLMModel = os.Getenv("LLM_MODEL")
	cfg.LLMBaseURL = os.Getenv("LLM_BASE_URL")
	cfg.LLMAPIKey = os.Getenv("LLM_API_KEY")
	cfg.RSSFeedURL = strings.TrimSpace(os.Getenv(envRSSFeedURL))

	feedsJSON := strings.TrimSpace(os.Getenv(envRSSFeedsJSON))
	if feedsJSON != "" {
		var decoded []rssFeedJSON
		if err := json.Unmarshal([]byte(feedsJSON), &decoded); err != nil {
			return Config{}, fmt.Errorf("%s: invalid JSON: %w", envRSSFeedsJSON, err)
		}
		if len(decoded) == 0 {
			return Config{}, fmt.Errorf("%s: must contain at least one feed object", envRSSFeedsJSON)
		}
		cfg.RSSFeeds = make([]RSSFeedEntry, 0, len(decoded))
		for index, row := range decoded {
			url := strings.TrimSpace(row.URL)
			if url == "" {
				return Config{}, fmt.Errorf("%s: entry %d: missing url", envRSSFeedsJSON, index)
			}
			tier := strings.TrimSpace(row.Tier)
			if _, err := domain.ParseSourceTier(tier); err != nil {
				return Config{}, fmt.Errorf("%s: entry %d: %w", envRSSFeedsJSON, index, err)
			}
			cfg.RSSFeeds = append(cfg.RSSFeeds, RSSFeedEntry{URL: url, Tier: tier})
		}
	} else if cfg.RSSFeedURL != "" {
		tier := strings.TrimSpace(os.Getenv(envRSSFeedTier))
		if tier == "" {
			tier = defaultSingleFeedTierName
		}
		if _, err := domain.ParseSourceTier(tier); err != nil {
			return Config{}, fmt.Errorf("%s: %w", envRSSFeedTier, err)
		}
		cfg.RSSFeeds = []RSSFeedEntry{{URL: cfg.RSSFeedURL, Tier: tier}}
	}

	return cfg, nil
}

// ValidatePhase1 returns an error if required settings for the RSS → Telegram MVP are missing.
func ValidatePhase1(cfg Config) error {
	var missing []string
	if len(cfg.RSSFeeds) == 0 {
		missing = append(missing, envRSSFeedURL+" or "+envRSSFeedsJSON)
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

// ValidateLLMSmoke returns an error if the environment is not ready for a minimal OpenAI chat completion (flag -llm-smoke).
func ValidateLLMSmoke(cfg Config) error {
	if strings.TrimSpace(cfg.LLMAPIKey) == "" {
		return fmt.Errorf("LLM_API_KEY is required for -llm-smoke")
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.LLMProvider))
	if provider != "" && provider != "openai" {
		return fmt.Errorf("LLM_PROVIDER must be openai or empty for -llm-smoke (got %q)", cfg.LLMProvider)
	}
	return nil
}
