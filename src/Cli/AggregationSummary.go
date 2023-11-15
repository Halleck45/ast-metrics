package Cli

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type AggregationSummary struct {
	isInteractive     bool
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
	parent            tea.Model
}

func (v AggregationSummary) GetScreenName() string {
	return "Overview"
}

func (v AggregationSummary) GetModel() tea.Model {
	m := modelAggregationSummary{parent: v.parent, files: v.files, projectAggregated: v.projectAggregated}
	return m
}

type modelAggregationSummary struct {
	parent            tea.Model
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
}

func (m modelAggregationSummary) Init() tea.Cmd { return nil }

func (m modelAggregationSummary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m.parent, tea.ClearScreen
		}
	}
	return m, nil
}

func (m modelAggregationSummary) View() string {

	// for the moment we aggregate by class only
	// @todo
	aggregatedByClass := m.projectAggregated.ByClass
	//aggregatedByFile := projectAggregated.ByFile
	combined := m.projectAggregated.Combined

	var percentageCloc int = 0
	var percentageLloc int = 0
	if combined.Loc > 0 {
		percentageCloc = 100 * combined.Cloc / combined.Loc
		percentageLloc = 100 * combined.Lloc / combined.Loc
	}

	in := `*This code is composed from ` +
		strconv.Itoa(combined.NbFiles) + ` files, ` +
		strconv.Itoa(combined.Loc) + ` lines of code, ` +
		strconv.Itoa(combined.Cloc) + ` (` + (strconv.Itoa(percentageCloc)) + `%) comment lines of code and ` +
		strconv.Itoa(combined.Lloc) + ` (` + (strconv.Itoa(percentageLloc)) + `%) logical lines of code.*

   ## Complexity

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

	return StyleTitle("Results overview").Render() + "\n" +
		out
}
