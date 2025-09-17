package ui

import (
	"strings"
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestComponentBarchartLcomRepartition_AsTerminalElement(t *testing.T) {
	lcom4 := int32(2)
	files := []*pb.File{
		{
			Stmts: &pb.Stmts{
				StmtClass: []*pb.StmtClass{
					{
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								ClassCohesion: &pb.ClassCohesion{Lcom4: &lcom4},
							},
						},
					},
				},
			},
		},
	}

	component := &ComponentBarchartLcomRepartition{
		Files:      files,
		Aggregated: analyzer.Aggregated{},
	}

	result := component.AsTerminalElement()
	if result == "" {
		t.Error("expected non-empty terminal element")
	}
}

func TestComponentBarchartLcomRepartition_AsHtml(t *testing.T) {
	component := &ComponentBarchartLcomRepartition{
		Files:      []*pb.File{},
		Aggregated: analyzer.Aggregated{},
	}

	result := component.AsHtml()
	if !strings.Contains(result, "chart-lcom4") {
		t.Error("expected HTML to contain chart-lcom4 identifier")
	}
}
