package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/victorcesc/neurofeed/internal/domain"
)

func TestHeadlineSummarizer_Summarize(t *testing.T) {
	summary, err := HeadlineSummarizer{}.Summarize(context.Background(), []domain.Article{
		{Title: "A", Link: "https://a"},
		{Title: "B", Link: "https://b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "A") || !strings.Contains(summary, "https://a") {
		t.Fatalf("unexpected summary %q", summary)
	}
}

func TestHeadlineSummarizer_Summarize_skipsNoLink(t *testing.T) {
	summary, err := HeadlineSummarizer{}.Summarize(context.Background(), []domain.Article{
		{Title: "No link"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary != "" {
		t.Fatalf("want empty, got %q", summary)
	}
}
