package Analyzer

import (
    "github.com/halleck45/ast-metrics/src/Storage"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    "os"
    "io/ioutil"
    "log"
    "fmt"
    "strconv"
    "sync"
    "github.com/golang/protobuf/proto"
    "github.com/pterm/pterm"
    "github.com/yargevad/filepathx"
)

func Start(progressbar *pterm.SpinnerPrinter) {

    workdir := Storage.Path()
    // List all ASTs files (*.bin) in the workdir
    astFiles, err := filepathx.Glob(workdir + "/**/*.bin")
    if err != nil {
        panic(err)
    }

    maxParallelCommands := os.Getenv("MAX_PARALLEL_COMMANDS")
    // if maxParallelCommands is empty, set default value
    if maxParallelCommands == "" {
        maxParallelCommands = "10"
    }
    // to int
    maxParallelCommandsInt, err := strconv.Atoi(maxParallelCommands)
    if err != nil {
        progressbar.Fail("Error while parsing MAX_PARALLEL_COMMANDS env variable")
    }

    // Wait for end of all goroutines
    var wg sync.WaitGroup

    nbParsingFiles := 0
    // in parallel, 8 process max, analyze each AST file running the runAnalysis function
    sem := make(chan struct{}, maxParallelCommandsInt)
    for _, file := range astFiles {
        wg.Add(1)
        nbParsingFiles++
        sem <- struct{}{}
        go func(file string) {
            defer wg.Done()
            executeFileAnalysis(file)

            // details is the number of files processed / total number of files
            details := strconv.Itoa(nbParsingFiles) + "/" + strconv.Itoa(len(astFiles))
            progressbar.UpdateText("Analyzing (" + details + ")")
            <-sem
        }(file)
    }

    // Wait for all goroutines to finish
    for i := 0; i < maxParallelCommandsInt; i++ {
        sem <- struct{}{}
    }

    wg.Wait()
    progressbar.UpdateText("")
    progressbar.Info("Analysis finished")
    progressbar.Stop()

}

func executeFileAnalysis(file string) {

    // load AST via ProtoBuf (using NodeType package)
    in, err := ioutil.ReadFile(file)
    if err != nil {
        log.Fatalln("Error reading file:", err)
    }
    pbFile := &pb.File{}
    if err := proto.Unmarshal(in, pbFile); err != nil {
        log.Fatalln("Failed to parse address pbFile:", err)
    }

    root := &ASTNode{children: pbFile.Stmts}
    cyclomaticVisitor := &ComplexityVisitor{}
    root.Accept(cyclomaticVisitor)

    root.Visit()
    complexity := cyclomaticVisitor.GetComplexity()

    if complexity > 0 {
        fmt.Println("Cyclomatic complexity of " + pbFile.Path + " is " + strconv.Itoa(complexity))
    }
}

type Visitor interface {
    Visit(stmts *pb.Stmts)
}

type ComplexityVisitor struct {
    complexity int
}
func (v *ComplexityVisitor) Visit(stmts *pb.Stmts) {
    // get Stmts

    // foreach type of statements
    for _, stmt := range stmts.StmtClass {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtNamespace {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtFunction {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtTrait {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtDecisionIf {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtDecisionElseIf {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtDecisionElse {
        v.Visit(stmt.Stmts)
    }
    for _, stmt := range stmts.StmtDecisionCase {
        v.Visit(stmt.Stmts)
    }

    // count if
    v.complexity += len(stmts.StmtDecisionIf) + len(stmts.StmtDecisionElseIf) + len(stmts.StmtDecisionElse) + len(stmts.StmtDecisionCase)
}
func (v *ComplexityVisitor) GetComplexity() int {
	return v.complexity
}


// Structure de l'arbre AST
type ASTNode struct {
	children *pb.Stmts
	Visitors []Visitor
}

func (n *ASTNode) Accept(visitor Visitor) {
    n.Visitors = append(n.Visitors, visitor)
}

func (n *ASTNode) Visit() {
    // todo: faire le parent aussi
	for _, v := range n.Visitors {
        v.Visit(n.children)
    }
}
