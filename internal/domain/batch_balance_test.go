package domain

import "testing"

func TestBalanceArticlesAcrossSubjects_roundRobin(t *testing.T) {
	t.Parallel()
	order := []string{"AI", "NBA", "Tech"}
	var articles []Article
	for i := range 8 {
		articles = append(articles, Article{Title: "ai", Link: "https://ai" + string(rune('0'+i)), Subject: "AI"})
	}
	for i := range 8 {
		articles = append(articles, Article{Title: "nba", Link: "https://nba" + string(rune('0'+i)), Subject: "NBA"})
	}
	for i := range 8 {
		articles = append(articles, Article{Title: "tech", Link: "https://tech" + string(rune('0'+i)), Subject: "Tech"})
	}
	got := BalanceArticlesAcrossSubjects(order, articles, 9)
	if len(got) != 9 {
		t.Fatalf("len %d", len(got))
	}
	wantSubjects := []string{"AI", "NBA", "Tech", "AI", "NBA", "Tech", "AI", "NBA", "Tech"}
	for i := range wantSubjects {
		if got[i].BucketSubject() != wantSubjects[i] {
			t.Fatalf("index %d: want %s got %s", i, wantSubjects[i], got[i].BucketSubject())
		}
	}
}

func TestBalanceArticlesAcrossSubjects_emptyOrderFallsBackToHead(t *testing.T) {
	t.Parallel()
	articles := []Article{
		{Title: "a", Link: "https://1", Subject: "X"},
		{Title: "b", Link: "https://2", Subject: "Y"},
	}
	got := BalanceArticlesAcrossSubjects(nil, articles, 1)
	if len(got) != 1 || got[0].Link != "https://1" {
		t.Fatalf("got %+v", got)
	}
}

func TestBalanceArticlesAcrossSubjects_orphanSubjectAppended(t *testing.T) {
	t.Parallel()
	order := []string{"AI"}
	articles := []Article{
		{Title: "a", Link: "https://a", Subject: "AI"},
		{Title: "o", Link: "https://o", Subject: "Other"},
	}
	got := BalanceArticlesAcrossSubjects(order, articles, 2)
	if len(got) != 2 {
		t.Fatalf("len %d", len(got))
	}
	if got[0].Subject != "AI" || got[1].Subject != "Other" {
		t.Fatalf("got %+v", got)
	}
}
