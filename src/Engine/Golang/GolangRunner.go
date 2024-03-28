package Golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/File"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
)

type GolangRunner struct {
	progressbar   *pterm.SpinnerPrinter
	configuration *Configuration.Configuration
	foundFiles    File.FileList
}

// IsRequired returns true if at least one Go file is found
func (r GolangRunner) IsRequired() bool {
	// If at least one Go file is found, we need to run PHP engine
	return len(r.getFileList().Files) > 0
}

// SetProgressbar sets the progressbar
func (r *GolangRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	(*r).progressbar = progressbar
}

// SetConfiguration sets the configuration
func (r *GolangRunner) SetConfiguration(configuration *Configuration.Configuration) {
	(*r).configuration = configuration
}

// Ensure ensures Go is ready to run.
func (r *GolangRunner) Ensure() error {
	return nil
}

// Finish cleans up the workspace
func (r GolangRunner) Finish() error {
	if r.progressbar != nil {
		r.progressbar.Stop()
	}
	return nil
}

// DumpAST dumps the AST of Go files in protobuf format
func (r GolangRunner) DumpAST() {

	cnt := 0
	for _, filePath := range r.getFileList().Files {

		cnt++
		if r.progressbar != nil {
			r.progressbar.UpdateText("🦫 Dumping AST of Go files (" + fmt.Sprintf("%d", cnt) + "/" + fmt.Sprintf("%d", len(r.getFileList().Files)) + ")")
		}

		hash, err := Engine.GetFileHash(filePath)
		if err != nil {
			log.Error(err)
		}
		binPath := Storage.OutputPath() + string(os.PathSeparator) + hash + ".bin"
		// if file exists, skip it
		if _, err := os.Stat(binPath); err == nil {
			continue
		}

		// Create protobuf object
		protoFile := ParseGoFile(filePath)

		// Dump protobuf object to destination
		err = Engine.DumpProtobuf(protoFile, binPath)
		if err != nil {
			log.Error(err)
		}
	}

	if r.progressbar != nil {
		r.progressbar.Info("🦫 Golang code dumped")
	}

}

func ParseGoFile(filePath string) *pb.File {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Error(err)
	}

	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(err)
	}
	linesOfFileString := string(fileContent)
	// make it slice of lines (one line per element)
	linesOfFile := strings.Split(linesOfFileString, "\n")

	var funcs []*pb.StmtFunction

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Function declaration
			funcNode := &pb.StmtFunction{}
			funcNode.Name = &pb.Name{
				Short: x.Name.Name,
				// package + short  @todo
				Qualified: x.Name.String(),
			}
			funcNode.Operators = []*pb.StmtOperator{}
			funcNode.Operands = []*pb.StmtOperand{}
			funcNode.Stmts = &pb.Stmts{}
			funcNode.Stmts.StmtDecisionIf = []*pb.StmtDecisionIf{}
			funcNode.Stmts.StmtDecisionSwitch = []*pb.StmtDecisionSwitch{}
			funcNode.Stmts.StmtDecisionCase = []*pb.StmtDecisionCase{}
			funcNode.Stmts.StmtLoop = []*pb.StmtLoop{}

			funcs = append(funcs, funcNode)

			// Add function parameters to operands
			for _, param := range x.Type.Params.List {
				for _, paramName := range param.Names {
					funcNode.Operands = append(funcNode.Operands, &pb.StmtOperand{Name: paramName.Name})
				}
			}

			// Go through the function body
			ast.Inspect(x.Body, func(n ast.Node) bool {

				switch y := n.(type) {
				case *ast.BinaryExpr:
					funcNode.Operators = append(funcNode.Operators, &pb.StmtOperator{Name: y.Op.String()})
				case *ast.Ident:
					funcNode.Operands = append(funcNode.Operands, &pb.StmtOperand{Name: y.Name})
				case *ast.IfStmt:
					funcNode.Stmts.StmtDecisionIf = append(funcNode.Stmts.StmtDecisionIf, &pb.StmtDecisionIf{})
				case *ast.SwitchStmt:
					funcNode.Stmts.StmtDecisionSwitch = append(funcNode.Stmts.StmtDecisionSwitch, &pb.StmtDecisionSwitch{})
				case *ast.CaseClause:
					funcNode.Stmts.StmtDecisionCase = append(funcNode.Stmts.StmtDecisionCase, &pb.StmtDecisionCase{})
				case *ast.ForStmt:
					funcNode.Stmts.StmtLoop = append(funcNode.Stmts.StmtLoop, &pb.StmtLoop{})
				case *ast.RangeStmt:
					funcNode.Stmts.StmtLoop = append(funcNode.Stmts.StmtLoop, &pb.StmtLoop{})

				}
				return true
			})

			// Count lines of code
			start := fset.Position(x.Pos()).Line
			end := fset.Position(x.End()).Line
			loc := Engine.GetLocPositionFromSource(linesOfFile, start, end)
			funcNode.LinesOfCode = loc
		}
		return true
	})

	stmts := pb.Stmts{}
	stmts.StmtFunction = funcs

	file := &pb.File{
		Path:                filePath,
		ProgrammingLanguage: "Golang",
		Stmts:               &stmts,
		LinesOfCode: &pb.LinesOfCode{
			LinesOfCode: int32(len(linesOfFile)),
		},
	}

	return file
}

// getFileList returns the list of PHP files to analyze, and caches it in memory
func (r *GolangRunner) getFileList() File.FileList {

	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".go")

	return r.foundFiles
}
