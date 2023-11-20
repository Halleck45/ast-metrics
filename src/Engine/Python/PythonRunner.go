package Python

import (
	"bytes"
	"os"
	"strings"

	"fmt"
	"log"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/File"
	"github.com/pterm/pterm"
)

type PythonRunner struct {
	progressbar   *pterm.SpinnerPrinter
	configuration *Configuration.Configuration
	foundFiles    File.FileList
}

// IsRequired returns true if at least one Go file is found
func (r PythonRunner) IsRequired() bool {
	// If at least one Go file is found, we need to run PHP engine
	return len(r.getFileList().Files) > 0
}

// SetProgressbar sets the progressbar
func (r *PythonRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	(*r).progressbar = progressbar
}

// SetConfiguration sets the configuration
func (r *PythonRunner) SetConfiguration(configuration *Configuration.Configuration) {
	(*r).configuration = configuration
}

// Ensure ensures Go is ready to run.
func (r *PythonRunner) Ensure() error {
	return nil
}

// Finish cleans up the workspace
func (r PythonRunner) Finish() error {
	r.progressbar.Stop()
	return nil
}

// DumpAST dumps the AST of python files in protobuf format
func (r PythonRunner) DumpAST() {

	cnt := 0
	for _, filePath := range r.getFileList().Files {

		cnt++
		r.progressbar.UpdateText("🐍 Dumping AST of Python files (" + fmt.Sprintf("%d", cnt) + "/" + fmt.Sprintf("%d", len(r.getFileList().Files)) + ")")

		hash, err := Engine.GetFileHash(filePath)
		if err != nil {
			log.Fatal(err)
		}
		binPath := Storage.OutputPath() + string(os.PathSeparator) + hash + ".bin"
		// if file exists, skip it
		if _, err := os.Stat(binPath); err == nil {
			continue
		}

		// Create protobuf object
		protoFile, _ := parsePythonFile(filePath)

		// Dump protobuf object to destination
		Engine.DumpProtobuf(protoFile, binPath)
	}

	r.progressbar.Info("🐍 Python code dumped")

}

func parsePythonFile(filename string) (*pb.File, error) {

	stmts := Engine.FactoryStmts()

	file := &pb.File{
		Path:                filename,
		ProgrammingLanguage: "Python",
		Stmts:               stmts,
	}

	sourceCode, err := os.ReadFile(filename)
	if err != nil {
		return file, err
	}

	Ast, err := parser.Parse(bytes.NewBufferString(string(sourceCode)), filename, py.ExecMode)
	if err != nil {
		return file, err
	}

	// Read file content. make it slice of lines (one line per element)
	linesOfFileString := string(sourceCode)
	linesOfFile := strings.Split(linesOfFileString, "\n")

	var classNode *pb.StmtClass

	// @see https://github.com/go-python/gpython/blob/main/ast/walk.go
	ast.Walk(Ast, func(node ast.Ast) bool {

		switch x := node.(type) {
		case *ast.FunctionDef:
			// Function declaration
			funcNode := &pb.StmtFunction{}
			qualifiedName := string(x.Name)
			if classNode != nil {
				qualifiedName = string(classNode.Name.Qualified) + "." + string(x.Name)
			}
			funcNode.Name = &pb.Name{
				Short:     string(x.Name),
				Qualified: qualifiedName,
			}
			funcNode.Operators = []*pb.StmtOperator{}
			funcNode.Operands = []*pb.StmtOperand{}
			funcNode.Stmts = Engine.FactoryStmts()

			if classNode != nil {
				classNode.Stmts.StmtFunction = append(classNode.Stmts.StmtFunction, funcNode)
			} else {
				file.Stmts.StmtFunction = append(file.Stmts.StmtFunction, funcNode)
			}

			// Add function parameters to operands list
			for _, param := range x.Args.Args {
				funcNode.Operands = append(funcNode.Operands, &pb.StmtOperand{
					Name: string(param.Arg),
				})
			}

			lastPosInFunction := x.GetLineno()

			ast.Walk(x, func(node ast.Ast) bool {

				// increase line number, in order to get the latest line of the function
				if node.GetLineno() > lastPosInFunction {
					lastPosInFunction = node.GetLineno()
				}

				switch x := node.(type) {
				case *ast.Name:
					// Variable usage
					// Library does not allow to get the context of the variable (for example if it is a function call or a variable declaration)
					// We store it as operand, and will remove it if it is a function call
					// operation := (*node.(*ast.Name)).Ctx
					identifier := string(x.Id)

					// get next char
					line := linesOfFile[x.GetLineno()-1]
					colOffset := x.GetColOffset()
					colOffset += len(identifier)
					if len(line) > colOffset {
						nextChar := string(line[colOffset])
						if nextChar == "(" {
							return true
						}

						if nextChar == "." {
							return true
						}
					}

					funcNode.Operands = append(funcNode.Operands, &pb.StmtOperand{Name: identifier})

				case *ast.If:
					funcNode.Stmts.StmtDecisionIf = append(funcNode.Stmts.StmtDecisionIf, &pb.StmtDecisionIf{})
				case *ast.For:
					funcNode.Stmts.StmtLoop = append(funcNode.Stmts.StmtLoop, &pb.StmtLoop{})
				case *ast.While:
					funcNode.Stmts.StmtLoop = append(funcNode.Stmts.StmtLoop, &pb.StmtLoop{})
				case *ast.With:
					funcNode.Stmts.StmtLoop = append(funcNode.Stmts.StmtLoop, &pb.StmtLoop{})
				case *ast.AugAssign:
					// x += 1
					funcNode.Operators = append(funcNode.Operators, &pb.StmtOperator{Name: string(x.Op.String())})
				case *ast.BinOp:
					// x + 1
					funcNode.Operators = append(funcNode.Operators, &pb.StmtOperator{Name: string(x.Op.String())})
				case *ast.BoolOp:
					// x and y
					funcNode.Operators = append(funcNode.Operators, &pb.StmtOperator{Name: string(x.Op.String())})
				case *ast.UnaryOp:
					// -x
					funcNode.Operators = append(funcNode.Operators, &pb.StmtOperator{Name: string(x.Op.String())})
				case *ast.Compare:
					// x == 1
					funcNode.Operators = append(funcNode.Operators, &pb.StmtOperator{Name: "=="})
				}
				return true
			})

			// Count lines of code
			loc := Engine.GetLocPositionFromSource(linesOfFile, x.GetLineno(), lastPosInFunction)
			funcNode.LinesOfCode = loc

			// increment loc for class
			if classNode != nil {
				classNode.LinesOfCode.LinesOfCode += loc.LinesOfCode
				classNode.LinesOfCode.LogicalLinesOfCode += loc.LogicalLinesOfCode
				classNode.LinesOfCode.CommentLinesOfCode += loc.CommentLinesOfCode
			}

		case *ast.ClassDef:
			// Class declaration
			classNode = &pb.StmtClass{}
			classNode.Stmts = Engine.FactoryStmts()
			classNode.Name = &pb.Name{
				Short:     string(x.Name),
				Qualified: string(x.Name),
			}
			classNode.LinesOfCode = &pb.LinesOfCode{}

			file.Stmts.StmtClass = append(file.Stmts.StmtClass, classNode)
		}

		return true
	})

	return file, nil
}

// getFileList returns the list of PHP files to analyze, and caches it in memory
func (r *PythonRunner) getFileList() File.FileList {

	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".py")

	return r.foundFiles
}
