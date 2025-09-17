package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestMaintainabilityRule_ViolationAndSuccess(t *testing.T) {
	min := 70
	rule := NewMaintainabilityRule(&min)

	makeFile := func(maintainability float64) *pb.File {
		return &pb.File{
			Stmts: &pb.Stmts{
				StmtClass: []*pb.StmtClass{
					{
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								Maintainability: &pb.Maintainability{
									MaintainabilityIndex: &maintainability,
								},
							},
						},
					},
				},
			},
		}
	}

	// Violation case
	fileLow := makeFile(50.0)
	errors := []issue.RequirementError{}
	rule.CheckFile(fileLow, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) == 0 {
		t.Fatal("expected a violation when maintainability < min")
	}

	err := errors[0]
	if err.Code != "maintainability" {
		t.Errorf("expected 'maintainability' code, got %s", err.Code)
	}
	if err.Severity != issue.SeverityHigh {
		t.Errorf("expected high severity, got %s", err.Severity)
	}

	// Success case
	errors = []issue.RequirementError{}
	fileOk := makeFile(80.0)
	rule.CheckFile(fileOk, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) != 0 {
		t.Fatal("expected no violation when maintainability >= min")
	}
}

func TestMaintainabilityRule_NilMin(t *testing.T) {
	rule := NewMaintainabilityRule(nil)
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Maintainability: &pb.Maintainability{
								MaintainabilityIndex: func() *float64 { v := 10.0; return &v }(),
							},
						},
					},
				},
			},
		},
	}

	errors := []issue.RequirementError{}
	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) != 0 {
		t.Error("expected no errors when min is nil")
	}
}

func TestMaintainabilityRule_NilMaintainability(t *testing.T) {
	min := 70
	rule := NewMaintainabilityRule(&min)
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Maintainability: nil,
						},
					},
				},
			},
		},
	}

	errors := []issue.RequirementError{}
	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) }, 
		func(string) {})

	if len(errors) != 0 {
		t.Error("expected no errors when maintainability is nil")
	}
}
