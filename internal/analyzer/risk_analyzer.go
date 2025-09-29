package analyzer

import (
	"math"

	"github.com/halleck45/ast-metrics/internal/analyzer/risk"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

type RiskAnalyzer struct {
	detectors []risk.Detector
}

func NewRiskAnalyzer() *RiskAnalyzer {
	return &RiskAnalyzer{detectors: []risk.Detector{
		&risk.TooManyGoClassesDetector{},
		&risk.TooBuggedDetector{},
		&risk.TooManyResponsibilityDetector{},
		&risk.TooManyEfferentCouplingDetector{},
	}}
}

// Detects risks for a single file using simple detectors.
// The returned items are not persisted in protobuf and are intended for reporting views.
func (v *RiskAnalyzer) DetectFileRisks(file *pb.File) []risk.RiskItem {
	items := make([]risk.RiskItem, 0)
	for _, d := range v.detectors {
		items = append(items, d.Detect(file)...)
	}
	return items
}

func (v *RiskAnalyzer) Analyze(project ProjectAggregated) {

	var maxComplexity float64 = 0
	var maxCyclomatic int32 = 0
	var maxCommits int = 0

	// get bounds
	for _, file := range project.Combined.ConcernedFiles {
		classes := engine.GetClassesInFile(file)

		if file.Commits == nil {
			continue
		}

		commits := file.Commits.Commits

		// OOP file
		for _, class := range classes {
			// Guard against nil pointers in class analysis
			if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Maintainability == nil || class.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
				continue
			}
			maintainability := 128 - *class.Stmts.Analyze.Maintainability.MaintainabilityIndex
			if maintainability > maxComplexity {
				maxComplexity = maintainability
			}
		}

		// all files (procedural and OOP)
		if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Complexity != nil && file.Stmts.Analyze.Complexity.Cyclomatic != nil {
			cyclomatic := *file.Stmts.Analyze.Complexity.Cyclomatic
			if cyclomatic > maxCyclomatic {
				maxCyclomatic = cyclomatic
			}
		}

		if len(commits) > maxCommits {
			maxCommits = len(commits)
		}
	}

	// From https://github.com/bmitch/churn-php/blob/master/src/Result/Result.php
	for _, file := range project.Combined.ConcernedFiles {

		// Ensure analysis structures exist
		if file.Stmts == nil {
			file.Stmts = &pb.Stmts{}
		}
		if file.Stmts.Analyze == nil {
			file.Stmts.Analyze = &pb.Analyze{}
		}
		if file.Stmts.Analyze.Risk == nil {
			file.Stmts.Analyze.Risk = &pb.Risk{Score: float64(0)}
		}

		nbCommits := 0
		if file.Commits != nil {
			nbCommits = len(file.Commits.Commits)
		}

		// OOP objects. We put risk on classes, according to the maintainability index.
		for _, class := range engine.GetClassesInFile(file) {

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
	build := func(agg Aggregated) []Suggestion {
		sugs := make([]Suggestion, 0)
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
					sugs = append(sugs, Suggestion{
						Summary:  "Refactor hotspot",
						Location: f.Path,
						Why:      "High combined risk and frequent changes (>= 3 commits)",
						DetailedExplanation: "This file exhibits both complexity and change frequency. Focus refactoring on complex or frequently modified parts, add tests, and reduce cyclomatic complexity.",
					})
					seen[msg] = true
				}
			}
			// Suggest MI improvements per class
			for _, cls := range engine.GetClassesInFile(f) {
				if cls.Stmts == nil || cls.Stmts.Analyze == nil || cls.Stmts.Analyze.Maintainability == nil || cls.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
					continue
				}
				if *cls.Stmts.Analyze.Maintainability.MaintainabilityIndex < 65 {
					name := cls.Name.GetQualified()
					msg := "Improve maintainability of " + name
					if !seen[msg] {
						sugs = append(sugs, Suggestion{
							Summary:  "Improve maintainability",
							Location: name,
							Why:      "Maintainability Index below 65",
							DetailedExplanation: "Refactor long methods, reduce nesting, and enhance cohesion. Consider splitting responsibilities and improving naming and tests.",
						})
						seen[msg] = true
					}
				}
			}
			// Suggest splitting when cyclomatic is high
			if f.Stmts.Analyze.Complexity != nil && f.Stmts.Analyze.Complexity.Cyclomatic != nil {
				if *f.Stmts.Analyze.Complexity.Cyclomatic > 50 && commits >= 3 {
					msg := "Split complex functions in " + f.Path
					if !seen[msg] {
						sugs = append(sugs, Suggestion{
							Summary:  "Split complex functions",
							Location: f.Path,
							Why:      "Cyclomatic complexity > 50 with frequent changes (>= 3 commits)",
							DetailedExplanation: "Identify the most complex functions and extract smaller, well-named functions. Aim to reduce decision points and increase readability.",
						})
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

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (v *RiskAnalyzer) GetRisk(maxCommits int32, maxComplexity float64, nbCommits int, complexity int) float64 {
	// Guard against invalid bounds
	mc := float64(maxCommits)
	mx := maxComplexity

	// Calculate distances from the top-right corner only on available axes
	var normalizedHorizontalDistance float64
	if mc > 0 {
		h := mc - float64(nbCommits)
		normalizedHorizontalDistance = h / mc
	}
	var normalizedVerticalDistance float64
	if mx > 0 {
		v := mx - float64(complexity)
		normalizedVerticalDistance = v / mx
	}

	// Clamp to [0,1]
	normalizedHorizontalDistance = clamp01(normalizedHorizontalDistance)
	normalizedVerticalDistance = clamp01(normalizedVerticalDistance)

	// Euclidean distance in normalized space
	distanceFromTopRightCorner := math.Sqrt(normalizedHorizontalDistance*normalizedHorizontalDistance + normalizedVerticalDistance*normalizedVerticalDistance)

	// Invert and clamp to [0,1]
	risk := 1 - distanceFromTopRightCorner
	if math.IsNaN(risk) || math.IsInf(risk, 0) {
		return 0
	}
	return clamp01(risk)
}
