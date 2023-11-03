package Cli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
	sortColumnIndex int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
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
        a, _ := strconv.Atoi(rows[i][m.sortColumnIndex])
        b, _ := strconv.Atoi(rows[j][m.sortColumnIndex])
        return a > b
    })
    m.table.SetRows(rows)

	return baseStyle.Render(m.table.View()) + "\n"
}

func TableForClasses(pbFiles []pb.File) {

    style := StyleTitle()
    fmt.Println(style.Render("Classes"))

     style = lipgloss.NewStyle().
            Italic(true).
            Foreground(lipgloss.Color("#666666")).
            PaddingLeft(0)

     help := `
     Use arrows to navigate and esc to quit.
     Press:
        - (n) to sort by name
        - (l) to sort by LLOC
        - (c) to sort by cyclomatic complexity
        - (m) to sort by number of methods
        - (i) to sort by maintainability index
    `
    fmt.Println(style.Render(help))


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
	for _, file := range pbFiles {
	    for _,class := range file.Stmts.StmtClass {

	        if class == nil {
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
                strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadVolume)),
                strconv.Itoa(int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)),
            })
        }

        for _,namespace := range file.Stmts.StmtNamespace {
            for _,class := range namespace.Stmts.StmtClass {

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
                        strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadVolume)),
                        strconv.Itoa(int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)),
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

	m := model{table: t, sortColumnIndex: 0}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}