package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RepoItem struct {
	Name          string
	FullName      string
	DefaultBranch string
	Selected      bool
}

type modelRepoSelector struct {
	repos    []RepoItem
	cursor   int
	quitting bool
	title    string
	helpText string
}

func (m modelRepoSelector) Init() tea.Cmd {
	return nil
}

func (m modelRepoSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.repos)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			m.repos[m.cursor].Selected = !m.repos[m.cursor].Selected
		case "a":
			// Select all
			for i := range m.repos {
				m.repos[i].Selected = true
			}
		case "d":
			// Deselect all
			for i := range m.repos {
				m.repos[i].Selected = false
			}
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m modelRepoSelector) View() string {
	if m.quitting {
		return ""
	}

	var s string

	// Simple explanatory text
	if m.helpText != "" {
		s += m.helpText + "\n\n"
	}

	// Count selected repos
	selectedCount := 0
	for _, repo := range m.repos {
		if repo.Selected {
			selectedCount++
		}
	}

	s += fmt.Sprintf("Selected: %d/%d repositories\n\n", selectedCount, len(m.repos))

	// Display repos
	for i, repo := range m.repos {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if repo.Selected {
			checkbox = "[x]"
		}

		// Style the current line
		line := fmt.Sprintf("%s %s %s", cursor, checkbox, repo.Name)

		if m.cursor == i {
			line = colorFg(line, "212")
		} else if repo.Selected {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("#43BF6D")).Render(line)
		}

		s += line + "\n"
	}

	s += "\n" + StyleHelp("Use ↑/↓ to navigate, Space to toggle, 'a' to select all, 'd' to deselect all, Enter to confirm, q to quit").Render() + "\n"

	return s
}

func AskUserToSelectRepos(repos []RepoItem, title, helpText string) []RepoItem {
	if len(repos) == 0 {
		return []RepoItem{}
	}

	m := modelRepoSelector{
		repos:    repos,
		cursor:   0,
		quitting: false,
		title:    title,
		helpText: helpText,
	}

	p := tea.NewProgram(&m)
	tm, err := p.Run()
	if err != nil {
		return []RepoItem{}
	}

	if tm == nil {
		return []RepoItem{}
	}

	mm, ok := tm.(modelRepoSelector)
	if !ok {
		return []RepoItem{}
	}

	// Return only selected repos
	selected := []RepoItem{}
	for _, repo := range mm.repos {
		if repo.Selected {
			selected = append(selected, repo)
		}
	}

	return selected
}
