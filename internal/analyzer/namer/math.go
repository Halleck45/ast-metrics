package namer

import "math"

func l2norm(a []float32) float32 {
	var sum float32
	for _, v := range a {
		sum += v * v
	}
	if sum == 0 {
		return 0
	}
	return float32(math.Sqrt(float64(sum)))
}

// cosineWithPreNorm computes cosine(a, b) given ||a|| and b.
func cosineWithPreNorm(a []float32, normA float32, b []float32) float32 {
	var dot, normB2 float32
	for i := range a {
		dot += a[i] * b[i]
		normB2 += b[i] * b[i]
	}
	if normA == 0 || normB2 == 0 {
		return 0
	}
	return dot / (normA * float32(math.Sqrt(float64(normB2))))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
