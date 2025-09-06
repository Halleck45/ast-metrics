package ui

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// ComponentBarchartLcomRepartition renders a barchart for LCOM4 repartition by class
type ComponentBarchartLcomRepartition struct {
	Aggregated analyzer.Aggregated
	Files      []*pb.File
}

// AsHtml renders the chart as HTML
func (c *ComponentBarchartLcomRepartition) AsHtml() string {
	data := c.GetData()
	return engine.HtmlChartLine(data, "Number of classes", "chart-lcom4")
}

// AsTerminalElement renders the chart in terminal
func (c *ComponentBarchartLcomRepartition) AsTerminalElement() string {
	dataOrdered := c.GetData()
	data := make(map[string]float64)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}
	graph := ComponentBarchart{data: data}
	graph.height = 5
	graph.barWidth = 8
	return graph.AsTerminalElement()
}

// GetData computes the repartition of classes by LCOM4 value
func (c *ComponentBarchartLcomRepartition) GetData() *orderedmap.OrderedMap[string, float64] {
	data := orderedmap.NewOrderedMap[string, float64]()

	// Buckets inspired by qualitative thresholds for cohesion
	// 1 = perfect cohesion; >1 indicates multiple components (worse)
	rangeOfLabels := []string{"= 1", "< 2", "< 4", ">= 4"}
	rangeOfValues := []int32{1, 2, 4, 2147483647}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	for _, file := range c.Files {
		classes := engine.GetClassesInFile(file)
		for _, class := range classes {
			if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.ClassCohesion == nil || class.Stmts.Analyze.ClassCohesion.Lcom4 == nil {
				continue
			}
			mesured := *class.Stmts.Analyze.ClassCohesion.Lcom4

			// map measured to bucket
			// special case for exactly 1
			if mesured == 1 {
				value, _ := data.Get("= 1")
				data.Set("= 1", value+1)
				continue
			}
			for i, r := range rangeOfValues {
				if i == 0 { // skip first (=1) already handled
					continue
				}
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
