package Cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/muesli/termenv"
)

// General stuff for styling the view
var (
	term = termenv.EnvColorProfile()
)

// ScreenHome is the home view
type ScreenHome struct {
	isInteractive bool
}

// modelChoices is the model for the home view
type modelChoices struct {
	files             []pb.File
	projectAggregated Analyzer.ProjectAggregated
	Choice            int

	// array of screens
	screens []Screen
}

// NewScreenHome creates a new ScreenHome
func NewScreenHome(isInteractive bool) *ScreenHome {
	return &ScreenHome{
		isInteractive: isInteractive,
	}
}

// Render renders the home view and runs the Tea program
func (r ScreenHome) Render(pbFiles []pb.File, aggregated Analyzer.ProjectAggregated) {

	// Prepare list of accepted screens
	m := modelChoices{files: pbFiles, projectAggregated: aggregated}
	fillInScreens(&m)

	if !r.isInteractive {
		// If not interactive, just display the first screen
		fmt.Println(m.screens[0].GetModel().View())
		return
	}

	options := tea.WithAltScreen()
	if _, err := tea.NewProgram(m, options).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// fillInScreens fills in the screens array, and is used to avoid a circular dependency when creating the screens and coming back to the main screen
func fillInScreens(modelChoices *modelChoices) {

	// Create the table screen
	viewTableClass := ScreenTableClass{isInteractive: true}
	viewTableClass.files = modelChoices.files
	viewTableClass.projectAggregated = modelChoices.projectAggregated
	viewTableClass.parent = modelChoices

	// Create the table screen
	summaryScreen := ScreenSummary{isInteractive: true}
	summaryScreen.files = modelChoices.files
	summaryScreen.projectAggregated = modelChoices.projectAggregated
	viewTableClass.parent = modelChoices

	// Create the screen list
	modelChoices.screens = []Screen{
		summaryScreen,
		viewTableClass,
	}
}

// Init initializes the Tea model
func (m modelChoices) Init() tea.Cmd {
	return nil
}

// The main view, which just calls the appropriate sub-view
func (m modelChoices) View() string {
	c := m.Choice

	tpl := StyleTitle("AST Metrics").Render() +
		"\n" + StyleSubTitle("AST Metrics is a language-agnostic static code analyzer. "+StyleUrl("https://github.com/Halleck45/ast-metrics").Render()).Render() +
		fmt.Sprintf("\n\nAll %d files have been analyzed. What do you want to do next?\n\n", len(m.files))

	choices := StyleHelp("Use arrows to navigate and esc to quit.").Render() + "\n\n"
	for i, s := range m.screens {
		label := s.GetScreenName()
		if i == c {
			choices += colorFg("[x] "+label, "212")
		} else {
			choices += fmt.Sprintf("[ ] %s", label)
		}

		choices += "\n"
	}
	tpl += StyleChoices(choices).Render()

	tpl += "\n\nAST Metrics is an Open Source project. Contributions are welcome!\nDo not hesitate to open issue at " + StyleUrl("https://github.com/Halleck45/ast-metrics/issues").Render() + ". ❤️  Thanks!\n"

	return StyleScreen(tpl).Render()
}

// Main update function.
func (m modelChoices) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Choice++
			if m.Choice > (len(m.screens) - 1) {
				m.Choice = len(m.screens) - 1
			}
		case "k", "up":
			m.Choice--
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "enter":
			index := m.Choice
			if len(m.screens) == 0 {
				// issue when navigating back to the main screen
				fillInScreens(&m)
			}
			return m.screens[index].GetModel(), tea.ClearScreen
		}
	}

	return m, nil
}

// Color a string's foreground with the given value.
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}
