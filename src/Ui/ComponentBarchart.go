package Ui

import (
	"github.com/evandro-slv/go-cli-charts/bar"
)

type ComponentBarchart struct {
	data     map[string]float64
	height   int
	barWidth int
}

func (c *ComponentBarchart) AsTerminalElement() string {

	if c.height == 0 {
		c.height = 3
	}

	if c.barWidth == 0 {
		c.barWidth = 8
	}

	data := make(map[string]float64)
	for key, value := range c.data {
		data[key] = float64(value)
	}

	graph := bar.Draw(data, bar.Options{
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
