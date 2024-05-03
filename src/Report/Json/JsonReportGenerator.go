package Report

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Pkg/Cleaner"
)

type JsonReportGenerator struct {
	ReportPath string
}

// This factory creates a new JsonReportGenerator
func NewJsonReportGenerator(ReportPath string) *JsonReportGenerator {
	return &JsonReportGenerator{
		ReportPath: ReportPath,
	}
}

// Generate generates a JSON report
func (j *JsonReportGenerator) Generate(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) error {
	report := j.buildReport(projectAggregated)

	err := Cleaner.CleanVal(report)
	if err != nil {
		return fmt.Errorf("can not clean report err: %s", err.Error())
	}

	// This code serializes the results to JSON
	jsonReport, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("can not serialize report to JSON err: %s", err.Error())
	}

	// This code writes the JSON report to a file
	err = ioutil.WriteFile(j.ReportPath, jsonReport, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can not save report to path %s err: %s", j.ReportPath, err.Error())
	}

	return nil
}

// The buildReport creates a JSON report using the
// `projectAggregated.Combined`. For concerned
// files, it only includes the `stmts.Analyze`.
func (j *JsonReportGenerator) buildReport(projectAggregated Analyzer.ProjectAggregated) *report {
	r := &report{}
	combined := projectAggregated.Combined

	r.ConcernedFiles = make([]file, len(combined.ConcernedFiles))
	for i, f := range combined.ConcernedFiles {
		r.ConcernedFiles[i] = file{
			Path: f.Path,
			Complexity: complexity{
				Cyclomatic: *f.Stmts.Analyze.Complexity.Cyclomatic,
			},
			Volume: volume{
				Loc:                     f.Stmts.Analyze.Volume.GetLoc(),
				Lloc:                    f.Stmts.Analyze.Volume.GetLloc(),
				Cloc:                    f.Stmts.Analyze.Volume.GetCloc(),
				HalsteadVolume:          f.Stmts.Analyze.Volume.GetHalsteadVolume(),
				HalsteadDifficulty:      f.Stmts.Analyze.Volume.GetHalsteadDifficulty(),
				HalsteadEffort:          f.Stmts.Analyze.Volume.GetHalsteadEffort(),
				HalsteadTime:            f.Stmts.Analyze.Volume.GetHalsteadTime(),
				HalsteadVocabulary:      f.Stmts.Analyze.Volume.GetHalsteadVocabulary(),
				HalsteadLength:          f.Stmts.Analyze.Volume.GetHalsteadLength(),
				HalsteadEstimatedLength: f.Stmts.Analyze.Volume.GetHalsteadEstimatedLength(),
			},
			Maintainability: maintainability{
				MaintainabilityIndex:                f.Stmts.Analyze.Maintainability.GetMaintainabilityIndex(),
				MaintainabilityIndexWithoutComments: f.Stmts.Analyze.Maintainability.GetMaintainabilityIndexWithoutComments(),
				CommentWeight:                       f.Stmts.Analyze.Maintainability.GetCommentWeight(),
			},
			Risk: risk{
				Score: f.Stmts.Analyze.Risk.GetScore(),
			},
			Coupling: coupling{
				Afferent:    f.Stmts.Analyze.Coupling.GetAfferent(),
				Efferent:    f.Stmts.Analyze.Coupling.GetEfferent(),
				Instability: f.Stmts.Analyze.Coupling.GetInstability(),
			},
		}
	}

	r.TopCommitters = make([]contributor, len(combined.TopCommitters))
	for i, committer := range combined.TopCommitters {
		r.TopCommitters[i] = contributor{Name: committer.Name, Count: committer.Count}
	}

	r.GitAnalysis = make([]gitAnalysis, len(combined.ResultOfGitAnalysis))
	for i, analysis := range combined.ResultOfGitAnalysis {
		r.GitAnalysis[i] = gitAnalysis{
			ProgrammingLanguage:     analysis.ProgrammingLanguage,
			ReportRootDir:           analysis.ReportRootDir,
			CountCommits:            analysis.CountCommits,
			CountCommiters:          analysis.CountCommiters,
			CountCommitsForLanguage: analysis.CountCommitsForLanguage,
			CountCommitsIgnored:     analysis.CountCommitsIgnored,
		}
	}

	// Other fields
	r.NbFiles = combined.NbFiles
	r.NbFunctions = combined.NbFunctions
	r.NbClasses = combined.NbClasses
	r.NbClassesWithCode = combined.NbClassesWithCode
	r.NbMethods = combined.NbMethods
	r.Loc = combined.Loc
	r.Cloc = combined.Cloc
	r.Lloc = combined.Lloc
	r.AverageMethodsPerClass = combined.AverageMethodsPerClass
	r.AverageLocPerMethod = combined.AverageLocPerMethod
	r.AverageLlocPerMethod = combined.AverageLlocPerMethod
	r.AverageClocPerMethod = combined.AverageClocPerMethod
	r.AverageCyclomaticComplexityPerMethod = combined.AverageCyclomaticComplexityPerMethod
	r.AverageCyclomaticComplexityPerClass = combined.AverageCyclomaticComplexityPerClass
	r.MinCyclomaticComplexity = combined.MinCyclomaticComplexity
	r.MaxCyclomaticComplexity = combined.MaxCyclomaticComplexity
	r.AverageHalsteadDifficulty = combined.AverageHalsteadDifficulty
	r.AverageHalsteadEffort = combined.AverageHalsteadEffort
	r.AverageHalsteadVolume = combined.AverageHalsteadVolume
	r.AverageHalsteadTime = combined.AverageHalsteadTime
	r.AverageHalsteadBugs = combined.AverageHalsteadBugs
	r.SumHalsteadDifficulty = combined.SumHalsteadDifficulty
	r.SumHalsteadEffort = combined.SumHalsteadEffort
	r.SumHalsteadVolume = combined.SumHalsteadVolume
	r.SumHalsteadTime = combined.SumHalsteadTime
	r.SumHalsteadBugs = combined.SumHalsteadBugs
	r.AverageMI = combined.AverageMI
	r.AverageMIwoc = combined.AverageMIwoc
	r.AverageMIcw = combined.AverageMIcw
	r.AverageMIPerMethod = combined.AverageMIPerMethod
	r.AverageMIwocPerMethod = combined.AverageMIwocPerMethod
	r.AverageMIcwPerMethod = combined.AverageMIcwPerMethod
	r.AverageAfferentCoupling = combined.AverageAfferentCoupling
	r.AverageEfferentCoupling = combined.AverageEfferentCoupling
	r.AverageInstability = combined.AverageInstability
	r.CommitCountForPeriod = combined.CommitCountForPeriod
	r.CommittedFilesCountForPeriod = combined.CommittedFilesCountForPeriod
	r.BusFactor = combined.BusFactor
	r.PackageRelations = combined.PackageRelations

	return r
}
