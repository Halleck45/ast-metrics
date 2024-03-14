package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

// ComponentBarchartLocByMethodRepartition is the barchart component for the loc repartition
type ComponentBarchartLocByMethodRepartition struct {
	aggregated Analyzer.Aggregated
	files      []*pb.File
}

// NewComponentBarchartLocByMethodRepartition is the constructor for the ComponentBarchartLocByMethodRepartition
func NewComponentBarchartLocByMethodRepartition(aggregated Analyzer.Aggregated, files []*pb.File) *ComponentBarchartLocByMethodRepartition {
	return &ComponentBarchartLocByMethodRepartition{
		aggregated: aggregated,
		files:      files,
	}
}

// Render is the method to render the component
func (c *ComponentBarchartLocByMethodRepartition) Render() string {
	dataOrdered := c.GetData()
	data := make(map[string]float64)
	for _, k := range dataOrdered.Keys() {
		value, _ := dataOrdered.Get(k)
		data[k] = value
	}
	graph := NewComponentBarchart(data)
	graph.height = 5
	graph.barWidth = 6
	return graph.Render()
}

// Render Html
func (c *ComponentBarchartLocByMethodRepartition) RenderHTML() string {
	data := c.GetData()
	return Engine.HtmlChartLine(data, "Number of files", "chart-loc-by-method")
}

func (c *ComponentBarchartLocByMethodRepartition) GetData() *orderedmap.OrderedMap[string, float64] {
	data := orderedmap.NewOrderedMap[string, float64]()

	rangeOfLabels := []string{"< 15", "< 35", "< 50", "> 50"}
	rangeOfValues := []int32{15, 35, 50, 999999}
	for _, r := range rangeOfLabels {
		data.Set(r, 0)
	}

	// repartition of files by LOC
	for _, file := range c.files {
		functions := Engine.GetFunctionsInFile(file)
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

// Update is the method to update the component
func (c *ComponentBarchartLocByMethodRepartition) Update(msg tea.Msg) {
}
