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

	if file.Stmts.Analyze.Coupling == nil {
		file.Stmts.Analyze.Coupling = &pb.Coupling{
			Afferent: 0,
			Efferent: 0,
		}
	}
}
