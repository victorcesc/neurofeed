package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/victorcesc/neurofeed/internal/domain"
)

const digestDescriptionRunes = 280

// digestLLMTemperature is kept low to reduce index/link mix-ups across batched articles.
const digestLLMTemperature = 0.2

// DigestSummarizer calls the OpenAI chat API with JSON output and returns Telegram HTML (Phase 4 layout).
type DigestSummarizer struct {
	client          *OpenAIChatClient
	maxArticles     int
	maxOutputTokens int
}

// NewDigestSummarizer wires an OpenAI client with caps from config (already validated positive).
func NewDigestSummarizer(client *OpenAIChatClient, maxArticles int, maxOutputTokens int) *DigestSummarizer {
	return &DigestSummarizer{
		client:          client,
		maxArticles:     maxArticles,
		maxOutputTokens: maxOutputTokens,
	}
}

// Summarize implements Summarizer.
func (digest *DigestSummarizer) Summarize(ctx context.Context, articles []domain.Article) (string, error) {
	return digest.SummarizeWithSubjectOrder(ctx, articles, nil)
}

// SummarizeWithSubjectOrder implements SubjectOrderedSummarizer. When subjectOrder is empty,
// section order falls back to domain.SubjectBucketOrder(articles).
func (digest *DigestSummarizer) SummarizeWithSubjectOrder(ctx context.Context, articles []domain.Article, subjectOrder []string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(articles) == 0 {
		return "", nil
	}

	temperature := digestLLMTemperature
	order := effectiveSectionOrder(subjectOrder, articles)
	var batch []domain.Article
	if domain.HasAnyConfiguredSubject(articles) {
		batch = domain.BalanceArticlesAcrossSubjects(order, articles, digest.maxArticles)
	} else {
		limit := len(articles)
		if limit > digest.maxArticles {
			limit = digest.maxArticles
		}
		batch = articles[:limit]
	}

	if domain.HasAnyConfiguredSubject(articles) {
		userContent := buildDigestUserContentSectioned(batch, order)
		raw, err := digest.client.ChatCompletionWithOptions(ctx, ChatCompletionInput{
			Messages: []ChatMessage{
				{Role: "system", Content: digestSystemPromptSections},
				{Role: "user", Content: userContent},
			},
			MaxTokens:        digest.maxOutputTokens,
			JSONResponseMode: true,
			Temperature:      &temperature,
		})
		if err != nil {
			return "", err
		}
		rows, err := parseDigestSectionsPicksJSON(raw)
		if err != nil {
			return "", err
		}
		perSection := articlesBySubjectOrder(batch, order)
		return assembleSectionedPhase4HTML(order, perSection, rows)
	}

	userContent := buildDigestUserContentFlat(batch)
	raw, err := digest.client.ChatCompletionWithOptions(ctx, ChatCompletionInput{
		Messages: []ChatMessage{
			{Role: "system", Content: digestSystemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens:        digest.maxOutputTokens,
		JSONResponseMode: true,
		Temperature:      &temperature,
	})
	if err != nil {
		return "", err
	}
	picks, err := parseDigestFlatPicksJSON(raw)
	if err != nil {
		return "", err
	}
	return assembleFlatPhase4HTML(batch, picks)
}

func effectiveSectionOrder(subjectOrder []string, batch []domain.Article) []string {
	if len(subjectOrder) > 0 {
		return subjectOrder
	}
	return domain.SubjectBucketOrder(batch)
}

func buildDigestUserContentFlat(articles []domain.Article) string {
	var builder strings.Builder
	builder.WriteString("Articles (numbered). Each line starts with the index, then title, link, source tier, optional snippet:\n\n")
	for index := range articles {
		article := articles[index]
		builder.WriteString(fmt.Sprintf("%d. Title: %s\n", index+1, strings.TrimSpace(article.Title)))
		builder.WriteString(fmt.Sprintf("   Link: %s\n", strings.TrimSpace(article.Link)))
		builder.WriteString(fmt.Sprintf("   Source: %s | Tier: %s\n", strings.TrimSpace(article.Source), article.SourceTier.String()))
		desc := strings.TrimSpace(article.Description)
		if desc != "" {
			runes := []rune(desc)
			if len(runes) > digestDescriptionRunes {
				desc = string(runes[:digestDescriptionRunes]) + "…"
			}
			builder.WriteString(fmt.Sprintf("   Snippet: %s\n", desc))
		}
		builder.WriteString("\n")
	}
	builder.WriteString(`Respond with JSON only: {"picks":[{"index":1,"link":"<exact URL from that item's Link line>","line1":"…","line2":"…"},…]} — one entry per chosen article; at most two picks; each "link" must copy the chosen article's Link exactly (same row as index).`)
	return builder.String()
}

func buildDigestUserContentSectioned(articles []domain.Article, order []string) string {
	var builder strings.Builder
	builder.WriteString("Output JSON with a \"sections\" array as specified in the system message.\n\n")
	builder.WriteString("Required section order and exact subject names:\n")
	for index, name := range order {
		if index > 0 {
			builder.WriteString(" | ")
		}
		builder.WriteString(name)
	}
	builder.WriteString("\n\nArticles grouped by topic:\n\n")
	for _, subjectName := range order {
		builder.WriteString("## ")
		builder.WriteString(subjectName)
		builder.WriteString("\n")
		count := 0
		for articleIndex := range articles {
			article := articles[articleIndex]
			if article.BucketSubject() != subjectName {
				continue
			}
			count++
			builder.WriteString(fmt.Sprintf("%d. Title: %s\n", count, strings.TrimSpace(article.Title)))
			builder.WriteString(fmt.Sprintf("   Link: %s\n", strings.TrimSpace(article.Link)))
			builder.WriteString(fmt.Sprintf("   Source: %s | Tier: %s\n", strings.TrimSpace(article.Source), article.SourceTier.String()))
			desc := strings.TrimSpace(article.Description)
			if desc != "" {
				runes := []rune(desc)
				if len(runes) > digestDescriptionRunes {
					desc = string(runes[:digestDescriptionRunes]) + "…"
				}
				builder.WriteString(fmt.Sprintf("   Snippet: %s\n", desc))
			}
			builder.WriteString("\n")
		}
		if count == 0 {
			builder.WriteString("(no articles in this run)\n\n")
		}
	}
	builder.WriteString(`Respond with JSON only: {"sections":[{"subject":"…","picks":[{"index":1,"link":"<exact URL from that item's Link line>","line1":"…","line2":"…"},…]},…]}`)
	return builder.String()
}
