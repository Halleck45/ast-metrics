package ui

import (
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	pb "github.com/ast-metrics/ast-metrics/pb"
)

func TestComponentBarchartCyclomaticByMethodRepartition_AsTerminalElement(t *testing.T) {
	component := &ComponentBarchartCyclomaticByMethodRepartition{
		Files:      []*pb.File{},
		Aggregated: analyzer.Aggregated{},
	}

	result := component.AsTerminalElement()
	if result == "" {
		t.Error("expected non-empty terminal element")
	}
}

func TestComponentBarchartCyclomaticByMethodRepartition_AsHtml(t *testing.T) {
	component := &ComponentBarchartCyclomaticByMethodRepartition{
		Files:      []*pb.File{},
		Aggregated: analyzer.Aggregated{},
	}

	result := component.AsHtml()
	if result == "" {
		t.Error("expected non-empty HTML element")
	}
}
