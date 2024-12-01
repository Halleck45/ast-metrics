package Engine

import (
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func EnsureNodeTypeIsComplete(file *pb.File) {

	if file.Stmts.Analyze == nil {
		file.Stmts.Analyze = &pb.Analyze{}
	}

	if file.LinesOfCode == nil && file.Stmts.Analyze.Volume != nil {
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
		zero := int32(0)
		file.Stmts.Analyze.Complexity = &pb.Complexity{
			Cyclomatic: &zero,
		}
	}

	// Transfert complexity from classes and functions to file itself
	classes := GetClassesInFile(file)
	if len(classes) == 0 {
		functions := GetFunctionsInFile(file)
		for _, function := range functions {
			if function.Stmts.Analyze == nil || function.Stmts.Analyze.Complexity == nil || function.Stmts.Analyze.Complexity.Cyclomatic == nil {
				continue
			}

			// increment complexity of file itself
			ccn := *function.Stmts.Analyze.Complexity.Cyclomatic + *file.Stmts.Analyze.Complexity.Cyclomatic
			file.Stmts.Analyze.Complexity.Cyclomatic = &ccn
		}
	} else {
		for _, class := range classes {
			if class.Stmts.Analyze == nil || class.Stmts.Analyze.Complexity == nil || class.Stmts.Analyze.Complexity.Cyclomatic == nil {
				continue
			}

			// increment complexity of file itself
			ccn := *class.Stmts.Analyze.Complexity.Cyclomatic + *file.Stmts.Analyze.Complexity.Cyclomatic
			file.Stmts.Analyze.Complexity.Cyclomatic = &ccn
		}
	}

	if file.Stmts.Analyze.Coupling == nil {
		file.Stmts.Analyze.Coupling = &pb.Coupling{
			Afferent: 0,
			Efferent: 0,
		}
	}
}
