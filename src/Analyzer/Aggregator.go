package Analyzer

import (
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    "math"
)

type Aggregated struct {
    NbFiles int
    NbFunctions int
    NbClasses int
    NbMethods int
    Loc int
    Cloc int
    Lloc int
    AverageMethodsPerClass float64
    AverageLocPerMethod float64
    AverageLlocPerMethod float64
    AverageClocPerMethod float64
    AverageCyclomaticComplexityPerMethod float64
    AverageCyclomaticComplexityPerClass float64
    MinCyclomaticComplexity int
    MaxCyclomaticComplexity int
    AverageHalsteadDifficulty float64
    AverageHalsteadEffort float64
    AverageHalsteadVolume float64
    AverageHalsteadTime float64
    AverageHalsteadBugs float64
    SumHalsteadDifficulty float64
    SumHalsteadEffort float64
    SumHalsteadVolume float64
    SumHalsteadTime float64
    SumHalsteadBugs float64
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

        // Average cyclomatic complexity per method
        if class.Stmts.StmtFunction != nil {
            for _, method := range class.Stmts.StmtFunction {
                if method.Stmts == nil {
                    continue
                }
                if method.Stmts.Analyze.Complexity != nil {
                    if method.Stmts.Analyze.Complexity.Cyclomatic != nil {
                        aggregated.AverageCyclomaticComplexityPerMethod += float64(*method.Stmts.Analyze.Complexity.Cyclomatic)
                    }
                }
            }
        }

        // lines of code
        if class.Stmts.Analyze.Volume != nil {
            if class.Stmts.Analyze.Volume.Loc != nil {
                aggregated.Loc += int(*class.Stmts.Analyze.Volume.Loc)
            }
            if class.Stmts.Analyze.Volume.Cloc != nil {
                aggregated.Cloc += int(*class.Stmts.Analyze.Volume.Cloc)
            }
            if class.Stmts.Analyze.Volume.Lloc != nil {
                aggregated.Lloc += int(*class.Stmts.Analyze.Volume.Lloc)
            }

            // average lines of code per method
            if class.Stmts.StmtFunction != nil {
                for _, method := range class.Stmts.StmtFunction {

                    if method.Stmts == nil {
                        continue
                    }

                    if method.Stmts.Analyze.Volume != nil {
                        if method.Stmts.Analyze.Volume.Loc != nil {
                            aggregated.AverageLocPerMethod += float64(*method.Stmts.Analyze.Volume.Loc)
                        }
                        if method.Stmts.Analyze.Volume.Cloc != nil {
                            aggregated.AverageClocPerMethod += float64(*method.Stmts.Analyze.Volume.Cloc)
                        }
                        if method.Stmts.Analyze.Volume.Lloc != nil {
                            aggregated.AverageLlocPerMethod += float64(*method.Stmts.Analyze.Volume.Lloc)
                        }
                    }
                }
            }
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

        // Halstead
        if class.Stmts.Analyze.Volume != nil {
            if class.Stmts.Analyze.Volume.HalsteadDifficulty != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadDifficulty)) {
                aggregated.AverageHalsteadDifficulty += float64(*class.Stmts.Analyze.Volume.HalsteadDifficulty)
                aggregated.SumHalsteadDifficulty += float64(*class.Stmts.Analyze.Volume.HalsteadDifficulty)
            }
            if class.Stmts.Analyze.Volume.HalsteadEffort != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadEffort)) {
                aggregated.AverageHalsteadEffort += float64(*class.Stmts.Analyze.Volume.HalsteadEffort)
                aggregated.SumHalsteadEffort += float64(*class.Stmts.Analyze.Volume.HalsteadEffort)
            }
            if class.Stmts.Analyze.Volume.HalsteadVolume != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadVolume)) {
                aggregated.AverageHalsteadVolume += float64(*class.Stmts.Analyze.Volume.HalsteadVolume)
                aggregated.SumHalsteadVolume += float64(*class.Stmts.Analyze.Volume.HalsteadVolume)
            }
            if class.Stmts.Analyze.Volume.HalsteadTime != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadTime)) {
                aggregated.AverageHalsteadTime += float64(*class.Stmts.Analyze.Volume.HalsteadTime)
                aggregated.SumHalsteadTime += float64(*class.Stmts.Analyze.Volume.HalsteadTime)
            }
        }
    }

    // averages
    aggregated.AverageMethodsPerClass = float64(aggregated.NbMethods) / float64(aggregated.NbClasses)
    aggregated.AverageCyclomaticComplexityPerClass = aggregated.AverageCyclomaticComplexityPerClass / float64(aggregated.NbClasses)
    aggregated.AverageHalsteadDifficulty = aggregated.AverageHalsteadDifficulty / float64(aggregated.NbClasses)
    aggregated.AverageHalsteadEffort = aggregated.AverageHalsteadEffort / float64(aggregated.NbClasses)
    aggregated.AverageHalsteadVolume = aggregated.AverageHalsteadVolume / float64(aggregated.NbClasses)
    aggregated.AverageHalsteadTime = aggregated.AverageHalsteadTime / float64(aggregated.NbClasses)

    aggregated.AverageLocPerMethod = aggregated.AverageLocPerMethod / float64(aggregated.NbMethods)
    aggregated.AverageClocPerMethod = aggregated.AverageClocPerMethod / float64(aggregated.NbMethods)
    aggregated.AverageLlocPerMethod = aggregated.AverageLlocPerMethod / float64(aggregated.NbMethods)

    aggregated.AverageCyclomaticComplexityPerMethod = aggregated.AverageCyclomaticComplexityPerMethod / float64(aggregated.NbMethods)

    return aggregated
}