package domain

import "strings"

// DefaultArticleSubject is the bucket label for articles whose feed had no "subject" in config.
const DefaultArticleSubject = "Geral"

// RawSubjectTrimmed returns the configured subject from RSS env JSON after trim (may be empty).
func (article Article) RawSubjectTrimmed() string {
	return strings.TrimSpace(article.Subject)
}

// BucketSubject returns the display bucket for grouping: explicit subject or DefaultArticleSubject.
func (article Article) BucketSubject() string {
	if s := article.RawSubjectTrimmed(); s != "" {
		return s
	}
	return DefaultArticleSubject
}

// HasAnyConfiguredSubject reports whether any article came from a feed with a non-empty subject field.
func HasAnyConfiguredSubject(articles []Article) bool {
	for index := range articles {
		if articles[index].RawSubjectTrimmed() != "" {
			return true
		}
	}
	return false
}

// SubjectBucketOrder returns distinct bucket labels in first-seen order across articles.
func SubjectBucketOrder(articles []Article) []string {
	var order []string
	seen := map[string]struct{}{}
	for index := range articles {
		key := articles[index].BucketSubject()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		order = append(order, key)
	}
	return order
}

// EnrichSubjectOrderWithArticles returns configOrder first (deduped), then any article bucket labels
// from SubjectBucketOrder that were not already listed (e.g. "Geral" when some feeds omit subject).
func EnrichSubjectOrderWithArticles(configOrder []string, articles []Article) []string {
	seen := map[string]struct{}{}
	var out []string
	for index := range configOrder {
		key := strings.TrimSpace(configOrder[index])
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	for _, key := range SubjectBucketOrder(articles) {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
