package Golang

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/File"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
	"google.golang.org/protobuf/proto"
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

	// Ensure outdir exists
	if _, err := os.Stat(r.getLocalOutDirectory()); os.IsNotExist(err) {
		if err := os.Mkdir(r.getLocalOutDirectory(), 0755); err != nil {
			return err
		}
	}
	return nil
}

// Finish cleans up the workspace
func (r GolangRunner) Finish() error {
	r.progressbar.Stop()
	return nil
}

// DumpAST dumps the AST of Go files in protobuf format
func (r GolangRunner) DumpAST() {

	cnt := 0
	for _, filePath := range r.getFileList().Files {

		cnt++
		r.progressbar.UpdateText("🦫 Dumping AST of Go files (" + fmt.Sprintf("%d", cnt) + "/" + fmt.Sprintf("%d", len(r.getFileList().Files)) + ")")

		hash, err := getFileHash(filePath)
		if err != nil {
			log.Fatal(err)
		}
		binPath := r.getLocalOutDirectory() + string(os.PathSeparator) + hash + ".bin"
		// if file exists, skip it
		if _, err := os.Stat(binPath); err == nil {
			continue
		}

		// Create protobuf object
		protoFile := parseGoFile(filePath)

		// Dump protobuf object to destination
		dumpProtobuf(protoFile, binPath)
	}

	r.progressbar.Info("🦫 Golang code dumped")

}

// Provides the hash of a file, in order to avoid to parse it twice
func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func parseGoFile(filePath string) *pb.File {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	linesOfFileString := string(fileContent)
	// make it slice of lines (one line per element)
	linesOfFile := strings.Split(linesOfFileString, "\n")

	var funcs []*pb.StmtFunction
	var loc, cloc, lloc, blankLines int
	//var operators []*pb.StmtOperator
	//var operands []*pb.StmtOperand

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
			loc = end - start + 1
			cloc = 0 //countComments(x)
			lloc = loc
			blankLines = 0

			// get blank lines (line breaks) and declaration line
			for i := start - 1; i < end; i++ {
				// trim it
				linesOfFile[i] = strings.TrimSpace(linesOfFile[i])

				if linesOfFile[i] == "" {
					lloc--
					blankLines++
				}

				// if beginning of line is not a comment, it's a declaration line
				if strings.HasPrefix(linesOfFile[i], "//") ||
					strings.HasPrefix(linesOfFile[i], "/*") ||
					strings.HasPrefix(linesOfFile[i], "*/") ||
					strings.HasPrefix(linesOfFile[i], "*") ||
					strings.HasPrefix(linesOfFile[i], "#") {
					// @todo issue here.
					// Please update it using the countComments() function
					lloc--
					cloc++
				}
			}

			funcNode.LinesOfCode = &pb.LinesOfCode{}
			funcNode.LinesOfCode.LinesOfCode = int32(loc)
			funcNode.LinesOfCode.CommentLinesOfCode = int32(cloc)
			// lloc = loc - (clocl + blank lines + declaration line)
			lloc = loc - (cloc + blankLines + 2)
			funcNode.LinesOfCode.LogicalLinesOfCode = int32(lloc)
		}
		return true
	})

	stmts := pb.Stmts{}
	stmts.StmtFunction = funcs

	file := &pb.File{
		Path:  filePath,
		Stmts: &stmts,
	}

	return file
}

func dumpProtobuf(file *pb.File, binPath string) {
	out, err := proto.Marshal(file)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(binPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		log.Fatal(err)
	}
}

func countComments(n ast.Node) int {
	var count int
	ast.Inspect(n, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.CommentGroup:
		case *ast.Comment:
			count++
		}
		return true
	})
	return count
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

// getLocalOutDirectory returns the path to the local output directory
func (r *GolangRunner) getLocalOutDirectory() string {
	return Storage.Path() + "/output"
}
