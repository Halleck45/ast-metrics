package namer

import (
	"sort"
)

type wordScore struct {
	word  string
	score float32
}

func (n *Namer) computeCentroid(tokens []string) ([]float32, float32) {
	centroid := make([]float32, n.dim)
	var count float32

	for _, t := range tokens {
		vec, ok := n.vectors[t]
		if !ok {
			continue
		}
		for i := 0; i < n.dim; i++ {
			centroid[i] += vec[i]
		}
		count++
	}

	if count == 0 {
		return centroid, 0
	}

	inv := 1 / count
	for i := 0; i < n.dim; i++ {
		centroid[i] *= inv
	}

	return centroid, count
}

func (n *Namer) computeScores(tokens []string, centroid []float32) []wordScore {
	seen := make(map[string]struct{}, len(tokens))
	scores := make([]wordScore, 0, len(tokens))

	centNorm := l2norm(centroid)
	if centNorm == 0 {
		return nil
	}

	for _, t := range tokens {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}

		vec, ok := n.vectors[t]
		if !ok {
			continue
		}

		score := cosineWithPreNorm(centroid, centNorm, vec)
		scores = append(scores, wordScore{word: t, score: score})
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].word < scores[j].word
		}
		return scores[i].score > scores[j].score
	})

	return scores
}
