package Cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Configuration"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Report"
)

type ScreenEnd struct {
	isInteractive     bool
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
	// program
	tea *tea.Program
	// reports
	configuration Configuration.Configuration
	reports       []Report.GeneratedReport
}

func NewScreenEnd(
	isInteractive bool,
	files []*pb.File,
	projectAggregated Analyzer.ProjectAggregated,
	configuration Configuration.Configuration,
	reports []Report.GeneratedReport,
) *ScreenEnd {
	return &ScreenEnd{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
		configuration:     configuration,
		reports:           reports,
	}
}

type modelEnd struct {
}

func (m modelEnd) Init() tea.Cmd {
	return nil
}

func (m *modelEnd) Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) {
}

func (m modelEnd) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m modelEnd) View() string {
	return ""
}

func (r *ScreenEnd) Render() {
	// List reports
	if r.configuration.Reports.HasReports() {

		fmt.Println("\n📁 These reports have been generated:")

		for _, report := range r.reports {
			fmt.Println("\n  ✔ " + report.Path + " (" + report.Type + ")")
			fmt.Println("\n        " + report.Description)
		}

		fmt.Println("")
	}

	// Tips if configuration file does not exist
	if !r.configuration.IsComingFromConfigFile {
		fmt.Println("\n💡 We noticed that you haven't yet created a configuration file. You can create a .ast-metrics.yaml configuration file by running: ast-metrics init")
		fmt.Println("")
	}

	fmt.Println("\n🌟 If you like AST Metrics, please consider starring the project on GitHub: https://github.com/Halleck45/ast-metrics/. Thanks!")
	fmt.Println("")

}

func (r *ScreenEnd) Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) {
}

func (r ScreenEnd) GetModel() tea.Model {
	return modelEnd{}
}

func (r ScreenEnd) GetScreenName() string {
	return "End"
}
