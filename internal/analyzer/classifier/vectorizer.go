package classifier

import (
	"fmt"
	"math"
)

func (v *TfidfVectorizer) TransformHashingTfidf(text string, out []float64, weight float64) error {
	if v == nil {
		return fmt.Errorf("vectorizer is nil")
	}
	if v.Type != "hashing_tfidf" {
		return fmt.Errorf("vectorizer type is %q, expected hashing_tfidf", v.Type)
	}
	if v.NFeatures <= 0 {
		return fmt.Errorf("vectorizer n_features is invalid: %d", v.NFeatures)
	}
	if len(out) != v.NFeatures {
		return fmt.Errorf("output slice size mismatch: got=%d expected=%d", len(out), v.NFeatures)
	}

	// Reset out (caller may reuse a backing array)
	for i := range out {
		out[i] = 0
	}

	stopWords := ""
	switch sw := v.StopWords.(type) {
	case string:
		stopWords = sw
	default:
		stopWords = ""
	}

	unigrams := tokenizeBasic(text, v.Lowercase, stopWords)
	if len(unigrams) == 0 {
		return nil
	}

	// Generate n-grams if configured (e.g. [1,3] for unigrams+bigrams+trigrams)
	minN, maxN := 1, 1
	if len(v.NgramRange) >= 2 {
		minN, maxN = v.NgramRange[0], v.NgramRange[1]
	}
	tokens := generateNgrams(unigrams, minN, maxN)

	// term frequency in hashed bins
	// sklearn uses abs(signed_hash) % n_features, not unsigned_hash % n_features
	for _, tok := range tokens {
		h := murmur3_32(tok)
		signed := int32(h)
		absH := int64(signed)
		if absH < 0 {
			absH = -absH
		}
		idx := int(absH % int64(v.NFeatures))

		sign := 1.0
		if v.AlternateSign {
			if signed < 0 {
				sign = -1.0
			}
		}

		out[idx] += sign
	}

	// apply sublinear tf + idf + weight
	for i := range out {
		val := out[i]
		if val == 0 {
			continue
		}

		sign := 1.0
		if val < 0 {
			sign = -1.0
			val = -val
		}

		if v.SublinearTf {
			val = 1.0 + math.Log(val)
		}
		val *= sign

		if v.UseIdf && i < len(v.Idf) {
			val *= v.Idf[i]
		}

		val *= weight
		out[i] = val
	}

	// L2 norm
	if v.Norm == "l2" {
		var sum float64
		for _, x := range out {
			sum += x * x
		}
		n := math.Sqrt(sum)
		if n > 0 {
			inv := 1.0 / n
			for i := range out {
				out[i] *= inv
			}
		}
	}

	return nil
}
