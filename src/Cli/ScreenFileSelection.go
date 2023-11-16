package Cli

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type modelFileSelection struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m modelFileSelection) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m modelFileSelection) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m modelFileSelection) View() string {
	if m.quitting {
		return ""
	}

	out := StyleTitle("AST Metrics").Render() +
		StyleHelp("\n\nYou did not provide any path to analyze. Please note that you can pass arguments to the command when you use AST Metrics."+"\n"+
			"Example: "+StyleCommand("ast-metrics analyze ./path/to/analyze1 ./path/to/analyze2\n").Render()+"\n\n").Render()

	if m.err != nil {
		out += m.filepicker.Styles.DisabledFile.Render(m.err.Error())
	} else if m.selectedFile == "" {
		out += "Pick a file or a directory to analyze, then press <esc> to continue:"
	} else {
		out += m.filepicker.Styles.Selected.Render("Selected file: " + m.selectedFile)
	}

	out += "\n\n" + m.filepicker.View() + "\n"

	out += "\n\n" + StyleHelp("Use arrows to navigate, enter to select, and esc to continue.").Render() + "\n\n"
	return StyleScreen(out).Render()
}

func AskUserToSelectFile() []string {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.CurrentDirectory = "."
	fp.ShowHidden = false

	m := modelFileSelection{
		filepicker: fp,
	}

	options := tea.WithAltScreen()
	tm, _ := tea.NewProgram(&m, options).Run()
	mm := tm.(modelFileSelection)
	return []string{mm.selectedFile}
}
