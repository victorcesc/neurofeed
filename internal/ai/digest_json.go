package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

const digestJSONMaxRunes = 12000

type digestJSONPayload struct {
	Digest string `json:"digest"`
}

// parseDigestJSON extracts and validates the digest field from model output (optionally wrapped in markdown fences).
func parseDigestJSON(raw string) (string, error) {
	trimmed := stripMarkdownJSONFence(strings.TrimSpace(raw))
	var payload digestJSONPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return "", fmt.Errorf("ai: parse digest json: %w", err)
	}
	digest := strings.TrimSpace(payload.Digest)
	if digest == "" {
		return "", fmt.Errorf("ai: digest json: empty digest")
	}
	runes := []rune(digest)
	if len(runes) > digestJSONMaxRunes {
		digest = string(runes[:digestJSONMaxRunes-1]) + "…"
	}
	return digest, nil
}

func stripMarkdownJSONFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	close := strings.LastIndex(s, "```")
	if close <= 3 {
		return strings.TrimSpace(strings.TrimPrefix(s, "```"))
	}
	inner := strings.TrimSpace(s[3:close])
	if strings.HasPrefix(inner, "json") {
		inner = strings.TrimSpace(inner[len("json"):])
	}
	inner = strings.TrimSpace(strings.TrimPrefix(inner, "\n"))
	return strings.TrimSpace(inner)
}
