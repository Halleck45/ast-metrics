package ui

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

// ComponentBarchartMaintainabilityIndexRepartition is the barchart component for the loc repartition
type ComponentBarchartMaintainabilityIndexRepartition struct {
	Aggregated analyzer.Aggregated
	Files      []*pb.File
}

// render as HTML
func (c *ComponentBarchartMaintainabilityIndexRepartition) AsHtml() string {
	data := c.GetData()
	return engine.HtmlChartLine(data, "Number of files", "chart-mi")
}

// Render is the method to render the component
func (c *ComponentBarchartMaintainabilityIndexRepartition) AsTerminalElement() string {
	dataOrdered := c.GetData()
	data := make(map[string]float64)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}
	graph := ComponentBarchart{data: data}
	graph.height = 5
	return graph.AsTerminalElement()
}

// GetData returns the data for the barchart
func (c *ComponentBarchartMaintainabilityIndexRepartition) GetData() *orderedmap.OrderedMap[string, float64] {
	data := orderedmap.NewOrderedMap[string, float64]()

	rangeOfLabels := []string{"🔴 < 64", "🟡 < 85", "🟢 > 85"}
	rangeOfValues := []float64{64, 85, 1000}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	// repartition of files by LOC
	for _, file := range c.Files {
		classes := engine.GetClassesInFile(file)

		if classes == nil || len(classes) == 0 {
			if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Maintainability != nil && file.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil {
				miForFile := file.Stmts.Analyze.Maintainability.MaintainabilityIndex
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
