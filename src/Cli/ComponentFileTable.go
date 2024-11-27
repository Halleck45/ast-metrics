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
	search          ComponentSearch
}

// index of cols
var colsRisks = map[string]int{
	"Name":       0,
	"Risk":       1,
	"Commits":    2,
	"Authors":    3,
	"LOC":        4,
	"Cyclomatic": 5,
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

	var baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Help and handers
	text := StyleHelp(`
		Use arrows to navigate and esc to quit.
		Press following keys to sort by column:
		   (n) by name     (l) by LOC		(r) by risk
		   (c) by cyclomatic complexity (g) by commits
		Press (ctrl+f) to search.

	   `).Render() + "\n"

	// search box
	text = text + v.search.Render()

	// The results
	if len(v.table.Rows()) == 0 {

		if v.search.HasSearch() {
			text += "No file found with the search: " + v.search.Expression + ".\n" + StyleHelp(`
				Press Ctrl+C or Esc to quit.
			`).Render()
		} else {
			text += "No file found.\n" + StyleHelp(`
				Press Ctrl+C or Esc to quit.
			`).Render()
		}
	} else {
		text = text + baseStyle.Render(v.table.View())
	}

	return text
}

func (v *ComponentFileTable) Init() {

	// search
	if &v.search == nil {
		v.search = ComponentSearch{Expression: ""}
	}

	columns := []table.Column{
		{Title: "File", Width: 60},
		{Title: "Risk", Width: 9},
		{Title: "Commits", Width: 9},
		{Title: "Authors", Width: 9},
		{Title: "LOC", Width: 9},
		{Title: "Cyclomatic", Width: 9},
	}

	rows := []table.Row{}
	for _, file := range v.files {

		// Search
		if v.search.HasSearch() {
			// search by file name
			if !v.search.Match(file.Path) {
				continue
			}
		}

		nbCommits := 0
		nbCommiters := 0
		if file.Commits != nil {
			nbCommits = int(file.Commits.Count)
			nbCommiters = int(file.Commits.CountCommiters)
		}

		cyclo := 0
		if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Complexity != nil {
			cyclo = int(*file.Stmts.Analyze.Complexity.Cyclomatic)
		}

		risk := float64(0.0)
		if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Risk != nil {
			risk = float64(file.Stmts.Analyze.Risk.Score)
		}

		// truncate filename, but to the left
		filename := file.Path
		if len(filename) > 60 {
			filename = "..." + filename[len(filename)-57:]
			// remove the extension

		}

		rows = append(rows, table.Row{
			filename,
			strconv.FormatFloat(float64(risk), 'f', 2, 32),
			strconv.Itoa(nbCommits),
			strconv.Itoa(nbCommiters),
			strconv.Itoa(int(file.LinesOfCode.GetLinesOfCode())),
			strconv.Itoa(int(cyclo)),
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

		// for floats
		a, _ := strconv.ParseFloat(rows[i][sortColumnIndex], 32)
		b, _ := strconv.ParseFloat(rows[j][sortColumnIndex], 32)
		if a != b {
			return a > b
		}

		// for integers
		aInt, _ := strconv.Atoi(rows[i][sortColumnIndex])
		bInt, _ := strconv.Atoi(rows[j][sortColumnIndex])
		return aInt > bInt
	})

	v.sortColumnIndex = sortColumnIndex
	v.table.SetRows(rows)
}

func (v *ComponentFileTable) SortByName() {
	v.Sort(colsRisks["Name"])
}

func (v *ComponentFileTable) SortByLoc() {
	v.Sort(colsRisks["LOC"])
}

func (v *ComponentFileTable) SortByCommits() {
	v.Sort(colsRisks["Commits"])
}

func (v *ComponentFileTable) SortByCyclomaticComplexity() {
	v.Sort(colsRisks["Cyclomatic"])
}

func (v *ComponentFileTable) SortByRisk() {
	v.Sort(colsRisks["Risk"])
}

func (v *ComponentFileTable) Update(msg tea.Msg) {

	v.search.Update(msg)
	if v.search.IsEnabled() {
		// Stop here, and make search do its job
		// Apply search
		v.Init()
		v.table, _ = v.table.Update(msg)
		return
	}

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
		case "r":
			v.SortByRisk()
		case "ctrl+f":
			v.search.Toggle("")
		}
	}

	v.table, _ = v.table.Update(msg)
}
