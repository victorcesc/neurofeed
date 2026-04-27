package domain

import (
	"reflect"
	"testing"
)

func TestNormalizeTitleKey(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Hello, World!", "helloworld"},
		{"  AI & ML — News  ", "aimlnews"},
		{"Same Title", "sametitle"},
		{"", ""},
	}
	for _, testCase := range tests {
		if got := NormalizeTitleKey(testCase.title); got != testCase.want {
			t.Errorf("NormalizeTitleKey(%q) = %q, want %q", testCase.title, got, testCase.want)
		}
	}
}

func TestDeduplicateArticlesByTitle(t *testing.T) {
	first := Article{Title: "Hello!", Link: "https://a.example/x", SourceTier: SourceTierNews}
	second := Article{Title: "hello?", Link: "https://b.example/y", SourceTier: SourceTierExpert}
	third := Article{Title: "Other", Link: "https://c.example/z", SourceTier: SourceTierPrimary}
	got := DeduplicateArticlesByTitle([]Article{first, second, third})
	want := []Article{first, third}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestDeduplicateArticlesByTitle_emptyTitleUsesLink(t *testing.T) {
	linkArticle := Article{Title: "", Link: "HTTPS://Same.EXAMPLE/path", SourceTier: SourceTierNews}
	dup := Article{Title: "   ", Link: "https://same.example/path", SourceTier: SourceTierExpert}
	got := DeduplicateArticlesByTitle([]Article{linkArticle, dup})
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if got[0].SourceTier != SourceTierNews {
		t.Fatalf("expected first article kept")
	}
}
