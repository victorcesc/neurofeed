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

func TestParseSourceTier(t *testing.T) {
	tests := []struct {
		input string
		want  SourceTier
	}{
		{"", SourceTierNews},
		{"  ", SourceTierNews},
		{"news", SourceTierNews},
		{"NEWS", SourceTierNews},
		{"primary", SourceTierPrimary},
		{"expert", SourceTierExpert},
		{"community", SourceTierCommunity},
	}
	for _, testCase := range tests {
		got, err := ParseSourceTier(testCase.input)
		if err != nil {
			t.Fatalf("ParseSourceTier(%q): %v", testCase.input, err)
		}
		if got != testCase.want {
			t.Fatalf("ParseSourceTier(%q) = %v, want %v", testCase.input, got, testCase.want)
		}
	}
	if _, err := ParseSourceTier("unknown-tier"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSourceTier_String(t *testing.T) {
	if got, want := SourceTierPrimary.String(), "primary"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if got := SourceTier(42).String(); got == "" || got == "primary" {
		t.Fatalf("unexpected String for unknown: %q", got)
	}
}
