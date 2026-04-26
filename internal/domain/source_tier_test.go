package domain

import "testing"

func TestSourceTier_ScoreWeight(t *testing.T) {
	tests := []struct {
		tier SourceTier
		want int
	}{
		{SourceTierUnspecified, 0},
		{SourceTierPrimary, DefaultTierWeightPrimary},
		{SourceTierExpert, DefaultTierWeightExpert},
		{SourceTierNews, DefaultTierWeightNews},
		{SourceTierCommunity, DefaultTierWeightCommunity},
		{SourceTier(99), 0},
	}
	for _, testCase := range tests {
		if actualWeight := testCase.tier.ScoreWeight(); actualWeight != testCase.want {
			t.Errorf("%d.ScoreWeight() = %d, want %d", testCase.tier, actualWeight, testCase.want)
		}
	}
}
