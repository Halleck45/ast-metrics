package treesitter

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
)

func TestVisitor_curStmts_FileLevel(t *testing.T) {
	visitor := &Visitor{
		file: &pb.File{
			Stmts: &pb.Stmts{},
		},
	}

	stmts := visitor.curStmts()
	if stmts != visitor.file.Stmts {
		t.Error("expected file-level stmts")
	}
}

func TestVisitor_curStmts_ClassLevel(t *testing.T) {
	class := &pb.StmtClass{
		Stmts: &pb.Stmts{},
	}
	
	visitor := &Visitor{
		file:     &pb.File{Stmts: &pb.Stmts{}},
		classStk: []*pb.StmtClass{class},
	}

	stmts := visitor.curStmts()
	if stmts != class.Stmts {
		t.Error("expected class-level stmts")
	}
}

func TestVisitor_curStmts_FunctionLevel(t *testing.T) {
	function := &pb.StmtFunction{
		Stmts: &pb.Stmts{},
	}
	
	visitor := &Visitor{
		file:    &pb.File{Stmts: &pb.Stmts{}},
		funcStk: []*pb.StmtFunction{function},
	}

	stmts := visitor.curStmts()
	if stmts != function.Stmts {
		t.Error("expected function-level stmts")
	}
}
