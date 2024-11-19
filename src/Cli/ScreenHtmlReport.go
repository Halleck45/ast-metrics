package Cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Report"
)

type ScreenHtmlReport struct {
	isInteractive     bool
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func NewScreenHtmlReport(isInteractive bool, files []*pb.File, projectAggregated Analyzer.ProjectAggregated) ScreenHtmlReport {
	return ScreenHtmlReport{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
	}
}

func (v ScreenHtmlReport) GetScreenName() string {
	return "Generate HTML report\n"
}

func (v ScreenHtmlReport) GetModel() tea.Model {
	m := modelScreenHtmlReport{files: v.files, projectAggregated: v.projectAggregated}
	return m
}

type modelScreenHtmlReport struct {
	parent            tea.Model
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
	generated         bool
}

func (m modelScreenHtmlReport) Init() tea.Cmd {
	return nil
}

func (m *ScreenHtmlReport) Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) {
	m.files = files
	m.projectAggregated = projectAggregated
}

func (m modelScreenHtmlReport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return NewScreenHome(true, m.files, m.projectAggregated).GetModel(), tea.ClearScreen
		}
	case DoRefreshModel:
		// refresh the model
		m.files = msg.files
		m.projectAggregated = msg.projectAggregated
		m.generated = false
		return m, nil
	}
	return m, nil
}

func (m modelScreenHtmlReport) View() string {

	directory := "ast-metrics-report"

	if !m.generated {
		// Generate report
		// report: html
		htmlReportGenerator := Report.NewHtmlReportGenerator(directory)
		_, err := htmlReportGenerator.Generate(m.files, m.projectAggregated)
		if err != nil {
			return fmt.Sprintf("Error generating report: %s", err)
		}
	}

	destination := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f0f0f0")).
		Render(directory + "/index.html")

	box := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Margin(2)

	return StyleScreen(
		StyleTitle("HTML report").Render() +
			box.Render("Report generated at: "+destination) +
			StyleHowToQuit("").Render()).Render()
}
