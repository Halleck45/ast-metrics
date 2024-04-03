package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

// ComponentBarchartCyclomaticByMethodRepartition is the barchart component for the loc repartition
type ComponentBarchartCyclomaticByMethodRepartition struct {
	aggregated Analyzer.Aggregated
	files      []*pb.File
}

// NewComponentBarchartCyclomaticByMethodRepartition is the constructor for the ComponentBarchartCyclomaticByMethodRepartition
func NewComponentBarchartCyclomaticByMethodRepartition(aggregated Analyzer.Aggregated, files []*pb.File) *ComponentBarchartCyclomaticByMethodRepartition {
	return &ComponentBarchartCyclomaticByMethodRepartition{
		aggregated: aggregated,
		files:      files,
	}
}

// Render is the method to render the component
func (c *ComponentBarchartCyclomaticByMethodRepartition) Render() string {

	dataOrdered := c.GetData()
	data := make(map[string]float64)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}

	graph := NewComponentBarchart(data)
	graph.height = 5
	graph.barWidth = 8
	return graph.Render()
}

func (c *ComponentBarchartCyclomaticByMethodRepartition) GetData() *orderedmap.OrderedMap[string, float64] {
	data := orderedmap.NewOrderedMap[string, float64]()

	rangeOfLabels := []string{"0-5", "5-20", "> 20"}
	rangeOfValues := []int32{5, 20, 999999}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	// repartition of classes by cyclomatic complexity
	for _, file := range c.files {
		classes := Engine.GetClassesInFile(file)
		for _, class := range classes {
			if class.Stmts.Analyze == nil {
				continue
			}
			mesured := *class.Stmts.Analyze.Complexity.Cyclomatic
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
func (c *ComponentBarchartCyclomaticByMethodRepartition) RenderHTML() string {
	data := c.GetData()
	return Engine.HtmlChartLine(data, "Number of files", "chart-loc")
}

// Update is the method to update the component
func (c *ComponentBarchartCyclomaticByMethodRepartition) Update(msg tea.Msg) {
}
