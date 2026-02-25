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

	pathHash, ok := decoded[0]["pathHash"].(string)
	if !ok || pathHash == "" {
		t.Fatalf("expected non-empty pathHash in json, got %#v", decoded[0]["pathHash"])
	}
}

func TestBuildFilesJSONPruned_PathHashStableAndDistinct(t *testing.T) {
	makeFile := func(path string) *pb.File {
		return &pb.File{
			Path:                path,
			ShortPath:           "internal/analyzer/community_aggregator.go",
			ProgrammingLanguage: "Golang",
			Stmts:               &pb.Stmts{},
		}
	}

	first := makeFile("/tmp/repo-a/internal/analyzer/community_aggregator.go")
	second := makeFile("/tmp/repo-b/internal/analyzer/community_aggregator.go")

	raw := buildFilesJSONPruned([]*pb.File{first, second}, "All")
	var decoded []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("expected 2 files in json, got %d", len(decoded))
	}

	hashA, okA := decoded[0]["pathHash"].(string)
	hashB, okB := decoded[1]["pathHash"].(string)
	if !okA || !okB || hashA == "" || hashB == "" {
		t.Fatalf("expected non-empty pathHash values, got %#v and %#v", decoded[0]["pathHash"], decoded[1]["pathHash"])
	}
	if hashA == hashB {
		t.Fatalf("expected different pathHash for different absolute paths, got same value %q", hashA)
	}

	raw2 := buildFilesJSONPruned([]*pb.File{first, second}, "All")
	var decoded2 []map[string]interface{}
	if err := json.Unmarshal([]byte(raw2), &decoded2); err != nil {
		t.Fatalf("expected valid json on second call, got error: %v", err)
	}
	hashA2, _ := decoded2[0]["pathHash"].(string)
	hashB2, _ := decoded2[1]["pathHash"].(string)
	if hashA != hashA2 || hashB != hashB2 {
		t.Fatalf("expected stable pathHash across calls, got (%q,%q) then (%q,%q)", hashA, hashB, hashA2, hashB2)
	}
}
