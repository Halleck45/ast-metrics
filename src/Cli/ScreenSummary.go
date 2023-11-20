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
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
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
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func (m modelScreenSummary) Init() tea.Cmd {
	return nil
}

func (m modelScreenSummary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return NewScreenHome(true, m.files, m.projectAggregated).GetModel(), tea.ClearScreen
		}
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

   ### Cyclomatic complexity

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

   ### Halstead metrics

   *Halstead metrics are software metrics introduced to empirically determine the complexity of a program.*

   | | Difficulty | Effort | Volume | Time |
   | --- | --- | --- | --- | --- |
    ` +
		` | Total` +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.SumHalsteadDifficulty) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.SumHalsteadEffort) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.SumHalsteadVolume) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.SumHalsteadTime) +
		"\n | Average per class" +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageHalsteadDifficulty) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageHalsteadEffort) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageHalsteadVolume) +
		` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageHalsteadTime) +
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
   | ` + DecorateMaintainabilityIndex(int(aggregatedByClass.AverageMI)) + ` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageMIwoc) + ` | ` + fmt.Sprintf("%.2f", aggregatedByClass.AverageMIcw) + ` |
   `
	out, _ := glamour.Render(in, "dark")

	return StyleScreen(StyleTitle("Results overview").Render() + "\n" + row1 + "\n" + out).Render()
}
