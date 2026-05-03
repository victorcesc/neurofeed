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
	"github.com/victorcesc/neurofeed/internal/domain"
)

func TestDigestSummarizer_Summarize_ok(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var decoded chatCompletionRequest
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.ResponseFormat == nil || decoded.ResponseFormat.Type != "json_object" {
			t.Fatalf("missing json_object: %+v", decoded.ResponseFormat)
		}
		if decoded.MaxTokens != 900 {
			t.Fatalf("max tokens %d", decoded.MaxTokens)
		}
		if decoded.Temperature == nil || *decoded.Temperature != digestLLMTemperature {
			t.Fatalf("temperature want %v got %+v", digestLLMTemperature, decoded.Temperature)
		}
		assistant := `{"picks":[{"index":1,"line1":"Item one","line2":"Impacto: test."}]}`
		responseBody, err := json.Marshal(map[string]any{
			"choices": []any{
				map[string]any{
					"message": map[string]any{
						"role":    "assistant",
						"content": assistant,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = writer.Write(responseBody)
	}))
	t.Cleanup(server.Close)

	cfg := config.Config{
		LLMBaseURL: strings.TrimSuffix(server.URL, "/") + "/v1",
		LLMAPIKey:  "secret",
		LLMModel:   "gpt-test",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	summarizer := NewDigestSummarizer(client, 5, 900)
	summary, err := summarizer.Summarize(context.Background(), []domain.Article{
		{Title: "T1", Link: "https://a", Source: "S", SourceTier: domain.SourceTierNews},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "Item one") || !strings.Contains(summary, "<b>") || !strings.Contains(summary, "<a href") {
		t.Fatalf("unexpected %q", summary)
	}
}

func TestDigestSummarizer_SummarizeWithSubjectOrder_includesEmptySection(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var decoded chatCompletionRequest
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(decoded.Messages[1].Content, "## NBA") {
			t.Fatalf("user payload missing NBA section: %s", decoded.Messages[1].Content)
		}
		assistant := `{"sections":[
			{"subject":"AI","picks":[{"index":1,"line1":"a","line2":"b"},{"index":2,"line1":"c","line2":"d"}]},
			{"subject":"NBA","picks":[]}
		]}`
		responseBody, err := json.Marshal(map[string]any{
			"choices": []any{
				map[string]any{"message": map[string]any{"role": "assistant", "content": assistant}},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = writer.Write(responseBody)
	}))
	t.Cleanup(server.Close)

	cfg := config.Config{
		LLMBaseURL: strings.TrimSuffix(server.URL, "/") + "/v1",
		LLMAPIKey:  "secret",
		LLMModel:   "gpt-test",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	summarizer := NewDigestSummarizer(client, 10, 900)
	articles := []domain.Article{
		{Title: "T1", Link: "https://a1", Source: "S", SourceTier: domain.SourceTierNews, Subject: "AI"},
		{Title: "T2", Link: "https://a2", Source: "S", SourceTier: domain.SourceTierNews, Subject: "AI"},
	}
	summary, err := summarizer.SummarizeWithSubjectOrder(context.Background(), articles, []string{"AI", "NBA"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "📌 AI") || !strings.Contains(summary, "📌 NBA") {
		t.Fatalf("missing headers: %q", summary)
	}
	if strings.Index(summary, "📌 NBA") <= strings.Index(summary, "📌 AI") {
		t.Fatalf("want AI before NBA: %q", summary)
	}
	if !strings.Contains(summary, "Nada de novo") {
		t.Fatalf("want placeholder for empty NBA: %q", summary)
	}
}

func TestDigestSummarizer_Summarize_sectionsPath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var decoded chatCompletionRequest
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatal(err)
		}
		if len(decoded.Messages) != 2 || !strings.Contains(decoded.Messages[0].Content, "topic section") {
			t.Fatalf("expected sections system prompt, got %q", decoded.Messages[0].Content)
		}
		if !strings.Contains(decoded.Messages[1].Content, "## AI") {
			t.Fatalf("user payload missing section header: %s", decoded.Messages[1].Content)
		}
		assistant := `{"sections":[{"subject":"AI","picks":[{"index":1,"line1":"Summary here","line2":"Impact line"}]}]}`
		responseBody, err := json.Marshal(map[string]any{
			"choices": []any{
				map[string]any{"message": map[string]any{"role": "assistant", "content": assistant}},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = writer.Write(responseBody)
	}))
	t.Cleanup(server.Close)

	cfg := config.Config{
		LLMBaseURL: strings.TrimSuffix(server.URL, "/") + "/v1",
		LLMAPIKey:  "secret",
		LLMModel:   "gpt-test",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	summarizer := NewDigestSummarizer(client, 5, 900)
	summary, err := summarizer.Summarize(context.Background(), []domain.Article{
		{Title: "T1", Link: "https://a", Source: "S", SourceTier: domain.SourceTierNews, Subject: "AI"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "Summary here") || !strings.Contains(summary, "📌 AI") || !strings.Contains(summary, "<a href") {
		t.Fatalf("unexpected %q", summary)
	}
}

func TestDigestSummarizer_Summarize_emptyArticles(t *testing.T) {
	t.Parallel()
	cfg := config.Config{
		LLMBaseURL: "https://example.com/v1",
		LLMAPIKey:  "k",
		LLMModel:   "m",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	summarizer := NewDigestSummarizer(client, 5, 900)
	summary, err := summarizer.Summarize(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary != "" {
		t.Fatalf("want empty")
	}
}

func TestDigestSummarizer_Summarize_capsArticles(t *testing.T) {
	t.Parallel()

	var seenArticles int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		var decoded chatCompletionRequest
		_ = json.Unmarshal(body, &decoded)
		user := decoded.Messages[1].Content
		seenArticles = strings.Count(user, "Title:")
		assistant := `{"picks":[{"index":1,"line1":"a","line2":"b"},{"index":2,"line1":"c","line2":"d"}]}`
		responseBody, err := json.Marshal(map[string]any{
			"choices": []any{
				map[string]any{"message": map[string]any{"role": "assistant", "content": assistant}},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = writer.Write(responseBody)
	}))
	t.Cleanup(server.Close)

	cfg := config.Config{
		LLMBaseURL: strings.TrimSuffix(server.URL, "/") + "/v1",
		LLMAPIKey:  "secret",
		LLMModel:   "gpt-test",
	}
	client, err := NewOpenAIChatClientFromConfig(cfg, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	summarizer := NewDigestSummarizer(client, 2, 500)
	articles := []domain.Article{
		{Title: "1", Link: "https://1"},
		{Title: "2", Link: "https://2"},
		{Title: "3", Link: "https://3"},
	}
	_, err = summarizer.Summarize(context.Background(), articles)
	if err != nil {
		t.Fatal(err)
	}
	if seenArticles != 2 {
		t.Fatalf("model saw %d titles want 2", seenArticles)
	}
}
