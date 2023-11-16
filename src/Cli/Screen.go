package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Screen interface {

	// Returns the Tea model used for the screen
	GetModel() tea.Model

	// Returns the name of the screen
	GetScreenName() string
}
