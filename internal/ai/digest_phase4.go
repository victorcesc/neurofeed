package ai

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"unicode/utf8"

	"github.com/victorcesc/neurofeed/internal/domain"
	"github.com/victorcesc/neurofeed/internal/notify"
)

const digestPhase4MaxRunes = 12000

const pickLineMaxRunes = 220

type digestPickJSON struct {
	Index int    `json:"index"`
	Link  string `json:"link,omitempty"` // exact article URL from the prompt; when set, resolution prefers this over index
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
}

type digestFlatPicksPayload struct {
	Picks []digestPickJSON `json:"picks"`
}

type digestSectionPicksRow struct {
	Subject string           `json:"subject"`
	Picks   []digestPickJSON `json:"picks"`
}

type digestSectionsPicksPayload struct {
	Sections []digestSectionPicksRow `json:"sections"`
}

// articlesBySubjectOrder returns, for each name in order, articles in batch with that bucket in batch order.
func articlesBySubjectOrder(batch []domain.Article, order []string) [][]domain.Article {
	out := make([][]domain.Article, len(order))
	for index := range order {
		subject := order[index]
		for articleIndex := range batch {
			article := batch[articleIndex]
			if article.BucketSubject() != subject {
				continue
			}
			out[index] = append(out[index], article)
		}
	}
	return out
}

func parseDigestFlatPicksJSON(raw string) ([]digestPickJSON, error) {
	trimmed := stripMarkdownJSONFence(strings.TrimSpace(raw))
	var payload digestFlatPicksPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, fmt.Errorf("ai: parse flat picks json: %w", err)
	}
	if len(payload.Picks) == 0 {
		return nil, fmt.Errorf("ai: flat picks: empty picks")
	}
	return payload.Picks, nil
}

func parseDigestSectionsPicksJSON(raw string) ([]digestSectionPicksRow, error) {
	trimmed := stripMarkdownJSONFence(strings.TrimSpace(raw))
	var payload digestSectionsPicksPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, fmt.Errorf("ai: parse section picks json: %w", err)
	}
	if len(payload.Sections) == 0 {
		return nil, fmt.Errorf("ai: section picks: empty sections")
	}
	return payload.Sections, nil
}

func sectionPicksLookup(rows []digestSectionPicksRow) map[string][]digestPickJSON {
	lookup := map[string][]digestPickJSON{}
	for index := range rows {
		row := rows[index]
		key := strings.ToLower(strings.TrimSpace(row.Subject))
		if key == "" {
			continue
		}
		lookup[key] = row.Picks
	}
	return lookup
}

type resolvedPick struct {
	Article domain.Article
	Line1   string
	Line2   string
}

func trimPickLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) > pickLineMaxRunes {
		return string(runes[:pickLineMaxRunes-1]) + "…"
	}
	return s
}

func defaultPickLines(article domain.Article) (string, string) {
	title := strings.TrimSpace(article.Title)
	if title == "" {
		title = "Artigo"
	}
	runes := []rune(title)
	if len(runes) > 140 {
		title = string(runes[:139]) + "…"
	}
	return title, "Impacto: ver o artigo."
}

func articleIndexByLink(sectionArts []domain.Article, link string) (int, bool) {
	link = strings.TrimSpace(link)
	if link == "" {
		return 0, false
	}
	for i := range sectionArts {
		if strings.TrimSpace(sectionArts[i].Link) == link {
			return i, true
		}
	}
	return 0, false
}

