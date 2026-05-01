package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/victorcesc/neurofeed/internal/config"
)

func TestOpenAIChatClient_ChatCompletion_ok(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method %s", request.Method)
		}
		if !strings.HasSuffix(request.URL.Path, "/v1/chat/completions") {
			t.Fatalf("path %s", request.URL.Path)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var decoded chatCompletionRequest
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Model != "gpt-test" || len(decoded.Messages) != 1 {
			t.Fatalf("payload %+v", decoded)
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("auth header missing")
		}
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":" OK \n"}}]}`))
	}))
	t.Cleanup(server.Close)

	cfg := config.Config{
		LLMBaseURL: strings.TrimSuffix(server.URL, "/") + "/v1",
		LLMAPIKey:  "secret",
		LLMModel:   "gpt-test",
	}
	httpClient := &http.Client{Timeout: 5 * time.Second}
	client, err := NewOpenAIChatClientFromConfig(cfg, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	reply, err := client.ChatCompletion(context.Background(), []ChatMessage{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if reply != "OK" {
		t.Fatalf("got %q", reply)
	}
}

func TestOpenAIChatClient_ChatCompletion_apiError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte(`{"error":{"message":"bad key","type":"invalid_request_error"}}`))
	}))
	t.Cleanup(server.Close)

	cfg := config.Config{
		LLMBaseURL: strings.TrimSuffix(server.URL, "/") + "/v1",
		LLMAPIKey:  "x",
		LLMModel:   "m",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ChatCompletion(context.Background(), []ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("got %v", err)
	}
}

func TestOpenAIChatClientFromConfig_defaultsAndTrim(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		LLMBaseURL: "https://example.com/v1/",
		LLMAPIKey:  "k",
		LLMModel:   "",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	if client.baseURL != "https://example.com/v1" {
		t.Fatalf("base %q", client.baseURL)
	}
	if client.model != "gpt-4o-mini" {
		t.Fatalf("model %q", client.model)
	}
}

func TestOpenAIChatClientFromConfig_errors(t *testing.T) {
	t.Parallel()

	_, err := NewOpenAIChatClientFromConfig(config.Config{LLMAPIKey: "k"}, nil)
	if err == nil {
		t.Fatal("expected nil client error")
	}
	_, err = NewOpenAIChatClientFromConfig(config.Config{LLMAPIKey: ""}, &http.Client{})
	if err == nil {
		t.Fatal("expected empty key error")
	}
}
