package ui

import (
	"time"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

// ComponentLineChartGitActivity is the barchart component for the loc repartition
type ComponentLineChartGitActivity struct {
	Aggregated analyzer.Aggregated
	Files      []*pb.File
}

// Render is the method to render the component
func (c *ComponentLineChartGitActivity) AsTerminalElement() string {
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
func (c *ComponentLineChartGitActivity) AsHtml() string {
	data := c.GetData()
	return engine.HtmlChartArea(data, "Number of commits", "chart-git")
}

func (c *ComponentLineChartGitActivity) GetData() *orderedmap.OrderedMap[string, float64] {
	//data := make(map[string]float64)*
	data := orderedmap.NewOrderedMap[string, float64]()

	// 1 year ago
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	// generate 12 months of labels
	for i := 1; i < 12; i++ {
		month := oneYearAgo.AddDate(0, i, 0)
		data.Set(month.Format("Jan"), 0)
	}
	// add current month
	data.Set(time.Now().Format("Jan"), 0)

	// count the number of files per month
	for _, file := range c.Files {
		if file.Commits == nil {
			continue
		}

		for _, commit := range file.Commits.Commits {
			// timestamp to date
			commitTime := time.Unix(commit.Date, 0)
			month := commitTime.Format("Jan")
			if value, ok := data.Get(month); ok {
				data.Set(month, value+1)
			} else {
				data.Set(month, 1)
			}
		}
	}

	return data
}
