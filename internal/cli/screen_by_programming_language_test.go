package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestScreenByProgrammingLanguageGetScreenName(t *testing.T) {
	screen := ScreenByProgrammingLanguage{
		programmingLangageName: "PHP",
		files: []*pb.File{
			{ProgrammingLanguage: "PHP"},
			{ProgrammingLanguage: "PHP"},
			{ProgrammingLanguage: "Python"},
		},
	}

	expected := "🐘 PHP (2 files)"
	got := screen.GetScreenName()

	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestScreenByProgrammingLanguageGetModel(t *testing.T) {
	screen := ScreenByProgrammingLanguage{
		isInteractive:          true,
		programmingLangageName: "Golang",
		files: []*pb.File{
			{ProgrammingLanguage: "Golang"},
			{ProgrammingLanguage: "Python"},
		},
		projectAggregated: analyzer.ProjectAggregated{},
	}

	model := screen.GetModel()

	if model == nil {
		t.Errorf("Expected model, got nil")
	}
}

func TestScreenByProgrammingLanguageModelByProgrammingLanguageUpdate(t *testing.T) {
	model := modelByProgrammingLanguage{
		programmingLangageName: "Golang",
		files: []*pb.File{
			{ProgrammingLanguage: "Golang"},
			{ProgrammingLanguage: "Python"},
		},
		projectAggregated:   analyzer.ProjectAggregated{},
		componentTableClass: NewComponentFileTable(true, []*pb.File{}),
	}

	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if updatedModel == nil {
		t.Errorf("Expected updated model, got nil")
	}
}

func TestModelByProgrammingLanguageView(t *testing.T) {
	model := modelByProgrammingLanguage{
		programmingLangageName: "Golang",
		files: []*pb.File{
			{ProgrammingLanguage: "Golang"},
			{ProgrammingLanguage: "Python"},
		},
		projectAggregated:   analyzer.ProjectAggregated{},
		componentTableClass: NewComponentFileTable(true, []*pb.File{}),
	}

	view := model.View()

	if view == "" {
		t.Errorf("Expected view, got empty string")
	}
}
