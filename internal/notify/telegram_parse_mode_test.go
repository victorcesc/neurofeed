package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelegramNotifier_Notify_HTMLParseMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["parse_mode"] != "HTML" {
			t.Fatalf("parse_mode: %v", payload["parse_mode"])
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	notifier := &TelegramNotifier{
		APIBaseURL: server.URL,
		Token:      "test-token",
		ChatID:     "1",
		Client:     server.Client(),
		ParseMode:  "HTML",
	}
	if err := notifier.Notify(context.Background(), "<b>x</b>"); err != nil {
		t.Fatal(err)
	}
}
