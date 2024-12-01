package Cli

import (
	"fmt"
	"math"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	osterm "golang.org/x/term"
)

var (
	subtle  = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	titleStyle = lipgloss.NewStyle().
			Padding(1).
			Italic(true).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#5A56E0"))

	subtitleStyle = lipgloss.NewStyle().
			Padding(1).
			Italic(true).
			Align(lipgloss.Center)
)

func windowWidth() int {
	width, _, _ := osterm.GetSize(0)
	return min(max(width-1, 120), 120)
}

func StyleTitle(text string) lipgloss.Style {
	return titleStyle.Width(windowWidth()).SetString(text)
}

func StyleHowToQuit(text string) lipgloss.Style {
	if text == "" {
		text = "Press 'Ctrl + C' to quit"
	}
	return lipgloss.NewStyle().
		Width(windowWidth()).
		SetString(text).
		MarginTop(2).Align(lipgloss.Center)
}

func StyleSubTitle(text string) lipgloss.Style {
	return subtitleStyle.Width(windowWidth()).SetString(text)
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

func StyleChoices(text string) lipgloss.Style {
	return lipgloss.NewStyle().
		SetString(text).
		MarginLeft(4)
}

func StyleCommand(text string) lipgloss.Style {
	return lipgloss.NewStyle().
		SetString(text).
		Foreground(lipgloss.Color("#666666"))
}
func StyleScreen(text string) lipgloss.Style {
	return lipgloss.NewStyle().
		SetString(text).
		Width(windowWidth()).
		MarginLeft(1).
		MarginTop(1)
}

func StyleNumberBox(number string, label string, sublabel string) lipgloss.Style {
	return lipgloss.NewStyle().
		PaddingTop(1).
		PaddingBottom(1).
		MarginRight(2).
		Border(lipgloss.RoundedBorder(), true).
		Background(lipgloss.Color("#2563eb")).
		Foreground(lipgloss.Color("#FFFFFF")).
		BorderForeground(lipgloss.Color("#1e3a8a")).
		Width(35).
		Height(13).
		Align(lipgloss.Center).
		SetString(fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Bold(true).SetString(number).Render(),
			lipgloss.NewStyle().Bold(false).SetString(label).Render(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Italic(true).SetString(sublabel).Render(),
		))

}

func DecorateMaintainabilityIndex(mi int, analyze *pb.Analyze) string {

	min := int32(1)

	if analyze != nil && analyze.Volume != nil && analyze.Volume.Lloc != nil && *analyze.Volume.Lloc < min {
		return "-"
	}

	if mi < 64 {
		return "🔴 " + strconv.Itoa(mi)
	}
	if mi < 85 {
		return "🟡 " + strconv.Itoa(mi)
	}

	return "🟢 " + strconv.Itoa(mi)
}

func Round(num float64) int {
	return int(num + float64(math.Copysign(0.5, float64(num))))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*float64(output))) / float64(output)
}
