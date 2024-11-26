package Ui

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ComponentBarchartCyclomaticByMethodRepartition struct {
	Aggregated Analyzer.Aggregated
	Files      []*pb.File
}

func (c *ComponentBarchartCyclomaticByMethodRepartition) AsTerminalElement() string {

	dataOrdered := c.GetData()
	data := make(map[string]float32)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}

	graph := ComponentBarchart{data: data}
	graph.height = 5
	graph.barWidth = 8
	return graph.AsTerminalElement()
}

func (c *ComponentBarchartCyclomaticByMethodRepartition) GetData() *orderedmap.OrderedMap[string, float32] {
	data := orderedmap.NewOrderedMap[string, float32]()

	rangeOfLabels := []string{"0-5", "5-20", "> 20"}
	rangeOfValues := []int32{5, 20, 999999}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	// repartition of classes by cyclomatic complexity
	for _, file := range c.Files {

		functions := Engine.GetFunctionsInFile(file)
		for _, function := range functions {
			if function.Stmts.Analyze == nil {
				continue
			}

			mesured := *function.Stmts.Analyze.Complexity.Cyclomatic
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

// render as HTML
func (c *ComponentBarchartCyclomaticByMethodRepartition) AsHtml() string {
	data := c.GetData()
	return Engine.HtmlChartLine(data, "Number of files", "chart-loc")
}
