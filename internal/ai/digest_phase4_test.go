package ai

import (
	"strings"
	"testing"

	"github.com/victorcesc/neurofeed/internal/domain"
)

func TestAssembleFlatPhase4HTML(t *testing.T) {
	t.Parallel()
	batch := []domain.Article{
		{Title: "One", Link: "https://one", Source: "S", SourceTier: domain.SourceTierNews},
		{Title: "Two", Link: "https://two", Source: "S", SourceTier: domain.SourceTierNews},
	}
	picks := []digestPickJSON{
		{Index: 1, Line1: "First fact", Line2: "First impact"},
		{Index: 2, Line1: "Second fact", Line2: "Second impact"},
	}
	htmlOut, err := assembleFlatPhase4HTML(batch, picks)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(htmlOut, "https://one") || !strings.Contains(htmlOut, "First fact") {
		t.Fatalf("got %q", htmlOut)
	}
}

func TestAssembleSectionedPhase4HTML_placeholder(t *testing.T) {
	t.Parallel()
	order := []string{"AI", "NBA"}
	perSection := [][]domain.Article{
		{{Title: "T", Link: "https://t", Subject: "AI"}},
		nil,
	}
	raw := `{"sections":[{"subject":"AI","picks":[{"index":1,"line1":"a","line2":"b"}]},{"subject":"NBA","picks":[]}]}`
	rows, err := parseDigestSectionsPicksJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	htmlOut, err := assembleSectionedPhase4HTML(order, perSection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(htmlOut, "📌 AI") || !strings.Contains(htmlOut, "Nada de novo") {
		t.Fatalf("got %q", htmlOut)
	}
}

func TestNormalizePicksForArticles_fillsFromOrder(t *testing.T) {
	t.Parallel()
	arts := []domain.Article{
		{Title: "A", Link: "https://a"},
		{Title: "B", Link: "https://b"},
	}
	out := normalizePicksForArticles(nil, arts, 2)
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
	if out[0].Article.Link != "https://a" || out[1].Article.Link != "https://b" {
		t.Fatalf("order %+v", out)
	}
}

func TestNormalizePicksForArticles_prefersLinkWhenIndexWrong(t *testing.T) {
	t.Parallel()
	arts := []domain.Article{
		{Title: "Celtics recap", Link: "https://espn/celtics"},
		{Title: "Lakers drama", Link: "https://espn/lakers"},
	}
	// Model summarized Lakers but mis-labeled index as 1 (first row).
	picks := []digestPickJSON{
		{Index: 1, Link: "https://espn/lakers", Line1: "Jeanie Buss spoke.", Line2: "Team optics shift."},
	}
	out := normalizePicksForArticles(picks, arts, 2)
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
	if out[0].Article.Link != "https://espn/lakers" || out[0].Line1 != "Jeanie Buss spoke." {
		t.Fatalf("first pick: %+v", out[0])
	}
	if out[1].Article.Link != "https://espn/celtics" {
		t.Fatalf("fallback fill want celtics, got %+v", out[1])
	}
}
