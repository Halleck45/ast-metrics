package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ScreenByProgrammingLanguage struct {
	isInteractive                bool
	programmingLangageName       string
	programmingLangageAggregated Analyzer.Aggregated
	files                        []pb.File
	projectAggregated            Analyzer.ProjectAggregated
}

type modelByProgrammingLanguage struct {
	programmingLangageName string
	componentTableClass    *ComponentTableClass
	files                  []pb.File
	projectAggregated      Analyzer.ProjectAggregated
}

func (m modelByProgrammingLanguage) Init() tea.Cmd {
	return nil
}

func (m modelByProgrammingLanguage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return NewScreenHome(true, m.files, m.projectAggregated).GetModel(), tea.ClearScreen
		}
	}

	m.componentTableClass.Update(msg)

	return m, cmd
}

func (m modelByProgrammingLanguage) View() string {

	// 1. Header
	header := NewComponentStatisticsOverview(m.files, m.projectAggregated.ByProgrammingLanguage[m.programmingLangageName])

	// 2. Table
	return StyleScreen(StyleTitle(m.programmingLangageName+" overview").Render() +
		header.Render() +
		"\n\n" + m.componentTableClass.Render()).Render()
}

func (v ScreenByProgrammingLanguage) GetScreenName() string {
	return v.programmingLangageName + " overview"
}

func (v ScreenByProgrammingLanguage) GetModel() tea.Model {

	// table of classes, but only for the programming language
	files := []pb.File{}
	for _, file := range v.files {
		if file.ProgrammingLanguage == v.programmingLangageName {
			files = append(files, file)
		}
	}
	table := NewComponentTableClass(v.isInteractive, files)

	m := modelByProgrammingLanguage{
		programmingLangageName: v.programmingLangageName,
		files:                  v.files,
		projectAggregated:      v.projectAggregated,
		componentTableClass:    table,
	}

	return m
}
