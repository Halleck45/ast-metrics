package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

// ComponentBarchartMaintainabilityIndexRepartition is the barchart component for the loc repartition
type ComponentBarchartMaintainabilityIndexRepartition struct {
	aggregated Analyzer.Aggregated
	files      []pb.File
}

// NewComponentBarchartMaintainabilityIndexRepartition is the constructor for the ComponentBarchartMaintainabilityIndexRepartition
func NewComponentBarchartMaintainabilityIndexRepartition(aggregated Analyzer.Aggregated, files []pb.File) *ComponentBarchartMaintainabilityIndexRepartition {
	return &ComponentBarchartMaintainabilityIndexRepartition{
		aggregated: aggregated,
		files:      files,
	}
}

// Render is the method to render the component
func (c *ComponentBarchartMaintainabilityIndexRepartition) Render() string {
	data := make(map[string]float64)

	rangeOfLabels := []string{"🔴 < 64", "🟡 < 85", "🟢 > 85"}
	rangeOfValues := []float32{64, 85, 1000}
	for _, r := range rangeOfLabels {
		data[r] = 0
	}

	// repartition of files by LOC
	for _, file := range c.files {
		classes := Engine.GetClassesInFile(&file)
		for _, class := range classes {
			if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Maintainability == nil {
				continue
			}
			mesured := class.Stmts.Analyze.Maintainability.MaintainabilityIndex
			for i, r := range rangeOfValues {
				if *mesured < r {
					data[rangeOfLabels[i]]++
					break
				}
			}
		}
	}

	graph := NewComponentBarchart(data)
	graph.height = 5
	return graph.Render()
}

// Update is the method to update the component
func (c *ComponentBarchartMaintainabilityIndexRepartition) Update(msg tea.Msg) {
}
