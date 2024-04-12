package Cli

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ScreenSummary struct {
	isInteractive     bool
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func NewScreenSummary(isInteractive bool, files []*pb.File, projectAggregated Analyzer.ProjectAggregated) ScreenSummary {
	return ScreenSummary{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
	}
}

func (v ScreenSummary) GetScreenName() string {
	return "Overview"
}

func (v ScreenSummary) GetModel() tea.Model {
	m := modelScreenSummary{files: v.files, projectAggregated: v.projectAggregated}
	return m
}

type modelScreenSummary struct {
	parent            tea.Model
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func (m modelScreenSummary) Init() tea.Cmd {
	return nil
}

func (m *ScreenSummary) Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) {
	m.files = files
	m.projectAggregated = projectAggregated
}

func (m modelScreenSummary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return NewScreenHome(true, m.files, m.projectAggregated).GetModel(), tea.ClearScreen
		}
	case DoRefreshModel:
		// refresh the model
		m.files = msg.files
		m.projectAggregated = msg.projectAggregated
		return m, nil
	}
	return m, nil
}

func (m modelScreenSummary) View() string {

	// for the moment we aggregate by class only
	// @todo
	aggregatedByClass := m.projectAggregated.ByClass
	combined := m.projectAggregated.Combined

	// Header (statistics overview)
	row1 := NewComponentStatisticsOverview(m.files, m.projectAggregated.Combined).Render()

	in := `## Complexity

   *Cyclomatic Complexity is a measure of the number of linearly independent paths through a program's source code.
   More you have paths, more your code is complex.*

   | Min | Max | Average per class | Average per method | 
   | --- | --- | --- | --- |
   | ` +
		strconv.Itoa(combined.MinCyclomaticComplexity) +
		` | ` + strconv.Itoa(combined.MaxCyclomaticComplexity) +
		` | ` + fmt.Sprintf("%.2f", combined.AverageCyclomaticComplexityPerClass) +
		` | ` + fmt.Sprintf("%.2f", combined.AverageCyclomaticComplexityPerMethod) +
		` |

   ### Classes and methods

   | Classes | Methods | Average methods per class | Average LOC per method |
   | --- | --- | --- | --- |` + "\n" +
		` | ` + strconv.Itoa(aggregatedByClass.NbClasses) +
		` | ` + strconv.Itoa(combined.NbMethods) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageMethodsPerClass) +
		` | ` + fmt.Sprintf("%.2f", combined.AverageLocPerMethod) +
		` |

   ## Maintainability

   *Maintainability Index is a software metric which measures how maintainable (easy to support and change) the source code is.
   If you have a high MI (>85), your code is easy to maintain.*

   | Maintainability index | MI without comments | Comment weight |
   | --- | --- | --- |
   | ` + DecorateMaintainabilityIndex(int(aggregatedByClass.AverageMI), nil) + ` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageMIwoc) + ` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageMIcw) + ` |
   `
	out, _ := glamour.Render(in, "dark")

	// tempporary disabled
	// out = ""

	return StyleScreen(StyleTitle("Results overview").Render() +
		"\n" + row1 +
		StyleHowToQuit("").Render() +
		"\n" + out).Render()
}
