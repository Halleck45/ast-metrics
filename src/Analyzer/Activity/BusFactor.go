package Analyzer

import (
	"sort"

	"github.com/halleck45/ast-metrics/src/Analyzer"
)

type BusFactor struct {
}

func NewBusFactor() *BusFactor {
	return &BusFactor{}
}

func (busFactor *BusFactor) Calculate(aggregate *Analyzer.Aggregated) {
	// Cf. https://chaoss.community/kb/metric-bus-factor/

	// count commits of all committers, by commiters
	files := aggregate.ConcernedFiles
	commits := make(map[string]int)
	for _, file := range files {

		if file.Commits == nil {
			continue
		}
		
		for _, commit := range file.Commits.Commits {

			// Exclude commits with no author or from noreply@github.com
			if commit.Author == "" || commit.Author == "noreply@github.com" {
				continue
			}

			if _, ok := commits[commit.Author]; !ok {
				commits[commit.Author] = 0
			}

			commits[commit.Author]++
		}
	}

	// sort committers by commits count
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range commits {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	// calculate bus factor
	resultBusFactor := 0
	// 50% sum of commits value
	midPercentBoundary := 0
	for _, kv := range ss {
		midPercentBoundary += kv.Value
	}
	midPercentBoundary = midPercentBoundary / 2

	sum := 0
	for _, kv := range ss {
		sum += kv.Value
		resultBusFactor++
		if sum >= midPercentBoundary {
			break
		}
	}

	// keep only top 3 committers
	aggregate.TopCommitters = make([]Analyzer.TopCommitter, 0)
	for i, kv := range ss {
		if i > 3 {
			break
		}
		commiter := Analyzer.TopCommitter{
			Name:  kv.Key,
			Count: kv.Value,
		}
		aggregate.TopCommitters = append(aggregate.TopCommitters, commiter)
	}

	// sort TopCommitters by commits count
	sort.Slice(aggregate.TopCommitters, func(i, j int) bool {
		return aggregate.TopCommitters[i].Count > aggregate.TopCommitters[j].Count
	})

	aggregate.BusFactor = resultBusFactor
}
