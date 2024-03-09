package Cli

import (
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ComponentFileTable struct {
	isInteractive   bool
	files           []*pb.File
	sortColumnIndex int
	table           table.Model
}

func NewComponentFileTable(isInteractive bool, files []*pb.File) *ComponentFileTable {
	v := &ComponentFileTable{
		isInteractive:   isInteractive,
		files:           files,
		sortColumnIndex: 0,
	}

	v.Init()
	return v
}

func (v *ComponentFileTable) Render() string {

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
		   (n) by name     (l) by LOC		(c) by cyclomatic complexity (g) by commits

	   `).Render() + "\n" +
		baseStyle.Render(v.table.View())
}

func (v *ComponentFileTable) Init() {

	columns := []table.Column{
		{Title: "File", Width: 75},
		{Title: "Commits", Width: 9},
		{Title: "Authors", Width: 9},
		{Title: "LOC", Width: 9},
		{Title: "Cyclomatic", Width: 9},
	}

	rows := []table.Row{}
	for _, file := range v.files {

		nbCommits := 0
		nbCommiters := 0
		if file.Commits != nil {
			nbCommits = int(file.Commits.CountCommits)
			nbCommiters = int(file.Commits.CountCommitters)
		}

		rows = append(rows, table.Row{
			file.Path,
			strconv.Itoa(nbCommits),
			strconv.Itoa(nbCommiters),
			strconv.Itoa(int(file.LinesOfCode.GetLinesOfCode())),
			strconv.Itoa(int(*file.Stmts.Analyze.Complexity.Cyclomatic)),
		})
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

func (v *ComponentFileTable) Sort(sortColumnIndex int) {
	// Sort rows by second column (Loc)
	rows := v.table.Rows()
	sort.Slice(rows, func(i, j int) bool {
		if sortColumnIndex == 0 {
			return rows[i][sortColumnIndex] < rows[j][sortColumnIndex]
		}

		a, _ := strconv.Atoi(rows[i][sortColumnIndex])
		b, _ := strconv.Atoi(rows[j][sortColumnIndex])
		return a > b
	})

	v.sortColumnIndex = 0
	v.table.SetRows(rows)
}

func (v *ComponentFileTable) SortByName() {
	v.Sort(0)
}

func (v *ComponentFileTable) SortByLoc() {
	v.Sort(2)
}

func (v *ComponentFileTable) SortByCommits() {
	v.Sort(1)
}

func (v *ComponentFileTable) SortByCyclomaticComplexity() {
	v.Sort(3)
}

func (v *ComponentFileTable) Update(msg tea.Msg) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			v.SortByCommits()
		case "l":
			v.SortByLoc()
		case "c":
			v.SortByCyclomaticComplexity()
		case "n":
			v.SortByName()
		}
	}

	v.table, _ = v.table.Update(msg)
}
