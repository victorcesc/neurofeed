package domain

import (
	"strings"
	"unicode"
)

// NormalizeTitleKey builds a deduplication key from a title: lowercased letters and digits only
// (punctuation and spacing removed), matching the product spec for basic title dedup.
func NormalizeTitleKey(title string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(strings.TrimSpace(title)) {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

// DeduplicateArticlesByTitle keeps the first article for each normalized title key.
// Articles whose normalized title is empty fall back to a normalized link key so items
// without titles still dedupe when the same URL appears from multiple feeds.
func DeduplicateArticlesByTitle(articles []Article) []Article {
	if len(articles) == 0 {
		return nil
	}
	seenDedupKeys := make(map[string]struct{}, len(articles))
	deduplicatedArticles := make([]Article, 0, len(articles))
	for articleIndex := range articles {
		article := articles[articleIndex]
		deduplicationKey := NormalizeTitleKey(article.Title)
		if deduplicationKey == "" {
			deduplicationKey = normalizeLinkDedupKey(article.Link)
		}
		if deduplicationKey == "" {
			deduplicatedArticles = append(deduplicatedArticles, article)
			continue
		}
		if _, alreadySeen := seenDedupKeys[deduplicationKey]; alreadySeen {
			continue
		}
		seenDedupKeys[deduplicationKey] = struct{}{}
		deduplicatedArticles = append(deduplicatedArticles, article)
	}
	return deduplicatedArticles
}

func normalizeLinkDedupKey(link string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(strings.TrimSpace(link)) {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
