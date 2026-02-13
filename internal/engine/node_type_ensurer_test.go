package engine

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
)

func int32Ptr(v int32) *int32 { return &v }

func TestEnsureNodeTypeIsComplete_SetsFileCyclomaticFromClasses(t *testing.T) {
	classA := &pb.StmtClass{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: int32Ptr(2)},
			},
		},
	}
	classB := &pb.StmtClass{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: int32Ptr(3)},
			},
		},
	}

	// classA is present both in namespace and file; it must be counted once.
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: int32Ptr(99)},
			},
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtClass: []*pb.StmtClass{classA},
					},
				},
			},
			StmtClass: []*pb.StmtClass{classA, classB},
		},
	}

	EnsureNodeTypeIsComplete(file)

	if file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
		t.Fatalf("expected file cyclomatic complexity to be set")
	}
	if got := *file.Stmts.Analyze.Complexity.Cyclomatic; got != 5 {
		t.Fatalf("expected file cyclomatic complexity 5 (sum of classes), got %d", got)
	}
}

func TestEnsureNodeTypeIsComplete_SetsFileCyclomaticFromFunctionsWhenNoClasses(t *testing.T) {
	fnA := &pb.StmtFunction{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: int32Ptr(1)},
			},
		},
	}
	fnB := &pb.StmtFunction{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: int32Ptr(4)},
			},
		},
	}

	// fnA is present both in namespace and file; it must be counted once.
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: int32Ptr(42)},
			},
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{fnA},
					},
				},
			},
			StmtFunction: []*pb.StmtFunction{fnA, fnB},
		},
	}

	EnsureNodeTypeIsComplete(file)

	if file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
		t.Fatalf("expected file cyclomatic complexity to be set")
	}
	if got := *file.Stmts.Analyze.Complexity.Cyclomatic; got != 5 {
		t.Fatalf("expected file cyclomatic complexity 5 (sum of functions), got %d", got)
	}
}
