package Cli

import (
	"github.com/charmbracelet/lipgloss"

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