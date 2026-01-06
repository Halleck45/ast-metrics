package namer

import (
	"sort"
	"strings"
)

// Names returns a human-friendly label for the given class names.
func (n *Namer) Names(classNames []string) string {
	size := len(classNames)
	if size == 0 {
		return "Miscellaneous"
	}

	// Small cluster → namespace heuristic.
	if size < 4 {
		if fallback := fallbackFromNamespaces(classNames); fallback != "" {
			return fallback
		}
	}

	tokens := n.tokenize(classNames)
	if len(tokens) == 0 {
		return "Miscellaneous"
	}

	centroid, count := n.computeCentroid(tokens)
	if count == 0 {
		if fallback := fallbackFromNamespaces(classNames); fallback != "" {
			return fallback
		}
		return "Miscellaneous"
	}

	scores := n.computeScores(tokens, centroid)
	if len(scores) == 0 {
		return "Miscellaneous"
	}

	wordCount := decideWordCount(scores)

	parts := make([]string, 0, wordCount)
	for i := 0; i < wordCount && i < len(scores); i++ {
		parts = append(parts, title(scores[i].word))
	}

	// Make output deterministic.
	sort.Strings(parts)
	return strings.Join(parts, " ")
}
