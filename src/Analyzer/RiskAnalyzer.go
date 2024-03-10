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

	maxComplexity := 0.0
	maxCommits := 0.0

	// get bounds
	for _, file := range project.Combined.ConcernedFiles {
		classes := Engine.GetClassesInFile(file)
		commits := file.Commits.Committers

		for _, class := range classes {
			maintainability := float64(128 - *class.Stmts.Analyze.Maintainability.MaintainabilityIndex)
			if maintainability > maxComplexity {
				maxComplexity = maintainability
			}
		}

		if float64(len(commits)) > maxCommits {
			maxCommits = float64(len(commits))
		}
	}

	// From https://github.com/bmitch/churn-php/blob/master/src/Result/Result.php
	for _, file := range project.Combined.ConcernedFiles {
		for _, class := range Engine.GetClassesInFile(file) {

			// Calculate the horizontal and vertical distance from the "top right" corner.
			horizontalDistance := maxCommits - float64(len(file.Commits.Committers))
			verticalDistance := maxComplexity - float64(128-*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)

			// Normalize these values over time, we first divide by the maximum values, to always end up with distances between 0 and 1.
			normalizedHorizontalDistance := horizontalDistance / maxCommits
			normalizedVerticalDistance := verticalDistance / maxComplexity

			// Calculate the distance of this class from the "top right" corner, using the simple formula A^2 + B^2 = C^2; or: C = sqrt(A^2 + B^2)).
			distanceFromTopRightCorner := math.Sqrt(math.Pow(normalizedHorizontalDistance, 2) + math.Pow(normalizedVerticalDistance, 2))

			// The resulting value will be between 0 and sqrt(2). A short distance is bad, so in order to end up with a high score, we invert the value by subtracting it from 1.
			risk := 1 - distanceFromTopRightCorner
			class.Stmts.Analyze.Risk = &pb.Risk{Score: float32(risk)}

			if file.Stmts.Analyze.Risk == nil {
				file.Stmts.Analyze.Risk = &pb.Risk{Score: float32(0)}
			}
			file.Stmts.Analyze.Risk.Score += float32(risk)
		}
	}
}
