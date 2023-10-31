package Analyzer

import (
    "github.com/halleck45/ast-metrics/src/Storage"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    Complexity "github.com/halleck45/ast-metrics/src/Analyzer/Complexity"
    Volume "github.com/halleck45/ast-metrics/src/Analyzer/Volume"
    "io/ioutil"
    "log"
    "strconv"
    "fmt"
    "sync"
    "github.com/golang/protobuf/proto"
    "github.com/pterm/pterm"
    "github.com/yargevad/filepathx"
)

func Start(progressbar *pterm.SpinnerPrinter) ([]pb.File){

    workdir := Storage.Path()
    // List all ASTs files (*.bin) in the workdir
    astFiles, err := filepathx.Glob(workdir + "/**/*.bin")
    if err != nil {
        panic(err)
    }

    // Wait for end of all goroutines
    var wg sync.WaitGroup

    // store results
    // channel should have value
    // https://stackoverflow.com/questions/58743038/why-does-this-goroutine-not-call-wg-done
    channelResult := make(chan pb.File, len(astFiles))

    nbParsingFiles := 0
    // in parallel, 8 process max, analyze each AST file running the runAnalysis function
    for _, file := range astFiles {
        wg.Add(1)
        nbParsingFiles++
        go func(file string) {
            executeFileAnalysis(file, channelResult)
            defer wg.Done()

            // details is the number of files processed / total number of files
            details := strconv.Itoa(nbParsingFiles) + "/" + strconv.Itoa(len(astFiles))
            progressbar.UpdateText("Analyzing (" + details + ")")
        }(file)
    }


    wg.Wait()
    progressbar.Info("AST Analysis finished")

    // Convert it to slice of pb.File
    allResults := make([]pb.File, 0, len(astFiles))
    for i := 0; i < len(astFiles); i++ {
        allResults = append(allResults, <-channelResult)
    }
    defer close(channelResult)

    return allResults
}

func executeFileAnalysis(file string, channelResult chan<- pb.File) {

    // load AST via ProtoBuf (using NodeType package)
    in, err := ioutil.ReadFile(file)
    if err != nil {
        log.Fatal("Error reading file:", err)
        return
    }

    // if file is empty, return
    if len(in) == 0 {
        log.Fatal("File is empty:", err)
        return
    }

    pbFile := &pb.File{}
    if err := proto.Unmarshal(in, pbFile); err != nil {
        log.Fatalln("Failed to parse address pbFile:", err)
        return
    }

    root := &ASTNode{children: pbFile.Stmts}

    // register visitors
    cyclomaticVisitor := &Complexity.ComplexityVisitor{}
    root.Accept(cyclomaticVisitor)

    locVisitor := &Volume.LocVisitor{}
    root.Accept(locVisitor)

    // visit AST
    root.Visit()

    channelResult <- *pbFile
    return
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

    return
    fmt.Println("\nClass: " + cl.Name.Qualified)
    fmt.Println("    Cyclomatic complexity: " + strconv.Itoa(int(*cl.Stmts.Analyze.Complexity.Cyclomatic)))
    fmt.Println("    Logical lines of code: " + strconv.Itoa(int(*cl.Stmts.Analyze.Volume.Lloc)))
    fmt.Println("    Comment lines of code: " + strconv.Itoa(int(*cl.Stmts.Analyze.Volume.Cloc)))
    fmt.Println("    Lines of code: " + strconv.Itoa(int(*cl.Stmts.Analyze.Volume.Loc)))
}