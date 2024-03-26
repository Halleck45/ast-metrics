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
	maxCyclomatic := 0.0
	maxCommits := 0.0

	// get bounds
	for _, file := range project.Combined.ConcernedFiles {
		classes := Engine.GetClassesInFile(file)

		if file.Commits == nil {
			continue
		}

		commits := file.Commits.Commits

		// OOP file
		for _, class := range classes {
			maintainability := float64(128 - *class.Stmts.Analyze.Maintainability.MaintainabilityIndex)
			if maintainability > maxComplexity {
				maxComplexity = maintainability
			}
		}

		// all files (procedural and OOP)
		cyclomatic := float64(*file.Stmts.Analyze.Complexity.Cyclomatic)
		if cyclomatic > maxCyclomatic {
			maxCyclomatic = cyclomatic
		}

		if float64(len(commits)) > maxCommits {
			maxCommits = float64(len(commits))
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

			risk := v.GetRisk(maxCommits, maxComplexity, nbCommits, int(128-*class.Stmts.Analyze.Maintainability.MaintainabilityIndex))
			file.Stmts.Analyze.Risk.Score += float32(risk)
		}

		// Procedural file. We put risk on the file itself, according to the cyclomatic complexity.
		if file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil {
			continue
		}

		cyclo := float64(*file.Stmts.Analyze.Complexity.Cyclomatic)
		risk := v.GetRisk(maxCommits, maxCyclomatic, nbCommits, int(cyclo))
		file.Stmts.Analyze.Risk.Score += float32(risk)
	}
}

func (v *RiskAnalyzer) GetRisk(maxCommits float64, maxComplexity float64, nbCommits int, complexity int) float32 {

	// Calculate the horizontal and vertical distance from the "top right" corner.
	horizontalDistance := maxCommits - float64(nbCommits)
	verticalDistance := maxComplexity - float64(complexity)

	// Normalize these values over time, we first divide by the maximum values, to always end up with distances between 0 and 1.
	normalizedHorizontalDistance := horizontalDistance / maxCommits
	normalizedVerticalDistance := verticalDistance / maxComplexity

	// Calculate the distance of this class from the "top right" corner, using the simple formula A^2 + B^2 = C^2; or: C = sqrt(A^2 + B^2)).
	distanceFromTopRightCorner := math.Sqrt(math.Pow(normalizedHorizontalDistance, 2) + math.Pow(normalizedVerticalDistance, 2))

	// The resulting value will be between 0 and sqrt(2). A short distance is bad, so in order to end up with a high score, we invert the value by subtracting it from 1.
	risk := 1 - distanceFromTopRightCorner

	return float32(risk)
}
