package config

import "testing"

func TestValidatePhase1(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := ValidatePhase1(Config{
			RSSFeedURL:       "https://example.com/feed.xml",
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
