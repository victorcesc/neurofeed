// Package domain holds core types and pure logic for articles, deduplication, and scoring.
package domain

import "time"

// Article is a normalized news item from any ingest source.
type Article struct {
	Title       string
	Link        string
	Description string
	Source      string
	Published   time.Time
}
