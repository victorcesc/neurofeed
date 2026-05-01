package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/victorcesc/neurofeed/internal/config"
)

const openAIChatCompletionsPath = "/chat/completions"

// ChatMessage is one entry in an OpenAI-compatible chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// OpenAIChatClient calls OpenAI-compatible POST /v1/chat/completions (base URL may point to a proxy).
type OpenAIChatClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIChatClientFromConfig builds a client from validated config fields.
// Empty LLM_BASE_URL defaults to https://api.openai.com/v1; empty LLM_MODEL defaults to gpt-4o-mini.
// httpClient must be non-nil (typically with Timeout set to cfg.LLMRequestTimeout).
func NewOpenAIChatClientFromConfig(cfg config.Config, httpClient *http.Client) (*OpenAIChatClient, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("ai: http client is nil")
	}
	base := strings.TrimSpace(cfg.LLMBaseURL)
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	base = strings.TrimRight(base, "/")
	model := strings.TrimSpace(cfg.LLMModel)
	if model == "" {
		model = "gpt-4o-mini"
	}
	key := strings.TrimSpace(cfg.LLMAPIKey)
	if key == "" {
		return nil, fmt.Errorf("ai: LLM_API_KEY is empty")
	}
	return &OpenAIChatClient{
		baseURL:    base,
		apiKey:     key,
		model:      model,
		httpClient: httpClient,
	}, nil
}

// ChatCompletion sends a chat completion request and returns the first assistant message content (trimmed).
func (openAI *OpenAIChatClient) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "", fmt.Errorf("ai: no messages")
	}
	requestURL := openAI.baseURL + openAIChatCompletionsPath
	payload := chatCompletionRequest{
		Model:     openAI.model,
		Messages:  messages,
		MaxTokens: 256,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ai: encode request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ai: build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+openAI.apiKey)
	request.Header.Set("User-Agent", "neurofeed/1.0 (+https://github.com/victorcesc/neurofeed)")

	response, err := openAI.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("ai: http: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("ai: read body: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("ai: chat completions: status %d: %s", response.StatusCode, truncateForError(responseBytes, 512))
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(responseBytes, &decoded); err != nil {
		return "", fmt.Errorf("ai: decode response: %w", err)
	}
	if decoded.Error != nil && decoded.Error.Message != "" {
		return "", fmt.Errorf("ai: api error: %s", decoded.Error.Message)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("ai: empty choices in response")
	}
	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("ai: empty assistant content")
	}
	return content, nil
}

func truncateForError(data []byte, max int) string {
	s := string(data)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