func normalizePicksForArticles(picks []digestPickJSON, sectionArts []domain.Article, want int) []resolvedPick {
	if len(sectionArts) == 0 {
		return nil
	}
	if want > len(sectionArts) {
		want = len(sectionArts)
	}
	if want > 2 {
		want = 2
	}
	if want < 1 {
		return nil
	}
	usedArticle := map[int]struct{}{}
	var resolved []resolvedPick
	for pickIndex := range picks {
		if len(resolved) >= want {
			break
		}
		p := picks[pickIndex]
		var articleIndex int
		var article domain.Article
		var haveArticle bool
		if idx, ok := articleIndexByLink(sectionArts, p.Link); ok {
			articleIndex = idx
			article = sectionArts[articleIndex]
			haveArticle = true
		} else if p.Index >= 1 && p.Index <= len(sectionArts) {
			articleIndex = p.Index - 1
			article = sectionArts[articleIndex]
			haveArticle = strings.TrimSpace(article.Link) != ""
		}
		if !haveArticle {
			continue
		}
		if _, seen := usedArticle[articleIndex]; seen {
			continue
		}
		line1 := trimPickLine(p.Line1)
		line2 := trimPickLine(p.Line2)
		if line1 == "" || line2 == "" {
			d1, d2 := defaultPickLines(article)
			if line1 == "" {
				line1 = d1
			}
			if line2 == "" {
				line2 = d2
			}
		}
		usedArticle[articleIndex] = struct{}{}
		resolved = append(resolved, resolvedPick{Article: article, Line1: line1, Line2: line2})
	}
	for articleIndex := range sectionArts {
		if len(resolved) >= want {
			break
		}
		if _, seen := usedArticle[articleIndex]; seen {
			continue
		}
		article := sectionArts[articleIndex]
		if strings.TrimSpace(article.Link) == "" {
			continue
		}
		line1, line2 := defaultPickLines(article)
		usedArticle[articleIndex] = struct{}{}
		resolved = append(resolved, resolvedPick{Article: article, Line1: line1, Line2: line2})
	}
	return resolved
}

func linkAnchorText(article domain.Article) string {
	title := strings.TrimSpace(article.Title)
	if title == "" {
		return "Artigo"
	}
	runes := []rune(title)
	if len(runes) > 72 {
		return string(runes[:71]) + "…"
	}
	return title
}

func formatPhase4DigestHTML(sections []struct {
	Subject string
	Items   []resolvedPick
	Empty   bool
}) string {
	var builder strings.Builder
	builder.WriteString("<b>")
	builder.WriteString(notify.EscapeTelegramHTML("🧠 Resumo do dia"))
	builder.WriteString("</b>")
	for index := range sections {
		block := sections[index]
		builder.WriteString("\n\n")
		builder.WriteString("<b>📌 ")
		builder.WriteString(notify.EscapeTelegramHTML(block.Subject))
		builder.WriteString("</b>\n")
		if block.Empty || len(block.Items) == 0 {
			builder.WriteString(notify.EscapeTelegramHTML("(Nada de novo neste período.)"))
			continue
		}
		for itemIndex := range block.Items {
			item := block.Items[itemIndex]
			if itemIndex > 0 {
				builder.WriteString("\n\n")
			}
			link := strings.TrimSpace(item.Article.Link)
			builder.WriteString("🔗 <a href=\"")
			builder.WriteString(html.EscapeString(link))
			builder.WriteString("\">")
			builder.WriteString(notify.EscapeTelegramHTML(linkAnchorText(item.Article)))
			builder.WriteString("</a>\n")
			builder.WriteString(notify.EscapeTelegramHTML(item.Line1))
			builder.WriteString("\n")
			builder.WriteString(notify.EscapeTelegramHTML(item.Line2))
		}
	}
	out := builder.String()
	runes := []rune(out)
	if len(runes) > digestPhase4MaxRunes {
		out = string(runes[:digestPhase4MaxRunes-1]) + "…"
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" || utf8.RuneCountInString(trimmed) < 12 {
		return ""
	}
	return out
}

func assembleFlatPhase4HTML(batch []domain.Article, picks []digestPickJSON) (string, error) {
	if len(batch) == 0 {
		return "", fmt.Errorf("ai: phase4 flat: no articles")
	}
	want := 2
	if len(batch) < want {
		want = len(batch)
	}
	normalized := normalizePicksForArticles(picks, batch, want)
	if len(normalized) == 0 {
		return "", fmt.Errorf("ai: phase4 flat: no usable picks")
	}
	htmlOut := formatPhase4DigestHTML([]struct {
		Subject string
		Items   []resolvedPick
		Empty   bool
	}{
		{Subject: domain.DefaultArticleSubject, Items: normalized, Empty: false},
	})
	if htmlOut == "" {
		return "", fmt.Errorf("ai: phase4 flat: assembled html empty")
	}
	return htmlOut, nil
}

