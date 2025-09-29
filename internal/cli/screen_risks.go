package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

type ScreenRisks struct {
	isInteractive     bool
	files             []*pb.File
	projectAggregated analyzer.ProjectAggregated
}

func NewScreenRisks(isInteractive bool, files []*pb.File, projectAggregated analyzer.ProjectAggregated) ScreenRisks {
	return ScreenRisks{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
	}
}

type modelRisks struct {
	table             *ComponentFileTable
	sortColumnIndex   int
	files             []*pb.File
	projectAggregated analyzer.ProjectAggregated
}

func (m modelRisks) Init() tea.Cmd { return nil }

func (m modelRisks) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return NewScreenHome(true, m.files, m.projectAggregated).GetModel(), tea.ClearScreen
		}
	case DoRefreshModel:
		// refresh the modelRisks
		m.files = msg.files
		m.projectAggregated = msg.projectAggregated
	}

	m.table.Update(msg)
	return m, cmd
}

func (m modelRisks) View() string {
	return StyleScreen(StyleTitle("Top candidates for refactoring").Render() + "\n" +
		"\n" + m.table.Render()).Render()
}

func (v ScreenRisks) GetScreenName() string {
	return "Top candidates for refactoring"
}

func (v *ScreenRisks) Reset(files []*pb.File, projectAggregated analyzer.ProjectAggregated) {
	v.files = files
	v.projectAggregated = projectAggregated
}

func (v ScreenRisks) GetModel() tea.Model {
	table := NewComponentFileTable(v.isInteractive, v.files)
	table.SortByRisk()
	m := modelRisks{table: table, sortColumnIndex: 0, files: v.files, projectAggregated: v.projectAggregated}
	return m
}
