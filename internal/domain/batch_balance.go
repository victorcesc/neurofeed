package domain

import "strings"

// BalanceArticlesAcrossSubjects builds a batch of at most maxTotal articles by round-robin across
// subjectOrder (e.g. configured digest sections). Preserves original order within each subject's queue.
// Subjects in order with no articles are skipped. Articles whose BucketSubject() is not in order are
// appended after round-robin in global appearance order until maxTotal is reached.
func BalanceArticlesAcrossSubjects(subjectOrder []string, articles []Article, maxTotal int) []Article {
	if maxTotal <= 0 || len(articles) == 0 {
		return nil
	}
	if len(subjectOrder) == 0 {
		if len(articles) <= maxTotal {
			return append([]Article(nil), articles...)
		}
		return append([]Article(nil), articles[:maxTotal]...)
	}

	queues := make(map[string][]Article, len(subjectOrder))
	for index := range articles {
		article := articles[index]
		key := article.BucketSubject()
		queues[key] = append(queues[key], article)
	}

	inOrder := map[string]struct{}{}
	for index := range subjectOrder {
		inOrder[subjectOrder[index]] = struct{}{}
	}

	heads := make(map[string]int, len(queues))
	for key := range queues {
		heads[key] = 0
	}

	var result []Article
	for len(result) < maxTotal {
		added := false
		for orderIndex := range subjectOrder {
			if len(result) >= maxTotal {
				break
			}
			subject := subjectOrder[orderIndex]
			queue := queues[subject]
			head := heads[subject]
			if head >= len(queue) {
				continue
			}
			result = append(result, queue[head])
			heads[subject] = head + 1
			added = true
		}
		if !added {
			break
		}
	}

	if len(result) >= maxTotal {
		return result
	}

	for index := range articles {
		if len(result) >= maxTotal {
			break
		}
		article := articles[index]
		if articleInSlice(result, article) {
			continue
		}
		if _, ok := inOrder[article.BucketSubject()]; ok {
			continue
		}
		result = append(result, article)
	}

	return result
}

func articleInSlice(list []Article, candidate Article) bool {
	link := trimLink(candidate.Link)
	for index := range list {
		if link != "" && trimLink(list[index].Link) == link {
			return true
		}
		if link == "" && list[index].Title == candidate.Title {
			return true
		}
	}
	return false
}

func trimLink(s string) string {
	return strings.TrimSpace(s)
}
