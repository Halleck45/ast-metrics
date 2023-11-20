package Cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ComponentStatisticsOverview struct {
	isInteractive bool
	files         []pb.File
	aggregated    Analyzer.Aggregated
}

func NewComponentStatisticsOverview(files []pb.File, aggregated Analyzer.Aggregated) *ComponentStatisticsOverview {
	return &ComponentStatisticsOverview{
		files:      files,
		aggregated: aggregated,
	}
}

func (v *ComponentStatisticsOverview) Render() string {

	// Screen is composed from differents boxes
	boxCcn := StyleNumberBox(
		fmt.Sprintf("%.2f", v.aggregated.AverageCyclomaticComplexityPerMethod),
		"Cycl. complexity per method",
		fmt.Sprintf("(min: %d, max: %d)", v.aggregated.MinCyclomaticComplexity, v.aggregated.MaxCyclomaticComplexity),
	)
	boxMethods := StyleNumberBox(
		fmt.Sprintf("%.2f", v.aggregated.AverageLocPerMethod),
		"Average LOC per method",
		"",
	)
	boxMaintainability := StyleNumberBox(
		DecorateMaintainabilityIndex(int(v.aggregated.AverageMI)),
		"Maintainability index",
		fmt.Sprintf("(MI without comments: %.2f, comment weight: %.2f)", v.aggregated.AverageMIwoc, v.aggregated.AverageMIcw),
	)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, boxCcn.Render(), boxMethods.Render(), boxMaintainability.Render())

	return row1
}

func (v *ComponentStatisticsOverview) Update(msg tea.Msg) {
	// pass
}
