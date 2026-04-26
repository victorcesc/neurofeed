package domain

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

// ScoreWeight returns the default bonus (or malus) for this tier in keyword-based scoring (Phase 3).
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
