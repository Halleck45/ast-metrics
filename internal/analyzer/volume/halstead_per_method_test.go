package analyzer

import (
	"testing"

	pb "github.com/ast-metrics/ast-metrics/pb"
	"google.golang.org/protobuf/proto"
)

// method builds a function statement holding the given operators and operands.
func method(name string, operators, operands []string) *pb.StmtFunction {
	fn := &pb.StmtFunction{
		Name:  &pb.Name{Short: name},
		Stmts: &pb.Stmts{Analyze: &pb.Analyze{Volume: &pb.Volume{}}},
	}
	for _, o := range operators {
		fn.Operators = append(fn.Operators, &pb.StmtOperator{Name: o})
	}
	for _, o := range operands {
		fn.Operands = append(fn.Operands, &pb.StmtOperand{Name: o})
	}
	return fn
}

// Each method must keep its own figures. They used to share the pointers taken
// on the visitor's local counters, so every method of a class ended up with the
// numbers of the last one visited: a class whose last method was trivial
// reported a volume of zero, and its maintainability index fell back to 7.
func TestHalsteadIsComputedPerMethodNotShared(t *testing.T) {
	dense := method("dense", []string{"+", "-", "*", "+"}, []string{"a", "b", "c", "a"})
	trivial := method("trivial", nil, nil)

	class := &pb.StmtClass{
		Name:  &pb.Name{Short: "C"},
		Stmts: &pb.Stmts{Analyze: &pb.Analyze{Volume: &pb.Volume{}}},
	}
	class.Stmts.StmtFunction = []*pb.StmtFunction{dense, trivial}

	visitor := &HalsteadMetricsVisitor{}
	visitor.Visit(class.Stmts, class.Stmts)

	// dense: 4 operators (3 distinct) + 4 operands (3 distinct) => N=8, n=6
	if got := dense.Stmts.Analyze.Volume.GetHalsteadLength(); got != 8 {
		t.Errorf("dense method: expected length 8, got %d", got)
	}
	if got := dense.Stmts.Analyze.Volume.GetHalsteadVocabulary(); got != 6 {
		t.Errorf("dense method: expected vocabulary 6, got %d", got)
	}
	if got := dense.Stmts.Analyze.Volume.GetHalsteadVolume(); got <= 0 {
		t.Errorf("dense method: expected a positive volume, got %v", got)
	}

	// the trivial method has nothing to measure, and must not have inherited
	// the figures of the dense one
	if got := trivial.Stmts.Analyze.Volume.GetHalsteadLength(); got != 0 {
		t.Errorf("trivial method: expected length 0, got %d", got)
	}
	if got := trivial.Stmts.Analyze.Volume.GetHalsteadVolume(); got != 0 {
		t.Errorf("trivial method: expected volume 0, got %v", got)
	}
}

// The class average is rounded, not truncated: an integer division reported
// zero for a class whose methods each held a symbol or two.
func TestHalsteadClassAverageIsRoundedNotTruncated(t *testing.T) {
	file := &pb.Stmts{Analyze: &pb.Analyze{Volume: &pb.Volume{}}}
	class := &pb.StmtClass{
		Name:  &pb.Name{Short: "Small"},
		Stmts: &pb.Stmts{Analyze: &pb.Analyze{Volume: &pb.Volume{}}},
	}
	// three methods holding one symbol each: the sum stays below the method
	// count, which used to truncate the average down to zero
	for _, name := range []string{"a", "b", "c"} {
		m := method(name, []string{"+"}, nil)
		m.Stmts.Analyze.Volume = &pb.Volume{
			HalsteadVocabulary:      proto.Int32(1),
			HalsteadLength:          proto.Int32(1),
			HalsteadEstimatedLength: proto.Float64(0),
			HalsteadVolume:          proto.Float64(0),
			HalsteadDifficulty:      proto.Float64(0),
			HalsteadEffort:          proto.Float64(0),
			HalsteadTime:            proto.Float64(0),
		}
		class.Stmts.StmtFunction = append(class.Stmts.StmtFunction, m)
	}
	file.StmtClass = []*pb.StmtClass{class}

	visitor := &HalsteadMetricsVisitor{}
	visitor.LeaveNode(file)

	if got := class.Stmts.Analyze.Volume.GetHalsteadLength(); got != 1 {
		t.Errorf("class average: expected length 1, got %d", got)
	}
	if got := class.Stmts.Analyze.Volume.GetHalsteadVocabulary(); got != 1 {
		t.Errorf("class average: expected vocabulary 1, got %d", got)
	}
}
