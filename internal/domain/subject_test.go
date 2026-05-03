package domain

import "testing"

func TestSubjectBucketOrder_andHasConfigured(t *testing.T) {
	t.Parallel()
	articles := []Article{
		{Subject: "NBA", Title: "a"},
		{Subject: "  AI  ", Title: "b"},
		{Subject: "", Title: "c"},
		{Subject: "NBA", Title: "d"},
	}
	if !HasAnyConfiguredSubject(articles) {
		t.Fatal("expected configured subject")
	}
	order := SubjectBucketOrder(articles)
	if len(order) != 3 || order[0] != "NBA" || order[1] != "AI" || order[2] != DefaultArticleSubject {
		t.Fatalf("order %#v", order)
	}
}

func TestHasAnyConfiguredSubject_false(t *testing.T) {
	t.Parallel()
	if HasAnyConfiguredSubject([]Article{{Title: "x"}, {Title: "y"}}) {
		t.Fatal("expected false")
	}
	if len(SubjectBucketOrder([]Article{{Title: "x"}})) != 1 {
		t.Fatal("single Geral bucket")
	}
}

func TestEnrichSubjectOrderWithArticles(t *testing.T) {
	t.Parallel()
	cfgOrder := []string{"AI", "NBA"}
	articles := []Article{{Subject: "AI", Title: "only ai"}}
	got := EnrichSubjectOrderWithArticles(cfgOrder, articles)
	if len(got) != 2 || got[0] != "AI" || got[1] != "NBA" {
		t.Fatalf("got %#v", got)
	}
	articles2 := []Article{{Subject: "AI", Title: "a"}, {Title: "g"}}
	got2 := EnrichSubjectOrderWithArticles(cfgOrder, articles2)
	if len(got2) != 3 || got2[2] != DefaultArticleSubject {
		t.Fatalf("want Geral appended, got %#v", got2)
	}
}
