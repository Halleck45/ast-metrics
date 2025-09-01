package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/report"
)

type ScreenEnd struct {
	isInteractive     bool
	files             []*pb.File
	projectAggregated analyzer.ProjectAggregated
	// program
	tea *tea.Program
	// reports
	Configuration configuration.Configuration
	reports       []report.GeneratedReport
}

func NewScreenEnd(
	isInteractive bool,
	files []*pb.File,
	projectAggregated analyzer.ProjectAggregated,
	configuration configuration.Configuration,
	reports []report.GeneratedReport,
) *ScreenEnd {
	return &ScreenEnd{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
		Configuration:     configuration,
		reports:           reports,
	}
}

type modelEnd struct {
}

func (m modelEnd) Init() tea.Cmd {
	return nil
}

func (m *modelEnd) Reset(files []*pb.File, projectAggregated analyzer.ProjectAggregated) {
}

func (m modelEnd) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m modelEnd) View() string {
	return ""
}

func (r *ScreenEnd) Render() {
	// List reports
	if r.Configuration.Reports.HasReports() {

		fmt.Println("\n📁 These reports have been generated:")

		for _, report := range r.reports {
			fmt.Println("\n  ✔ " + report.Path + " (" + report.Type + ")")
			fmt.Println("\n        " + report.Description)
		}

		fmt.Println("")
	}

	// Tips if configuration file does not exist
	if !r.Configuration.IsComingFromConfigFile {
		fmt.Println("\n💡 We noticed that you haven't yet created a configuration file. You can create a .ast-metrics.yaml configuration file by running: ast-metrics init")
		fmt.Println("")
	}

	fmt.Println("\n🌟 If you like AST Metrics, please consider starring the project on GitHub: https://github.com/Halleck45/ast-metrics/. Thanks!")
	fmt.Println("")

}

func (r *ScreenEnd) Reset(files []*pb.File, projectAggregated analyzer.ProjectAggregated) {
}

func (r ScreenEnd) GetModel() tea.Model {
	return modelEnd{}
}

func (r ScreenEnd) GetScreenName() string {
	return "End"
}
