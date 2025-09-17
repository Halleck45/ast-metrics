package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestAfferentCouplingRule_ViolationAndSuccess(t *testing.T) {
	max := 5
	r := NewAfferentCouplingRule(&max)

	makeFile := func(aff int32) *pb.File {
		return &pb.File{Stmts: &pb.Stmts{Analyze: &pb.Analyze{Coupling: &pb.Coupling{Afferent: aff}}}}
	}

	// Violation case
	fileHigh := makeFile(10)
	errs := []issue.RequirementError{}
	r.CheckFile(fileHigh, func(e issue.RequirementError) { errs = append(errs, e) }, func(string) {})
	if len(errs) == 0 {
		t.Fatalf("expected a violation when afferent > max")
	}

	// Success case
	errs = []issue.RequirementError{}
	fileOk := makeFile(3)
	r.CheckFile(fileOk, func(e issue.RequirementError) { errs = append(errs, e) }, func(string) {})
	if len(errs) != 0 {
		t.Fatalf("expected no violation when afferent <= max")
	}
}
