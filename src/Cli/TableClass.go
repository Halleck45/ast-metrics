package Cli

import (
	"fmt"
	"os"
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
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func TableForClasses(pbFiles []pb.File) {
	columns := []table.Column{
		{Title: "Class", Width: 20},
		{Title: "LLoc", Width: 10},
		{Title: "Cyclomatic", Width: 15},
		{Title: "Halstead Length", Width: 15},
		{Title: "Halstead Volume", Width: 15},
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
                strconv.Itoa(int(*class.Stmts.Analyze.Volume.Loc)),
                strconv.Itoa(int(*class.Stmts.Analyze.Complexity.Cyclomatic)),
                strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadLength)),
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
                        strconv.Itoa(int(*class.Stmts.Analyze.Volume.Loc)),
                        strconv.Itoa(int(*class.Stmts.Analyze.Complexity.Cyclomatic)),
                        strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadLength)),
                        strconv.Itoa(int(*class.Stmts.Analyze.Volume.HalsteadVolume)),
                    })
                }
            }
	}

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

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}