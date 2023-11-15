package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Screen interface {
	GetModel() tea.Model
	GetScreenName() string
}
