package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evandro-slv/go-cli-charts/bar"
)

// ComponentBarchart is the barchart component
type ComponentBarchart struct {
	data     map[string]float64
	height   int
	barWidth int
}

// NewComponentBarchart is the constructor for the ComponentBarchart
func NewComponentBarchart(data map[string]float64) *ComponentBarchart {
	return &ComponentBarchart{
		data: data,
	}
}

// Render is the method to render the component
func (c *ComponentBarchart) Render() string {

	if c.height == 0 {
		c.height = 3
	}

	if c.barWidth == 0 {
		c.barWidth = 8
	}

	graph := bar.Draw(c.data, bar.Options{
		Chart: bar.Chart{
			Height: c.height,
		},
		Bars: bar.Bars{
			Width: c.barWidth,
			Margin: bar.Margin{
				Left:  1,
				Right: 1,
			},
		},
		UI: bar.UI{
			YLabel: bar.YLabel{
				Hide: true,
			},
		},
	})
	return graph
}

// Update is the method to update the component
func (c *ComponentBarchart) Update(msg tea.Msg) {
}
