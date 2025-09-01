package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/stretchr/testify/assert"
)

func TestNewScreenHome(t *testing.T) {
	isInteractive := true
	files := []*pb.File{}
	projectAggregated := analyzer.ProjectAggregated{}

	screenHome := NewScreenHome(isInteractive, files, projectAggregated)

	if screenHome.isInteractive != isInteractive {
		t.Errorf("Expected isInteractive to be %v, but got %v", isInteractive, screenHome.isInteractive)
	}

	if len(screenHome.files) != len(files) {
		t.Errorf("Expected files to be %v, but got %v", files, screenHome.files)
	}
}

func TestGetModel(t *testing.T) {
	isInteractive := true
	files := []*pb.File{}
	projectAggregated := analyzer.ProjectAggregated{}

	screenHome := NewScreenHome(isInteractive, files, projectAggregated)
	model := screenHome.GetModel()

	// sending the "enter" key
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, ok := model.(modelScreenSummary)
	assert.True(t, ok)
}

func TestInit(t *testing.T) {
	model := modelChoices{}
	cmd := model.Init()

	if cmd != nil {
		t.Errorf("Expected cmd to be nil, but got %v", cmd)
	}
}
