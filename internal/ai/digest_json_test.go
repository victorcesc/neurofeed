package ai

import (
	"strings"
	"testing"
)

func TestParseDigestJSON_plain(t *testing.T) {
	t.Parallel()
	out, err := parseDigestJSON(`{"digest":"Hello"}`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hello" {
		t.Fatalf("got %q", out)
	}
}

func TestParseDigestJSON_fenced(t *testing.T) {
	t.Parallel()
	raw := "```json\n{\"digest\":\"X\"}\n```"
	out, err := parseDigestJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out != "X" {
		t.Fatalf("got %q", out)
	}
}

func TestParseDigestJSON_errors(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{``, `{`, `{"digest":""}`, `{"foo":"bar"}`} {
		_, err := parseDigestJSON(raw)
		if err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestParseDigestJSON_truncatesHugeDigest(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", digestJSONMaxRunes+50)
	raw := `{"digest":"` + long + `"}`
	out, err := parseDigestJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len([]rune(out)) != digestJSONMaxRunes {
		t.Fatalf("len runes %d", len([]rune(out)))
	}
	if !strings.HasSuffix(out, "…") {
		t.Fatalf("suffix")
	}
}
