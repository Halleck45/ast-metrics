package Analyzer

import (
	"math"

	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type RiskAnalyzer struct {
}

func NewRiskAnalyzer() *RiskAnalyzer {
	return &RiskAnalyzer{}
}

func (v *RiskAnalyzer) Analyze(project ProjectAggregated) {

	var maxComplexity float64 = 0
	var maxCyclomatic int32 = 0
	var maxCommits int = 0

	// get bounds
	for _, file := range project.Combined.ConcernedFiles {
		classes := Engine.GetClassesInFile(file)

		if file.Commits == nil {
			continue
		}

		commits := file.Commits.Commits

		// OOP file
		for _, class := range classes {
			maintainability := 128 - *class.Stmts.Analyze.Maintainability.MaintainabilityIndex
			if maintainability > maxComplexity {
				maxComplexity = maintainability
			}
		}

		// all files (procedural and OOP)
		cyclomatic := *file.Stmts.Analyze.Complexity.Cyclomatic
		if cyclomatic > maxCyclomatic {
			maxCyclomatic = cyclomatic
		}

		if len(commits) > maxCommits {
			maxCommits = len(commits)
		}
	}

	// From https://github.com/bmitch/churn-php/blob/master/src/Result/Result.php
	for _, file := range project.Combined.ConcernedFiles {

		if file.Stmts.Analyze.Risk == nil {
			file.Stmts.Analyze.Risk = &pb.Risk{Score: float64(0)}
		}

		nbCommits := 0
		if file.Commits != nil {
			nbCommits = len(file.Commits.Commits)
		}

		// OOP objects. We put risk on classes, according to the maintainability index.
		for _, class := range Engine.GetClassesInFile(file) {

			if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Maintainability == nil {
				continue
			}

			risk := v.GetRisk(int32(maxCommits), maxComplexity, nbCommits, int(128-*class.Stmts.Analyze.Maintainability.MaintainabilityIndex))
			file.Stmts.Analyze.Risk.Score += float64(risk)
		}
		
		// Procedural file. We put risk on the file itself, according to the cyclomatic complexity.
		if file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil {
			continue
		}

		cyclo := *file.Stmts.Analyze.Complexity.Cyclomatic
		risk := v.GetRisk(int32(maxCommits), float64(maxCyclomatic), nbCommits, int(cyclo))
		file.Stmts.Analyze.Risk.Score += float64(risk)
	}

	// Build suggestions for Combined and per language
	build := func(agg Aggregated) []string {
		sugs := make([]string, 0)
		seen := make(map[string]bool)
		for _, f := range agg.ConcernedFiles {
			if f == nil || f.Stmts == nil || f.Stmts.Analyze == nil || f.Stmts.Analyze.Risk == nil {
				continue
			}
			risk := f.Stmts.Analyze.Risk.Score
			commits := 0
			if f.Commits != nil {
				commits = len(f.Commits.Commits)
			}
			// Suggest hotspots
			if risk >= 0.7 && commits >= 3 {
				msg := "Refactor hotspot: " + f.Path
				if !seen[msg] {
					sugs = append(sugs, msg)
					seen[msg] = true
				}
			}
			// Suggest MI improvements per class
			for _, cls := range Engine.GetClassesInFile(f) {
				if cls.Stmts == nil || cls.Stmts.Analyze == nil || cls.Stmts.Analyze.Maintainability == nil || cls.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
					continue
				}
				if *cls.Stmts.Analyze.Maintainability.MaintainabilityIndex < 65 {
					name := cls.Name.GetQualified()
					msg := "Improve maintainability of " + name
					if !seen[msg] {
						sugs = append(sugs, msg)
						seen[msg] = true
					}
				}
			}
			// Suggest splitting when cyclomatic is high
			if f.Stmts.Analyze.Complexity != nil && f.Stmts.Analyze.Complexity.Cyclomatic != nil {
				if *f.Stmts.Analyze.Complexity.Cyclomatic > 50 && commits >= 3 {
					msg := "Split complex functions in " + f.Path
					if !seen[msg] {
						sugs = append(sugs, msg)
						seen[msg] = true
					}
				}
			}
			if len(sugs) >= 10 {
				break
			}
		}
		return sugs
	}
	project.Combined.Suggestions = build(project.Combined)
	for lng, agg := range project.ByProgrammingLanguage {
		a := agg
		a.Suggestions = build(agg)
		project.ByProgrammingLanguage[lng] = a
	}
}

func (v *RiskAnalyzer) GetRisk(maxCommits int32, maxComplexity float64, nbCommits int, complexity int) float64 {

	// Calculate the horizontal and vertical distance from the "top right" corner.
	horizontalDistance := float64(maxCommits) - float64(nbCommits)
	verticalDistance := maxComplexity - float64(complexity)

	// Normalize these values over time, we first divide by the maximum values, to always end up with distances between 0 and 1.
	normalizedHorizontalDistance := horizontalDistance / float64(maxCommits)
	normalizedVerticalDistance := verticalDistance / maxComplexity

	// Calculate the distance of this class from the "top right" corner, using the simple formula A^2 + B^2 = C^2; or: C = sqrt(A^2 + B^2)).
	distanceFromTopRightCorner := math.Sqrt(math.Pow(float64(normalizedHorizontalDistance), 2) + math.Pow(float64(normalizedVerticalDistance), 2))

	// The resulting value will be between 0 and sqrt(2). A short distance is bad, so in order to end up with a high score, we invert the value by subtracting it from 1.
	risk := 1 - distanceFromTopRightCorner

	return float64(risk)
}
