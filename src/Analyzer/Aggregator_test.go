package Analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"google.golang.org/protobuf/proto"
)

func TestConsolidate(t *testing.T) {

	aggregator := Aggregator{}
	aggregated := Aggregated{
		NbMethods:                           10,
		NbClasses:                           5,
		AverageCyclomaticComplexityPerClass: 20,
		AverageHalsteadDifficulty:           30,
		AverageHalsteadEffort:               40,
		AverageHalsteadVolume:               50,
		AverageHalsteadTime:                 60,
		AverageLocPerMethod:                 70,
		AverageClocPerMethod:                80,
		AverageLlocPerMethod:                90,
		AverageMI:                           100,
		AverageMIwoc:                        110,
		AverageMIcw:                         120,
	}

	aggregator.consolidate(&aggregated)

	if aggregated.AverageMethodsPerClass != 2 {
		t.Errorf("Expected 2, got %f", aggregated.AverageMethodsPerClass)
	}

	if aggregated.AverageCyclomaticComplexityPerClass != 4 {
		t.Errorf("Expected 4, got %f", aggregated.AverageCyclomaticComplexityPerClass)
	}

	if aggregated.AverageHalsteadDifficulty != 6 {
		t.Errorf("Expected 6, got %f", aggregated.AverageHalsteadDifficulty)
	}

	if aggregated.AverageHalsteadEffort != 8 {
		t.Errorf("Expected 8, got %f", aggregated.AverageHalsteadEffort)
	}

	if aggregated.AverageHalsteadVolume != 10 {
		t.Errorf("Expected 10, got %f", aggregated.AverageHalsteadVolume)
	}

	if aggregated.AverageHalsteadTime != 12 {
		t.Errorf("Expected 12, got %f", aggregated.AverageHalsteadTime)
	}

	if aggregated.AverageLocPerMethod != 7 {
		t.Errorf("Expected 7, got %f", aggregated.AverageLocPerMethod)
	}

	if aggregated.AverageClocPerMethod != 8 {
		t.Errorf("Expected 8, got %f", aggregated.AverageClocPerMethod)
	}

	if aggregated.AverageLlocPerMethod != 9 {
		t.Errorf("Expected 9, got %f", aggregated.AverageLlocPerMethod)
	}

	if aggregated.AverageMI != 20 {
		t.Errorf("Expected 20, got %f", aggregated.AverageMI)
	}

	if aggregated.AverageMIwoc != 22 {
		t.Errorf("Expected 22, got %f", aggregated.AverageMIwoc)
	}

	if aggregated.AverageMIcw != 24 {
		t.Errorf("Expected 24, got %f", aggregated.AverageMIcw)
	}
}

func TestCalculate(t *testing.T) {
	aggregator := Aggregator{}
	stmts := pb.Stmts{
		StmtFunction: []*pb.StmtFunction{
			{
				Stmts: &pb.Stmts{
					Analyze: &pb.Analyze{
						Complexity: &pb.Complexity{
							Cyclomatic: proto.Int32(10),
						},
					},
				},
			},
			{
				Stmts: &pb.Stmts{
					Analyze: &pb.Analyze{
						Complexity: &pb.Complexity{
							Cyclomatic: proto.Int32(20),
						},
					},
				},
			},
		},
		StmtClass: []*pb.StmtClass{
			{}, {}, {},
		},
		Analyze: &pb.Analyze{
			Volume: &pb.Volume{
				Loc:  proto.Int32(100),
				Cloc: proto.Int32(200),
				Lloc: proto.Int32(300),
			},
		},
	}
	aggregated := Aggregated{}

	aggregator.calculate(&stmts, &aggregated)

	if aggregated.NbMethods != 2 {
		t.Errorf("Expected 2, got %d", aggregated.NbMethods)
	}

	if aggregated.NbClasses != 3 {
		t.Errorf("Expected 3, got %d", aggregated.NbClasses)
	}

	if aggregated.AverageCyclomaticComplexityPerMethod != 30 {
		t.Errorf("Expected 30, got %f", aggregated.AverageCyclomaticComplexityPerMethod)
	}

	if aggregated.Loc != 100 {
		t.Errorf("Expected 100, got %d", aggregated.Loc)
	}

	if aggregated.Cloc != 200 {
		t.Errorf("Expected 200, got %d", aggregated.Cloc)
	}

	if aggregated.Lloc != 300 {
		t.Errorf("Expected 300, got %d", aggregated.Lloc)
	}
}
