package domain

import (
	"fmt"
	"strings"
)

// SourceTier classifies a feed on the editorial / signal ladder used for scoring.
// Config maps each RSS URL to a tier; ingest stamps Article.SourceTier from that mapping.
type SourceTier int

const (
	// SourceTierUnspecified means no tier was configured (neutral in score weight).
	SourceTierUnspecified SourceTier = iota
	// SourceTierPrimary: primários — reguladores, datasets, filings, release notes oficiais.
	SourceTierPrimary
	// SourceTierExpert: síntese especializada — revisões, documentação de referência de domínio.
	SourceTierExpert
	// SourceTierNews: notícias — veículos com corpo editorial que costumam linkar primários.
	SourceTierNews
	// SourceTierCommunity: comentário, redes, newsletters opinativas sem primário obrigatório.
	SourceTierCommunity
)

// Default score contribution per tier. Profiles or deployment config may override these values
// when building the scorer (same keys, custom ints).
const (
	DefaultTierWeightPrimary   = 4
	DefaultTierWeightExpert    = 3
	DefaultTierWeightNews      = 2
	DefaultTierWeightCommunity = -1
)

// ScoreWeight returns the default bonus (or malus) for this tier when relevance scoring exists (planned product spec).
func (sourceTier SourceTier) ScoreWeight() int {
	switch sourceTier {
	case SourceTierPrimary:
		return DefaultTierWeightPrimary
	case SourceTierExpert:
		return DefaultTierWeightExpert
	case SourceTierNews:
		return DefaultTierWeightNews
	case SourceTierCommunity:
		return DefaultTierWeightCommunity
	default:
		return 0
	}
}

// String returns the tier name for logs (config-style), or unknown(N) for unexpected values.
func (sourceTier SourceTier) String() string {
	switch sourceTier {
	case SourceTierUnspecified:
		return "unspecified"
	case SourceTierPrimary:
		return "primary"
	case SourceTierExpert:
		return "expert"
	case SourceTierNews:
		return "news"
	case SourceTierCommunity:
		return "community"
	default:
		return fmt.Sprintf("unknown(%d)", int(sourceTier))
	}
}

// ParseSourceTier maps a config string to SourceTier. Empty or whitespace defaults to news.
// Accepts case-insensitive names: primary, expert, news, community.
func ParseSourceTier(value string) (SourceTier, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "news":
		return SourceTierNews, nil
	case "primary":
		return SourceTierPrimary, nil
	case "expert":
		return SourceTierExpert, nil
	case "community":
		return SourceTierCommunity, nil
	default:
		return SourceTierUnspecified, fmt.Errorf("unknown source tier %q (want primary, expert, news, or community)", value)
	}
}
