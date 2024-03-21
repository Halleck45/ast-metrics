package Cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type Screen interface {

	// Returns the Tea model used for the screen
	GetModel() tea.Model

	// Returns the name of the screen
	GetScreenName() string

	Reset(files []*pb.File, projectAggregated Analyzer.ProjectAggregated)
}
