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
	if !strings.Contains(summary, "<a href") || !strings.Contains(summary, "https://a") || !strings.Contains(summary, "A") {
		t.Fatalf("unexpected summary %q", summary)
	}
}

func TestHeadlineSummarizer_Summarize_sectioned(t *testing.T) {
	summary, err := HeadlineSummarizer{}.Summarize(context.Background(), []domain.Article{
		{Title: "A1", Link: "https://a1", Subject: "AI"},
		{Title: "B1", Link: "https://b1", Subject: "NBA"},
		{Title: "A2", Link: "https://a2", Subject: "AI"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "📌 AI") || !strings.Contains(summary, "📌 NBA") || !strings.Contains(summary, "<b>") {
		t.Fatalf("missing headers: %q", summary)
	}
	aiPos := strings.Index(summary, "📌 AI")
	nbaPos := strings.Index(summary, "📌 NBA")
	if aiPos == -1 || nbaPos == -1 || aiPos > nbaPos {
		t.Fatalf("want AI section before NBA in feed order, got %q", summary)
	}
}

func TestHeadlineSummarizer_SummarizeWithSubjectOrder_emptyConfiguredSection(t *testing.T) {
	t.Parallel()
	summary, err := HeadlineSummarizer{}.SummarizeWithSubjectOrder(context.Background(), []domain.Article{
		{Title: "Only AI", Link: "https://ai", Subject: "AI"},
	}, []string{"AI", "NBA"})
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(summary, "📌 AI", "📌 NBA") {
		t.Fatalf("missing section order: %q", summary)
	}
	if !containsInOrder(summary, "📌 AI", "https://ai") {
		t.Fatalf("expected AI link before NBA header: %q", summary)
	}
	if !containsInOrder(summary, "https://ai", "📌 NBA") {
		t.Fatalf("expected NBA header after AI block: %q", summary)
	}
	if !containsInOrder(summary, "📌 NBA", "Nada de novo") {
		t.Fatalf("expected empty NBA placeholder: %q", summary)
	}
}

func containsInOrder(haystack, a, b string) bool {
	ia := strings.Index(haystack, a)
	ib := strings.Index(haystack, b)
	return ia >= 0 && ib > ia
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
