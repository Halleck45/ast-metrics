package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestEfferentCouplingRule_ViolationAndSuccess(t *testing.T) {
	max := 5
	rule := NewEfferentCouplingRule(&max)

	makeFile := func(eff int32) *pb.File {
		return &pb.File{
			Stmts: &pb.Stmts{
				Analyze: &pb.Analyze{
					Coupling: &pb.Coupling{Efferent: eff},
				},
			},
		}
	}

	// Violation case
	fileHigh := makeFile(10)
	errors := []issue.RequirementError{}
	rule.CheckFile(fileHigh, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) == 0 {
		t.Fatal("expected a violation when efferent > max")
	}

	err := errors[0]
	if err.Code != "efferent_coupling" {
		t.Errorf("expected 'efferent_coupling' code, got %s", err.Code)
	}

	// Success case
	errors = []issue.RequirementError{}
	fileOk := makeFile(3)
	rule.CheckFile(fileOk, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) != 0 {
		t.Fatal("expected no violation when efferent <= max")
	}
}

func TestEfferentCouplingRule_NilMax(t *testing.T) {
	rule := NewEfferentCouplingRule(nil)
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Coupling: &pb.Coupling{Efferent: 100},
			},
		},
	}

	errors := []issue.RequirementError{}
	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) != 0 {
		t.Error("expected no errors when max is nil")
	}
}
