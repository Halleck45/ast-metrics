package Golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/File"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
	"golang.org/x/mod/modfile"
)

type GolangRunner struct {
	progressbar      *pterm.SpinnerPrinter
	configuration    *Configuration.Configuration
	foundFiles       File.FileList
	currentGoModFile *modfile.File
	currentGoModPath string
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

		hash, err := Storage.GetFileHash(filePath)
		if err != nil {
			log.Error(err)
		}
		binPath := r.configuration.Storage.AstDirectory() + string(os.PathSeparator) + hash + ".bin"
		// if file exists, skip it
		if _, err := os.Stat(binPath); err == nil {
			continue
		}

		// Find the mod file sible to the file
		// make it realpath
		realPath, err := filepath.Abs(filePath)
		if err == nil {
			r.currentGoModFile, err = r.SearchModfile(realPath)
			if err != nil {
				log.Error(err)
			}
		}

		// Create protobuf object
		protoFile := r.ParseGoFile(filePath)
		protoFile.Checksum = hash

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

func (r *GolangRunner) SearchModfile(path string) (*modfile.File, error) {

	// Avoid duplicate search
	if r.currentGoModFile != nil {
		// if directory is a subdirectory of the current mod file, return it
		if strings.Contains(path, r.currentGoModPath) {
			return r.currentGoModFile, nil
		}
	}

	goModFile := path + string(os.PathSeparator) + "go.mod"

	if _, err := os.Stat(goModFile); err == nil {

		fileBytes, err := os.ReadFile(goModFile)
		if err != nil {
			return nil, err
		}
		f, err := modfile.Parse("go.mod", fileBytes, nil)
		if err != nil {
			return nil, err
		}

		r.currentGoModFile = f
		r.currentGoModPath = path

		return f, nil
	}

	// Search in parent directory
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) <= 2 {
		return nil, fmt.Errorf("go.mod file not found")
	}
	parts = parts[:len(parts)-1]
	parentDirectory := strings.Join(parts, string(os.PathSeparator))
	return r.SearchModfile(parentDirectory)
}

func (r GolangRunner) ParseGoFile(filePath string) *pb.File {
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
	importedPackages := make(map[string]string)
	currentPackage := ""
	if r.currentGoModFile != nil {
		currentPackage = r.currentGoModFile.Module.Mod.Path
	}

	stmts := pb.Stmts{}
	file := &pb.File{
		Path:                filePath,
		ProgrammingLanguage: "Golang",
		Stmts:               &stmts,
		LinesOfCode: &pb.LinesOfCode{
			LinesOfCode: int32(len(linesOfFile)),
		},
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {

		case *ast.ImportSpec:
			importedPackage := x.Path.Value
			alias := ""
			if x.Name != nil {
				alias = x.Name.Name
			}

			// Remove quotes
			importedPackage = importedPackage[1 : len(importedPackage)-1]

			// if alias is empty, it means the package is imported with its default name
			if alias == "" {
				alias = importedPackage[strings.LastIndex(importedPackage, "/")+1:]
			}

			// Add to imported packages
			importedPackages[alias] = importedPackage

			// Skip system packages
			if !strings.Contains(importedPackage, "github.com") {
				return true
			}

			if file.Stmts.StmtExternalDependencies == nil {
				file.Stmts.StmtExternalDependencies = []*pb.StmtExternalDependency{}
			}
			dependency := &pb.StmtExternalDependency{
				ClassName: alias,
				Namespace: importedPackage,
			}

			if currentPackage != "" {
				dependency.From = currentPackage
			}

			file.Stmts.StmtExternalDependencies = append(file.Stmts.StmtExternalDependencies, dependency)

		case *ast.File:
			// Get the full package name
			// File declaration
			currentPackage += x.Name.Name

		case *ast.Package:
			currentPackage += x.Name
			if file.Stmts.StmtNamespace == nil {
				file.Stmts.StmtNamespace = []*pb.StmtNamespace{}
			}
			file.Stmts.StmtNamespace = append(file.Stmts.StmtNamespace, &pb.StmtNamespace{
				Name: &pb.Name{
					Short:     x.Name,
					Qualified: x.Name,
				},
			})

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
			funcNode.Stmts.StmtNamespace = []*pb.StmtNamespace{}
			funcNode.Stmts.StmtLoop = []*pb.StmtLoop{}
			funcNode.Stmts.StmtExternalDependencies = []*pb.StmtExternalDependency{}

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
					// Vérifier si l'identifiant fait référence à un autre paquet
					for _, imported := range importedPackages {
						if y.Name == imported {
							funcNode.Stmts.StmtExternalDependencies = append(funcNode.Stmts.StmtExternalDependencies, &pb.StmtExternalDependency{
								ClassName: imported,
								Namespace: imported,
								From:      currentPackage,
							})
						}
					}
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
				case *ast.Package:
					namespace := &pb.StmtNamespace{}
					namespace.Name = &pb.Name{
						Short:     y.Name,
						Qualified: y.Name,
					}
					funcNode.Stmts.StmtNamespace = append(funcNode.Stmts.StmtNamespace, namespace)
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

	file.Stmts.StmtFunction = funcs

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
