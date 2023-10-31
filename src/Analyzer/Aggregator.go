package Analyzer

import (
    pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type Aggregated struct {
    NbFiles int
    NbFunctions int
    NbClasses int
    NbMethods int
    AverageMethodsPerClass float64
    AverageLocPerMethod float64
    AverageLlocPerMethod float64
    AverageClocPerMethod float64
    AverageCyclomaticComplexityPerMethod float64
    AverageCyclomaticComplexityPerClass float64
    MinCyclomaticComplexity int
    MaxCyclomaticComplexity int
}

func Aggregates(files []pb.File) Aggregated {
    aggregated := Aggregated{NbClasses: 0, NbMethods: 0, AverageMethodsPerClass: 0}

    // get classes
    var classes []pb.StmtClass
    for _, file := range files {
        for _, stmt := range file.Stmts.StmtClass {
            classes = append(classes, *stmt)
        }
        for _, stmt := range file.Stmts.StmtNamespace {
            for _, s := range stmt.Stmts.StmtClass {
                classes = append(classes, *s)
            }
        }
    }

    aggregated.NbFiles = len(files)
    aggregated.NbClasses = len(classes)

    for _, class := range classes {
        if class.Stmts == nil {
            continue
        }
        // methods per class
        if class.Stmts.StmtFunction != nil {
            aggregated.NbMethods += len(class.Stmts.StmtFunction)
        }

        // cyclomatic complexity per class
        if class.Stmts.Analyze.Complexity.Cyclomatic != nil {
            aggregated.AverageCyclomaticComplexityPerClass += float64(*class.Stmts.Analyze.Complexity.Cyclomatic)
            if aggregated.MinCyclomaticComplexity == 0 || int(*class.Stmts.Analyze.Complexity.Cyclomatic) < aggregated.MinCyclomaticComplexity {
                aggregated.MinCyclomaticComplexity = int(*class.Stmts.Analyze.Complexity.Cyclomatic)
            }
            if aggregated.MaxCyclomaticComplexity == 0 || int(*class.Stmts.Analyze.Complexity.Cyclomatic) > aggregated.MaxCyclomaticComplexity {
                aggregated.MaxCyclomaticComplexity = int(*class.Stmts.Analyze.Complexity.Cyclomatic)
            }
        }
    }
    aggregated.AverageMethodsPerClass = float64(aggregated.NbMethods) / float64(aggregated.NbClasses)

    return aggregated
}