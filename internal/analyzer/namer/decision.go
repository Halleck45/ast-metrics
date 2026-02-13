package namer

func decideWordCount(scores []wordScore) int {
	if len(scores) == 1 {
		return 1
	}

	top := min(5, len(scores))
	mean := float32(0)
	for i := 0; i < top; i++ {
		mean += scores[i].score
	}
	mean /= float32(top)

	varTop := float32(0)
	for i := 0; i < top; i++ {
		d := scores[i].score - mean
		varTop += d * d
	}
	varTop /= float32(top)

	gap := scores[0].score - scores[1].score

	switch {
	case gap > 0.25 && varTop < 0.02:
		return 1
	case varTop > 0.06 && gap < 0.10:
		return 3
	default:
		return 2
	}
}
