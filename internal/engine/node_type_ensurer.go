package engine

import (
	pb "github.com/halleck45/ast-metrics/pb"
)

func EnsureNodeTypeIsComplete(file *pb.File) {

	if file.Stmts.Analyze == nil {
		file.Stmts.Analyze = &pb.Analyze{}
	}

	// Ensure file-level Volume is populated from file.LinesOfCode when missing
	if file.Stmts.Analyze.Volume == nil {
		file.Stmts.Analyze.Volume = &pb.Volume{}
	}
	if file.LinesOfCode != nil {
		if file.Stmts.Analyze.Volume.Loc == nil || *file.Stmts.Analyze.Volume.Loc == 0 {
			v := file.LinesOfCode.LinesOfCode
			file.Stmts.Analyze.Volume.Loc = &v
		}
		if file.Stmts.Analyze.Volume.Lloc == nil || *file.Stmts.Analyze.Volume.Lloc == 0 {
			v := file.LinesOfCode.LogicalLinesOfCode
			file.Stmts.Analyze.Volume.Lloc = &v
		}
		if file.Stmts.Analyze.Volume.Cloc == nil || *file.Stmts.Analyze.Volume.Cloc == 0 {
			v := file.LinesOfCode.CommentLinesOfCode
			file.Stmts.Analyze.Volume.Cloc = &v
		}
	}

	if file.LinesOfCode == nil && file.Stmts.Analyze.Volume != nil && file.Stmts.Analyze.Volume.Loc != nil && file.Stmts.Analyze.Volume.Lloc != nil && file.Stmts.Analyze.Volume.Cloc != nil {
		file.LinesOfCode = &pb.LinesOfCode{
			LinesOfCode:        *file.Stmts.Analyze.Volume.Loc,
			CommentLinesOfCode: *file.Stmts.Analyze.Volume.Cloc,
			LogicalLinesOfCode: *file.Stmts.Analyze.Volume.Lloc,
		}
	}

	if file.Stmts.Analyze == nil {
		file.Stmts.Analyze = &pb.Analyze{}
	}

	if file.Stmts.Analyze.Complexity == nil {
		file.Stmts.Analyze.Complexity = &pb.Complexity{}
	}

	// Set file-level cyclomatic complexity from contained symbols:
	// - If the file contains classes, use the sum of class complexities.
	// - If no class complexity is available, use the sum of top-level function complexities.
	classes := GetClassesInFile(file)
	var fileCyclomatic int32 = 0
	classComplexityCount := 0
	if len(classes) > 0 {
		for _, class := range classes {
			if class == nil || class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Complexity == nil || class.Stmts.Analyze.Complexity.Cyclomatic == nil {
				continue
			}
			fileCyclomatic += *class.Stmts.Analyze.Complexity.Cyclomatic
			classComplexityCount++
		}
	}
	if len(classes) == 0 || classComplexityCount == 0 {
		fileCyclomatic = 0
		functions := GetFunctionsInFile(file)
		for _, function := range functions {
			if function == nil || function.Stmts == nil || function.Stmts.Analyze == nil || function.Stmts.Analyze.Complexity == nil || function.Stmts.Analyze.Complexity.Cyclomatic == nil {
				continue
			}
			fileCyclomatic += *function.Stmts.Analyze.Complexity.Cyclomatic
		}
	}
	file.Stmts.Analyze.Complexity.Cyclomatic = &fileCyclomatic

	if file.Stmts.Analyze.Coupling == nil {
		file.Stmts.Analyze.Coupling = &pb.Coupling{
			Afferent: 0,
			Efferent: 0,
		}
	}
}
