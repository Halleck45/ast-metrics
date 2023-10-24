package Analyzer

import (
    "github.com/halleck45/ast-metrics/src/Storage"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    Complexity "github.com/halleck45/ast-metrics/src/Analyzer/Complexity"
    Volume "github.com/halleck45/ast-metrics/src/Analyzer/Volume"
    "os"
    "io/ioutil"
    "log"
    "strconv"
    "fmt"
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

    // register visitors
    cyclomaticVisitor := &Complexity.ComplexityVisitor{}
    root.Accept(cyclomaticVisitor)

    locVisitor := &Volume.LocVisitor{}
    root.Accept(locVisitor)

    // visit AST
    root.Visit()

    // Now pbFile contains the AST and analyze results
    // We dump it to a file with ProtoBuf
    //out, err := proto.Marshal(pbFile)
    //if err != nil {
    //    log.Fatalln("Failed to encode pbFile:", err)
    //}
    // Write the new file back to disk into "file"
    //ioutil.WriteFile(file, out, 0644)

    // debug for the moment
    // iterate over nodes of pbFile
    for _, stmt := range pbFile.Stmts.StmtClass {
        temporaryDisplayStatForClass(*stmt)
    }
    for _, stmt := range pbFile.Stmts.StmtNamespace {
        for _, s := range stmt.Stmts.StmtClass {
            temporaryDisplayStatForClass(*s)
        }
    }
}

func temporaryDisplayStatForClass(cl pb.StmtClass ) {

    if cl.Stmts == nil {
        return
    }

    fmt.Println("\nClass: " + cl.Name.Qualified)
    fmt.Println("    Cyclomatic complexity: " + strconv.Itoa(int(*cl.Stmts.Analyze.Complexity.Cyclomatic)))
    fmt.Println("    Logical lines of code: " + strconv.Itoa(int(*cl.Stmts.Analyze.Volume.Lloc)))
    fmt.Println("    Comment lines of code: " + strconv.Itoa(int(*cl.Stmts.Analyze.Volume.Cloc)))
    fmt.Println("    Lines of code: " + strconv.Itoa(int(*cl.Stmts.Analyze.Volume.Loc)))
}