// Package config loads and validates runtime configuration from the environment.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/victorcesc/neurofeed/internal/domain"
)

const (
	defaultHTTPTimeout          = 30 * time.Second
	defaultLLMRequestTimeout    = 60 * time.Second
	defaultLLMMaxDigestArticles = 12
	defaultLLMMaxOutputTokens   = 2500
	defaultRSSMaxItemsPerFeed   = 2
	minLLMMaxDigestArticles     = 1
	maxLLMMaxDigestArticles     = 40
	minLLMMaxOutputTokens       = 256
	maxLLMMaxOutputTokens       = 8192
	minRSSMaxItemsPerFeed       = 0
	maxRSSMaxItemsPerFeed       = 50
	envLLMRequestTimeout        = "NEUROFEED_LLM_TIMEOUT"
	envLLMMaxDigestArticles     = "NEUROFEED_LLM_MAX_ARTICLES"
	envLLMMaxOutputTokens       = "NEUROFEED_LLM_MAX_OUTPUT_TOKENS"
	envRSSMaxItemsPerFeed       = "NEUROFEED_RSS_ITEMS_PER_FEED"
	envRSSFeedsJSON             = "NEUROFEED_RSS_FEEDS"
	envRSSFeedURL               = "RSS_FEED_URL"
	envRSSFeedTier              = "RSS_FEED_TIER"
	defaultSingleFeedTierName   = "news"
	maxRSSFeedSubjectRunes      = 64
	envRSSFeedSubject           = "RSS_FEED_SUBJECT"
)

// RSSFeedEntry is one RSS source with an optional tier string (primary, expert, news, community).
// Empty tier defaults to news when resolved via domain.ParseSourceTier.
type RSSFeedEntry struct {
	URL     string
	Tier    string
	Subject string
}

// Config holds validated settings for the neurofeed pipeline.
// Secrets are read from the environment only; never commit real values.
type Config struct {
	HTTPClientTimeout time.Duration
	// LLMRequestTimeout bounds each LLM HTTP call (chat completions). Separate from RSS/Telegram client timeout.
	LLMRequestTimeout time.Duration
	// LLMMaxDigestArticles caps how many articles are sent to the digest model (after dedup, in feed order).
	LLMMaxDigestArticles int
	// LLMMaxOutputTokens is the chat completion max_tokens budget for digest summarization.
	LLMMaxOutputTokens int

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
	// RSSMaxItemsPerFeed keeps the N newest items per feed URL after parse (0 = no cap). Default 2.
	RSSMaxItemsPerFeed int
}

type rssFeedJSON struct {
	URL     string `json:"url"`
	Tier    string `json:"tier"`
	Subject string `json:"subject"`
}

// Load reads configuration from the process environment.
func Load() (Config, error) {
	cfg := Config{
		HTTPClientTimeout:    defaultHTTPTimeout,
		LLMRequestTimeout:    defaultLLMRequestTimeout,
		LLMMaxDigestArticles: defaultLLMMaxDigestArticles,
		LLMMaxOutputTokens:   defaultLLMMaxOutputTokens,
		RSSMaxItemsPerFeed:   defaultRSSMaxItemsPerFeed,
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

	if maxArticlesValue := strings.TrimSpace(os.Getenv(envLLMMaxDigestArticles)); maxArticlesValue != "" {
		maxArticles, err := strconv.Atoi(maxArticlesValue)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", envLLMMaxDigestArticles, err)
		}
		if maxArticles < minLLMMaxDigestArticles || maxArticles > maxLLMMaxDigestArticles {
			return Config{}, fmt.Errorf("%s must be between %d and %d", envLLMMaxDigestArticles, minLLMMaxDigestArticles, maxLLMMaxDigestArticles)
		}
		cfg.LLMMaxDigestArticles = maxArticles
	}

	if maxOutputTokensValue := strings.TrimSpace(os.Getenv(envLLMMaxOutputTokens)); maxOutputTokensValue != "" {
		maxOutputTokens, err := strconv.Atoi(maxOutputTokensValue)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", envLLMMaxOutputTokens, err)
		}
		if maxOutputTokens < minLLMMaxOutputTokens || maxOutputTokens > maxLLMMaxOutputTokens {
			return Config{}, fmt.Errorf("%s must be between %d and %d", envLLMMaxOutputTokens, minLLMMaxOutputTokens, maxLLMMaxOutputTokens)
		}
		cfg.LLMMaxOutputTokens = maxOutputTokens
	}

	if rssItemsValue := strings.TrimSpace(os.Getenv(envRSSMaxItemsPerFeed)); rssItemsValue != "" {
		rssItems, err := strconv.Atoi(rssItemsValue)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", envRSSMaxItemsPerFeed, err)
		}
		if rssItems < minRSSMaxItemsPerFeed || rssItems > maxRSSMaxItemsPerFeed {
			return Config{}, fmt.Errorf("%s must be between %d and %d (0 disables the per-feed cap)", envRSSMaxItemsPerFeed, minRSSMaxItemsPerFeed, maxRSSMaxItemsPerFeed)
		}
		cfg.RSSMaxItemsPerFeed = rssItems
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
			subject := strings.TrimSpace(row.Subject)
			if subject != "" && utf8.RuneCountInString(subject) > maxRSSFeedSubjectRunes {
				return Config{}, fmt.Errorf("%s: entry %d: subject exceeds %d characters", envRSSFeedsJSON, index, maxRSSFeedSubjectRunes)
			}
			cfg.RSSFeeds = append(cfg.RSSFeeds, RSSFeedEntry{URL: url, Tier: tier, Subject: subject})
		}
	} else if cfg.RSSFeedURL != "" {
		tier := strings.TrimSpace(os.Getenv(envRSSFeedTier))
		if tier == "" {
			tier = defaultSingleFeedTierName
		}
		if _, err := domain.ParseSourceTier(tier); err != nil {
			return Config{}, fmt.Errorf("%s: %w", envRSSFeedTier, err)
		}
		singleSubject := strings.TrimSpace(os.Getenv(envRSSFeedSubject))
		if singleSubject != "" && utf8.RuneCountInString(singleSubject) > maxRSSFeedSubjectRunes {
			return Config{}, fmt.Errorf("%s exceeds %d characters", envRSSFeedSubject, maxRSSFeedSubjectRunes)
		}
		cfg.RSSFeeds = []RSSFeedEntry{{URL: cfg.RSSFeedURL, Tier: tier, Subject: singleSubject}}
	}

	return cfg, nil
}

// DigestSubjectSections returns distinct non-empty feed subject labels in configured order
// (first index in RSSFeeds wins). Used so Telegram digests list every configured topic even
// when a feed produced no items in this run.
func (cfg Config) DigestSubjectSections() []string {
	var out []string
	seen := map[string]struct{}{}
	for index := range cfg.RSSFeeds {
		subject := strings.TrimSpace(cfg.RSSFeeds[index].Subject)
		if subject == "" {
			continue
		}
		if _, ok := seen[subject]; ok {
			continue
		}
		seen[subject] = struct{}{}
		out = append(out, subject)
	}
	return out
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
