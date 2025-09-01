package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type Screen interface {

	// Returns the Tea model used for the screen
	GetModel() tea.Model

	// Returns the name of the screen
	GetScreenName() string

	Reset(files []*pb.File, projectAggregated analyzer.ProjectAggregated)
}
