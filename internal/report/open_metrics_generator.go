package report

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/bsm/openmetrics"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

type OpenMetricsReportGenerator struct {
	// The path where the report will be generated
	ReportPath string
}

func NewOpenMetricsReportGenerator(reportPath string) Reporter {
	return &OpenMetricsReportGenerator{
		ReportPath: reportPath,
	}
}

func (v *OpenMetricsReportGenerator) Generate(files []*pb.File, projectAggregated analyzer.ProjectAggregated) ([]GeneratedReport, error) {

	if v.ReportPath == "" {
		return nil, nil
	}

	reg := openmetrics.NewConsistentRegistry(func() time.Time {
		return time.Now()
	})

	// Prepare series
	ccn := reg.Gauge(openmetrics.Desc{
		Name:   "cyclomatic_complexity",
		Help:   "Cyclomatic complexity of the code",
		Labels: []string{"path"},
	})
	loc := reg.Gauge(openmetrics.Desc{
		Name:   "lines_of_code",
		Help:   "Lines of code",
		Labels: []string{"path"},
	})
	lloc := reg.Gauge(openmetrics.Desc{
		Name:   "logical_lines_of_code",
		Help:   "Logical lines of code",
		Labels: []string{"path"},
	})
	cloc := reg.Gauge(openmetrics.Desc{
		Name:   "comment_lines_of_code",
		Help:   "Comment lines of code",
		Labels: []string{"path"},
	})
	maintanability := reg.Gauge(openmetrics.Desc{
		Name:   "maintainability",
		Help:   "Maintainability index",
		Labels: []string{"path"},
	})
	maintanabilityWithoutComments := reg.Gauge(openmetrics.Desc{
		Name:   "maintainability_without_comments",
		Help:   "Maintainability index without comments",
		Labels: []string{"path"},
	})
	numberOfFunctions := reg.Gauge(openmetrics.Desc{
		Name:   "number_of_functions",
		Help:   "Number of functions",
		Labels: []string{"path"},
	})
	numberOfClasses := reg.Gauge(openmetrics.Desc{
		Name:   "number_of_classes",
		Help:   "Number of classes",
		Labels: []string{"path"},
	})
	afferentCoupling := reg.Gauge(openmetrics.Desc{
		Name:   "afferent_coupling",
		Help:   "Afferent coupling",
		Labels: []string{"path"},
	})
	efferentCoupling := reg.Gauge(openmetrics.Desc{
		Name:   "efferent_coupling",
		Help:   "Efferent coupling",
		Labels: []string{"path"},
	})

	// Add data to the series
	for _, file := range files {
		if file.Stmts == nil || file.Stmts.Analyze == nil {
			continue
		}

		if file.Stmts.Analyze.Complexity != nil {
			ccn.With(file.Path).Set(float64(*file.Stmts.Analyze.Complexity.Cyclomatic))
		}

		if file.Stmts.Analyze.Volume != nil {
			loc.With(file.Path).Set(float64(*file.Stmts.Analyze.Volume.Loc))
			lloc.With(file.Path).Set(float64(*file.Stmts.Analyze.Volume.Lloc))
			cloc.With(file.Path).Set(float64(*file.Stmts.Analyze.Volume.Cloc))
		}

		if file.Stmts.Analyze.Maintainability != nil && file.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil {
			maintanability.With(file.Path).Set(float64(*file.Stmts.Analyze.Maintainability.MaintainabilityIndex))
			if file.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments != nil {
				maintanabilityWithoutComments.With(file.Path).Set(float64(*file.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments))
			}
		}

		numberOfFunctions.With(file.Path).Set(float64(len(engine.GetFunctionsInFile(file))))
		numberOfClasses.With(file.Path).Set(float64(len(engine.GetClassesInFile(file))))

		if file.Stmts.Analyze.Coupling != nil {
			afferentCoupling.With(file.Path).Set(float64(file.Stmts.Analyze.Coupling.Afferent))
			efferentCoupling.With(file.Path).Set(float64(file.Stmts.Analyze.Coupling.Efferent))
		}
	}

	// Write the report
	var buf bytes.Buffer
	if _, err := reg.WriteTo(&buf); err != nil {
		panic(err)
	}
	file, err := os.Create(v.ReportPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, &buf)
	if err != nil {
		return nil, err
	}

	// Return the created report, in order to inform the user
	reports := []GeneratedReport{
		{
			Path:        v.ReportPath,
			Type:        "file",
			Description: "The openmetrics report allows to monitor the project with prometheus, or with specific tools that can read openmetrics format (like Gitlab CI).",
			Icon:        "📄",
		},
	}
	return reports, nil

}
