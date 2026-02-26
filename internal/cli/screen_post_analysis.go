package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/halleck45/ast-metrics/internal/report"
)

// PostAnalysisChoice represents the user's selection after analysis.
type PostAnalysisChoice int

const (
	PostAnalysisOpenHTML PostAnalysisChoice = iota
	PostAnalysisExplore
	PostAnalysisQuit
)

type modelPostAnalysis struct {
	files             []*pb.File
	projectAggregated analyzer.ProjectAggregated
	cursor            int
	choice            PostAnalysisChoice
	quitting          bool
}

func (m modelPostAnalysis) Init() tea.Cmd {
	return nil
}

func (m modelPostAnalysis) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.choice = PostAnalysisQuit
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				m.choice = PostAnalysisOpenHTML
			} else {
				m.choice = PostAnalysisExplore
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m modelPostAnalysis) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#73F59F")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(ScreenHeader("Analysis complete"))
	b.WriteString(fmt.Sprintf("\n  All %d files have been analyzed. What do you want to do?\n\n", len(m.files)))

	items := []struct {
		label string
		desc  string
	}{
		{"Open HTML report in browser", "Generate and open a detailed report"},
		{"Explore in terminal", "Browse results interactively"},
	}

	for i, item := range items {
		if i == m.cursor {
			b.WriteString(checkStyle.Render("  [x] "+item.label) + "  " + descStyle.Render(item.desc) + "\n")
		} else {
			b.WriteString(normalStyle.Render("  [ ] "+item.label) + "  " + descStyle.Render(item.desc) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(StyleHelp("  Use arrows to navigate, enter to select, q to quit.").Render())
	b.WriteString("\n")

	return b.String()
}

// AskPostAnalysis shows the post-analysis choice screen.
// Returns the user's choice.
func AskPostAnalysis(files []*pb.File, projectAggregated analyzer.ProjectAggregated) PostAnalysisChoice {
	m := modelPostAnalysis{
		files:             files,
		projectAggregated: projectAggregated,
		choice:            PostAnalysisQuit,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return PostAnalysisQuit
	}

	mm, ok := result.(modelPostAnalysis)
	if !ok {
		return PostAnalysisQuit
	}

	return mm.choice
}

// GenerateAndOpenHTMLReport generates the HTML report, opens it in the browser,
// and shows a confirmation screen that waits for a keypress before returning.
func GenerateAndOpenHTMLReport(files []*pb.File, projectAggregated analyzer.ProjectAggregated) {
	directory := "ast-metrics-report"

	htmlReportGenerator := report.NewHtmlReportGenerator(directory)
	_, err := htmlReportGenerator.Generate(files, projectAggregated)

	// Clear screen before showing the confirmation
	fmt.Print("\033[H\033[2J")

	fmt.Print(ScreenHeader("HTML Report"))
	fmt.Println()

	if err != nil {
		PrintError("Error generating report: " + err.Error())
		fmt.Println()
		PressAnyKeyToContinue()
		return
	}

	htmlPath := filepath.Join(directory, "index.html")
	openErr := report.OpenHtmlReport(htmlPath)

	if openErr != nil {
		PrintWarning("Could not open the browser automatically.")
		fmt.Println()
		PrintInfo(report.GetOpenInstructions(htmlPath))
	} else {
		absPath, _ := filepath.Abs(htmlPath)
		PrintSuccess("Report generated and opened in your browser.")
		if absPath != "" {
			fmt.Println()
			PrintInfo(absPath)
		}
	}

	fmt.Println()
	PressAnyKeyToContinue()
}
