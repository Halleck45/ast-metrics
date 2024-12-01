package Cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Ui"
)

type ComponentStatisticsOverview struct {
	isInteractive bool
	files         []*pb.File
	aggregated    Analyzer.Aggregated
}

func NewComponentStatisticsOverview(files []*pb.File, aggregated Analyzer.Aggregated) *ComponentStatisticsOverview {
	return &ComponentStatisticsOverview{
		files:      files,
		aggregated: aggregated,
	}
}

func (v *ComponentStatisticsOverview) Render() string {

	// Cyclomatic complexity repartition
	chartRepartitionCyclomatic := Ui.ComponentBarchartCyclomaticByMethodRepartition{
		Aggregated: v.aggregated,
		Files:      v.files,
	}
	boxCcn := StyleNumberBox(
		fmt.Sprintf("%.2f", v.aggregated.CyclomaticComplexityPerMethod.Avg),
		"Cycl. complexity per method",
		chartRepartitionCyclomatic.AsTerminalElement(),
	)

	// LOC repartition
	chartRepartitionLocByMethod := Ui.ComponentBarchartLocByMethodRepartition{
		Aggregated: v.aggregated,
		Files:      v.files,
	}
	boxMethods := StyleNumberBox(
		fmt.Sprintf("%.2f", v.aggregated.LocPerMethod.Avg),
		"Average LOC per method",
		chartRepartitionLocByMethod.AsTerminalElement()+"     ",
	)

	// MI repartition
	chartRepartitionMI := Ui.ComponentBarchartMaintainabilityIndexRepartition{
		Aggregated: v.aggregated,
		Files:      v.files,
	}
	boxMaintainability := StyleNumberBox(
		DecorateMaintainabilityIndex(int(v.aggregated.MaintainabilityIndex.Avg), nil),
		"Maintainability index",
		chartRepartitionMI.AsTerminalElement(),
	)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, boxCcn.Render(), boxMethods.Render(), boxMaintainability.Render())

	return row1
}

func (v *ComponentStatisticsOverview) Update(msg tea.Msg) {
	// pass
}
