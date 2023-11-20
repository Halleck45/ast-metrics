package Cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ComponentTableClass struct {
	isInteractive   bool
	files           []pb.File
	sortColumnIndex int
	table           table.Model
}

func NewComponentTableClass(isInteractive bool, files []pb.File) *ComponentTableClass {
	v := &ComponentTableClass{
		isInteractive:   isInteractive,
		files:           files,
		sortColumnIndex: 0,
	}

	v.Init()
	return v
}

func (v *ComponentTableClass) Render() string {

	if len(v.table.Rows()) == 0 {
		return "No class found.\n" + StyleHelp(`
			Press q or esc to quit.
		`).Render()
	}

	var baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return StyleHelp(`
		Use arrows to navigate and esc to quit.
		Press following keys to sort by column:
		   (n) by name     (l) by LLOC		(c) by cyclomatic complexity
		   (m) by methods  (i) by maintainability index

	   `).Render() + "\n" +
		baseStyle.Render(v.table.View())
}

func (v *ComponentTableClass) Init() {

	columns := []table.Column{
		{Title: "Class", Width: 30},
		{Title: "Methods", Width: 8},
		{Title: "LLoc", Width: 8},
		{Title: "Cyclomatic", Width: 8},
		{Title: "Halst. Length", Width: 8},
		{Title: "Halst. Volume", Width: 8},
		{Title: "Maint. Index", Width: 8},
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
				DecorateMaintainabilityIndex(int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex), class.Stmts.Analyze),
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
					DecorateMaintainabilityIndex(int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex), class.Stmts.Analyze),
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
		table.WithHeight(15),
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

	v.table = t
}

func (v *ComponentTableClass) Sort(sortColumnIndex int) {
	// Sort rows by second column (Loc)
	rows := v.table.Rows()
	sort.Slice(rows, func(i, j int) bool {
		if sortColumnIndex == 0 {
			return rows[i][sortColumnIndex] < rows[j][sortColumnIndex]
		}

		if sortColumnIndex == 6 {
			// replace emoji
			a := strings.Replace(rows[i][sortColumnIndex], "🔴 ", "", 1)
			a = strings.Replace(a, "🟡 ", "", 1)
			a = strings.Replace(a, "🟢 ", "", 1)
			b := strings.Replace(rows[j][sortColumnIndex], "🔴 ", "", 1)
			b = strings.Replace(b, "🟡 ", "", 1)
			b = strings.Replace(b, "🟢 ", "", 1)

			aInt, _ := strconv.Atoi(a)
			bInt, _ := strconv.Atoi(b)
			return aInt < bInt
		}

		a, _ := strconv.Atoi(rows[i][sortColumnIndex])
		b, _ := strconv.Atoi(rows[j][sortColumnIndex])
		return a > b
	})

	v.sortColumnIndex = 0
	v.table.SetRows(rows)
}

func (v *ComponentTableClass) SortByName() {
	v.Sort(0)
}

func (v *ComponentTableClass) SortByLloc() {
	v.Sort(2)
}

func (v *ComponentTableClass) SortByNumberOfMethods() {
	v.Sort(3)
}

func (v *ComponentTableClass) SortByMaintainabilityIndex() {
	v.Sort(6)
}

func (v *ComponentTableClass) SortByCyclomaticComplexity() {
	v.Sort(3)
}

func (v *ComponentTableClass) Update(msg tea.Msg) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i":
			v.SortByMaintainabilityIndex()
		case "c":
			v.SortByCyclomaticComplexity()
		case "l":
			v.SortByLloc()
		case "m":
			v.SortByNumberOfMethods()
		case "n":
			v.SortByName()
		}
	}

	v.table, _ = v.table.Update(msg)
}
