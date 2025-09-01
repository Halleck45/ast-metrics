package analyzer

import (
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzerStart(t *testing.T) {
	protoFile := pb.File{
		ProgrammingLanguage: "Go",
		Stmts: &pb.Stmts{
			StmtFunction: []*pb.StmtFunction{
				{
					Stmts: &pb.Stmts{},
				},
			},
			StmtClass: []*pb.StmtClass{
				// class
				{
					Stmts: &pb.Stmts{},
				},
				// class
				{
					Stmts: &pb.Stmts{},
				},
				// class
				{
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{},
					},
				},
				// class
				{
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{},
					},
				},
			},
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{
								Stmts: &pb.Stmts{},
							},
						},
						StmtClass: []*pb.StmtClass{
							// class
							{
								Stmts: &pb.Stmts{},
							},
							// class
							{
								Stmts: &pb.Stmts{},
							},
						},
					},
				},
			},
		},
	}

	// Dump protobuf object to destination
	storage := storage.NewWithName("test")
	storage.Ensure()
	binPath := storage.AstDirectory() + string(os.PathSeparator) + "tmp.bin"
	err := engine.DumpProtobuf(&protoFile, binPath)
	if err != nil {
		t.Error("Error dumping protobuf object", err)
	}

	// Ensure file exists
	if _, err := os.Stat(binPath); err != nil {
		t.Error("File not found", binPath)
	}

	// Start analysis
	parsedFiles := Start(storage, nil)

	// Now first parsed file should be the same as the one we dumped, + analysis
	assert.Equal(t, "Go", parsedFiles[0].ProgrammingLanguage)
	ccn := parsedFiles[0].Stmts.Analyze.Complexity.Cyclomatic
	assert.Greater(t, int(*ccn), 0)
}
