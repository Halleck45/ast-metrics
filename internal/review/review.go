// Package review compares two analyzed versions of a project (base and head)
// and turns the difference into a small set of actionable findings, suitable
// for a pull request check. Unchanged debt is never reported.
package review

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	pb "github.com/halleck45/ast-metrics/pb"
)

// MethodologyVersion identifies the formulas and thresholds used to compute
// findings. It must be bumped whenever a threshold or a formula changes, so
// that users can distinguish a code evolution from a calculation evolution.
const MethodologyVersion = "1.0"

type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

type Kind string

const (
	KindRegression  Kind = "regression"
	KindImprovement Kind = "improvement"
)

// Finding is a single review conclusion, anchored to a file (and when
// possible a function) with an explanation and a suggested action.
type Finding struct {
	Kind       Kind     `json:"kind"`
	Severity   Severity `json:"severity"`
	Rule       string   `json:"rule"`
	File       string   `json:"file"`
	Line       int      `json:"line,omitempty"`
	Subject    string   `json:"subject,omitempty"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
	Before     float64  `json:"before"`
	After      float64  `json:"after"`
}

type Summary struct {
	FilesChanged int `json:"filesChanged"`
	FilesAdded   int `json:"filesAdded"`
	FilesDeleted int `json:"filesDeleted"`
	High         int `json:"high"`
	Medium       int `json:"medium"`
	Low          int `json:"low"`
	Improvements int `json:"improvements"`
}

type Result struct {
	MethodologyVersion string    `json:"methodologyVersion"`
	BaseRef            string    `json:"baseRef"`
	BaseSha            string    `json:"baseSha,omitempty"`
	HeadSha            string    `json:"headSha,omitempty"`
	Summary            Summary   `json:"summary"`
	Regressions        []Finding `json:"regressions"`
	Improvements       []Finding `json:"improvements"`
	Gate               string    `json:"gate"`
}

// Options gathers every threshold used by the review. Defaults are the
// methodology; changing them requires a MethodologyVersion bump.
type Options struct {
	// A function at or above this cyclomatic complexity is worth reporting
	ComplexityMedium int
	// A function at or above this cyclomatic complexity is a high issue
	ComplexityHigh int
	// A complexity increase of at least this much is a high issue
	ComplexityJumpHigh int
	// Minimum drop of the maintainability index to report a file
	MaintainabilityDrop float64
	// Below this maintainability index a file is considered hard to maintain
	MaintainabilityLow float64
	// Below this maintainability index a file is considered critical
	MaintainabilityCritical float64
	// Minimum increase of efferent coupling to report a file
	CouplingIncrease int
	// Minimum complexity decrease to report an improvement
	ImprovementComplexityDrop int
	// Minimum maintainability gain to report an improvement
	ImprovementMaintainabilityGain float64
}

func DefaultOptions() Options {
	return Options{
		ComplexityMedium:               10,
		ComplexityHigh:                 25,
		ComplexityJumpHigh:             10,
		MaintainabilityDrop:            5,
		MaintainabilityLow:             85,
		MaintainabilityCritical:        65,
		CouplingIncrease:               3,
		ImprovementComplexityDrop:      3,
		ImprovementMaintainabilityGain: 5,
	}
}

// Compare analyzes the difference between head (current code) and base
// (target branch) and returns deterministic findings. headRoot and baseRoot
// are used to relativize file paths so both versions can be matched.
func Compare(headFiles []*pb.File, baseFiles []*pb.File, headRoot string, baseRoot string, opts Options) Result {
	result := Result{
		MethodologyVersion: MethodologyVersion,
		Regressions:        []Finding{},
		Improvements:       []Finding{},
	}

	baseByPath := map[string]*pb.File{}
	for _, f := range baseFiles {
		baseByPath[relativize(f.Path, baseRoot)] = f
	}

	seen := map[string]bool{}
	for _, head := range headFiles {
		path := relativize(head.Path, headRoot)
		seen[path] = true
		base, exists := baseByPath[path]

		if !exists {
			result.Summary.FilesAdded++
			result.Regressions = append(result.Regressions, findingsForNewFile(head, path, opts)...)
			continue
		}

		if head.Checksum != "" && head.Checksum == base.Checksum {
			continue
		}
		result.Summary.FilesChanged++

		regressions, improvements := findingsForModifiedFile(head, base, path, opts)
		result.Regressions = append(result.Regressions, regressions...)
		result.Improvements = append(result.Improvements, improvements...)
	}

	for _, base := range baseFiles {
		if !seen[relativize(base.Path, baseRoot)] {
			result.Summary.FilesDeleted++
		}
	}

	sortFindings(result.Regressions)
	sortFindings(result.Improvements)
	result.Summary.High = countBySeverity(result.Regressions, SeverityHigh)
	result.Summary.Medium = countBySeverity(result.Regressions, SeverityMedium)
	result.Summary.Low = countBySeverity(result.Regressions, SeverityLow)
	result.Summary.Improvements = len(result.Improvements)

	return result
}

// AppendFindings merges extra findings (e.g. lint diff) into the result and
// recomputes counters, keeping the output deterministic.
func (r *Result) AppendFindings(findings []Finding) {
	for _, f := range findings {
		if f.Kind == KindImprovement {
			r.Improvements = append(r.Improvements, f)
		} else {
			r.Regressions = append(r.Regressions, f)
		}
	}
	sortFindings(r.Regressions)
	sortFindings(r.Improvements)
	r.Summary.High = countBySeverity(r.Regressions, SeverityHigh)
	r.Summary.Medium = countBySeverity(r.Regressions, SeverityMedium)
	r.Summary.Low = countBySeverity(r.Regressions, SeverityLow)
	r.Summary.Improvements = len(r.Improvements)
}

// HasRegressionAtLeast reports whether at least one regression has the given
// severity or a more severe one.
func (r *Result) HasRegressionAtLeast(level Severity) bool {
	rank := severityRank(level)
	for _, f := range r.Regressions {
		if severityRank(f.Severity) >= rank {
			return true
		}
	}
	return false
}

func findingsForNewFile(head *pb.File, path string, opts Options) []Finding {
	findings := []Finding{}
	for key, fn := range collectFunctions(head) {
		ccn := cyclomaticOf(fn)
		if ccn < int32(opts.ComplexityMedium) {
			continue
		}
		severity := SeverityMedium
		if ccn >= int32(opts.ComplexityHigh) {
			severity = SeverityHigh
		}
		findings = append(findings, Finding{
			Kind:       KindRegression,
			Severity:   severity,
			Rule:       "new-complex-function",
			File:       path,
			Line:       lineOf(fn),
			Subject:    key,
			Message:    fmt.Sprintf("New function with cyclomatic complexity %d (threshold: %d)", ccn, opts.ComplexityMedium),
			Suggestion: "Extract smaller, well-named functions to reduce decision points",
			Before:     0,
			After:      float64(ccn),
		})
	}
	return findings
}

func findingsForModifiedFile(head *pb.File, base *pb.File, path string, opts Options) ([]Finding, []Finding) {
	regressions := []Finding{}
	improvements := []Finding{}

	headFunctions := collectFunctions(head)
	baseFunctions := collectFunctions(base)

	for key, headFn := range headFunctions {
		headCcn := cyclomaticOf(headFn)
		baseFn, existed := baseFunctions[key]

		if !existed {
			if headCcn >= int32(opts.ComplexityMedium) {
				severity := SeverityMedium
				if headCcn >= int32(opts.ComplexityHigh) {
					severity = SeverityHigh
				}
				regressions = append(regressions, Finding{
					Kind:       KindRegression,
					Severity:   severity,
					Rule:       "new-complex-function",
					File:       path,
					Line:       lineOf(headFn),
					Subject:    key,
					Message:    fmt.Sprintf("New function with cyclomatic complexity %d (threshold: %d)", headCcn, opts.ComplexityMedium),
					Suggestion: "Extract smaller, well-named functions to reduce decision points",
					Before:     0,
					After:      float64(headCcn),
				})
			}
			continue
		}

		baseCcn := cyclomaticOf(baseFn)
		if headCcn > baseCcn && headCcn >= int32(opts.ComplexityMedium) {
			severity := SeverityMedium
			if headCcn >= int32(opts.ComplexityHigh) || int(headCcn-baseCcn) >= opts.ComplexityJumpHigh {
				severity = SeverityHigh
			}
			regressions = append(regressions, Finding{
				Kind:       KindRegression,
				Severity:   severity,
				Rule:       "complexity-regression",
				File:       path,
				Line:       lineOf(headFn),
				Subject:    key,
				Message:    fmt.Sprintf("Cyclomatic complexity: %d -> %d (threshold: %d)", baseCcn, headCcn, opts.ComplexityMedium),
				Suggestion: "Extract smaller, well-named functions to reduce decision points",
				Before:     float64(baseCcn),
				After:      float64(headCcn),
			})
		}
		if int(baseCcn-headCcn) >= opts.ImprovementComplexityDrop {
			improvements = append(improvements, Finding{
				Kind:     KindImprovement,
				Severity: SeverityLow,
				Rule:     "complexity-improvement",
				File:     path,
				Line:     lineOf(headFn),
				Subject:  key,
				Message:  fmt.Sprintf("Cyclomatic complexity: %d -> %d", baseCcn, headCcn),
				Before:   float64(baseCcn),
				After:    float64(headCcn),
			})
		}
	}

	// Maintainability index, at file level
	headMi, headOk := maintainabilityOf(head)
	baseMi, baseOk := maintainabilityOf(base)
	if headOk && baseOk {
		drop := baseMi - headMi
		if drop >= opts.MaintainabilityDrop && headMi < opts.MaintainabilityLow {
			severity := SeverityMedium
			if headMi < opts.MaintainabilityCritical {
				severity = SeverityHigh
			}
			regressions = append(regressions, Finding{
				Kind:       KindRegression,
				Severity:   severity,
				Rule:       "maintainability-regression",
				File:       path,
				Message:    fmt.Sprintf("Maintainability index: %.0f -> %.0f (below %.0f is hard to maintain)", baseMi, headMi, opts.MaintainabilityLow),
				Suggestion: "Refactor long or deeply nested functions and reduce responsibilities in this file",
				Before:     baseMi,
				After:      headMi,
			})
		}
		if headMi-baseMi >= opts.ImprovementMaintainabilityGain && baseMi < opts.MaintainabilityLow {
			improvements = append(improvements, Finding{
				Kind:     KindImprovement,
				Severity: SeverityLow,
				Rule:     "maintainability-improvement",
				File:     path,
				Message:  fmt.Sprintf("Maintainability index: %.0f -> %.0f", baseMi, headMi),
				Before:   baseMi,
				After:    headMi,
			})
		}
	}

	// Efferent coupling, at file level
	headCoupling, headHasCoupling := efferentOf(head)
	baseCoupling, baseHasCoupling := efferentOf(base)
	if headHasCoupling && baseHasCoupling && int(headCoupling-baseCoupling) >= opts.CouplingIncrease {
		regressions = append(regressions, Finding{
			Kind:       KindRegression,
			Severity:   SeverityMedium,
			Rule:       "coupling-regression",
			File:       path,
			Message:    fmt.Sprintf("Efferent coupling: %d -> %d", baseCoupling, headCoupling),
			Suggestion: "Review the new outgoing dependencies; consider inverting or removing them",
			Before:     float64(baseCoupling),
			After:      float64(headCoupling),
		})
	}

	return regressions, improvements
}

// collectFunctions returns every function and method of a file, indexed by a
// stable key (Class::method or function name).
func collectFunctions(file *pb.File) map[string]*pb.StmtFunction {
	functions := map[string]*pb.StmtFunction{}
	if file == nil || file.Stmts == nil {
		return functions
	}
	var walk func(stmts *pb.Stmts, prefix string)
	walk = func(stmts *pb.Stmts, prefix string) {
		if stmts == nil {
			return
		}
		for _, fn := range stmts.StmtFunction {
			if fn == nil || fn.Name == nil {
				continue
			}
			name := fn.Name.Short
			if name == "" {
				name = fn.Name.Qualified
			}
			key := name
			if prefix != "" {
				key = prefix + "::" + name
			}
			functions[key] = fn
		}
		for _, class := range stmts.StmtClass {
			if class == nil {
				continue
			}
			className := ""
			if class.Name != nil {
				className = class.Name.Short
				if className == "" {
					className = class.Name.Qualified
				}
			}
			walk(class.Stmts, className)
		}
		for _, ns := range stmts.StmtNamespace {
			if ns == nil {
				continue
			}
			walk(ns.Stmts, prefix)
		}
	}
	walk(file.Stmts, "")
	return functions
}

func cyclomaticOf(fn *pb.StmtFunction) int32 {
	if fn == nil || fn.Stmts == nil || fn.Stmts.Analyze == nil || fn.Stmts.Analyze.Complexity == nil || fn.Stmts.Analyze.Complexity.Cyclomatic == nil {
		return 0
	}
	return *fn.Stmts.Analyze.Complexity.Cyclomatic
}

func lineOf(fn *pb.StmtFunction) int {
	if fn == nil || fn.Location == nil {
		return 0
	}
	return int(fn.Location.StartLine)
}

func maintainabilityOf(file *pb.File) (float64, bool) {
	if file == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Maintainability == nil || file.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
		return 0, false
	}
	mi := *file.Stmts.Analyze.Maintainability.MaintainabilityIndex
	// NaN guard
	if mi != mi {
		return 0, false
	}
	return mi, true
}

func efferentOf(file *pb.File) (int32, bool) {
	if file == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Coupling == nil {
		return 0, false
	}
	return file.Stmts.Analyze.Coupling.Efferent, true
}

func relativize(path string, root string) string {
	if root == "" {
		return path
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}

func severityRank(s Severity) int {
	switch s {
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	}
	return 0
}

func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if severityRank(findings[i].Severity) != severityRank(findings[j].Severity) {
			return severityRank(findings[i].Severity) > severityRank(findings[j].Severity)
		}
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].Subject != findings[j].Subject {
			return findings[i].Subject < findings[j].Subject
		}
		return findings[i].Rule < findings[j].Rule
	})
}

func countBySeverity(findings []Finding, s Severity) int {
	count := 0
	for _, f := range findings {
		if f.Severity == s {
			count++
		}
	}
	return count
}
