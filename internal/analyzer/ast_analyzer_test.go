package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeFiles(t *testing.T) {
	protoFile := &pb.File{
		ProgrammingLanguage: "Go",
		Stmts: &pb.Stmts{
			StmtFunction: []*pb.StmtFunction{
				{
					Stmts: &pb.Stmts{},
				},
			},
			StmtClass: []*pb.StmtClass{
				{Stmts: &pb.Stmts{}},
				{Stmts: &pb.Stmts{}},
				{Stmts: &pb.Stmts{Analyze: &pb.Analyze{}}},
				{Stmts: &pb.Stmts{Analyze: &pb.Analyze{}}},
			},
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{Stmts: &pb.Stmts{}},
						},
						StmtClass: []*pb.StmtClass{
							{Stmts: &pb.Stmts{}},
							{Stmts: &pb.Stmts{}},
						},
					},
				},
			},
		},
	}

	// Analyze in-memory
	results := AnalyzeFiles([]*pb.File{protoFile}, nil)

	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Go", results[0].ProgrammingLanguage)
	ccn := results[0].Stmts.Analyze.Complexity.Cyclomatic
	assert.NotNil(t, ccn)
	// File contains empty classes plus two empty functions outside classes.
	// With file CCN = classes + functions outside classes, expected value is 2.
	assert.Equal(t, 2, int(*ccn))
}

func TestRecomputeFileCyclomatic_SumsClassesAndFunctionsOutsideClasses(t *testing.T) {
	classCyclo := int32(4)
	methodCyclo := int32(3)
	functionCyclo := int32(2)

	method := &pb.StmtFunction{
		Name: &pb.Name{Short: "M", Qualified: "Acme\\C::M"},
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: &methodCyclo},
			},
		},
	}

	class := &pb.StmtClass{
		Name: &pb.Name{Short: "C", Qualified: "Acme\\C"},
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: &classCyclo},
			},
			StmtFunction: []*pb.StmtFunction{method},
		},
	}

	outsideFn := &pb.StmtFunction{
		Name: &pb.Name{Short: "F", Qualified: "Acme\\F"},
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: &functionCyclo},
			},
		},
	}

	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtClass:    []*pb.StmtClass{class},
						StmtFunction: []*pb.StmtFunction{method, outsideFn},
					},
				},
			},
		},
	}

	recomputeFileCyclomatic(file)

	if file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
		t.Fatalf("expected file cyclomatic complexity to be set")
	}
	assert.Equal(t, int32(6), *file.Stmts.Analyze.Complexity.Cyclomatic)
}

func TestRecomputeFileCyclomatic_UsesFunctionsWhenNoClasses(t *testing.T) {
	c1 := int32(1)
	c2 := int32(2)
	fn1 := &pb.StmtFunction{
		Name: &pb.Name{Short: "A", Qualified: "Acme\\A"},
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: &c1},
			},
		},
	}
	fn2 := &pb.StmtFunction{
		Name: &pb.Name{Short: "B", Qualified: "Acme\\B"},
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{Cyclomatic: &c2},
			},
		},
	}
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtFunction: []*pb.StmtFunction{fn1, fn2},
		},
	}

	recomputeFileCyclomatic(file)

	assert.Equal(t, int32(3), *file.Stmts.Analyze.Complexity.Cyclomatic)
}

func TestRecomputeFileCyclomatic_UsesClassesWhenNoFunctions(t *testing.T) {
	classCyclo := int32(5)
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtClass: []*pb.StmtClass{
							{
								Name: &pb.Name{Short: "OnlyClass", Qualified: "Acme\\OnlyClass"},
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Complexity: &pb.Complexity{Cyclomatic: &classCyclo},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	recomputeFileCyclomatic(file)

	assert.Equal(t, int32(5), *file.Stmts.Analyze.Complexity.Cyclomatic)
}
