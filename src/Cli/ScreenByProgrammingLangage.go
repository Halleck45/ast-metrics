package Cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ScreenByProgrammingLanguage struct {
	isInteractive                bool
	programmingLangageName       string
	programmingLangageAggregated Analyzer.Aggregated
	files                        []*pb.File
	projectAggregated            Analyzer.ProjectAggregated
}

type modelByProgrammingLanguage struct {
	programmingLangageName string
	componentTableClass    Component
	files                  []*pb.File
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
	case DoRefreshModel:
		// refresh the model
		m.files = msg.files
		m.projectAggregated = msg.projectAggregated
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

func (v *ScreenByProgrammingLanguage) Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) {
	v.files = files
	v.projectAggregated = projectAggregated
}

func (v ScreenByProgrammingLanguage) GetScreenName() string {
	// @todo use dynamic emoji
	emoji := "  "
	switch v.programmingLangageName {
	case "PHP":
		emoji = "🐘 "
	case "Python":
		emoji = "🐍 "
	case "Golang":
		emoji = "🐹 "
	}

	count := 0
	for _, file := range v.files {
		if file.ProgrammingLanguage == v.programmingLangageName {
			count++
		}
	}

	return fmt.Sprintf("%s%s (%d files)", emoji, v.programmingLangageName, count)
}

func (v ScreenByProgrammingLanguage) GetModel() tea.Model {

	// table of classes, but only for the programming language
	files := []*pb.File{}
	for _, file := range v.files {
		if file.ProgrammingLanguage == v.programmingLangageName {
			files = append(files, file)
		}
	}

	// for no OOP language, we display the file table
	// @todo: make it dynamic
	var table Component
	if v.programmingLangageName == "Golang" {
		table = NewComponentFileTable(v.isInteractive, files)
	} else {
		table = NewComponentTableClass(v.isInteractive, files)
	}

	m := modelByProgrammingLanguage{
		programmingLangageName: v.programmingLangageName,
		files:                  v.files,
		projectAggregated:      v.projectAggregated,
		componentTableClass:    table,
	}

	return m
}
