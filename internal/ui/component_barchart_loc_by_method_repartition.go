package ui

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// ComponentBarchartLocByMethodRepartition is the barchart component for the loc repartition
type ComponentBarchartLocByMethodRepartition struct {
	Aggregated analyzer.Aggregated
	Files      []*pb.File
}

// Render is the method to render the component
func (c *ComponentBarchartLocByMethodRepartition) AsTerminalElement() string {
	dataOrdered := c.GetData()
	data := make(map[string]float64)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}
	graph := ComponentBarchart{data: data}
	graph.height = 5
	graph.barWidth = 6
	return graph.AsTerminalElement()
}

// Render Html
func (c *ComponentBarchartLocByMethodRepartition) AsHtml() string {
	data := c.GetData()
	return engine.HtmlChartLine(data, "Number of files", "chart-loc-by-method")
}

func (c *ComponentBarchartLocByMethodRepartition) GetData() *orderedmap.OrderedMap[string, float64] {
	data := orderedmap.NewOrderedMap[string, float64]()

	rangeOfLabels := []string{"< 15", "< 35", "< 50", "> 50"}
	rangeOfValues := []int32{15, 35, 50, 999999}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	// repartition of files by LOC
	for _, file := range c.Files {
		functions := engine.GetFunctionsInFile(file)
		for _, funct := range functions {
			if funct.Stmts.Analyze == nil {
				continue
			}
			mesured := *funct.Stmts.Analyze.Volume.Loc
			for i, r := range rangeOfValues {
				if mesured < r {
					value, _ := data.Get(rangeOfLabels[i])
					data.Set(rangeOfLabels[i], value+1)
					break
				}
			}
		}
	}

	return data
}
