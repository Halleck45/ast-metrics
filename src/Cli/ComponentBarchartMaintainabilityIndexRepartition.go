package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

// ComponentBarchartMaintainabilityIndexRepartition is the barchart component for the loc repartition
type ComponentBarchartMaintainabilityIndexRepartition struct {
	aggregated Analyzer.Aggregated
	files      []*pb.File
}

// NewComponentBarchartMaintainabilityIndexRepartition is the constructor for the ComponentBarchartMaintainabilityIndexRepartition
func NewComponentBarchartMaintainabilityIndexRepartition(aggregated Analyzer.Aggregated, files []*pb.File) *ComponentBarchartMaintainabilityIndexRepartition {
	return &ComponentBarchartMaintainabilityIndexRepartition{
		aggregated: aggregated,
		files:      files,
	}
}

// render as HTML
func (c *ComponentBarchartMaintainabilityIndexRepartition) RenderHTML() string {
	data := c.GetData()
	return Engine.HtmlChartLine(data, "Number of files", "chart-mi")
}

// Render is the method to render the component
func (c *ComponentBarchartMaintainabilityIndexRepartition) Render() string {
	dataOrdered := c.GetData()
	data := make(map[string]float64)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}
	graph := NewComponentBarchart(data)
	graph.height = 5
	return graph.Render()
}

// GetData returns the data for the barchart
func (c *ComponentBarchartMaintainabilityIndexRepartition) GetData() *orderedmap.OrderedMap[string, float64] {
	data := orderedmap.NewOrderedMap[string, float64]()

	rangeOfLabels := []string{"🔴 < 64", "🟡 < 85", "🟢 > 85"}
	rangeOfValues := []float32{64, 85, 1000}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	// repartition of files by LOC
	for _, file := range c.files {
		classes := Engine.GetClassesInFile(file)

		if classes == nil || len(classes) == 0 {
			miForFile := file.Stmts.Analyze.Maintainability.MaintainabilityIndex
			if miForFile != nil {
				for i, r := range rangeOfValues {
					if *miForFile < r {
						value, _ := data.Get(rangeOfLabels[i])
						data.Set(rangeOfLabels[i], value+1)
						break
					}
				}
			}
		}

		for _, class := range classes {
			if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Maintainability == nil {
				continue
			}
			mesured := class.Stmts.Analyze.Maintainability.MaintainabilityIndex
			for i, r := range rangeOfValues {
				if *mesured < r {
					value, _ := data.Get(rangeOfLabels[i])
					data.Set(rangeOfLabels[i], value+1)
					break
				}
			}
		}
	}

	return data
}

// Update is the method to update the component
func (c *ComponentBarchartMaintainabilityIndexRepartition) Update(msg tea.Msg) {
}
