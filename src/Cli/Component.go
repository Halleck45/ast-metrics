package Cli

import tea "github.com/charmbracelet/bubbletea"

type Component interface {

	// Returns the content of the component
	Render() string

	// Updates attached model
	Update(msg tea.Msg)
}
