package report

import (
	"encoding/json"
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
)

func TestPruneFile_KeepsClassesAndOutsideFunctions(t *testing.T) {
	method := &pb.StmtFunction{
		Name:  &pb.Name{Short: "M", Qualified: "analyzer\\CommunityAggregator::M"},
		Stmts: &pb.Stmts{Analyze: &pb.Analyze{Complexity: &pb.Complexity{}}},
	}
	outside := &pb.StmtFunction{
		Name:  &pb.Name{Short: "density", Qualified: "analyzer\\density"},
		Stmts: &pb.Stmts{Analyze: &pb.Analyze{Complexity: &pb.Complexity{}}},
	}

	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtClass: []*pb.StmtClass{
							{
								Name: &pb.Name{Short: "CommunityAggregator", Qualified: "analyzer\\CommunityAggregator"},
								Stmts: &pb.Stmts{
									StmtFunction: []*pb.StmtFunction{method},
								},
							},
						},
						StmtFunction: []*pb.StmtFunction{method, outside},
					},
				},
			},
		},
	}

	pruneFile(file)

	if file.Stmts == nil {
		t.Fatalf("expected file stmts")
	}
	if file.Stmts.StmtNamespace != nil {
		t.Fatalf("expected namespace to be pruned")
	}

	if len(file.Stmts.StmtClass) != 1 {
		t.Fatalf("expected 1 class/struct, got %d", len(file.Stmts.StmtClass))
	}
	if len(file.Stmts.StmtClass[0].Stmts.StmtFunction) != 1 {
		t.Fatalf("expected 1 method in class/struct, got %d", len(file.Stmts.StmtClass[0].Stmts.StmtFunction))
	}

	if len(file.Stmts.StmtFunction) != 1 {
		t.Fatalf("expected 1 outside function, got %d", len(file.Stmts.StmtFunction))
	}
	if file.Stmts.StmtFunction[0].Name == nil || file.Stmts.StmtFunction[0].Name.Qualified != "analyzer\\density" {
		t.Fatalf("unexpected outside function: %+v", file.Stmts.StmtFunction[0].Name)
	}
}

func TestBuildFilesJSONPruned_ContainsOutsideFunctions(t *testing.T) {
	outside := &pb.StmtFunction{
		Name:  &pb.Name{Short: "density", Qualified: "analyzer\\density"},
		Stmts: &pb.Stmts{Analyze: &pb.Analyze{Complexity: &pb.Complexity{}}},
	}
	file := &pb.File{
		Path:                "/tmp/community_aggregator.go",
		ShortPath:           "internal/analyzer/community_aggregator.go",
		ProgrammingLanguage: "Golang",
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{outside},
					},
				},
			},
		},
	}

	raw := buildFilesJSONPruned([]*pb.File{file}, "All")
	var decoded []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected one file in json, got %d", len(decoded))
	}

	stmtsVal, ok := decoded[0]["stmts"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected stmts object in json")
	}
	functionsVal, ok := stmtsVal["stmtFunction"].([]interface{})
	if !ok {
		t.Fatalf("expected stmtFunction array in json")
	}
	if len(functionsVal) != 1 {
		t.Fatalf("expected 1 outside function in json, got %d", len(functionsVal))
	}
}
