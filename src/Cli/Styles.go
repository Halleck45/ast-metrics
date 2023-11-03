package Cli

import (
	"github.com/charmbracelet/lipgloss"
	"strconv"
)
func StyleTitle() lipgloss.Style {
    return lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("#FAFAFA")).
            Background(lipgloss.Color("#7D56F4")).
            Underline(true).
            PaddingTop(1).
            PaddingBottom(1).
            MarginTop(2).
            Align(lipgloss.Center).
            Width(80)
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