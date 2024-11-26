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

	var maxComplexity float32 = 0
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
			file.Stmts.Analyze.Risk = &pb.Risk{Score: float32(0)}
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
			file.Stmts.Analyze.Risk.Score += float32(risk)
		}

		// Procedural file. We put risk on the file itself, according to the cyclomatic complexity.
		if file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil {
			continue
		}

		cyclo := *file.Stmts.Analyze.Complexity.Cyclomatic
		risk := v.GetRisk(int32(maxCommits), float32(maxCyclomatic), nbCommits, int(cyclo))
		file.Stmts.Analyze.Risk.Score += float32(risk)
	}
}

func (v *RiskAnalyzer) GetRisk(maxCommits int32, maxComplexity float32, nbCommits int, complexity int) float32 {

	// Calculate the horizontal and vertical distance from the "top right" corner.
	horizontalDistance := float32(maxCommits) - float32(nbCommits)
	verticalDistance := maxComplexity - float32(complexity)

	// Normalize these values over time, we first divide by the maximum values, to always end up with distances between 0 and 1.
	normalizedHorizontalDistance := horizontalDistance / float32(maxCommits)
	normalizedVerticalDistance := verticalDistance / maxComplexity

	// Calculate the distance of this class from the "top right" corner, using the simple formula A^2 + B^2 = C^2; or: C = sqrt(A^2 + B^2)).
	distanceFromTopRightCorner := math.Sqrt(math.Pow(float64(normalizedHorizontalDistance), 2) + math.Pow(float64(normalizedVerticalDistance), 2))

	// The resulting value will be between 0 and sqrt(2). A short distance is bad, so in order to end up with a high score, we invert the value by subtracting it from 1.
	risk := 1 - distanceFromTopRightCorner

	return float32(risk)
}
