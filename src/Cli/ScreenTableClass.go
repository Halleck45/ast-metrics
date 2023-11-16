package Cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ScreenTableClass struct {
	isInteractive     bool
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
	parent            tea.Model
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table           table.Model
	sortColumnIndex int
	parent          tea.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m.parent, tea.ClearScreen
		case "i":
			// sort by maintainability index
			m.sortColumnIndex = 6
			return m, nil
		case "c":
			// sort by cyclomatic complexity
			m.sortColumnIndex = 3
			return m, nil
		case "l":
			// sort by LLOC
			m.sortColumnIndex = 2
			return m, nil
		case "m":
			// sort by number of methods
			m.sortColumnIndex = 1
			return m, nil
		case "n":
			// sort by name
			m.sortColumnIndex = 0
			return m, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {

	// Sort rows by second column (Loc)
	rows := m.table.Rows()
	sort.Slice(rows, func(i, j int) bool {
		if m.sortColumnIndex == 0 {
			return rows[i][m.sortColumnIndex] < rows[j][m.sortColumnIndex]
		}

		if m.sortColumnIndex == 6 {
			// replace emoji
			a := strings.Replace(rows[i][m.sortColumnIndex], "🔴 ", "", 1)
			a = strings.Replace(a, "🟡 ", "", 1)
			a = strings.Replace(a, "🟢 ", "", 1)
			b := strings.Replace(rows[j][m.sortColumnIndex], "🔴 ", "", 1)
			b = strings.Replace(b, "🟡 ", "", 1)
			b = strings.Replace(b, "🟢 ", "", 1)

			aInt, _ := strconv.Atoi(a)
			bInt, _ := strconv.Atoi(b)
			return aInt < bInt
		}

		a, _ := strconv.Atoi(rows[i][m.sortColumnIndex])
		b, _ := strconv.Atoi(rows[j][m.sortColumnIndex])
		return a > b
	})
	m.table.SetRows(rows)

	return StyleScreen(StyleTitle("Classes").Render() + "\n" +
		StyleHelp(`
		Use arrows to navigate and esc to quit.
		Press:
		   - (n) to sort by name
		   - (l) to sort by LLOC
		   - (c) to sort by cyclomatic complexity
		   - (m) to sort by number of methods
		   - (i) to sort by maintainability index

	   `).Render() + "\n\n" +
		baseStyle.Render(m.table.View()) + "\n").Render()
}

func (v ScreenTableClass) GetScreenName() string {
	return "Details for classes"
}

func (v ScreenTableClass) GetModel() tea.Model {

	columns := []table.Column{
		{Title: "Class", Width: 30},
		{Title: "Methods", Width: 10},
		{Title: "LLoc", Width: 10},
		{Title: "Cyclomatic", Width: 15},
		{Title: "Halstead Length", Width: 15},
		{Title: "Halstead Volume", Width: 15},
		{Title: "Maintainability Index", Width: 15},
	}

	rows := []table.Row{}
	for _, file := range v.files {
		for _, class := range file.Stmts.StmtClass {

			if class == nil {
				continue
			}
			if class.Stmts == nil {
				continue
			}
			if class.Stmts.Analyze == nil {
				continue
			}

			rows = append(rows, table.Row{
				class.Name.Qualified,
				strconv.Itoa(len(class.Stmts.StmtFunction)),
				strconv.Itoa(int(*class.Stmts.Analyze.Volume.Loc)),
				strconv.Itoa(int(*class.Stmts.Analyze.Complexity.Cyclomatic)),
				strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadLength)),
				fmt.Sprintf("%.2f", ToFixed(float64(*class.Stmts.Analyze.Volume.HalsteadVolume), 2)),
				DecorateMaintainabilityIndex(int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)),
			})
		}

		for _, namespace := range file.Stmts.StmtNamespace {
			for _, class := range namespace.Stmts.StmtClass {

				if class == nil {
					continue
				}
				if class.Stmts == nil {
					continue
				}

				rows = append(rows, table.Row{
					class.Name.Qualified,
					strconv.Itoa(len(class.Stmts.StmtFunction)),
					strconv.Itoa(int(*class.Stmts.Analyze.Volume.Loc)),
					strconv.Itoa(int(*class.Stmts.Analyze.Complexity.Cyclomatic)),
					strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadLength)),
					fmt.Sprintf("%.2f", ToFixed(float64(*class.Stmts.Analyze.Volume.HalsteadVolume), 2)),
					DecorateMaintainabilityIndex(int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)),
				})
			}
		}
	}

	// sort by name by default
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(25),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{table: t, sortColumnIndex: 0, parent: v.parent}

	return m
}
