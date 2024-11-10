package Cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
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
	isInteractive     bool
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
	// program
	tea *tea.Program
	// modelChoices
	modelChoices modelChoices
	// watchers
	FileWatcher  *fsnotify.Watcher
	currentModel tea.Model
}

// modelChoices is the model for the home view
type modelChoices struct {
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
	Choice            int

	// array of screens
	screens []Screen

	// Watcher
	FileWatcher *fsnotify.Watcher
}

type DoRefreshModel struct {
	files             []*pb.File
	projectAggregated Analyzer.ProjectAggregated
}

// NewScreenHome creates a new ScreenHome
func NewScreenHome(isInteractive bool, files []*pb.File, projectAggregated Analyzer.ProjectAggregated) *ScreenHome {
	return &ScreenHome{
		isInteractive:     isInteractive,
		files:             files,
		projectAggregated: projectAggregated,
	}
}

// Render renders the home view and runs the Tea program
func (r *ScreenHome) Render() {

	if r.tea != nil {
		// If already running, just return
		// send an update msg
		//r.tea.Send(DoRefreshModel{files: r.files, projectAggregated: r.projectAggregated})
		return
	}

	// Prepare list of accepted screens
	m := modelChoices{files: r.files, projectAggregated: r.projectAggregated, FileWatcher: r.FileWatcher}
	fillInScreens(&m)
	r.currentModel = m

	if !r.isInteractive {
		// If not interactive
		var style = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
		fmt.Println(style.Render("No interactive mode detected."))
		return
	}

	options := tea.WithAltScreen()
	r.tea = tea.NewProgram(m, options)
	if _, err := r.tea.Run(); err != nil {
		fmt.Println("Error running program:", err)
		r.tea.RestoreTerminal()
		os.Exit(1)
	}
}

func (r *ScreenHome) Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) {

	r.files = files
	r.projectAggregated = projectAggregated

	// Update all screens
	for _, s := range r.modelChoices.screens {
		s.Reset(files, projectAggregated)
	}

	if r.tea == nil {
		return
	}

	// Send update command to tea application
	r.tea.Send(DoRefreshModel{files: files, projectAggregated: projectAggregated})
	r.currentModel.Update(DoRefreshModel{files: files, projectAggregated: projectAggregated})
}

// Get Tea model
func (r ScreenHome) GetModel() modelChoices {

	// Prepare list of accepted screens
	m := modelChoices{files: r.files, projectAggregated: r.projectAggregated}
	fillInScreens(&m)

	return m
}

// fillInScreens fills in the screens array, and is used to avoid a circular dependency when creating the screens and coming back to the main screen
func fillInScreens(modelChoices *modelChoices) {

	if len(modelChoices.screens) > 0 {
		// we need to refresh screen only when --watch is set
		// return
	}

	// Create the table screen
	viewTableClass := NewScreenTableClass(true, modelChoices.files, modelChoices.projectAggregated)

	// Create the table screen
	summaryScreen := NewScreenSummary(true, modelChoices.files, modelChoices.projectAggregated)

	// Create the Risk screen
	viewRisks := NewScreenRisks(true, modelChoices.files, modelChoices.projectAggregated)

	// Create the html report screen
	viewHtmlReport := NewScreenHtmlReport(true, modelChoices.files, modelChoices.projectAggregated)

	// Create the screen list
	modelChoices.screens = []Screen{
		&summaryScreen,
		&viewHtmlReport,
		&viewTableClass,
		&viewRisks,
	}

	// Append one screen per programming language
	for languageName, lang := range modelChoices.projectAggregated.ByProgrammingLanguage {
		viewByProgrammingLanguage := ScreenByProgrammingLanguage{isInteractive: true}
		viewByProgrammingLanguage.programmingLangageName = languageName
		viewByProgrammingLanguage.programmingLangageAggregated = lang
		viewByProgrammingLanguage.files = modelChoices.files
		viewByProgrammingLanguage.projectAggregated = modelChoices.projectAggregated

		modelChoices.screens = append(modelChoices.screens, &viewByProgrammingLanguage)
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
		"\n" + StyleSubTitle("AST Metrics is a language-agnostic static code analyzer. \n"+StyleUrl("https://github.com/Halleck45/ast-metrics").Render()).Render() +
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

			// Check if we have a file watcher
			if m.FileWatcher != nil {
				m.FileWatcher.Close()
			}

			return m, tea.Quit
		}
	}

	// issue when navigating back to the main screen
	fillInScreens(&m)

	switch msg := msg.(type) {
	case DoRefreshModel:
		// refresh the model, and the models of the sub screens
		m.files = msg.files
		m.projectAggregated = msg.projectAggregated
		for _, s := range m.screens {
			s.GetModel().Update(msg)
		}

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
			m.Choice = 0
			return m.screens[index].GetModel(), tea.ClearScreen
		}
	}

	return m, nil
}

// Color a string's foreground with the given value.
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}
