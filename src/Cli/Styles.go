package Cli

import (
	"math"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	divider   = lipgloss.NewStyle().
			SetString("•").
			Padding(0, 1).
			Foreground(subtle).
			String()

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(1).
			Italic(true).
			Align(lipgloss.Center).
			Width(80).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#5A56E0"))

	subtitleStyle = lipgloss.NewStyle().
			Padding(1).
			Italic(true).
			Width(80).
			Align(lipgloss.Center)

	descStyle = lipgloss.NewStyle().MarginTop(1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

	url = lipgloss.NewStyle().Foreground(special).Render
)

func StyleTitle(text string) lipgloss.Style {
	return titleStyle.SetString(text)
}

func StyleSubTitle(text string) lipgloss.Style {
	return subtitleStyle.SetString(text)
}

func StyleUrl(text string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(special).SetString(text)
}

func StyleHelp(text string) lipgloss.Style {
	return lipgloss.NewStyle().
		SetString(text).
		Italic(true).
		Foreground(lipgloss.Color("#666666"))
}

func DecorateMaintainabilityIndex(mi int) string {
	if mi < 64 {
		return "🔴 " + strconv.Itoa(mi)
	}
	if mi < 85 {
		return "🟡 " + strconv.Itoa(mi)
	}

	return "🟢 " + strconv.Itoa(mi)
}

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}
