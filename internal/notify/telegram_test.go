package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTelegramNotifier_Notify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.Contains(request.URL.Path, "/bot") || !strings.HasSuffix(request.URL.Path, "/sendMessage") {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["chat_id"] != "99" {
			t.Fatalf("chat_id: %v", payload["chat_id"])
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	notifier := &TelegramNotifier{
		APIBaseURL: server.URL,
		Token:      "test-token",
		ChatID:     "99",
		Client:     server.Client(),
	}
	if err := notifier.Notify(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
}

func TestTelegramNotifier_Notify_apiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":false,"description":"bad"}`))
	}))
	t.Cleanup(server.Close)

	notifier := &TelegramNotifier{
		APIBaseURL: server.URL,
		Token:      "t",
		ChatID:     "1",
		Client:     server.Client(),
	}
	if err := notifier.Notify(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
