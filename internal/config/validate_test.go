package config

import "testing"

func TestValidateLLMSmoke(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := ValidateLLMSmoke(Config{LLMAPIKey: "k", LLMProvider: ""})
		if err != nil {
			t.Fatal(err)
		}
		err = ValidateLLMSmoke(Config{LLMAPIKey: "k", LLMProvider: "openai"})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("missing key", func(t *testing.T) {
		err := ValidateLLMSmoke(Config{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("wrong provider", func(t *testing.T) {
		err := ValidateLLMSmoke(Config{LLMAPIKey: "k", LLMProvider: "anthropic"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestValidatePhase1(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := ValidatePhase1(Config{
			RSSFeeds: []RSSFeedEntry{
				{URL: "https://example.com/feed.xml", Tier: "news"},
			},
			TelegramBotToken: "t",
			TelegramChatID:   "1",
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("missing", func(t *testing.T) {
		err := ValidatePhase1(Config{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
