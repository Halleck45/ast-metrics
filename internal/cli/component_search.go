package cli

import (
	"fmt"
	"regexp"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ComponentSearch struct {
	Expression string
	input      *textinput.Model
}

func NewComponentSearch(expression string) *ComponentSearch {
	return &ComponentSearch{
		Expression: expression,
	}
}

func (v *ComponentSearch) Match(value string) bool {
	res1, e := regexp.MatchString("(?i)"+v.Expression, value)
	return e == nil && res1
}

func (v *ComponentSearch) Update(msg tea.Msg) {
	if v.input == nil {
		return
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			v.input.Blur()
			v.input = nil
			v.Expression = ""
			return
		case tea.KeyEnter:
			v.input.Blur()
			v.input = nil
			return
		}
	}

	// Issue with the input. We temporary update keys manually
	// v.input.Update(msg)
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()

		if k == "backspace" {
			if len(v.input.Value()) == 0 {
				v.Toggle("")
				return
			}
			v.input.SetValue(v.input.Value()[:len(v.input.Value())-1])
			v.Expression = v.input.Value()
			return
		}

		if k == "space" {
			v.input.SetValue(v.input.Value() + " ")
			return
		}

		// if not a alpha numeric char, return
		if !regexp.MustCompile(`^[a-zA-Z0-9]{1}$`).MatchString(k) {
			return
		}

		v.input.SetValue(fmt.Sprintf("%s%s", v.input.Value(), k))
		v.input.SetCursor(len(v.input.Value()))

		v.Expression = v.input.Value()
		return
	}
}

func (v *ComponentSearch) IsEnabled() bool {
	return v.input != nil
}

func (v *ComponentSearch) HasSearch() bool {
	return v.Expression != ""
}

func (v *ComponentSearch) Toggle(expression string) {

	if v.input != nil {
		v.input = nil
		return
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	v.input = &ti
	v.Expression = expression

	// Focus
	v.input.Focus()
}

func (v *ComponentSearch) Render() string {

	if v.Expression != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("🔍 " + v.Expression + "\n")
	}

	if v.input == nil {
		return ""
	}

	return v.input.View()
}