func assembleSectionedPhase4HTML(order []string, perSection [][]domain.Article, rows []digestSectionPicksRow) (string, error) {
	if len(order) == 0 {
		return "", fmt.Errorf("ai: phase4 sections: empty order")
	}
	lookup := sectionPicksLookup(rows)
	blocks := make([]struct {
		Subject string
		Items   []resolvedPick
		Empty   bool
	}, 0, len(order))
	for index := range order {
		subject := order[index]
		sectionArts := perSection[index]
		if len(sectionArts) == 0 {
			blocks = append(blocks, struct {
				Subject string
				Items   []resolvedPick
				Empty   bool
			}{Subject: subject, Empty: true})
			continue
		}
		want := 2
		if len(sectionArts) < want {
			want = len(sectionArts)
		}
		picks := lookup[strings.ToLower(strings.TrimSpace(subject))]
		items := normalizePicksForArticles(picks, sectionArts, want)
		if len(items) == 0 {
			blocks = append(blocks, struct {
				Subject string
				Items   []resolvedPick
				Empty   bool
			}{Subject: subject, Empty: true})
			continue
		}
		blocks = append(blocks, struct {
			Subject string
			Items   []resolvedPick
			Empty   bool
		}{Subject: subject, Items: items, Empty: false})
	}
	htmlOut := formatPhase4DigestHTML(blocks)
	if htmlOut == "" {
		return "", fmt.Errorf("ai: phase4 sections: assembled html empty")
	}
	return htmlOut, nil
}

// FormatStaticHeadlineHTML builds Phase-4 Telegram HTML without the LLM (up to two links per subject; two lines from titles).
func FormatStaticHeadlineHTML(batch []domain.Article, order []string, sectioned bool) string {
	if sectioned {
		perSection := articlesBySubjectOrder(batch, order)
		blocks := make([]struct {
			Subject string
			Items   []resolvedPick
			Empty   bool
		}, 0, len(order))
		for index := range order {
			subject := order[index]
			sectionArts := perSection[index]
			if len(sectionArts) == 0 {
				blocks = append(blocks, struct {
					Subject string
					Items   []resolvedPick
					Empty   bool
				}{Subject: subject, Empty: true})
				continue
			}
			want := 2
			if len(sectionArts) < want {
				want = len(sectionArts)
			}
			var items []resolvedPick
			for articleIndex := 0; articleIndex < len(sectionArts) && len(items) < want; articleIndex++ {
				article := sectionArts[articleIndex]
				if strings.TrimSpace(article.Link) == "" {
					continue
				}
				line1, line2 := defaultPickLines(article)
				items = append(items, resolvedPick{Article: article, Line1: line1, Line2: line2})
			}
			if len(items) == 0 {
				blocks = append(blocks, struct {
					Subject string
					Items   []resolvedPick
					Empty   bool
				}{Subject: subject, Empty: true})
				continue
			}
			blocks = append(blocks, struct {
				Subject string
				Items   []resolvedPick
				Empty   bool
			}{Subject: subject, Items: items, Empty: false})
		}
		return formatPhase4DigestHTML(blocks)
	}
	withLinks := filterArticlesWithLink(batch)
	if len(withLinks) == 0 {
		return ""
	}
	want := 2
	if len(withLinks) < want {
		want = len(withLinks)
	}
	var items []resolvedPick
	for index := 0; index < len(withLinks) && len(items) < want; index++ {
		article := withLinks[index]
		line1, line2 := defaultPickLines(article)
		items = append(items, resolvedPick{Article: article, Line1: line1, Line2: line2})
	}
	if len(items) == 0 {
		return ""
	}
	return formatPhase4DigestHTML([]struct {
		Subject string
		Items   []resolvedPick
		Empty   bool
	}{
		{Subject: domain.DefaultArticleSubject, Items: items, Empty: false},
	})
}

func filterArticlesWithLink(batch []domain.Article) []domain.Article {
	var out []domain.Article
	for index := range batch {
		if strings.TrimSpace(batch[index].Link) != "" {
			out = append(out, batch[index])
		}
	}
	return out
}
