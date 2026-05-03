// Package domain holds core types and pure logic for articles, deduplication, and source tiers.
package domain

import "time"

// Article is a normalized news item from any ingest source.
type Article struct {
	Title       string
	Link        string
	Description string
	Source      string
	SourceTier  SourceTier
	// Subject is an optional section label from NEUROFEED_RSS_FEEDS JSON ("subject" per feed). Empty means bucket DefaultArticleSubject for grouping.
	Subject   string
	Published time.Time
}
