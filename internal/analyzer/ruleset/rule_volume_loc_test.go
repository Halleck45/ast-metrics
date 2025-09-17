package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestVolumeLocRule_ViolationAndSuccess(t *testing.T) {
	max := 1000
	rule := NewLocRule(&max)

	makeFile := func(loc int32) *pb.File {
		return &pb.File{
			Stmts: &pb.Stmts{
				Analyze: &pb.Analyze{
					Volume: &pb.Volume{Loc: &loc},
				},
			},
		}
	}

	// Violation case
	fileHigh := makeFile(1500)
	errors := []issue.RequirementError{}
	rule.CheckFile(fileHigh, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) == 0 {
		t.Fatal("expected a violation when LOC > max")
	}

	err := errors[0]
	if err.Code != "max_loc" {
		t.Errorf("expected 'max_loc' code, got %s", err.Code)
	}
	if err.Severity != issue.SeverityMedium {
		t.Errorf("expected medium severity, got %s", err.Severity)
	}

	// Success case
	errors = []issue.RequirementError{}
	fileOk := makeFile(500)
	rule.CheckFile(fileOk, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) != 0 {
		t.Fatal("expected no violation when LOC <= max")
	}
}

func TestVolumeLocRule_NilMax(t *testing.T) {
	rule := NewLocRule(nil)
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Volume: &pb.Volume{Loc: func() *int32 { v := int32(5000); return &v }()},
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
