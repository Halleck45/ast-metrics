package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ScreenTableClass struct {
	isInteractive     bool
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func NewScreenTableClass(isInteractive bool, files []pb.File, projectAggregated Analyzer.ProjectAggregated) ScreenTableClass {
	return ScreenTableClass{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
	}
}

type model struct {
	table             *ComponentTableClass
	sortColumnIndex   int
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return NewScreenHome(true, m.files, m.projectAggregated).GetModel(), tea.ClearScreen
		}
	}

	m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return StyleScreen(StyleTitle("Classes").Render() + "\n" +
		"\n" + m.table.Render()).Render()
}

func (v ScreenTableClass) GetScreenName() string {
	return "Classes and object oriented statistics"
}

func (v ScreenTableClass) GetModel() tea.Model {
	table := NewComponentTableClass(v.isInteractive, v.files)
	m := model{table: table, sortColumnIndex: 0, files: v.files, projectAggregated: v.projectAggregated}
	return m
}
